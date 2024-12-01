package groq

import (
	"TLExtractor/consts"
	"TLExtractor/environment"
	"TLExtractor/groq/types"
	schemeTypes "TLExtractor/telegram/scheme/types"
	"encoding/json"
	"fmt"
	"github.com/Laky-64/http"
	"regexp"
	"strings"
)

func GenerateDescriptions(differences *schemeTypes.TLFullDifferences) (map[string]string, error) {
	resultsNeeded, data, err := generateRequest(differences, "llm_descriptions")
	if err != nil {
		return nil, err
	}
	request, err := http.ExecuteRequest(
		fmt.Sprintf("%s/chat/completions", consts.GroqAPI),
		http.BearerToken(environment.CredentialsStorage.GroqToken),
		http.Method("POST"),
		http.Body(data),
		http.Headers(map[string]string{
			"Content-Type": "application/json",
		}),
	)
	if err != nil {
		return nil, err
	}
	var result types.Result
	err = json.Unmarshal(request.Body, &result)
	if err != nil {
		return nil, err
	}
	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("no completions found")
	}
	generatedResponse := strings.Split(result.Choices[0].Message.Content, "\n")
	generatedDescriptions := make(map[string]string)
	for _, response := range generatedResponse {
		if descInfo := regexp.MustCompile(`Added\s(.+?):\s(.+)`).FindStringSubmatch(response); len(descInfo) == 3 {
			generatedDescriptions[strings.TrimSpace(descInfo[1])] = strings.TrimSpace(descInfo[2])
		}
	}
	if len(generatedDescriptions) != resultsNeeded {
		return nil, fmt.Errorf("generated descriptions length does not match methods length")
	}
	return generatedDescriptions, nil
}
