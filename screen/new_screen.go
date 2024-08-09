package screen

import (
	"TLExtractor/consts"
	"TLExtractor/environment"
	"fmt"
	"os/exec"
)

func newScreen(execPath string) (string, error) {
	err := exec.Command(
		"screen",
		"-dmS",
		consts.ServiceName,
		execPath,
	).Run()
	if err != nil {
		return "", err
	}
	activeScreens, err := getScreens()
	if err != nil {
		return "", err
	}
	for _, screenInfo := range activeScreens {
		if screenInfo.Name == consts.ServiceName {
			environment.LocalStorage.ScreenPid = screenInfo.PID
			environment.LocalStorage.Commit()
			return screenInfo.PID, nil
		}
	}
	return "", fmt.Errorf("unable to find screen pid")
}
