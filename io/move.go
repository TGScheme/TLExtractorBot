package io

import (
	"os"
	"path/filepath"
)

func Move(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		if err = os.MkdirAll(filepath.Join(dst, filepath.Dir(rel)), os.ModePerm); err != nil {
			return err
		}
		return os.Rename(path, filepath.Join(dst, rel))
	})
}
