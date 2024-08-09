package screen

import (
	"TLExtractor/environment"
)

func isRunning() bool {
	screens, err := getScreens()
	if err != nil {
		return false
	}
	for _, screen := range screens {
		if screen.PID == environment.LocalStorage.ScreenPid {
			return true
		}
	}
	return false
}
