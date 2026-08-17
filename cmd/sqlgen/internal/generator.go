package internal

import (
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
)

var (
	currentDir     = filepath.Join("cmd", "sqlgen")
	outputDir      = filepath.Join("internal", "db")
	modelsDir      = filepath.Join(outputDir, "models")
	sqlcJsonFile   = filepath.Join(currentDir, "sqlc.json")
	currentPackage = strings.Join(strings.Split(reflect.TypeOf(SQLCConfig{}).PkgPath(), "/")[:3], "/") +
		"/" + strings.ReplaceAll(outputDir, string(os.PathSeparator), "/")
)

func Generate() {
	checkErr(exec.Command("sqlc", "generate").Run())

	config := LoadSQLCConfig(sqlcJsonFile)
	publicSchema := GetPublicSchema(config)

	templates := ParseTemplates(publicSchema, filepath.Join(currentDir, "templates"))
	checkErr(os.MkdirAll(modelsDir, os.ModePerm))

	GenerateModelFiles(publicSchema, templates.Model, modelsDir)
	GenerateEnumFile(publicSchema, templates.Types, modelsDir)
	GenerateStoreFile(publicSchema, config, templates.Store, outputDir)
	GenerateDBFile(publicSchema, config, templates.DB, outputDir)

	_ = os.Remove(sqlcJsonFile)
}
