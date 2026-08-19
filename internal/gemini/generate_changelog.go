package gemini

import (
	"encoding/json"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/Laky-64/gologging"
	"github.com/TGScheme/TLExtractorBot/internal/assets"
	"github.com/TGScheme/TLExtractorBot/internal/telegram/scheme"
	schemeTypes "github.com/TGScheme/TLExtractorBot/internal/telegram/scheme/types"
	"google.golang.org/genai"
)

const (
	maxContextPerType = 4
	maxContextLines   = 400
)

var (
	flagPrefix     = regexp.MustCompile(`^flags\d*\.\d+\?`)
	vectorWrapper  = regexp.MustCompile(`(?i)^vector<(.+)>$`)
	primitiveTypes = []string{
		"#", "int", "long", "string", "bytes", "double", "true", "false",
		"Bool", "int128", "int256", "X", "!X", "Object", "Type", "t",
	}
)

func (ctx *Client) GenerateChangelog(req ChangelogRequest) (*Changelog, error) {
	if ctx.model == "" {
		gologging.Warn("gemini: no model set, skipping changelog (pick one with /model)")
		return nil, nil
	}
	if req.Differences == nil {
		return nil, nil
	}

	added, changed, removed := collectDifferences(req.Differences)
	if len(added)+len(changed)+len(removed) == 0 {
		return nil, nil
	}

	resp, err := ctx.apiClient.Models.GenerateContent(
		ctx.ctx,
		ctx.model,
		[]*genai.Content{{
			Role: genai.RoleUser,
			Parts: []*genai.Part{{
				Text: assets.Render("llm_changelog", map[string]any{
					"layer":        req.Layer,
					"source":       req.Source,
					"version_name": req.VersionName,
					"build_number": req.BuildNumber,
					"is_patch":     req.IsPatch,
					"added":        added,
					"changed":      changed,
					"removed":      removed,
					"context":      buildContext(req.Scheme, req.Differences),
				}),
			}},
		}},
		&genai.GenerateContentConfig{
			Temperature:      genai.Ptr(float32(0.35)),
			TopP:             genai.Ptr(float32(0.95)),
			MaxOutputTokens:  65536,
			ResponseMIMEType: "application/json",
			ResponseSchema:   changelogSchema,
			SystemInstruction: &genai.Content{
				Role: genai.RoleUser,
				Parts: []*genai.Part{{
					Text: assets.Templates["llm_changelog_prompt"],
				}},
			},
		},
	)
	if err != nil {
		return nil, err
	}
	if len(resp.Candidates) == 0 {
		return nil, fmt.Errorf("no candidates")
	}
	raw := strings.TrimSpace(resp.Text())
	if raw == "" {
		return nil, fmt.Errorf("empty response")
	}

	var changelog Changelog
	if err = json.Unmarshal([]byte(raw), &changelog); err != nil {
		return nil, fmt.Errorf("invalid json response: %w", err)
	}
	normalize(&changelog, len(added))
	return &changelog, nil
}

func normalize(changelog *Changelog, expectedItems int) {
	changelog.Lead = sanitize(changelog.Lead)
	var sections []ChangelogSection
	for _, section := range changelog.Sections {
		section.Title = sanitize(section.Title)
		section.Paragraphs = sanitizeAll(section.Paragraphs)
		section.Highlights = sanitizeAll(section.Highlights)
		if section.Title == "" || len(section.Paragraphs)+len(section.Highlights) == 0 {
			continue
		}
		sections = append(sections, section)
	}
	changelog.Sections = sections

	changelog.Descriptions = make(map[string]string)
	for _, item := range changelog.Items {
		name := strings.TrimSpace(item.Name)
		name = strings.TrimPrefix(name, "Added ")
		name = strings.TrimSuffix(name, ":")
		if index := strings.Index(name, "#"); index > 0 {
			name = name[:index]
		}
		description := sanitize(item.Description)
		if name == "" || description == "" {
			continue
		}
		changelog.Descriptions[name] = description
	}
	if len(changelog.Descriptions) != expectedItems {
		gologging.Warn(fmt.Sprintf(
			"gemini: got %d descriptions for %d new objects",
			len(changelog.Descriptions), expectedItems,
		))
	}
}

func collectDifferences(differences *schemeTypes.TLFullDifferences) (added, changed, removed []string) {
	for _, api := range []struct {
		scheme *schemeTypes.TLSchemeDifferences
		isE2E  bool
	}{{differences.MainApi, false}, {differences.E2EApi, true}} {
		if api.scheme == nil {
			continue
		}
		for _, group := range [][]schemeTypes.TLObjDifferences{
			api.scheme.MethodsDifference,
			api.scheme.ConstructorsDifference,
		} {
			for _, diff := range group {
				line := tlLine(diff.Object)
				if api.isE2E {
					line = "[e2e] " + line
				}
				switch {
				case diff.IsNew:
					added = append(added, line)
				case diff.IsDeleted:
					removed = append(removed, line)
				default:
					changed = append(changed, fmt.Sprintf("%s (%s)", line, describeDelta(diff)))
				}
			}
		}
	}
	return added, changed, removed
}

func describeDelta(diff schemeTypes.TLObjDifferences) string {
	var parts []string
	if len(diff.NewFields) > 0 {
		parts = append(parts, "added: "+strings.Join(diff.NewFields, ", "))
	}
	if len(diff.RemovedFields) > 0 {
		parts = append(parts, "removed: "+strings.Join(diff.RemovedFields, ", "))
	}
	for _, field := range diff.ChangedFields {
		parts = append(parts, fmt.Sprintf("%s: %s -> %s", field.Name, field.OldType, field.NewType))
	}
	if diff.ChangedResult != nil {
		parts = append(parts, fmt.Sprintf("returns: %s -> %s", diff.ChangedResult.OldType, diff.ChangedResult.NewType))
	}
	return strings.Join(parts, "; ")
}

func buildContext(fullScheme *schemeTypes.TLFullScheme, differences *schemeTypes.TLFullDifferences) []string {
	if fullScheme == nil {
		return nil
	}
	byResult := make(map[string][]string)
	for _, api := range []schemeTypes.TLScheme{fullScheme.MainApi, fullScheme.E2EApi} {
		for _, object := range append(api.GetConstructors(), api.GetMethods()...) {
			result := object.Result()
			if len(byResult[result]) < maxContextPerType {
				byResult[result] = append(byResult[result], tlLine(object))
			}
		}
	}

	var context []string
	seen := make(map[string]bool)
	appendType := func(name string) {
		if seen[name] || slices.Contains(primitiveTypes, name) || len(context) >= maxContextLines {
			return
		}
		seen[name] = true
		for _, line := range byResult[name] {
			context = append(context, line)
		}
	}
	for _, api := range []*schemeTypes.TLSchemeDifferences{differences.MainApi, differences.E2EApi} {
		if api == nil {
			continue
		}
		for _, group := range [][]schemeTypes.TLObjDifferences{
			api.MethodsDifference,
			api.ConstructorsDifference,
		} {
			for _, diff := range group {
				if diff.IsDeleted {
					continue
				}
				appendType(baseType(diff.Object.Result()))
				for _, param := range diff.Object.Parameters() {
					appendType(baseType(param.Type))
				}
			}
		}
	}
	return context
}

func tlLine(object schemeTypes.TLInterface) string {
	line := fmt.Sprintf("%s#%s", object.Package(), scheme.ParseConstructor(object.Constructor()))
	for _, param := range object.Parameters() {
		line += fmt.Sprintf(" %s:%s", param.Name, param.Type)
	}
	return line + " = " + object.Result() + ";"
}

func baseType(fieldType string) string {
	fieldType = flagPrefix.ReplaceAllString(strings.TrimSpace(fieldType), "")
	for {
		match := vectorWrapper.FindStringSubmatch(fieldType)
		if match == nil {
			break
		}
		fieldType = match[1]
	}
	return strings.TrimLeft(fieldType, "%!")
}

var changelogSchema = &genai.Schema{
	Type: genai.TypeObject,
	Properties: map[string]*genai.Schema{
		"lead": {
			Type:        genai.TypeString,
			Description: "Opening paragraph of the article, one or two sentences.",
		},
		"sections": {
			Type:     genai.TypeArray,
			MaxItems: genai.Ptr(int64(6)),
			Items: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"title": {
						Type:        genai.TypeString,
						Description: "Short feature-style heading, no API names.",
					},
					"paragraphs": {
						Type:     genai.TypeArray,
						MaxItems: genai.Ptr(int64(3)),
						Items:    &genai.Schema{Type: genai.TypeString},
					},
					"highlights": {
						Type:     genai.TypeArray,
						MaxItems: genai.Ptr(int64(6)),
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
	Required:         []string{"lead", "sections", "items"},
	PropertyOrdering: []string{"lead", "sections", "items"},
}
