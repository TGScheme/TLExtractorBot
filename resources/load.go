package resources

import (
	"TLExtractor/consts"
	"embed"
	"fmt"
	"github.com/flosch/pongo2/v6"
	"path"
	"path/filepath"
	"strings"
)

func Load(langFolder embed.FS) error {
	files, _ := langFolder.ReadDir("templates")
	consts.Templates = make(map[string]string)
	consts.Resources = make(map[string][]byte)
	for _, file := range files {
		ext := filepath.Ext(file.Name())
		readFile, err := langFolder.ReadFile(path.Join("templates", file.Name()))
		if err != nil {
			return err
		}
		if ext == ".gohtml" {
			fileName := file.Name()[:len(file.Name())-len(ext)]
			consts.Templates[fileName] = string(readFile)
		} else {
			consts.Resources[file.Name()] = readFile
		}
	}
	for key, value := range consts.Templates {
		var foundImport, foundImportType bool
		var builtLine, builtText, importName string
		for _, char := range value {
			if char == '\n' {
				builtText += builtLine + "\n"
				builtLine = ""
			} else if strings.TrimSpace(builtLine) == "import" {
				foundImport = true
				builtLine = ""
			} else if foundImport {
				if char == ' ' {
					continue
				} else if char == '"' {
					foundImportType = !foundImportType
					if !foundImportType {
						if importName == key {
							return fmt.Errorf("recursive import in %s.gohtml", key)
						}
						if res, ok := consts.Templates[importName]; ok {
							builtText += res
							foundImport = false
							importName = ""
						} else {
							return fmt.Errorf("import %s not found in %s.gohtml", importName, key)
						}
					}
				} else if foundImportType {
					importName += string(char)
				}
			} else {
				builtLine += string(char)
			}
		}
		builtText += builtLine
		if foundImportType || foundImport {
			return fmt.Errorf("import not closed in %s.gohtml", key)
		}
		consts.Templates[key] = strings.TrimSpace(builtText)
		_, err := pongo2.FromString(consts.Templates[key])
		if err != nil {
			return err
		}
	}
	return nil
}
