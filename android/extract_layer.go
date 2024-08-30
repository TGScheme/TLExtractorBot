package android

import (
	"TLExtractor/consts"
	"TLExtractor/environment"
	"os"
	"path"
	"regexp"
	"strconv"
)

func ExtractLayer() (int, error) {
	readFile, err := os.ReadFile(path.Join(environment.EnvFolder, consts.TempSources, "TLRPC$Message.java"))
	if err != nil {
		return -1, err
	}
	layer := regexp.MustCompile(`this.layer = ([0-9]+);`).FindAllStringSubmatch(string(readFile), -1)[0][1]
	return strconv.Atoi(layer)
}
