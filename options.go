package countdown

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"io"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/image/colornames"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

type Option func(*Generator) error

func WithWidth(width int) Option {
	return func(g *Generator) error {
		g.Width = width
		return nil
	}
}

func WithHeight(height int) Option {
	return func(g *Generator) error {
		g.Height = height
		return nil
	}
}

func WithFontSize(size float64) Option {
	return func(g *Generator) error {
		g.FontSize = size
		return nil
	}
}

func WithFontPath(path string) Option {
	return func(g *Generator) error {
		if path == "" {
			return nil
		}

		var err error
		g.FontFace, err = loadFont(path, g.FontSize)
		if err != nil {
			return fmt.Errorf("failed to load font: %v", err)
		}
		return nil
	}
}

func WithFontOpenTypeData(data []byte) Option {
	return func(g *Generator) error {
		var err error
		g.FontFace, err = loadOpenTypeFont(data, g.FontSize)
		if err != nil {
			return fmt.Errorf("failed to load font: %v", err)
		}
		return nil
	}
}

func WithBackgroundColor(c string) Option {
	return func(g *Generator) error {
		col, err := parseColor(c)
		if err != nil {
			return fmt.Errorf("failed to parse color: %v", err)
		}
		g.BackgroundColor = col
		return nil
	}
}

func WithBackgroundImagePath(path string) Option {
	return func(g *Generator) error {
		if path == "" {
			return nil
		}

		var err error
		g.BackgroundImage, err = loadImage(path)
		if err != nil {
			return fmt.Errorf("failed to load image: %v", err)
		}
		return nil
	}
}

func WithBackgroundImageData(data []byte) Option {
	return func(g *Generator) error {
		var err error
		g.BackgroundImage, err = loadImageData(data)
		if err != nil {
			return fmt.Errorf("failed to load image: %v", err)
		}
		return nil
	}
}

func WithTextColor(c string) Option {
	return func(g *Generator) error {
		col, err := parseColor(c)
		if err != nil {
			return fmt.Errorf("failed to parse color: %v", err)
		}
		g.TextColor = col
		return nil
	}
}

func WithTimeFrom(d time.Duration) Option {
	return func(g *Generator) error {
		g.TimeFrom = d
		return nil
	}
}

func WithTargetTime(t int) Option {
	return func(g *Generator) error {
		if t == 0 {
			return nil
		}
		g.TimeFrom = time.Until(time.Unix(int64(t), 0))
		if g.TimeFrom < 0 {
			return fmt.Errorf("target time is in the past")
		}
		return nil
	}
}

func WithMaxFrames(max int) Option {
	return func(g *Generator) error {
		g.MaxFrames = max
		return nil
	}
}

func WithColonCompensation(y int) Option {
	return func(g *Generator) error {
		g.ColonCompensation = y
		return nil
	}
}

func WithColonCompensationAuto() Option {
	return func(g *Generator) error {
		g.ColonCompoensationAuto = true
		return nil
	}
}

func WithPaletteMaxColors(n int) Option {
	return func(g *Generator) error {
		if n <= 0 {
			return nil
		}
		g.PaletteMaxColors = n
		return nil
	}
}

func WithPalleteMaxColorsAuto() Option {
	return func(g *Generator) error {
		g.PaletteMaxColorsAuto = true
		return nil
	}
}

func WithoutLeadingZeros() Option {
	return func(g *Generator) error {
		g.NoLeadingZeros = true
		return nil
	}
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
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %v", err)
	}

	return &img, nil
}

func loadImageData(data []byte) (*image.Image, error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %v", err)
	}

	return &img, nil
}
