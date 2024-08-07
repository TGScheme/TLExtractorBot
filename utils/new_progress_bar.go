package utils

import (
	"fmt"
	"github.com/cheggaaa/pb/v3"
	"github.com/fatih/color"
)

func NewProgressBar(size int) *pb.ProgressBar {
	bar := pb.StartNew(size)
	active := color.New(38, 2, 249, 38, 114).SprintFunc()
	inactive := color.New(38, 2, 58, 58, 58).SprintFunc()
	progress := color.New(38, 2, 72, 139, 41).SprintFunc()
	speed := color.New(38, 2, 228, 65, 49).SprintFunc()
	eta := color.New(38, 2, 4, 154, 159).SprintFunc()
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
