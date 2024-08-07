package screen

import (
	"TLExtractor/screen/types"
	"os/exec"
	"regexp"
)

func getScreens() ([]types.ScreenInfo, error) {
	res, err := exec.Command(
		"screen",
		"-ls",
	).Output()
	if err != nil && len(res) == 0 {
		return nil, err
	}
	var screens []types.ScreenInfo
	rgxScreens := regexp.MustCompile(`\d+\.(\S+)`).FindAllStringSubmatch(string(res), -1)
	for _, screen := range rgxScreens {
		screens = append(screens, types.ScreenInfo{
			PID:  screen[0],
			Name: screen[1],
		})
	}
	return screens, nil
}
