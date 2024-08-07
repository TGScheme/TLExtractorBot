package java

import (
	"TLExtractor/consts"
	"TLExtractor/io"
	"TLExtractor/java/types"
	"errors"
	"os"
	"path"
)

func GetRawClasses() ([]*types.RawClass, error) {
	dir, err := io.GetFiles(path.Join(consts.BasePath, consts.TempSources))
	if err != nil {
		return nil, err
	}
	newDir, err := io.GetFiles(path.Join(consts.BasePath, consts.TempSources, "tl"))
	if err != nil {
		return nil, err
	}
	dir = append(dir, newDir...)

	tempList := make(map[string]*types.RawClass)
	for _, file := range dir {
		if file.IsDir {
			continue
		}
		readFile, err := os.ReadFile(path.Join(file.FullPath, file.Name))
		if err != nil {
			return nil, err
		}
		info, err := ParseClass(file.Name, string(readFile))
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
