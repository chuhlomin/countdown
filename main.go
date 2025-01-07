package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/image/colornames"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

func main() {
	log.Println("Starting...")

	if err := run(); err != nil {
		log.Fatalf("ERROR: %v", err)
	}

	log.Println("Done.")
}

func run() error {
	fontPath := flag.String("f", "", "path to font file")
	fontSize := flag.Float64("s", 48, "font size")
	backgroundColor := flag.String("bg", "black", "background color")
	backgroundImage := flag.String("bi", "", "path to background image (optional)")
	textColor := flag.String("c", "white", "text color")
	timeFrom := flag.Duration("from", 0, "start time")
	maxFramex := flag.Int("max", 0, "max frames")
	width := flag.Int("w", 600, "image width")
	height := flag.Int("h", 400, "image height")
	out := flag.String("o", "output.gif", "output file")
	colonCompensation := flag.Int("cy", 0, "compensate for colon Y position")
	colonCompensationAuto := flag.Bool("ca", false, "auto compensate for colon Y position")
	flag.Parse()

	var (
		fontFace font.Face
		err      error
	)
	if *fontPath == "" {
		fontFace = basicfont.Face7x13
		log.Println("Font path is not provided, using basicfont")
	} else {
		fontFace, err = loadFont(*fontPath, *fontSize)
		if err != nil {
			return fmt.Errorf("failed to load font: %v", err)
		}
	}

	bg, err := parseColor(*backgroundColor)
	if err != nil {
		return fmt.Errorf("failed to parse color: %v", err)
	}

	c, err := parseColor(*textColor)
	if err != nil {
		return fmt.Errorf("failed to parse color: %v", err)
	}

	var count int

	var frames []image.Image

	fontDrawer := &font.Drawer{
		Src:  image.NewUniform(c),
		Face: fontFace,
	}

	if *colonCompensationAuto {
		// for most fonts, the colon is placed at the bottom of the cell, and has x-height height
		// to center it vertically, we need to move it up by (capHeight - xHeight) / 2
		*colonCompensation = (fontFace.Metrics().CapHeight.Ceil() - fontFace.Metrics().XHeight.Ceil()) / 2
	}

	var bi *image.Image
	if *backgroundImage != "" {
		bi, err = loadImage(*backgroundImage)
		if err != nil {
			return fmt.Errorf("failed to load background image: %v", err)
		}
	}

	for *timeFrom > 0 && (*maxFramex == 0 || count < *maxFramex) {
		frame, err := renderFrame(
			*width,
			*height,
			fontDrawer,
			fontFace,
			bg,
			c,
			bi,
			timeFrom,
			*colonCompensation,
		)
		if err != nil {
			return fmt.Errorf("failed to render frame: %v", err)
		}

		frames = append(frames, frame)

		// decrease timeFrom by 1 second
		*timeFrom = *timeFrom - time.Second
		count++
	}

	return saveFile(frames, *out)
}

func loadFont(path string, size float64) (font.Face, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer f.Close()

	fontData, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %v", err)
	}

	// switch between opentype and truetype based on file extension
	switch ext := filepath.Ext(path); ext {
	case ".otf", ".ttf":
		return loadOpenTypeFont(fontData, size)
	default:
		return nil, fmt.Errorf("unsupported font format: %s", ext)
	}
}

func loadOpenTypeFont(data []byte, size float64) (font.Face, error) {
	otFontFace, err := opentype.Parse(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse font: %v", err)
	}

	return opentype.NewFace(otFontFace, &opentype.FaceOptions{
		Size:    size,
		DPI:     72,
		Hinting: font.HintingFull,
	})
}

func parseColor(colorNameOrCode string) (color.Color, error) {
	if color, ok := colornames.Map[colorNameOrCode]; ok {
		return color, nil
	}

	return parseHexColor(colorNameOrCode)
}

var errInvalidColorHexFormat = fmt.Errorf("invalid color format")

func parseHexColor(hex string) (c color.RGBA, err error) {
	if hex[0] != '#' {
		return c, errInvalidColorHexFormat
	}

	hexToByte := func(b byte) byte {
		switch {
		case b >= '0' && b <= '9':
			return b - '0'
		case b >= 'a' && b <= 'f':
			return b - 'a' + 10
		case b >= 'A' && b <= 'F':
			return b - 'A' + 10
		}
		err = errInvalidColorHexFormat
		return 0
	}

	c.A = 0xff
	switch len(hex) {
	case 9:
		c.R = hexToByte(hex[1])<<4 + hexToByte(hex[2])
		c.G = hexToByte(hex[3])<<4 + hexToByte(hex[4])
		c.B = hexToByte(hex[5])<<4 + hexToByte(hex[6])
		c.A = hexToByte(hex[7])<<4 + hexToByte(hex[8])
	case 7:
		c.R = hexToByte(hex[1])<<4 + hexToByte(hex[2])
		c.G = hexToByte(hex[3])<<4 + hexToByte(hex[4])
		c.B = hexToByte(hex[5])<<4 + hexToByte(hex[6])
	case 4:
		c.R = hexToByte(hex[1]) * 17
		c.G = hexToByte(hex[2]) * 17
		c.B = hexToByte(hex[3]) * 17
	default:
		err = errInvalidColorHexFormat
	}
	return
}

func loadImage(path string) (*image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}

	img, _, err := image.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %v", err)
	}

	return &img, nil
}

func renderFrame(
	width, height int,
	d *font.Drawer,
	fontFace font.Face,
	bg, c color.Color,
	bi *image.Image,
	timerDuration *time.Duration,
	colonCompensation int,
) (image.Image, error) {
	// create image 600×400 pixels with black background and white text
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(img, img.Bounds(), &image.Uniform{bg}, image.ZP, draw.Src)

	if bi != nil {
		draw.Draw(img, img.Bounds(), *bi, image.ZP, draw.Over)
	}

	d.Dst = img

	// not all fonts support tabular numbers,
	// so to avoid text jumping, we need to split it into parts
	// and draw each part separately, keeping ":" at the same position
	parts := formatTime(timerDuration)

	colonWidth := d.MeasureString(":")
	maxDigitsWidth, digit := findMaxDigitsWidth(d)
	totalWidth := d.MeasureString(strings.Repeat(":", len(parts)-1))
	for _, part := range parts {
		totalWidth += d.MeasureString(strings.Repeat(digit, len(part)))
	}

	x := (fixed.I(img.Bounds().Dx()) - totalWidth) / 2
	y := fixed.I(img.Bounds().Dy()+fontFace.Metrics().CapHeight.Ceil()) / 2
	d.Dot = fixed.Point26_6{X: x, Y: y}

	for i, part := range parts {
		d.Dot.X = x

		if i > 0 {
			d.Dot.Y -= fixed.I(colonCompensation)
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

func formatTime(d *time.Duration) []string {
	// format time as 00:00:00 if it's more than 1 hour
	// or 00:00 if it's less than 1 hour

	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60

	if h > 0 {
		return []string{
			fmt.Sprintf("%02d", h),
			fmt.Sprintf("%02d", m),
			fmt.Sprintf("%02d", s),
		}
	}

	return []string{
		fmt.Sprintf("%02d", m),
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

func saveFile(frames []image.Image, path string) error {
	// save image to file as GIF
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}

	g := &gif.GIF{
		Image:     make([]*image.Paletted, len(frames)),
		Delay:     make([]int, len(frames)),
		LoopCount: -1,
	}

	palette := choosePalette(frames)
	log.Printf("Palette has %d colors", len(palette))

	for i, frame := range frames {
		g.Image[i] = image.NewPaletted(frame.Bounds(), palette)
		draw.FloydSteinberg.Draw(g.Image[i], frame.Bounds(), frame, image.ZP)
		g.Delay[i] = 100
	}

	if err := gif.EncodeAll(f, g); err != nil {
		return fmt.Errorf("failed to encode image: %v", err)
	}

	// get the size of the file
	fi, err := f.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %v", err)
	}

	log.Printf("File saved: %s (%s)", path, humanizeBytes(fi.Size()))

	if err := f.Close(); err != nil {
		return fmt.Errorf("failed to close file: %v", err)
	}

	return nil
}

func choosePalette(frames []image.Image) color.Palette {
	colorsMap := map[color.Color]interface{}{}

	for _, frame := range frames {
		for i := 0; i < len(frame.(*image.RGBA).Pix); i += 4 {
			colorsMap[color.RGBA{
				frame.(*image.RGBA).Pix[i],
				frame.(*image.RGBA).Pix[i+1],
				frame.(*image.RGBA).Pix[i+2],
				frame.(*image.RGBA).Pix[i+3],
			}] = nil
		}
	}

	colors := make([]color.Color, 0, len(colorsMap))
	for color := range colorsMap {
		colors = append(colors, color)
	}

	return color.Palette(colors)
}

func humanizeBytes(size int64) string {
	units := []string{"B", "KB", "MB", "GB", "TB", "PB", "EB", "ZB", "YB"}
	unit := 0
	for size >= 1024 {
		size /= 1024
		unit++
	}
	return fmt.Sprintf("%d %s", size, units[unit])
}
