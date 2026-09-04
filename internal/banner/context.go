package banner

import (
	"embed"
	"fmt"
	"sync"

	"github.com/tdewolff/canvas"
)

//go:embed fonts/*.ttf
var fontFiles embed.FS

type Input struct {
	Layer        int
	Title        string
	Source       string
	ChangesLabel string
	Highlight    string
	Changes      string
	IsStable     bool
}

var (
	fontOnce   sync.Once
	fontFamily *canvas.FontFamily
	fontErr    error
)

func fonts() (*canvas.FontFamily, error) {
	fontOnce.Do(func() {
		family := canvas.NewFontFamily("inter")
		for name, style := range map[string]canvas.FontStyle{
			"Inter-SemiBold.ttf":  canvas.FontSemiBold,
			"Inter-Bold.ttf":      canvas.FontBold,
			"Inter-ExtraBold.ttf": canvas.FontExtraBold,
			"Inter-Black.ttf":     canvas.FontBlack,
		} {
			data, err := fontFiles.ReadFile("fonts/" + name)
			if err != nil {
				fontErr = fmt.Errorf("read %s: %w", name, err)
				return
			}
			if err = family.LoadFont(data, 0, style); err != nil {
				fontErr = fmt.Errorf("load %s: %w", name, err)
				return
			}
		}
		fontFamily = family
	})
	return fontFamily, fontErr
}
