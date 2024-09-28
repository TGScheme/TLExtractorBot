package java

import (
	"TLExtractor/consts"
	"TLExtractor/environment"
	"TLExtractor/io"
	"TLExtractor/java/types"
	"errors"
	"os"
	"path"
	"slices"
	"strings"
)

func GetRawClasses(isLegacy bool) ([]*types.RawClass, error) {
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
	var replaceClasses []string
	for _, file := range dir {
		if file.IsDir {
			continue
		}
		className := strings.TrimSuffix(file.Name, ".java")
		if !isLegacy && (className == "TLRPC" || strings.HasSuffix(file.FullPath, "tl")) {
			replaceClasses = append(replaceClasses, className)
		}
	}

	for _, file := range dir {
		if file.IsDir {
			continue
		}
		readFile, err := os.ReadFile(path.Join(file.FullPath, file.Name))
		if err != nil {
			return nil, err
		}
		className := strings.TrimSuffix(file.Name, ".java")
		if slices.Contains(replaceClasses, className) {
			newFiles := SplitClasses(className, string(readFile), replaceClasses)
			for name, content := range newFiles {
				contentFiles[name] = content
			}
		} else {
			contentFiles[file.Name] = string(readFile)
		}
	}

	tempList := make(map[string]*types.RawClass)
	for name, content := range contentFiles {
		info, err := ParseClass(name, content)
		if errors.Is(err, consts.NotTLRPC) ||
			errors.Is(err, consts.OldLayer) {
			continue
		} else if err != nil {
			return nil, err
		}
		tempList[info.FullName()] = info
	}

	var tlList []*types.RawClass
	for _, tl := range tempList {
		if extendedData := tempList[tl.ParentClass]; extendedData != nil {
			tl.ParentLink = extendedData
		}
		tl.Vars = getDeclaredVars(tl)
		tlList = append(tlList, tl)
	}

	return tlList, nil
}
