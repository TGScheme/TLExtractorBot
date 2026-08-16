package java

import (
	"TLExtractor/consts"
	"TLExtractor/environment"
	"os"
	"path"
)

func IsUnified() bool {
	tlrpcFile := path.Join(environment.EnvFolder, consts.TempSources, "TLRPC.java")
	_, err := os.Stat(tlrpcFile)
	return err == nil
}
