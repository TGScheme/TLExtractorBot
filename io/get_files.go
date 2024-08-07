package io

import (
	"TLExtractor/io/types"
	"os"
)

func GetFiles(path string) ([]types.FileInfo, error) {
	var files []types.FileInfo
	dir, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	for _, file := range dir {
		files = append(files, types.FileInfo{
			Name:     file.Name(),
			IsDir:    file.IsDir(),
			FullPath: path,
		})
	}
	return files, nil
}
