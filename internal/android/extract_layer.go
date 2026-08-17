package android

import (
	"github.com/TGScheme/TLExtractorBot/internal/consts"
	"github.com/TGScheme/TLExtractorBot/internal/java"
	"os"
	"path"
	"regexp"
	"strconv"
)

func extractLayer(workDir string) (int, error) {
	var fileName string
	if !java.IsUnified(workDir) {
		fileName = "TLRPC$Message.java"
	} else {
		fileName = "TLRPC.java"
	}
	readFile, err := os.ReadFile(path.Join(workDir, consts.TempSources, fileName))
	if err != nil {
		return -1, err
	}
	layer := regexp.MustCompile(`this.layer = ([0-9]+);`).FindAllStringSubmatch(string(readFile), -1)[0][1]
	return strconv.Atoi(layer)
}
