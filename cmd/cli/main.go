package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/chuhlomin/countdown"
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
	timeFrom := flag.Duration("from", 0, "duration to start countdown from")
	targetTime := flag.Int("t", 0, "target time in Unix format")
	maxFrames := flag.Int("max", 0, "max frames")
	width := flag.Int("w", 600, "image width")
	height := flag.Int("h", 400, "image height")
	out := flag.String("o", "output.gif", "output file")
	colonCompensation := flag.Int("cy", 0, "compensate for colon Y position")
	colonCompensationAuto := flag.Bool("ca", false, "auto compensate for colon Y position")
	paletteMaxColors := flag.Int("pm", 0, "max colors in palette")
	flag.Parse()

	if *fontPath == "" {
		log.Println("Font path is not provided, using basicfont")
	}

	opts := []countdown.Option{
		countdown.WithWidth(*width),
		countdown.WithHeight(*height),
		countdown.WithFontSize(*fontSize),
		countdown.WithFontPath(*fontPath),
		countdown.WithBackgroundColor(*backgroundColor),
		countdown.WithBackgroundImage(*backgroundImage),
		countdown.WithTextColor(*textColor),
		countdown.WithTimeFrom(*timeFrom),
		countdown.WithTargetTime(*targetTime),
		countdown.WithMaxFrames(*maxFrames),
		countdown.WithColonCompensation(*colonCompensation),
		countdown.WithPaletteMaxColors(*paletteMaxColors),
	}

	if *colonCompensationAuto {
		opts = append(opts, countdown.WithColonCompensationAuto())
	}

	gen, err := countdown.NewGenerator(opts...)
	if err != nil {
		return fmt.Errorf("failed to create generator: %v", err)
	}

	f, err := os.Create(*out)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}

	if err := gen.Write(f); err != nil {
		return fmt.Errorf("failed to generate GIF: %v", err)
	}

	// get the size of the file
	fi, err := f.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %v", err)
	}

	log.Printf("File saved: %s (%s)", *out, humanizeBytes(fi.Size()))

	if err := f.Close(); err != nil {
		return fmt.Errorf("failed to close file: %v", err)
	}

	return nil
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
