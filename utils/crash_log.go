package utils

import (
	"fmt"
	"github.com/fatih/color"
	"log"
)

func CrashLog(err error, fatal bool) {
	if err == nil {
		return
	}
	red := color.New(color.Bold, 38, 2, 239, 81, 98).SprintFunc()
	funcName := log.Println
	if fatal {
		funcName = log.Fatal
	}
	funcName(
		fmt.Sprintf(
			"%s %s",
			red("error:"),
			err,
		),
	)
}
