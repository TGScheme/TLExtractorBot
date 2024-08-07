package utils

import (
	"os"
)

var IsDebugMode = false

func CheckDebugMode() {
	for _, arg := range os.Args {
		if arg == "--debug" {
			IsDebugMode = true
			break
		}
	}
}
