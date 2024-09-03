package ios

import (
	"TLExtractor/consts"
	"archive/zip"
	"bytes"
	"fmt"
	"github.com/Laky-64/goswift/proxy"
	"github.com/blacktop/go-macho"
	"io"
)

func Decompile() (*proxy.Context, error) {
	reader, err := zip.OpenReader("../Telegram.ipa")
	if err != nil {
		return nil, err
	}
	var data []byte
	for _, file := range reader.File {
		if file.Name == consts.IOSMtProtoPath {
			open, err := file.Open()
			if err != nil {
				return nil, err
			}
			data, err = io.ReadAll(open)
			if err != nil {
				return nil, err
			}
			if err = open.Close(); err != nil {
				return nil, err
			}
		}
	}
	if err = reader.Close(); err != nil {
		return nil, err
	}
	if data == nil {
		return nil, fmt.Errorf("file %q not found", consts.IOSMtProtoPath)
	}
	file, err := macho.NewFile(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	proxyFile, err := proxy.New(file)
	if err != nil {
		return nil, err
	}
	if err = proxyFile.PreCache(); err != nil {
		return nil, err
	}
	return proxyFile, nil
}
