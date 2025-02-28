package gemini

import (
	"TLExtractor/assets"
	"TLExtractor/environment"
	"TLExtractor/tui"
	tuiTypes "TLExtractor/tui/types"
	"context"
	"errors"
	"fmt"
	"github.com/charmbracelet/huh"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"maps"
	"slices"
	"strings"
)

func init() {
	geminiApp := tui.NewMiniApp("gemini")
	geminiApp.SetLoadingMessage("Logging in to Gemini API...")
	secretToken := environment.CredentialsStorage.GeminiToken
	llmModel := environment.CredentialsStorage.LLMModel
	Client = &clientContext{
		ctx: context.Background(),
	}
	loadModel := func() {
		Client.generativeModel = Client.apiClient.GenerativeModel(llmModel)
		Client.generativeModel.SetTemperature(0)
		Client.generativeModel.SetTopK(64)
		Client.generativeModel.SetTopP(0.95)
		Client.generativeModel.SetMaxOutputTokens(65536)
		Client.generativeModel.ResponseMIMEType = "text/plain"
		Client.generativeModel.SystemInstruction = &genai.Content{
			Parts: []genai.Part{
				genai.Text(assets.Templates["llm_descriptions_prompt"]),
			},
		}
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
		client, errLogin := genai.NewClient(Client.ctx, option.WithAPIKey(secretToken))
		if errLogin != nil {
			return errLogin
		}
		models := client.ListModels(Client.ctx)
		foundCurrentModel := false
		for {
			model, err := models.Next()
			if errors.Is(err, iterator.Done) {
				break
			} else if err != nil {
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
		if checkType == tuiTypes.InitCheck && len(llmModel) > 0 {
			loadModel()
		}
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
		loadModel()
		return nil
	}, tuiTypes.SubmitCheck)
}
