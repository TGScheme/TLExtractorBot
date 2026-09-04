package banner

import (
	"image/color"

	"github.com/tdewolff/canvas"
)

const (
	alignLeft = iota
	alignRight
)

func (c *composition) text(x, y, size float64, style canvas.FontStyle, colour string, opacity, tracking float64, align int, runs ...string) {
	face := c.family.Face(size*ptPerPx, canvas.Hex(colour), style, canvas.FontNormal)
	if align == alignRight {
		x -= c.width(face, tracking, runs...)
	}
	for _, run := range runs {
		x = c.run(face, x, y, canvas.Hex(colour), opacity, tracking, run)
	}
}

func (c *composition) coloured(x, y, size float64, style canvas.FontStyle, opacity, tracking float64, align int, runs []textRun) {
	if align == alignRight {
		var total float64
		for _, run := range runs {
			total += c.width(c.family.Face(size*ptPerPx, canvas.Hex(run.Colour), style, canvas.FontNormal), tracking, run.Text)
		}
		x -= total
	}
	for _, run := range runs {
		face := c.family.Face(size*ptPerPx, canvas.Hex(run.Colour), style, canvas.FontNormal)
		x = c.run(face, x, y, canvas.Hex(run.Colour), opacity, tracking, run.Text)
	}
}

type textRun struct {
	Text   string
	Colour string
}

func (c *composition) run(face *canvas.FontFace, x, y float64, colour color.RGBA, opacity, tracking float64, run string) float64 {
	for _, letter := range run {
		path, advance := face.ToPath(string(letter))
		c.ctx.SetFillColor(fade(colour, opacity))
		c.ctx.SetStrokeColor(canvas.Transparent)
		c.ctx.DrawPath(x, flip(y), path)
		x += advance + tracking
	}
	return x
}

func (c *composition) width(face *canvas.FontFace, tracking float64, runs ...string) float64 {
	var total float64
	for _, run := range runs {
		for _, letter := range run {
			_, advance := face.ToPath(string(letter))
			total += advance + tracking
		}
	}
	return total
}
