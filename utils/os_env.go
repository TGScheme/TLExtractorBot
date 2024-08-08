package utils

import (
	"TLExtractor/consts"
	"fmt"
	"github.com/fatih/color"
	"os"
	"runtime"
	"slices"
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
				consts.EnvFolder = os.Args[i+1]
			} else {
				ShellFlags = append(ShellFlags, name)
			}
		}
	}
	if slices.Contains(ShellFlags, "debug") {
		debugColors := color.New(38, 2, 72, 139, 41).SprintFunc()
		fmt.Println(debugColors("//////////////////////////"))
		fmt.Println(debugColors("/////// DEBUG MODE ///////"))
		fmt.Println(debugColors("//////////////////////////\n"))
		if !slices.Contains(ShellFlags, "config") {
			consts.EnvFolder = ".env_debug"
		}
		consts.SchemeRepoName = "Schema-Tests"
	}
}
