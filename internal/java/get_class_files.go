package java

import (
	"bytes"
	"fmt"
	"github.com/TGScheme/TLExtractorBot/internal/consts"
	"github.com/TGScheme/TLExtractorBot/internal/fsutil"
	"os"
	"path"
	"strings"

	sitter "github.com/tree-sitter/go-tree-sitter"
)

func getClassFiles(workDir string, language *sitter.Language) (map[string]string, error) {
	dir, err := fsutil.GetFiles(path.Join(workDir, consts.TempSources))
	if err != nil {
		return nil, err
	}
	newDir, err := fsutil.GetFiles(path.Join(workDir, consts.TempSources, "tl"))
	if err != nil {
		return nil, err
	}
	dir = append(dir, newDir...)

	contentFiles := make(map[string]string)

	query, queryErr := sitter.NewQuery(language, consts.JavaClassQuery)
	if queryErr != nil {
		return nil, queryErr
	}
	defer query.Close()

	if !IsUnified(workDir) {
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
			if err != nil {
				return nil, err
			}
			parser := sitter.NewParser()
			if err = parser.SetLanguage(language); err != nil {
				return nil, err
			}
			tree := parser.Parse(readFile, nil)
			cursor := sitter.NewQueryCursor()
			matches := cursor.Matches(query, tree.RootNode(), readFile)
			match := matches.Next()
			if match == nil || len(match.Captures) == 0 {
				tree.Close()
				continue
			}
			tmpClass := match.Captures[0].Node.Utf8Text(readFile)
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
