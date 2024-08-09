package environment

import (
	"TLExtractor/consts"
	"TLExtractor/logging"
	"encoding/json"
	"fmt"
	"github.com/fatih/color"
	"golang.org/x/term"
	"os"
	"path"
	"slices"
	"strings"
)

var ShellFlags []string

func init() {
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
		width, _, _ := term.GetSize(0)
		width = int(float64(width) * 0.2)
		//goland:noinspection GoBoolExpressions
		if width%2 != 0 && len(consts.DebugModeMessage)%2 == 0 {
			width--
		}
		borderSlashes := debugColors(strings.Repeat("/", width))
		fmt.Println(borderSlashes)
		messageSlash := strings.Repeat("/", width/2-1-len(consts.DebugModeMessage)/2)
		fmt.Println(
			debugColors(
				fmt.Sprintf(
					"%s %s %s",
					messageSlash,
					consts.DebugModeMessage,
					messageSlash,
				),
			),
		)
		fmt.Println(borderSlashes + "\n")
		if !slices.Contains(ShellFlags, "config") {
			consts.EnvFolder = ".env_debug"
		}
		consts.SchemeRepoName = "Schema-Tests"
	}
	if err := os.MkdirAll(consts.EnvFolder, os.ModePerm); err != nil && !os.IsExist(err) {
		logging.Fatal(err)
	}
	file, _ := os.ReadFile(path.Join(consts.EnvFolder, consts.StorageFolder))
	_ = json.Unmarshal(file, &LocalStorage)
	file, _ = os.ReadFile(path.Join(consts.EnvFolder, consts.CredentialsFolder))
	_ = json.Unmarshal(file, &CredentialsStorage)
}
