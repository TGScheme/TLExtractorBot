package java

import (
	"github.com/TGScheme/TLExtractorBot/internal/consts"
	"github.com/TGScheme/TLExtractorBot/internal/java/types"
	"strings"

	sitter "github.com/tree-sitter/go-tree-sitter"
	java "github.com/tree-sitter/tree-sitter-java/bindings/go"
)

func GetClasses(workDir string) (map[string]map[string]*types.AstClass, *sitter.Language, error) {
	language := sitter.NewLanguage(java.Language())
	files, err := getClassFiles(workDir, language)
	if err != nil {
		return nil, nil, err
	}

	query, queryErr := sitter.NewQuery(language, consts.JavaClassWithNameQuery)
	if queryErr != nil {
		return nil, nil, queryErr
	}
	defer query.Close()
	classVarsQuery, queryErr := sitter.NewQuery(language, consts.ExtractJavaVars)
	if queryErr != nil {
		return nil, nil, queryErr
	}
	defer classVarsQuery.Close()
	classMethQuery, queryErr := sitter.NewQuery(language, consts.ExtractJavaFunctions)
	if queryErr != nil {
		return nil, nil, queryErr
	}
	defer classMethQuery.Close()

	allClassesReturn := make(map[string]map[string]*types.AstClass)

	for parent, file := range files {
		if !strings.HasPrefix(parent, "TL") {
			continue
		}
		parser := sitter.NewParser()
		err = parser.SetLanguage(language)
		if err != nil {
			return nil, nil, err
		}
		fileBytes := []byte(file)
		tree := parser.Parse(fileBytes, nil)
		cursor := sitter.NewQueryCursor()
		matches := cursor.Matches(query, tree.RootNode(), fileBytes)
		for data := matches.Next(); data != nil; data = matches.Next() {
			className := data.Captures[1].Node.Utf8Text(fileBytes)
			if _, ok := allClassesReturn[parent]; !ok {
				allClassesReturn[parent] = make(map[string]*types.AstClass)
			}

			parserClass := sitter.NewParser()
			err = parserClass.SetLanguage(parser.Language())
			if err != nil {
				return nil, nil, err
			}
			classBytes := []byte(data.Captures[0].Node.Utf8Text(fileBytes))
			classTree := parserClass.Parse(classBytes, nil)

			classVarsCursor := sitter.NewQueryCursor()
			classVarsMatches := classVarsCursor.Matches(classVarsQuery, classTree.RootNode(), classBytes)
			classVars := make(map[string]*types.AstVar)
			for varData := classVarsMatches.Next(); varData != nil; varData = classVarsMatches.Next() {
				info, err := GetVarInfo(varData.Captures[0].Node, classBytes)
				if err != nil {
					return nil, nil, err
				}
				if info == nil {
					continue
				}
				if len(varData.Captures) > 2 {
					info.Value = varData.Captures[2].Node.Utf8Text(classBytes)
				}
				classVars[varData.Captures[1].Node.Utf8Text(classBytes)] = info
			}

			classVarsCursor.Close()

			classMethCursor := sitter.NewQueryCursor()
			classMethMatches := classMethCursor.Matches(classMethQuery, classTree.RootNode(), classBytes)
			classMethods := make(map[string]string)
			for methData := classMethMatches.Next(); methData != nil; methData = classMethMatches.Next() {
				methName := methData.Captures[1].Node.Utf8Text(classBytes)
				switch methName {
				case "readParams", "TLdeserialize", "deserializeResponse", "deserializeResponseT", "serializeToStream":

					body := methData.Captures[2].Node.Utf8Text(classBytes)
					if len(body) > len(classMethods[methName]) {
						classMethods[methName] = body
					}
				}
			}

			classMethCursor.Close()

			classTree.Close()
			parserClass.Close()
			extendNode := data.Captures[2].Node
			var extendPackageName, extendClassName string
			switch extendNode.Kind() {
			case "scoped_type_identifier":
				extendPackageName = extendNode.Child(0).Utf8Text(fileBytes)
				extendClassName = extendNode.Child(2).Utf8Text(fileBytes)
			case "type_identifier":
				extendPackageName = parent
				extendClassName = extendNode.Utf8Text(fileBytes)
			case "generic_type":

				base := extendNode.Child(0)
				switch base.Kind() {
				case "scoped_type_identifier":
					extendPackageName = base.Child(0).Utf8Text(fileBytes)
					extendClassName = base.Child(2).Utf8Text(fileBytes)
				case "type_identifier":
					extendPackageName = parent
					extendClassName = base.Utf8Text(fileBytes)
				}
			}
			allClassesReturn[parent][className] = &types.AstClass{
				Vars:           classVars,
				Functions:      classMethods,
				ExtendsPackage: extendPackageName,
				ExtendsName:    extendClassName,
			}
		}
		tree.Close()
		parser.Close()
	}

	classesReturn := make(map[string]map[string]*types.AstClass)

	for parent, classMap := range allClassesReturn {
		if _, ok := classesReturn[parent]; !ok {
			classesReturn[parent] = make(map[string]*types.AstClass)
		}
		for className, classInfo := range classMap {
			if _, ok := classInfo.Vars["constructor"]; !ok {
				continue
			}

			classesReturn[parent][className] = &types.AstClass{
				Vars:           classInfo.Vars,
				Functions:      classInfo.Functions,
				ExtendsPackage: classInfo.ExtendsPackage,
				ExtendsName:    classInfo.ExtendsName,
			}

			extendedVars, extendedFunctions := appendAllExtends(
				parent, className, allClassesReturn,
				classInfo.ExtendsPackage, classInfo.ExtendsName,
			)
			for varName, varInfo := range extendedVars {
				if _, okVar := classesReturn[parent][className].Vars[varName]; !okVar {
					classesReturn[parent][className].Vars[varName] = varInfo
				}
			}
			for functionName, body := range extendedFunctions {
				if _, okFunction := classesReturn[parent][className].Functions[functionName]; !okFunction {
					classesReturn[parent][className].Functions[functionName] = body
				}
			}
		}
	}

	return classesReturn, language, nil
}
