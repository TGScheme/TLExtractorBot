package types

type TLSchemeDifferences struct {
	MethodsDifference      []TLObjDifferences
	ConstructorsDifference []TLObjDifferences
}

func (scheme TLSchemeDifferences) GetConstructors() []TLInterface {
	var objs []TLInterface
	for _, diff := range scheme.ConstructorsDifference {
		objs = append(objs, diff.Object)
	}
	return objs
}

func (scheme TLSchemeDifferences) GetMethods() []TLInterface {
	var objs []TLInterface
	for _, diff := range scheme.MethodsDifference {
		objs = append(objs, diff.Object)
	}
	return objs
}
