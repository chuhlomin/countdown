package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"io"
	"log"
	"os"
	"path/filepath"
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
	textColor := flag.String("c", "white", "text color")
	timeFrom := flag.Duration("from", 0, "start time")
	maxFramex := flag.Int("max", 0, "max frames")
	out := flag.String("o", "output.gif", "output file")
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

	for *timeFrom > 0 && (*maxFramex == 0 || count < *maxFramex) {
		frame, err := renderFrame(fontFace, bg, c, timeFrom)
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

func renderFrame(
	fontFace font.Face,
	bg, c color.Color,
	timerDuration *time.Duration,
) (image.Image, error) {
	// create image 600×400 pixels with black background and white text
	img := image.NewRGBA(image.Rect(0, 0, 600, 400))
	draw.Draw(img, img.Bounds(), &image.Uniform{bg}, image.ZP, draw.Src)

	text := formatTime(timerDuration)

	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(c),
		Face: fontFace,
	}
	d.Dot = fixed.Point26_6{
		X: (fixed.I(img.Bounds().Dx()) - d.MeasureString(text)) / 2,
		Y: fixed.I(img.Bounds().Dy()+fontFace.Metrics().CapHeight.Ceil()) / 2,
	}
	d.DrawString(text)

	return img, nil
}

func formatTime(d *time.Duration) string {
	// format time as 00:00:00 if it's more than 1 hour
	// or 00:00 if it's less than 1 hour

	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60

	if h > 0 {
		return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
	}

	return fmt.Sprintf("%02d:%02d", m, s)
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

	pallete := choosePalette(frames)
	log.Printf("Pallete has %d colors", len(pallete))

	for i, frame := range frames {
		g.Image[i] = image.NewPaletted(frame.Bounds(), pallete)
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
