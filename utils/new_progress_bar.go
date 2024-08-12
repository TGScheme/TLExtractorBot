package utils

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"github.com/cheggaaa/pb/v3"
)

func NewProgressBar(size int) *pb.ProgressBar {
	bar := pb.StartNew(size)
	active := lipgloss.NewStyle().Foreground(lipgloss.Color("#f92672")).Render
	inactive := lipgloss.NewStyle().Foreground(lipgloss.Color("#3a3a3a")).Render
	progress := lipgloss.NewStyle().Foreground(lipgloss.Color("#488b29")).Render
	speed := lipgloss.NewStyle().Foreground(lipgloss.Color("#e44131")).Render
	eta := lipgloss.NewStyle().Foreground(lipgloss.Color("#049a9f")).Render
	tmpl := fmt.Sprintf(
		" {{bar . \" \" \"%s\" \"%s\" \"%s\" \" \"}} {{counters . \"%s\" \"%s/%s\"}} {{speed . \"%s\" \"%s\"}} eta {{rtime . \"%s\"}}",
		active("━"),
		inactive("╺"),
		inactive("━"),
		progress("%s/%s"),
		progress("%s"),
		inactive("%s"),
		speed("%s/s"),
		inactive("??/s"),
		eta("%s"),
	)
	bar.SetTemplateString(tmpl)
	bar.SetMaxWidth(bar.Width() * 35 / 100)
	bar.Set(pb.SIBytesPrefix, true)
	bar.Set(pb.Bytes, true)
	return bar
}
