package internal

import (
	"path/filepath"
	"slices"
	"text/template"
)

type Templates struct {
	DB    *template.Template
	Model *template.Template
	Types *template.Template
	Store *template.Template
}

func ParseTemplates(schema Schema, templateDir string) *Templates {
	functions := template.FuncMap{
		"ToPascalCase":            ToPascalCase,
		"ToCamelCase":             ToCamelCase,
		"ToGoType":                ToGoType,
		"ToGoCase":                ToGoCase,
		"Singular":                Singular,
		"GetQueryOptions":         GetQueryOptions,
		"GetSprintfFormatFromKey": GetSprintfFormatFromKey,
		"GetSprintfFormat": func(query *Column) string {
			return internalGoType(*query, true, false)
		},
		"IsBulkQuery":         IsBulkQuery,
		"GetArrays":           GetArrays,
		"GetParamsOrdered":    GetParamsOrdered,
		"FilterColumnsByKeys": FilterColumnsByKeys,
		"contains": func(s []string, e string) bool {
			return slices.Contains(s, e)
		},
		"sub": func(a ...int) int {
			result := a[0]
			for _, v := range a[1:] {
				result -= v
			}
			return result
		},
		"bool_to_int": func(b bool) int {
			if b {
				return 1
			}
			return 0
		},
		"FindCompatible": FindCompatible(schema.Tables),
	}
	tDB := parseTemplate("db.tpl", templateDir, functions)
	tModel := parseTemplate("model.tpl", templateDir, functions)
	tTypes := parseTemplate("types.tpl", templateDir, functions)
	tStore := parseTemplate("store.tpl", templateDir, functions)

	return &Templates{
		DB:    tDB,
		Model: tModel,
		Types: tTypes,
		Store: tStore,
	}
}

func parseTemplate(name, dir string, functions template.FuncMap) *template.Template {
	tpl, err := template.New(name).Funcs(functions).ParseFiles(filepath.Join(dir, name))
	checkErr(err)
	return tpl
}
