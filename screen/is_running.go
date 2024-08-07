package screen

import (
	"TLExtractor/utils"
)

func isRunning() bool {
	screens, err := getScreens()
	if err != nil {
		return false
	}
	for _, screen := range screens {
		if screen.PID == utils.LocalStorage.ScreenPid {
			return true
		}
	}
	return false
}
