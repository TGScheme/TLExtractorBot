package gemini

import (
	"TLExtractor/environment"
	"TLExtractor/tui"
	tuiTypes "TLExtractor/tui/types"
	"context"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/charmbracelet/huh"
	"google.golang.org/genai"
)

func init() {
	geminiApp := tui.NewMiniApp("gemini")
	geminiApp.SetLoadingMessage("Logging in to Gemini API...")
	secretToken := environment.CredentialsStorage.GeminiToken
	llmModel := environment.CredentialsStorage.LLMModel
	Client = &clientContext{
		ctx: context.Background(),
	}
	foundModels := make(map[string]string)
	geminiApp.SetFields(
		huh.NewInput().
			Title("Enter your Gemini secret token:").
			Description("You can find your Gemini secret token in the Google Cloud Console.").
			Placeholder("AlzaB1c2DefG3h4gI5jK6l-M7nO8pQ9").
			EchoMode(huh.EchoModePassword).
			Validate(tui.Validate("Gemini secret token", tuiTypes.NoCheck)).
			Value(&secretToken),
	)
	geminiApp.SetCheckFunc(func(checkType tuiTypes.CheckType) error {
		if checkType == tuiTypes.InitCheck {
			if len(secretToken) == 0 || len(llmModel) == 0 {
				return errors.New("secret token and LLM model must be set")
			}
		}
		client, errLogin := genai.NewClient(Client.ctx, &genai.ClientConfig{
			APIKey: secretToken,
		})
		if errLogin != nil {
			return errLogin
		}
		models := client.Models.All(Client.ctx)
		foundCurrentModel := false
		for model, err := range models {
			if err != nil {
				secretToken = ""
				environment.CredentialsStorage.GeminiToken = ""
				environment.CredentialsStorage.Commit()
				return err
			}
			modelID := strings.ReplaceAll(model.Name, "models/", "")
			if !strings.HasPrefix(modelID, "gemini") {
				continue
			}
			foundModels[fmt.Sprintf("%s (%s)", model.DisplayName, modelID)] = modelID
			if modelID == llmModel {
				foundCurrentModel = true
			}
		}
		if !foundCurrentModel && checkType == tuiTypes.InitCheck {
			environment.CredentialsStorage.LLMModel = ""
			environment.CredentialsStorage.Commit()
			return fmt.Errorf("selected LLM model is not found")
		}
		Client.apiClient = client
		environment.CredentialsStorage.GeminiToken = secretToken
		environment.CredentialsStorage.Commit()
		return nil
	}, tuiTypes.InitCheck, tuiTypes.SubmitCheck)

	modelPage := geminiApp.NewAppPage()
	var selectedModel string
	modelPage.SetFields(
		huh.NewSelect[string]().
			Title(fmt.Sprintf("Select LLM model:")).
			Value(&selectedModel).
			OptionsFunc(func() []huh.Option[string] {
				orderedKeys := slices.Collect(maps.Keys(foundModels))
				slices.Sort(orderedKeys)
				slices.Reverse(orderedKeys)
				return huh.NewOptions(orderedKeys...)
			}, "model"),
	)
	modelPage.SetCheckFunc(func(checkType tuiTypes.CheckType) error {
		llmModel = foundModels[selectedModel]
		environment.CredentialsStorage.LLMModel = llmModel
		environment.CredentialsStorage.Commit()
		return nil
	}, tuiTypes.SubmitCheck)
}
