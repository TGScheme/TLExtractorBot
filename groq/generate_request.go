package groq

import (
	"TLExtractor/environment"
	"TLExtractor/groq/types"
	"TLExtractor/telegram/scheme"
	schemeTypes "TLExtractor/telegram/scheme/types"
	"encoding/json"
	"fmt"
)

func generateRequest(differences *schemeTypes.TLFullDifferences, llmMessage string) (int, []byte, error) {
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
	completionsData := types.Completions{
		Messages: []types.Messages{
			{
				Role: "user",
				Content: environment.FormatVar(
					llmMessage,
					map[string]interface{}{
						"prompt_constructors":  promptConstructors,
						"context_constructors": contextConstructors,
					},
				),
			},
		},
		Model:       environment.CredentialsStorage.LLMModel,
		Temperature: 0,
		TopP:        1,
		Stream:      false,
	}
	data, err := json.Marshal(completionsData)
	if err != nil {
		return 0, nil, err
	}
	return len(promptConstructors), data, nil
}
