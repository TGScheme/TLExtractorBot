package java

import (
	"os"
	"path"

	"github.com/TGScheme/TLExtractorBot/internal/consts"
)

func IsUnified(workDir string) bool {
	tlrpcFile := path.Join(workDir, consts.TempSources, "TLRPC.java")
	_, err := os.Stat(tlrpcFile)
	return err == nil
}
