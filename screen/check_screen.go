package screen

import (
	"TLExtractor/consts"
	"TLExtractor/io"
	"TLExtractor/utils"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
)

func CheckScreen() (bool, error) {
	currPath, _ := filepath.Abs(os.Args[0])
	fullTempBin, _ := filepath.Abs(path.Join(consts.BasePath, consts.TempBins))
	isScreen := path.Dir(currPath) == fullTempBin
	action := "started"
	if !isScreen && !slices.Contains(utils.ShellFlags, "debug") && !utils.IsWindows() {
		if slices.Contains(utils.ShellFlags, "kill") {
			if isRunning() {
				if err := killScreen(); err != nil {
					return false, err
				}
				fmt.Println("Service killed")
			} else {
				fmt.Println("Service is not running")
			}
			return true, nil
		}
		if isRunning() {
			var answer string
			if slices.Contains(utils.ShellFlags, "skip") {
				answer = "n"
			} else {
				fmt.Print("Service is already running, do you want to restart it? (y/n): ")
				_ = io.Scanln(&answer)
			}
			if slices.Contains([]string{"y", "yes"}, strings.ToLower(answer)) {
				action = "restarted"
				if err := killScreen(); err != nil {
					return false, err
				}
			} else {
				return true, nil
			}
		}
		file, err := os.ReadFile(currPath)
		if err != nil {
			return false, err
		}
		if err = os.MkdirAll(path.Join(consts.BasePath, consts.TempBins), os.ModePerm); err != nil && !errors.Is(err, os.ErrExist) {
			return false, err
		}
		filePath := path.Join(consts.BasePath, consts.TempBins, fmt.Sprintf("exc%s", path.Ext(path.Base(os.Args[0]))))
		if err = os.WriteFile(filePath, file, os.ModePerm); err != nil {
			return false, err
		}
		pid, err := newScreen(filePath)
		if err != nil {
			return false, err
		}
		fmt.Println(fmt.Sprintf("Service %s successfully with PID:", action), pid)
		return true, nil
	}
	return false, nil
}
