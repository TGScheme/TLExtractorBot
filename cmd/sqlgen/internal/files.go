package internal

import (
	"bytes"
	"go/format"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"text/template"

	"github.com/Laky-64/gologging"
)

func GenerateModelFiles(schema Schema, t *template.Template, outDir string) {
	for _, table := range schema.Tables {
		if len(table.Columns) == 0 {
			continue
		}
		filename := Singular(table.Rel.Name) + "_gen.go"
		outPath := filepath.Join(outDir, filename)
		writeTemplateFile(outPath, t, map[string]any{
			"Table":   table,
			"Imports": DetectImports(table.Columns, false),
		})
	}
}

func GenerateEnumFile(schema Schema, t *template.Template, outDir string) {
	outPath := filepath.Join(outDir, "types_gen.go")
	writeTemplateFile(outPath, t, map[string]any{
		"Enums": schema.Enums,
	})
}

func GenerateStoreFile(schema Schema, config SQLCConfig, t *template.Template, outDir string) {
	stores := make(map[string][]Query)
	for _, query := range config.Queries {
		queryName := strings.TrimSuffix(query.FileName, ".sql")
		stores[queryName] = append(stores[queryName], query)
	}
	for storeName, queries := range stores {
		filename := storeName + "_store_gen.go"
		outPath := filepath.Join(outDir, filename)
		writeTemplateFile(outPath, t, map[string]any{
			"StoreName":       storeName,
			"Queries":         queries,
			"Imports":         DetectQueryImports(schema.Tables, queries),
			"SharedReturning": DetectSharedReturning(queries),
		})
	}
}

func GenerateDBFile(schema Schema, config SQLCConfig, t *template.Template, outDir string) {
	outPath := filepath.Join(outDir, "db_gen.go")
	queries := make(map[string]struct{})
	for _, query := range config.Queries {
		queries[strings.TrimSuffix(query.FileName, ".sql")] = struct{}{}
	}
	writeTemplateFile(outPath, t, map[string]any{
		"Queries": slices.Collect(maps.Keys(queries)),
		"Enums":   schema.Enums,
	})
}

func writeTemplateFile(path string, tpl *template.Template, data any) {
	var b bytes.Buffer
	checkErr(tpl.Execute(&b, data))
	formatted, err := format.Source(b.Bytes())
	checkErr(err)
	checkErr(os.WriteFile(path, formatted, os.ModePerm))
}

func checkErr(err error) {
	if err != nil {
		gologging.Fatal(err)
	}
}
