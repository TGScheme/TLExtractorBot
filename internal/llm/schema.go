package llm

import (
	"strings"

	"google.golang.org/genai"
)

var changelogSchema = &genai.Schema{
	Type: genai.TypeObject,
	Properties: map[string]*genai.Schema{
		"title": {
			Type:        genai.TypeString,
			Description: "Title of the article, naming the biggest changes.",
		},
		"lead": {
			Type:        genai.TypeString,
			Description: "Opening paragraph of the article, one or two sentences.",
		},
		"sections": {
			Type:     genai.TypeArray,
			MaxItems: new(int64(maxSections)),
			Items: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"title": {
						Type:        genai.TypeString,
						Description: "Short feature-style heading, no API names.",
					},
					"paragraphs": {
						Type:     genai.TypeArray,
						MaxItems: new(int64(maxParagraphs)),
						Items:    &genai.Schema{Type: genai.TypeString},
					},
					"highlights": {
						Type:     genai.TypeArray,
						MaxItems: new(int64(maxHighlights)),
						Items:    &genai.Schema{Type: genai.TypeString},
					},
				},
				Required:         []string{"title", "paragraphs"},
				PropertyOrdering: []string{"title", "paragraphs", "highlights"},
			},
		},
		"items": {
			Type: genai.TypeArray,
			Items: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"name": {
						Type:        genai.TypeString,
						Description: "Name of the new object, without the constructor id.",
					},
					"description": {
						Type:        genai.TypeString,
						Description: "One sentence, Telegram API documentation style.",
					},
				},
				Required:         []string{"name", "description"},
				PropertyOrdering: []string{"name", "description"},
			},
		},
	},
	Required:         []string{"title", "lead", "sections", "items"},
	PropertyOrdering: []string{"title", "lead", "sections", "items"},
}

func jsonSchema(schema *genai.Schema) map[string]any {
	out := map[string]any{"type": strings.ToLower(string(schema.Type))}
	if schema.Description != "" {
		out["description"] = schema.Description
	}
	if schema.Items != nil {
		out["items"] = jsonSchema(schema.Items)
	}
	if len(schema.Properties) > 0 {
		properties := make(map[string]any, len(schema.Properties))
		required := make([]string, 0, len(schema.Properties))
		for _, name := range schema.PropertyOrdering {
			properties[name] = jsonSchema(schema.Properties[name])
			required = append(required, name)
		}
		for name, property := range schema.Properties {
			if _, ok := properties[name]; !ok {
				properties[name] = jsonSchema(property)
				required = append(required, name)
			}
		}
		out["properties"] = properties
		out["required"] = required
		out["additionalProperties"] = false
	}
	return out
}
