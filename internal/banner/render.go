package banner

import (
	"bytes"
	"fmt"
	"image/color"
	"image/jpeg"
	"image/png"
	"strings"

	"github.com/tdewolff/canvas"
	"github.com/tdewolff/canvas/renderers/rasterizer"
)

const (
	bannerWidth  = 1280.0
	bannerHeight = 720.0
	ptPerPx      = 72.0 / 25.4
	marginLeft   = 80.0
	marginRight  = 1200.0
	titleSize    = 94.0
	titleLead    = 92.0
	titleTop     = 336.0
	titleTail    = "and More"
	pixelRatio   = 2.0
	compactRatio = 1400.0 / bannerWidth
	compactLevel = 84
)

type composition struct {
	ctx    *canvas.Context
	family *canvas.FontFamily
}

func Render(input Input) ([]byte, error) {
	target, err := draw(input)
	if err != nil {
		return nil, err
	}
	image := rasterizer.Draw(target, canvas.DPMM(pixelRatio), canvas.DefaultColorSpace)
	var out bytes.Buffer
	if err = png.Encode(&out, image); err != nil {
		return nil, fmt.Errorf("encode banner: %w", err)
	}
	return out.Bytes(), nil
}

func RenderCompact(input Input) ([]byte, error) {
	target, err := draw(input)
	if err != nil {
		return nil, err
	}
	image := rasterizer.Draw(target, canvas.DPMM(compactRatio), canvas.DefaultColorSpace)
	var out bytes.Buffer
	if err = jpeg.Encode(&out, image, &jpeg.Options{Quality: compactLevel}); err != nil {
		return nil, fmt.Errorf("encode compact banner: %w", err)
	}
	return out.Bytes(), nil
}

func draw(input Input) (*canvas.Canvas, error) {
	family, err := fonts()
	if err != nil {
		return nil, err
	}
	target := canvas.New(bannerWidth, bannerHeight)
	ctx := canvas.NewContext(target)
	c := &composition{ctx: ctx, family: family}

	c.background()
	c.ghostLayer(input.Layer)
	c.mark()
	c.header(input)
	c.title(input.Title)
	c.footer(input)
	return target, nil
}

func flip(y float64) float64 {
	return bannerHeight - y
}

func (c *composition) background() {
	gradient := canvas.NewLinearGradient(canvas.Point{X: 64, Y: 720}, canvas.Point{X: 1216, Y: 0})
	gradient.Add(0, canvas.Hex("#16302A"))
	gradient.Add(0.5, canvas.Hex("#0C1A16"))
	gradient.Add(1, canvas.Hex("#060D0B"))
	c.ctx.SetFillGradient(gradient)
	c.ctx.SetStrokeColor(canvas.Transparent)
	c.ctx.DrawPath(0, 0, canvas.Rectangle(bannerWidth, bannerHeight))
	c.ctx.SetFillGradient(nil)
}

func (c *composition) ghostLayer(layer int) {
	face := c.family.Face(392*ptPerPx, canvas.Hex("#25C87C"), canvas.FontBlack, canvas.FontNormal)
	digits := fmt.Sprintf("%d", layer)
	width := c.width(face, -22, digits)
	c.ctx.SetFillColor(canvas.Transparent)
	c.ctx.SetStrokeColor(fade(canvas.Hex("#25C87C"), 0.42))
	c.ctx.SetStrokeWidth(3)
	x := 1216 - width
	for _, letter := range digits {
		path, advance := face.ToPath(string(letter))
		c.ctx.DrawPath(x, flip(556), path)
		x += advance - 22
	}
	c.ctx.SetStrokeColor(canvas.Transparent)
	c.ctx.SetStrokeWidth(0)
}

func (c *composition) mark() {
	glyph, err := canvas.ParseSVGPath("M17 25h56 M45 25v50h34")
	if err != nil {
		return
	}
	node, err := canvas.ParseSVGPath("M66 41h6a6 6 0 0 1 6 6v6a6 6 0 0 1-6 6h-6a6 6 0 0 1-6-6v-6a6 6 0 0 1 6-6z")
	if err != nil {
		return
	}
	scaled := glyph.Copy().Scale(0.46, -0.46)
	c.ctx.SetFillColor(canvas.Transparent)
	c.ctx.SetStrokeColor(canvas.Hex("#25C87C"))
	c.ctx.SetStrokeWidth(13 * 0.46)
	c.ctx.SetStrokeCapper(canvas.RoundCap)
	c.ctx.SetStrokeJoiner(canvas.RoundJoin)
	c.ctx.DrawPath(marginLeft, flip(72), scaled)
	c.ctx.SetStrokeColor(canvas.Transparent)
	c.ctx.SetStrokeWidth(0)
	c.ctx.SetFillColor(canvas.Hex("#25C87C"))
	c.ctx.DrawPath(marginLeft, flip(72), node.Copy().Scale(0.46, -0.46))
}

func (c *composition) header(input Input) {
	c.text(128, 103, 17, canvas.FontBold, "#FFFFFF", 0.92, 1.6, alignLeft, "TL SCHEMA")
	state := "PREVIEW"
	if input.IsStable {
		state = "STABLE"
	}
	c.text(marginRight, 103, 17, canvas.FontSemiBold, "#25C87C", 1, 1.6, alignRight,
		fmt.Sprintf("LAYER %d · %s", input.Layer, state))
}

func (c *composition) title(title string) {
	lines, tail := wrap(title)
	size, tracking := titleSize, -4.2
	for size > 48 {
		face := c.family.Face(size*ptPerPx, canvas.Hex("#FFFFFF"), canvas.FontExtraBold, canvas.FontNormal)
		widest := 0.0
		for _, line := range lines {
			if width := c.width(face, tracking, line); width > widest {
				widest = width
			}
		}
		if widest <= 960 {
			break
		}
		size -= 2
		tracking = -size * 0.045
	}
	y := titleTop
	for index, line := range lines {
		colour, opacity := "#FFFFFF", 1.0
		if tail && index == len(lines)-1 {
			colour, opacity = "#D8F3E7", 0.30
		}
		c.text(marginLeft, y, size, canvas.FontExtraBold, colour, opacity, tracking, alignLeft, line)
		y += size * (titleLead / titleSize)
	}
}

func (c *composition) footer(input Input) {
	c.ctx.SetFillColor(fade(canvas.Hex("#FFFFFF"), 0.12))
	c.ctx.SetStrokeColor(canvas.Transparent)
	c.ctx.DrawPath(marginLeft, flip(586), canvas.Rectangle(marginRight-marginLeft, 1))

	c.text(marginLeft, 620, 13, canvas.FontSemiBold, "#FFFFFF", 0.40, 1.8, alignLeft, "SOURCE")
	c.text(marginLeft, 652, 21, canvas.FontSemiBold, "#FFFFFF", 0.94, 0, alignLeft, input.Source)
	c.text(marginRight, 620, 13, canvas.FontSemiBold, "#FFFFFF", 0.40, 1.8, alignRight, "CONSTRUCTORS")
	c.coloured(marginRight, 652, 21, canvas.FontSemiBold, 0.94, 0, alignRight, []textRun{
		{Text: fmt.Sprintf("%d", input.Added), Colour: "#25C87C"},
		{Text: fmt.Sprintf(" added · %d changed · %d removed", input.Changed, input.Removed), Colour: "#FFFFFF"},
	})
}

func wrap(title string) ([]string, bool) {
	tail := false
	body := title
	if index := strings.LastIndex(title, titleTail); index > 0 {
		body = strings.TrimSpace(title[:index])
		tail = true
	}
	words := strings.Fields(body)
	var lines []string
	var current string
	limit := (len(body) + 1) / 2
	for _, word := range words {
		if current == "" {
			current = word
			continue
		}
		if len(current)+1+len(word) > limit && len(lines) == 0 {
			lines = append(lines, current)
			current = word
			continue
		}
		current += " " + word
	}
	if current != "" {
		lines = append(lines, current)
	}
	if tail {
		lines = append(lines, titleTail)
	}
	return lines, tail
}

func fade(base color.Color, opacity float64) color.RGBA {
	r, g, b, _ := base.RGBA()
	return color.RGBA{
		R: uint8(float64(r>>8) * opacity),
		G: uint8(float64(g>>8) * opacity),
		B: uint8(float64(b>>8) * opacity),
		A: uint8(255 * opacity),
	}
}
