package utils

import (
	"TLExtractor/consts"
	"os"
	"runtime"
	"strings"
)

var ShellFlags []string

func IsWindows() bool {
	return runtime.GOOS == "windows"
}

func LoadShellFlags() {
	for i, arg := range os.Args {
		if strings.Contains(arg, "--") {
			name := arg[2:]
			if name == "config" && i+1 < len(os.Args) && !strings.Contains(os.Args[i+1], "--") {
				consts.BasePath = os.Args[i+1]
			} else {
				ShellFlags = append(ShellFlags, name)
			}
		}
	}
}
