package environment

import (
	"TLExtractor/consts"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/Laky-64/gologging"
	"github.com/charmbracelet/lipgloss"
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
	EnvFolder = ".env"
)

func init() {
	flag.BoolVar(&Uninstall, "U", false, "Uninstall service")
	flag.StringVar(&EnvFolder, "C", EnvFolder, "Configuration folder")
	flag.Parse()
	excPath, err := os.Executable()
	if err != nil {
		gologging.Fatal(err)
	}
	tmpPathFolders := strings.Split(excPath, string(os.PathSeparator))
	Debug = len(tmpPathFolders) > 4 && strings.HasPrefix(tmpPathFolders[len(tmpPathFolders)-4], "go-build")
	StartTime = time.Now()
	if Debug {
		termWidth, _, _ := term.GetSize(int(os.Stdout.Fd()))
		termWidth = int(math.Max(float64(termWidth)*0.20, consts.MinTermWidth))
		var debugStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#488b29")).
			Foreground(lipgloss.Color("#488b29")).
			Align(lipgloss.Center).
			Width(termWidth)
		fmt.Println(debugStyle.Render(consts.DebugModeMessage))
		EnvFolder = path.Join(EnvFolder, "..", ".env_debug")
		consts.SchemeRepoName = "Schema-Tests"
	}
	EnvFolder, _ = filepath.Abs(EnvFolder)
	if err = os.MkdirAll(EnvFolder, os.ModePerm); err != nil && !os.IsExist(err) {
		gologging.Fatal(err)
	}
	file, err := os.ReadFile(path.Join(EnvFolder, consts.StorageFolder))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		gologging.Fatal(err)
	}
	_ = json.Unmarshal(file, &LocalStorage)
	file, err = os.ReadFile(path.Join(EnvFolder, consts.CredentialsFolder))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		gologging.Fatal(err)
	}
	_ = json.Unmarshal(file, &CredentialsStorage)
}
