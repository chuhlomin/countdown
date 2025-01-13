package main

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/chuhlomin/countdown"
)

func HandlerRoot() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
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
		if err == countdown.ErrTargetTimeInPast {
			// delete "t" from the query
			query := req.URL.Query()
			query.Del("t")
			req.URL.RawQuery = query.Encode()
			opts, err = processRequest(req)
			if err != nil {
				http.Error(w, fmt.Sprintf("failed to process request: %v", err), http.StatusBadRequest)
				return
			}

			gen, err = countdown.NewGenerator(
				append(
					opts,
					countdown.WithTimeFrom(0),
					countdown.WithMaxFrames(0),
				)...,
			)
		}

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
	"pm":   func(s string) (interface{}, error) { return strconv.Atoi(s) },
	"t":    func(s string) (interface{}, error) { return strconv.Atoi(s) },
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
	"pm":   func(v interface{}) countdown.Option { return countdown.WithPaletteMaxColors(v.(int)) },
	"t":    func(v interface{}) countdown.Option { return countdown.WithTargetTime(v.(int)) },
	"no0":  func(v interface{}) countdown.Option { return countdown.WithoutLeadingZeros() },
}

func processRequest(req *http.Request) ([]countdown.Option, error) {
	var (
		opts []countdown.Option
		err  error
	)

	// it's important to handle font size before font face
	// because font face depends on font size
	if v, ok := req.URL.Query()["s"]; ok {
		err := maybeAddOption(&opts, "s", v[0])
		if err != nil {
			return nil, err
		}
	}

	for k, v := range req.URL.Query() {
		err = maybeAddOption(&opts, k, v[0])
		if err != nil {
			return nil, err
		}
	}

	return opts, nil
}

func maybeAddOption(opts *[]countdown.Option, key string, value string) error {
	var (
		val interface{} = value
		err error
	)

	if parser, ok := parseMap[key]; ok {
		val, err = parser(value)
		if err != nil {
			return fmt.Errorf("failed to parse value for %s: %v", key, err)
		}
	}

	if applier, ok := applyMap[key]; ok {
		*opts = append(*opts, applier(val))
	}
	return nil
}
