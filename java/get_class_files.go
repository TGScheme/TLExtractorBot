package java

import (
	"TLExtractor/consts"
	"TLExtractor/environment"
	"TLExtractor/io"
	"bytes"
	"fmt"
	"os"
	"path"
	"strings"

	sitter "github.com/tree-sitter/go-tree-sitter"
)

func getClassFiles(language *sitter.Language) (map[string]string, error) {
	dir, err := io.GetFiles(path.Join(environment.EnvFolder, consts.TempSources))
	if err != nil {
		return nil, err
	}
	newDir, err := io.GetFiles(path.Join(environment.EnvFolder, consts.TempSources, "tl"))
	if err != nil {
		return nil, err
	}
	dir = append(dir, newDir...)

	contentFiles := make(map[string]string)

	if !IsUnified() {
		for _, file := range dir {
			if file.IsDir {
				continue
			}
			baseClassName := strings.Split(file.Name, "$")[0]
			if _, ok := contentFiles[baseClassName]; !ok {
				contentFiles[baseClassName] = fmt.Sprintf(
					"public class %s {\n",
					baseClassName,
				)
			}
			readFile, err := os.ReadFile(path.Join(file.FullPath, file.Name))
			parser := sitter.NewParser()
			err = parser.SetLanguage(language)
			if err != nil {
				return nil, err
			}
			tree := parser.Parse(readFile, nil)
			query, _ := sitter.NewQuery(language, consts.JavaClassQuery)
			cursor := sitter.NewQueryCursor()
			matches := cursor.Matches(query, tree.RootNode(), readFile)
			tmpClass := matches.Next().Captures[0].Node.Utf8Text(readFile)
			tmpClass = strings.ReplaceAll(tmpClass, fmt.Sprintf("%s$", baseClassName), "")
			tmpClass = strings.ReplaceAll(tmpClass, "$", ".")
			contentFiles[baseClassName] += tmpClass + "\n\n"
			tree.Close()
		}
		for name, content := range contentFiles {
			contentFiles[name] = fmt.Sprintf("%s\n}", content)
		}
	} else {
		for _, file := range dir {
			if file.IsDir {
				continue
			}
			readFile, err := os.ReadFile(path.Join(file.FullPath, file.Name))
			if err != nil {
				return nil, err
			}
			readFile = bytes.ReplaceAll(readFile, []byte("$$"), []byte("."))
			readFile = bytes.ReplaceAll(readFile, []byte("$"), []byte("."))
			contentFiles[strings.ReplaceAll(file.Name, ".java", "")] = string(readFile)
		}
	}

	return contentFiles, nil
}
