package gemini

import (
	"TLExtractor/environment"
	"TLExtractor/telegram/scheme"
	schemeTypes "TLExtractor/telegram/scheme/types"
	"fmt"
	"github.com/google/generative-ai-go/genai"
	"regexp"
	"strings"
)

func (ctx *clientContext) GenerateDescriptions(differences *schemeTypes.TLFullDifferences) (map[string]string, error) {
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
	session := ctx.generativeModel.StartChat()
	session.History = []*genai.Content{}
	resp, err := session.SendMessage(
		ctx.ctx,
		genai.Text(
			environment.FormatVar(
				"llm_descriptions",
				map[string]interface{}{
					"prompt_constructors":  promptConstructors,
					"context_constructors": contextConstructors,
				},
			),
		),
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

	generatedResponse := strings.Split(string(resp.Candidates[0].Content.Parts[0].(genai.Text)), "\n")
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
