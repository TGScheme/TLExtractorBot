package screen

import (
	"TLExtractor/consts"
	"os/exec"
)

func killScreen() error {
	err := exec.Command(
		"screen",
		"-S",
		consts.ServiceName,
		"-X",
		"stuff",
		"^C",
	).Run()
	if err != nil {
		return err
	}
	return nil
}
