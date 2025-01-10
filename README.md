# countdown

Tiny Go library that generates a GIF image with timer countdown.

Example:

![output.gif](https://github.com/user-attachments/assets/3866f8c6-e035-4d2c-bc85-d696b80ca139)

## Usage

```go
package main

import (
	...

	"github.com/chuhlomin/countdown"
)

func main() {
	...

	gen, err := countdown.NewGenerator(
		countdown.WithFontPath(fontPath),
		countdown.WithBackgroundImage(image),
		countdown.WithTimeFrom(2 * time.Hour),
		...
	)

	...
	err = gen.Generate(writer)
	...
}


```

## cli

At `cmd/cli` there is a simple CLI app that uses the library.

Available flags:

```
  -bg string
    	background color (default "black")
  -bi string
   	path to background image (optional)
  -c string
    	text color (default "white")
  -ca
   	auto compensate for colon Y position
  -cy int
   	compensate for colon Y position
  -f string
    	path to font file
  -from duration
   	duration to start countdown from
  -h int
     	image height (default 400)
  -max int
    	max frames
  -o string
    	output file (default "output.gif")
  -pm int
   	max colors in palette
  -s float
    	font size (default 48)
  -t int
   	target time in Unix format
  -w int
   	image width (default 600)
```

If `-f` flag is not provided, the app will use the default Face7x13 font.

If `-max` flag is not provided, the app will generate all frames until the end of the countdown.

If `-ca` flag is provided, `-cy` flag will be ignored.

`-t` is an alternative to `-from` flag. If both are provided, latter will be used.

Example:

```
go run ./cmd/cli \
  -f fonts/Gorton\ Digital\ Regular.otf \
  -s 120 \
  -bg "#8af" \
  -c "white" \
  -from "2h30s" \
  -ca \
  -max 100
```

## server

At `cmd/server` there is a simple HTTP server that uses the library.

Start it with:

```
go run ./cmd/server
```

Then open `http://localhost:8080/?from=1m` in your browser.

It supports almost the same flags as the CLI app, but they should be passed as query parameters, e.g.:

```
http://localhost:8080/?s=100&f=Gorton%20Digital%20Light.otf&bg=%23E2D9C5&c=%23141414&from=2h&max=10&ca&bi=retro.png
```

(assuming you have `cmd/server/fonts` directory)

Docker-compose file is provided for the server:

```
docker-compose up
```

Then open `http://localhost:8080/?s=140&f=Gidolinya-Regular.otf&bg=%23E2D9C5&c=%23141414&from=2h&max=10&ca&bi=retro.png` in your browser.
