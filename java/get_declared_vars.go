package java

import javaTypes "TLExtractor/java/types"

func getDeclaredVars(class *javaTypes.RawClass) map[string]string {
	declaredVars := make(map[string]string)
	if class.ParentLink != nil {
		for key, value := range getDeclaredVars(class.ParentLink) {
			declaredVars[key] = value
		}
	}
	for _, line := range class.Content {
		if res := GetVarDeclaration(line); res != nil {
			declaredVars[res.Name] = res.Type
		}
	}
	return declaredVars
}
