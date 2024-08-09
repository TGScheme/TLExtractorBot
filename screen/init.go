package screen

import (
	"TLExtractor/consts"
	"TLExtractor/environment"
	"TLExtractor/io"
	"TLExtractor/logging"
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
)

func init() {
	currPath, _ := filepath.Abs(os.Args[0])
	fullTempBin, _ := filepath.Abs(path.Join(consts.EnvFolder, consts.TempBins))
	isScreen := path.Dir(currPath) == fullTempBin
	action := "started"
	if !isScreen && !slices.Contains(environment.ShellFlags, "debug") && runtime.GOOS != "windows" {
		if slices.Contains(environment.ShellFlags, "kill") {
			if isRunning() {
				if err := killScreen(); err != nil {
					log.Fatal(err)
				}
				logging.Info("Service killed")
			} else {
				logging.Error("Service is not running")
			}
			os.Exit(0)
		}
		if isRunning() {
			var answer string
			if slices.Contains(environment.ShellFlags, "skip") {
				answer = "n"
			} else {
				fmt.Print("Service is already running, do you want to restart it? (y/n): ")
				_ = io.Scanln(&answer)
			}
			if slices.Contains([]string{"y", "yes"}, strings.ToLower(answer)) {
				action = "restarted"
				if err := killScreen(); err != nil {
					log.Fatal(err)
				}
			} else {
				os.Exit(0)
			}
		}
		file, err := os.ReadFile(currPath)
		if err != nil {
			log.Fatal(err)
		}
		if err = os.MkdirAll(path.Join(consts.EnvFolder, consts.TempBins), os.ModePerm); err != nil && !errors.Is(err, os.ErrExist) {
			log.Fatal(err)
		}
		filePath := path.Join(consts.EnvFolder, consts.TempBins, fmt.Sprintf("exc%s", path.Ext(path.Base(os.Args[0]))))
		if err = os.WriteFile(filePath, file, os.ModePerm); err != nil {
			log.Fatal(err)
		}
		pid, err := newScreen(filePath)
		if err != nil {
			log.Fatal(err)
		}
		logging.Info("screen:", fmt.Sprintf("Service %s successfully with PID:", action), pid)
		os.Exit(0)
	}
}
