package java

import "maps"

import "github.com/TGScheme/TLExtractorBot/internal/java/types"

func appendAllExtends(parent, className string, allClassesReturn map[string]map[string]*types.AstClass, extendsPackage, extendsName string) map[string]*types.AstVar {
	var result = make(map[string]*types.AstVar)
	if classInfoExtend, ok := allClassesReturn[extendsPackage][extendsName]; ok {
		result = appendAllExtends(parent, className, allClassesReturn, classInfoExtend.ExtendsPackage, classInfoExtend.ExtendsName)
		maps.Copy(result, classInfoExtend.Vars)
	}
	return result
}
