package countdown

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"sort"
	"strings"
	"time"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

type Generator struct {
	fontFace               font.Face
	backgroundColor        color.Color
	textColor              color.Color
	backgroundImage        *image.Image
	timeFrom               time.Duration
	fontSize               float64
	width                  int
	height                 int
	maxFrames              int
	colonCompensation      int
	paletteMaxColors       int
	paletteMaxColorsAuto   bool
	colonCompoensationAuto bool
	noLeadingZeros         bool
}

func NewGenerator(opts ...Option) (*Generator, error) {
	g := &Generator{
		width:           600,
		height:          400,
		fontSize:        48,
		fontFace:        basicfont.Face7x13,
		backgroundColor: color.Black,
		textColor:       color.White,
	}
	for _, opt := range opts {
		err := opt(g)
		if err != nil {
			return nil, err
		}
	}
	return g, nil
}

func (g *Generator) Write(w io.Writer) error {
	var count int

	var frames []image.Image

	fontDrawer := &font.Drawer{
		Src:  image.NewUniform(g.textColor),
		Face: g.fontFace,
	}

	if g.colonCompoensationAuto {
		// for most fonts, the colon is placed at the bottom of the cell, and has x-height height
		// to center it vertically, we need to move it up by (capHeight - xHeight) / 2
		g.colonCompensation = (g.fontFace.Metrics().CapHeight.Ceil() - g.fontFace.Metrics().XHeight.Ceil()) / 2
	}

	for g.timeFrom > 0 && (g.maxFrames == 0 || count < g.maxFrames) {
		frame, err := g.renderFrame(fontDrawer)
		if err != nil {
			return fmt.Errorf("failed to render frame: %v", err)
		}

		frames = append(frames, frame)

		// decrease timeFrom by 1 second
		g.timeFrom = g.timeFrom - time.Second
		count++
	}

	gw := &gif.GIF{
		Image:     make([]*image.Paletted, len(frames)),
		Delay:     make([]int, len(frames)),
		LoopCount: -1,
	}

	palette := choosePalette(frames, g.paletteMaxColors, g.paletteMaxColorsAuto)

	for i, frame := range frames {
		gw.Image[i] = image.NewPaletted(frame.Bounds(), palette)
		draw.FloydSteinberg.Draw(gw.Image[i], frame.Bounds(), frame, image.Point{})
		gw.Delay[i] = 100
	}

	if err := gif.EncodeAll(w, gw); err != nil {
		return fmt.Errorf("failed to encode image: %v", err)
	}

	return nil
}

func (g *Generator) renderFrame(d *font.Drawer) (image.Image, error) {
	// create image 600×400 pixels with black background and white text
	img := image.NewRGBA(image.Rect(0, 0, g.width, g.height))
	draw.Draw(img, img.Bounds(), &image.Uniform{g.backgroundColor}, image.Point{}, draw.Src)

	if g.backgroundImage != nil {
		draw.Draw(img, img.Bounds(), *g.backgroundImage, image.Point{}, draw.Over)
	}

	d.Dst = img

	// not all fonts support tabular numbers,
	// so to avoid text jumping, we need to split it into parts
	// and draw each part separately, keeping ":" at the same position
	parts := formatTime(g.timeFrom, g.noLeadingZeros)

	colonWidth := d.MeasureString(":")
	maxDigitsWidth, digit := findMaxDigitsWidth(d)
	totalWidth := d.MeasureString(strings.Repeat(":", len(parts)-1))
	for _, part := range parts {
		totalWidth += d.MeasureString(strings.Repeat(digit, len(part)))
	}

	x := (fixed.I(img.Bounds().Dx()) - totalWidth) / 2
	y := fixed.I(img.Bounds().Dy()+g.fontFace.Metrics().CapHeight.Ceil()) / 2
	d.Dot = fixed.Point26_6{X: x, Y: y}

	for i, part := range parts {
		d.Dot.X = x

		if i > 0 {
			d.Dot.Y -= fixed.I(g.colonCompensation)
			d.DrawString(":")
			d.Dot.Y = y
			x += colonWidth
		}

		for _, r := range part {
			// align digits to the center of the "cell"
			d.Dot.X = x + (maxDigitsWidth-d.MeasureString(string(r)))/2
			x += maxDigitsWidth
			d.DrawString(string(r))
		}
	}

	return img, nil
}

func formatTime(d time.Duration, noLeadingZeros bool) []string {
	// format time as 00:00:00 if it's more than 1 hour
	// or 00:00 if it's less than 1 hour

	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60

	firstPartFormat := "%02d"
	if noLeadingZeros {
		firstPartFormat = "%d"
	}

	if h > 0 {
		return []string{
			fmt.Sprintf(firstPartFormat, h),
			fmt.Sprintf("%02d", m),
			fmt.Sprintf("%02d", s),
		}
	}

	return []string{
		fmt.Sprintf(firstPartFormat, m),
		fmt.Sprintf("%02d", s),
	}
}

func findMaxDigitsWidth(d *font.Drawer) (fixed.Int26_6, string) {
	var (
		max  fixed.Int26_6
		maxS string
	)
	for i := 0; i < 10; i++ {
		s := fmt.Sprintf("%d", i)
		w := d.MeasureString(s)
		if w > max {
			max = w
			maxS = s
		}
	}
	return max, maxS
}

func choosePalette(frames []image.Image, max int, auto bool) color.Palette {
	colorsMap := map[color.Color]int{}

	for _, frame := range frames {
		for i := 0; i < len(frame.(*image.RGBA).Pix); i += 4 {
			colorsMap[color.RGBA{
				frame.(*image.RGBA).Pix[i],
				frame.(*image.RGBA).Pix[i+1],
				frame.(*image.RGBA).Pix[i+2],
				frame.(*image.RGBA).Pix[i+3],
			}]++
		}
	}

	if !auto && (max == 0 || len(colorsMap) <= max) {
		// return all colors
		colors := make([]color.Color, 0, len(colorsMap))
		for color := range colorsMap {
			colors = append(colors, color)
		}
		return color.Palette(colors)
	}

	// sort colors by frequency
	// and choose the most frequent ones
	type colorFreq struct {
		color color.Color
		freq  int
	}

	colorsFreq := make([]colorFreq, 0, len(colorsMap))
	for color, freq := range colorsMap {
		colorsFreq = append(colorsFreq, colorFreq{color, freq})
	}

	sort.Slice(colorsFreq, func(i, j int) bool {
		return colorsFreq[i].freq > colorsFreq[j].freq
	})

	if auto {
		// pick first 10% of most frequent colors
		max = len(colorsFreq) / 10
	}

	colors := make([]color.Color, 0, max)
	for i := 0; i < max; i++ {
		colors = append(colors, colorsFreq[i].color)
	}

	return color.Palette(colors)
}
