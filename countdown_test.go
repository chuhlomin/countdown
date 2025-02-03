package countdown

import (
	"bytes"
	"flag"
	"image/gif"
	"os"
	"path/filepath"
	"testing"
	"time"

	"golang.org/x/image/font/gofont/gobold"
)

// update is a flag to update golden files
// Usage: go test -update ./...
var update = flag.Bool("update", false, "update golden files")

func TestGenerator_Write(t *testing.T) {
	tests := []struct {
		name    string
		opts    []Option
		golden  string
		wantErr bool
	}{
		{
			name: "basic_countdown",
			opts: []Option{
				WithWidth(200),
				WithHeight(100),
				WithTimeFrom(10 * time.Second),
				WithBackgroundColor("black"),
				WithTextColor("white"),
				WithMaxFrames(3), // Limit frames for testing
			},
			golden: "basic_countdown.gif",
		},
		{
			name: "custom_colors",
			opts: []Option{
				WithWidth(200),
				WithHeight(100),
				WithTimeFrom(5 * time.Second),
				WithBackgroundColor("#afa"),
				WithTextColor("darkviolet"),
				WithMaxFrames(3),
			},
			golden: "custom_colors.gif",
		},
		{
			name: "invalid_target_time",
			opts: []Option{
				WithTargetTime(1000), // Past time
			},
			wantErr: true,
		},
		{
			name: "with_custom_font",
			opts: []Option{
				WithWidth(200),
				WithHeight(100),
				WithTimeFrom(5 * time.Second),
				WithMaxFrames(3),
				WithFontOpenTypeData(gobold.TTF),
			},
			golden: "with_custom_font.gif",
		},
		{
			name: "with_invalid_font",
			opts: []Option{
				WithWidth(200),
				WithHeight(100),
				WithTimeFrom(5 * time.Second),
				WithMaxFrames(3),
				WithFontOpenTypeData([]byte{0x00, 0x01, 0x02}),
			},
			wantErr: true,
		},
		{
			name: "without_leading_zeros",
			opts: []Option{
				WithWidth(200),
				WithHeight(100),
				WithTimeFrom(5 * time.Second),
				WithMaxFrames(3),
				WithFontOpenTypeData(gobold.TTF),
				WithoutLeadingZeros(),
			},
			golden: "without_leading_zeros.gif",
		},
		{
			name: "with_colon_compensation_auto",
			opts: []Option{
				WithWidth(200),
				WithHeight(100),
				WithTimeFrom(5 * time.Second),
				WithMaxFrames(3),
				WithFontOpenTypeData(gobold.TTF),
				WithColonCompensationAuto(),
			},
			golden: "with_colon_compensation_auto.gif",
		},
		{
			name: "with_background_image",
			opts: []Option{
				WithWidth(200),
				WithHeight(100),
				WithTimeFrom(5 * time.Second),
				WithMaxFrames(3),
				WithFontOpenTypeData(gobold.TTF),
				WithTextColor("black"),
				WithBackgroundImagePath("testdata/bg.png"),
			},
			golden: "with_background_image.gif",
		},
		{
			name: "with_palette_max_colors",
			opts: []Option{
				WithWidth(200),
				WithHeight(100),
				WithTimeFrom(5 * time.Second),
				WithMaxFrames(3),
				WithFontOpenTypeData(gobold.TTF),
				WithTextColor("black"),
				WithBackgroundImagePath("testdata/bg.png"),
				WithPaletteMaxColors(64),
			},
			golden: "with_palette_max_colors.gif",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g, err := NewGenerator(tt.opts...)
			if err != nil {
				if !tt.wantErr {
					t.Fatalf("NewGenerator() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}

			var buf bytes.Buffer
			if err := g.Write(&buf); err != nil {
				if !tt.wantErr {
					t.Fatalf("Generator.Write() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}

			if tt.wantErr {
				t.Fatal("Generator.Write() expected error")
			}

			// Compare with golden file
			golden := filepath.Join("testdata", tt.golden)
			if *update {
				if err := os.WriteFile(golden, buf.Bytes(), 0644); err != nil {
					t.Fatalf("failed to update golden file: %v", err)
				}
			}

			expected, err := os.ReadFile(golden)
			if err != nil {
				t.Fatalf("failed to read golden file: %v", err)
			}

			// Compare GIF metadata
			actualGif, err := gif.DecodeAll(bytes.NewReader(buf.Bytes()))
			if err != nil {
				t.Fatalf("failed to decode actual GIF: %v", err)
			}

			expectedGif, err := gif.DecodeAll(bytes.NewReader(expected))
			if err != nil {
				t.Fatalf("failed to decode expected GIF: %v", err)
			}

			compareGIFs(t, expectedGif, actualGif)
		})
	}
}

func compareGIFs(t *testing.T, expected, actual *gif.GIF) {
	t.Helper()

	if len(expected.Image) != len(actual.Image) {
		t.Errorf("frame count mismatch: got %d, want %d", len(actual.Image), len(expected.Image))
		return
	}

	if len(expected.Delay) != len(actual.Delay) {
		t.Errorf("delay count mismatch: got %d, want %d", len(actual.Delay), len(expected.Delay))
		return
	}

	for i := range expected.Image {
		if expected.Delay[i] != actual.Delay[i] {
			t.Errorf("frame %d delay mismatch: got %d, want %d", i, actual.Delay[i], expected.Delay[i])
		}

		bounds := expected.Image[i].Bounds()
		if bounds != actual.Image[i].Bounds() {
			t.Errorf("frame %d bounds mismatch: got %v, want %v", i, actual.Image[i].Bounds(), bounds)
			continue
		}

		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				expectedColor := expected.Image[i].At(x, y)
				actualColor := actual.Image[i].At(x, y)
				if expectedColor != actualColor {
					t.Errorf("frame %d pixel mismatch at (%d,%d): got %v, want %v", i, x, y, actualColor, expectedColor)
				}
			}
		}
	}
}

func TestParseColor(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"named_color", "white", false},
		{"hex_rgb", "#FF0000", false},
		{"hex_rgba", "#FF0000FF", false},
		{"hex_short", "#FFF", false},
		{"invalid_hex", "#GGG", true},
		{"invalid_format", "invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseColor(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseColor() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFormatTime(t *testing.T) {
	tests := []struct {
		name           string
		duration       time.Duration
		noLeadingZeros bool
		want           []string
	}{
		{
			name:     "hours",
			duration: 2*time.Hour + 30*time.Minute + 45*time.Second,
			want:     []string{"02", "30", "45"},
		},
		{
			name:     "minutes",
			duration: 30*time.Minute + 45*time.Second,
			want:     []string{"30", "45"},
		},
		{
			name:           "no_leading_zeros",
			duration:       2*time.Hour + 5*time.Minute + 45*time.Second,
			noLeadingZeros: true,
			want:           []string{"2", "05", "45"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatTime(tt.duration, tt.noLeadingZeros)
			if !compareStringSlices(got, tt.want) {
				t.Errorf("formatTime() = %v, want %v", got, tt.want)
			}
		})
	}
}

func compareStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
