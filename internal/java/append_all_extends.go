package java

import "maps"

import "github.com/TGScheme/TLExtractorBot/internal/java/types"

func appendAllExtends(
	parent, className string,
	allClassesReturn map[string]map[string]*types.AstClass,
	extendsPackage, extendsName string,
) (map[string]*types.AstVar, map[string]string) {
	var vars = make(map[string]*types.AstVar)
	var functions = make(map[string]string)
	if classInfoExtend, ok := allClassesReturn[extendsPackage][extendsName]; ok {
		vars, functions = appendAllExtends(
			parent, className, allClassesReturn,
			classInfoExtend.ExtendsPackage, classInfoExtend.ExtendsName,
		)
		maps.Copy(vars, classInfoExtend.Vars)
		maps.Copy(functions, classInfoExtend.Functions)
	}
	return vars, functions
}
