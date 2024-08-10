package environment

import (
	"TLExtractor/consts"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/Laky-64/gologging"
	"github.com/fatih/color"
	"golang.org/x/term"
	"math"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

var (
	Debug     bool
	Uninstall bool
	StartTime time.Time
)

func init() {
	flag.BoolVar(&Uninstall, "U", false, "Uninstall service")
	flag.StringVar(&consts.EnvFolder, "C", consts.EnvFolder, "Configuration folder")
	flag.Parse()
	excPath, err := os.Executable()
	if err != nil {
		gologging.Fatal(err)
	}
	tmpPathFolders := strings.Split(excPath, string(os.PathSeparator))
	Debug = len(tmpPathFolders) > 4 && strings.HasPrefix(tmpPathFolders[len(tmpPathFolders)-4], "go-build")
	StartTime = time.Now()
	if Debug {
		debugColors := color.New(38, 2, 72, 139, 41).SprintFunc()
		termWidth, _, _ := term.GetSize(0)
		termWidth = int(float64(termWidth) * 0.2)
		termWidth = int(math.Max(float64(termWidth), consts.MinTermWidth))
		//goland:noinspection GoBoolExpressions
		if termWidth%2 != 0 && len(consts.DebugModeMessage)%2 == 0 {
			termWidth--
		}
		borderSlashes := debugColors(strings.Repeat("/", termWidth))
		fmt.Println(borderSlashes)
		messageSlash := strings.Repeat("/", termWidth/2-1-len(consts.DebugModeMessage)/2)
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
		consts.EnvFolder = path.Join(consts.EnvFolder, "..", ".env_debug")
		consts.SchemeRepoName = "Schema-Tests"
	}
	consts.EnvFolder, _ = filepath.Abs(consts.EnvFolder)
	if err = os.MkdirAll(consts.EnvFolder, os.ModePerm); err != nil && !os.IsExist(err) {
		gologging.Fatal(err)
	}
	file, err := os.ReadFile(path.Join(consts.EnvFolder, consts.StorageFolder))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		gologging.Fatal(err)
	}
	_ = json.Unmarshal(file, &LocalStorage)
	file, err = os.ReadFile(path.Join(consts.EnvFolder, consts.CredentialsFolder))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		gologging.Fatal(err)
	}
	_ = json.Unmarshal(file, &CredentialsStorage)
}
