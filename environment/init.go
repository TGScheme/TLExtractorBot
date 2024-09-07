package environment

import (
	"TLExtractor/consts"
	"encoding/json"
	"errors"
	"flag"
	"github.com/Laky-64/gologging"
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
		EnvFolder = path.Join(EnvFolder, "..", ".env_debug")
		consts.SchemeRepoName = "Schema-Tests"
		gologging.SetLevel(gologging.DebugLevel)
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
