package android

import (
	"TLExtractor/consts"
	"TLExtractor/environment"
	"os"
	"path"
	"regexp"
	"strconv"
)

func extractLayer() (int, error) {
	var fileName string
	if isLegacyScheme() {
		fileName = "TLRPC$Message.java"
	} else {
		fileName = "TLRPC.java"
	}
	readFile, err := os.ReadFile(path.Join(environment.EnvFolder, consts.TempSources, fileName))
	if err != nil {
		return -1, err
	}
	layer := regexp.MustCompile(`this.layer = ([0-9]+);`).FindAllStringSubmatch(string(readFile), -1)[0][1]
	return strconv.Atoi(layer)
}
