package groq

import (
	"TLExtractor/consts"
	"TLExtractor/environment"
	"TLExtractor/groq/types"
	"TLExtractor/tui"
	tuiTypes "TLExtractor/tui/types"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Laky-64/http"
	"github.com/charmbracelet/huh"
	"maps"
	"slices"
)

func init() {
	groqApp := tui.NewMiniApp("groq")
	groqApp.SetLoadingMessage("Logging in to GroqCloud...")
	secretToken := environment.CredentialsStorage.GroqToken
	llmModel := environment.CredentialsStorage.LLMModel
	foundModels := make(map[string]string)
	groqApp.SetFields(
		huh.NewInput().
			Title("Enter your GroqCloud secret token:").
			Description("You can find your GroqCloud secret token in the GroqCloud settings.").
			Validate(tui.Validate("GroqCloud secret token", tuiTypes.NoCheck)).
			Value(&secretToken),
	)
	groqApp.SetCheckFunc(func(checkType tuiTypes.CheckType) error {
		if checkType == tuiTypes.InitCheck {
			if len(secretToken) == 0 || len(llmModel) == 0 {
				return errors.New("secret token and LLM model must be set")
			}
		}
		request, err := http.ExecuteRequest(
			fmt.Sprintf("%s/models", consts.GroqAPI),
			http.BearerToken(secretToken),
		)
		if err != nil {
			return err
		}
		var groqModels types.Models
		err = json.Unmarshal(request.Body, &groqModels)
		if err != nil {
			return err
		}
		for _, model := range groqModels.Data {
			foundModels[fmt.Sprintf("%s (%s)", model.ID, model.OwnedBy)] = model.ID
		}
		if len(foundModels) > 0 && checkType == tuiTypes.InitCheck {
			var found bool
			for _, model := range groqModels.Data {
				if model.ID == llmModel {
					found = true
					break
				}
			}
			if !found {
				return errors.New("selected LLM model is not found")
			}
		}
		environment.CredentialsStorage.GroqToken = secretToken
		environment.CredentialsStorage.Commit()
		return nil
	}, tuiTypes.InitCheck, tuiTypes.SubmitCheck)

	modelPage := groqApp.NewAppPage()
	var selectedModel string
	modelPage.SetFields(
		huh.NewSelect[string]().
			Title(fmt.Sprintf("Select LLM model:")).
			Value(&selectedModel).
			OptionsFunc(func() []huh.Option[string] {
				return huh.NewOptions(slices.Collect(maps.Keys(foundModels))...)
			}, "model"),
	)
	modelPage.SetCheckFunc(func(checkType tuiTypes.CheckType) error {
		llmModel = foundModels[selectedModel]
		environment.CredentialsStorage.LLMModel = llmModel
		environment.CredentialsStorage.Commit()
		return nil
	}, tuiTypes.SubmitCheck)
}
