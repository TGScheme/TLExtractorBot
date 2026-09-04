package assets

import (
	"embed"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/Laky-64/gologging"
	"github.com/flosch/pongo2/v6"
)

var Templates map[string]string

//go:embed *.gohtml
var assetsFolder embed.FS

func init() {
	Templates = make(map[string]string)
	files, _ := assetsFolder.ReadDir(".")
	for _, file := range files {
		readFile, err := assetsFolder.ReadFile(file.Name())
		if err != nil {
			gologging.Fatal(err)
		}
		name := file.Name()
		Templates[name[:len(name)-len(filepath.Ext(name))]] = string(readFile)
	}
	for key, value := range Templates {
		var foundImport, foundImportType bool
		var builtLine, builtText, importName string
		for _, char := range value {
			if char == '\n' {
				builtText += builtLine + "\r\n"
				builtLine = ""
			} else if char == '\r' {
				continue
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
							gologging.Fatal(fmt.Errorf("recursive import in %s.gohtml", key))
						}
						if res, ok := Templates[importName]; ok {
							builtText += res
							foundImport = false
							importName = ""
						} else {
							gologging.Fatal(fmt.Errorf("import %s not found in %s.gohtml", importName, key))
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
			gologging.Fatal(fmt.Errorf("import not closed in %s.gohtml", key))
		}
		Templates[key] = strings.TrimSpace(builtText)
		_, err := pongo2.FromString(Templates[key])
		if err != nil {
			gologging.Fatal("Error in template", fmt.Sprintf("\"%s.gohtml\":", key), err)
		}
	}
}
