package types

import "github.com/fatih/color"

type LogLevelInfo struct {
	Icon       byte
	Background *color.Color
	Foreground *color.Color
	TextColor  *color.Color
}
