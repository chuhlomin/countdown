package main

import (
	"bytes"
	"embed"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/chuhlomin/countdown"
)

//go:embed fonts/*
var fonts embed.FS

//go:embed images/*
var img embed.FS

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/" {
			http.NotFound(w, req)
			return
		}

		opts, err := processRequest(req)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to process request: %v", err), http.StatusBadRequest)
			return
		}

		gen, err := countdown.NewGenerator(opts...)

		if err != nil {
			http.Error(w, fmt.Sprintf("failed to create generator: %v", err), http.StatusInternalServerError)
			return
		}

		buf := new(bytes.Buffer)
		if err := gen.Write(buf); err != nil {
			http.Error(w, fmt.Sprintf("failed to generate GIF: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "image/gif")
		w.Header().Set("Content-Length", fmt.Sprintf("%d", buf.Len()))
		w.Header().Set("Cache-Control", "no-store")
		w.Write(buf.Bytes())
	})

	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}

var parseMap = map[string]func(string) (interface{}, error){
	"f":    func(s string) (interface{}, error) { return fonts.ReadFile(fmt.Sprintf("fonts/%s", s)) },
	"bi":   func(s string) (interface{}, error) { return img.ReadFile(fmt.Sprintf("images/%s", s)) },
	"s":    func(s string) (interface{}, error) { return strconv.ParseFloat(s, 64) },
	"from": func(s string) (interface{}, error) { return time.ParseDuration(s) },
	"max":  func(s string) (interface{}, error) { return strconv.Atoi(s) },
	"w":    func(s string) (interface{}, error) { return strconv.Atoi(s) },
	"h":    func(s string) (interface{}, error) { return strconv.Atoi(s) },
	"cy":   func(s string) (interface{}, error) { return strconv.Atoi(s) },
}

var applyMap = map[string]func(interface{}) countdown.Option{
	"f":    func(v interface{}) countdown.Option { return countdown.WithFontOpenTypeData(v.([]byte)) },
	"s":    func(v interface{}) countdown.Option { return countdown.WithFontSize(v.(float64)) },
	"bg":   func(v interface{}) countdown.Option { return countdown.WithBackgroundColor(v.(string)) },
	"bi":   func(v interface{}) countdown.Option { return countdown.WithBackgroundImageData(v.([]byte)) },
	"c":    func(v interface{}) countdown.Option { return countdown.WithTextColor(v.(string)) },
	"from": func(v interface{}) countdown.Option { return countdown.WithTimeFrom(v.(time.Duration)) },
	"max":  func(v interface{}) countdown.Option { return countdown.WithMaxFrames(v.(int)) },
	"w":    func(v interface{}) countdown.Option { return countdown.WithWidth(v.(int)) },
	"h":    func(v interface{}) countdown.Option { return countdown.WithHeight(v.(int)) },
	"cy":   func(v interface{}) countdown.Option { return countdown.WithColonCompensation(v.(int)) },
	"ca":   func(v interface{}) countdown.Option { return countdown.WithColonCompensationAuto() },
}

func processRequest(req *http.Request) ([]countdown.Option, error) {
	var opts []countdown.Option

	var (
		val interface{}
		err error
	)

	for k, v := range req.URL.Query() {
		val = v[0]
		if parser, ok := parseMap[k]; ok {
			val, err = parser(v[0])
			if err != nil {
				return nil, fmt.Errorf("failed to parse value for %s: %v", k, err)
			}
		}

		if applier, ok := applyMap[k]; ok {
			opts = append(opts, applier(val))
		}
	}

	return opts, nil
}
