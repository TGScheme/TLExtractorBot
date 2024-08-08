package package_manager

import (
	"TLExtractor/consts"
	io2 "TLExtractor/io"
	"TLExtractor/utils/package_manager/types"
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path"
)

func install(info types.PackageInfo) error {
	r, err := zip.OpenReader(
		path.Join(
			consts.EnvFolder,
			consts.TempBins,
			info.FileName,
		),
	)
	if err != nil {
		return err
	}
	dirPackage := path.Join(consts.EnvFolder, consts.TempPackages, info.GetFullName())
	if _, err = os.Stat(dirPackage); err == nil {
		_ = os.RemoveAll(dirPackage)
	}
	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			continue
		}
		filePath := path.Join(dirPackage, f.Name)
		err = os.MkdirAll(path.Dir(filePath), os.ModePerm)
		if err != nil {
			return err
		}
		rc, err := f.Open()
		if err != nil {
			return err
		}
		w, err := os.Create(filePath)
		if err != nil {
			return err
		}
		if _, err = io.Copy(w, rc); err != nil {
			return err
		}
		if err = w.Close(); err != nil {
			return err
		}
	}
	_ = r.Close()
	if pkg, err := FindPackage(info.Name); err == nil && pkg != nil {
		fmt.Println("  Attempting uninstall: " + pkg.Name)
		fmt.Println(fmt.Sprintf("    Found existing installation: %s %s", pkg.Name, pkg.Version))
		fmt.Println(fmt.Sprintf("    Uninstalling %s:", pkg.GetFullName()))
		err := os.RemoveAll(pkg.Path)
		if err != nil {
			return err
		}
		fmt.Println(fmt.Sprintf("      Successfully uninstalled %s", pkg.GetFullName()))
	}
	if err = io2.Move(dirPackage, path.Join(consts.EnvFolder, consts.PackagesFolder, info.GetFullName())); err != nil {
		return err
	}
	if err = os.RemoveAll(dirPackage); err != nil {
		return err
	}
	fmt.Println(fmt.Sprintf("Successfully installed %s", info.GetFullName()))
	return nil
}
