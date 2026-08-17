package gemini

import (
	"fmt"
	"github.com/Laky-64/gologging"
	"github.com/TGScheme/TLExtractorBot/internal/assets"
	"github.com/TGScheme/TLExtractorBot/internal/telegram/scheme"
	schemeTypes "github.com/TGScheme/TLExtractorBot/internal/telegram/scheme/types"
	"regexp"
	"strings"

	"google.golang.org/genai"
)

func (ctx *Client) GenerateDescriptions(differences *schemeTypes.TLFullDifferences) (map[string]string, error) {
	if ctx.model == "" {
		gologging.Warn("gemini: no model set, skipping descriptions (pick one with /model)")
		return nil, nil
	}

	var promptConstructors []string
	var contextConstructors []string
	appendMethods := func(methods []schemeTypes.TLObjDifferences) {
		if methods == nil {
			return
		}
		for _, constructor := range methods {
			constructorString := fmt.Sprintf("%s#%s", constructor.Object.Package(), scheme.ParseConstructor(constructor.Object.Constructor()))
			for _, param := range constructor.Object.Parameters() {
				constructorString += fmt.Sprintf(" %s:%s", param.Name, param.Type)
			}
			constructorString += " = "
			constructorString += constructor.Object.Result() + ";"
			if constructor.IsNew {
				promptConstructors = append(promptConstructors, constructorString)
			} else if !constructor.IsDeleted {
				contextConstructors = append(contextConstructors, constructorString)
			}
		}
	}
	if differences.MainApi != nil {
		appendMethods(differences.MainApi.MethodsDifference)
		appendMethods(differences.MainApi.ConstructorsDifference)
	}
	if differences.E2EApi != nil {
		appendMethods(differences.E2EApi.MethodsDifference)
		appendMethods(differences.E2EApi.ConstructorsDifference)
	}
	session, err := ctx.apiClient.Chats.Create(
		ctx.ctx,
		ctx.model,
		&genai.GenerateContentConfig{
			Temperature:      genai.Ptr(float32(0)),
			TopK:             genai.Ptr(float32(64)),
			TopP:             genai.Ptr(float32(0.95)),
			MaxOutputTokens:  65536,
			ResponseMIMEType: "text/plain",
			SystemInstruction: &genai.Content{
				Role: genai.RoleUser,
				Parts: []*genai.Part{
					{
						Text: assets.Templates["llm_descriptions_prompt"],
					},
				},
			},
		},
		[]*genai.Content{},
	)
	if err != nil {
		return nil, err
	}
	resp, err := session.SendMessage(
		ctx.ctx,
		genai.Part{
			Text: assets.Render(
				"llm_descriptions",
				map[string]interface{}{
					"prompt_constructors":  promptConstructors,
					"context_constructors": contextConstructors,
				},
			),
		},
	)
	if err != nil {
		return nil, err
	}
	if len(resp.Candidates) == 0 {
		return nil, fmt.Errorf("no candidates")
	}
	if len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no parts")
	}
	if len(resp.Candidates[0].Content.Parts) > 1 {
		return nil, fmt.Errorf("too many parts")
	}

	generatedResponse := strings.Split(resp.Candidates[0].Content.Parts[0].Text, "\n")
	generatedDescriptions := make(map[string]string)
	for _, response := range generatedResponse {
		if descInfo := regexp.MustCompile(`Added\s(.+?):\s(.+)`).FindStringSubmatch(response); len(descInfo) == 3 {
			generatedDescriptions[strings.TrimSpace(descInfo[1])] = strings.TrimSpace(descInfo[2])
		}
	}
	if len(generatedDescriptions) != len(promptConstructors) {
		return nil, fmt.Errorf("generated descriptions length does not match methods length")
	}
	return generatedDescriptions, nil
}
