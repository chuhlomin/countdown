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

## Options

| Option                      | CLI flag | GET parameter | Description                          | Default      |
| --------------------------- | -------- | ------------- | ------------------------------------ | ------------ |
| `WithBackgroundColor`       | `-bg`    | `bg`          | Background color                     | "black"      |
| `WithBackgroundImageData`   |          |               | Background image bytes (optional)    |              |
| `WithBackgroundImagePath`   | `-bi`    | `bi`          | Path to background image (optional)  |              |
| `WithColonCompensationAuto` | `-ca`    | `ca`          | Auto compensate for colon Y position | false        |
| `WithColonCompensation`     | `-cy`    | `cy`          | Compensate for colon Y position      | 0            |
| `WithFontOpenTypeData`      |          |               | OpenType font bytes                  |              |
| `WithFontPath`              | `-f`     | `f`           | Path to font file                    |              |
| `WithFontSize`              | `-s`     | `s`           | Font size                            | 48           |
| `WithImageHeight`           | `-h`     | `h`           | Image height                         | 400          |
| `WithImageWidth`            | `-w`     | `w`           | Image width                          | 600          |
| `WithMaxFrames`             | `-max`   | `max`         | Max frames                           |              |
| `WithoutLeadingZeros`       | `-no0`   | `no0`         | Do not show leading zeros            | false        |
| `WithPaletteMaxColors`      | `-pm`    | `pm`          | Max colors in palette                | 256          |
| `WithPalleteMaxColorsAuto`  | `-pma`   | `pma`         | Auto calculate optimal palette size  | false        |
| `WithTargetTime`            | `-t`     | `t`           | Target time in Unix format           |              |
| `WithTextColor`             | `-c`     | `c`           | Text color                           | "white"      |
| `WithTimeFrom`              | `-from`  | `from`        | Duration to start countdown from     |              |
|                             | `-o`     |               | Output file                          | "output.gif" |

If font is not provided, the app will use the default fixed-size `Face7x13` font.

If `WithMaxFrames` is not provided, the app will generate all frames until the end of the countdown.

If `WithColonCompensationAuto` flag is provided, `WithColonCompensation` flag will be ignored.

`WithTargetTime` is an alternative to `WithTimeFrom` option. If both are provided, latter will be used.

Examples of options effect:

| Option                      | false                                                                                                  | true                                                                                             |
| --------------------------- | ------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------ |
| `WithColonCompensationAuto` | ![noCA](https://github.com/user-attachments/assets/617efae8-fb08-4c2d-94bc-1c6d221b29bf)               | ![ca](https://github.com/user-attachments/assets/3e247c24-3cee-4017-9bc4-ec4f0d170d89)           |
| `WithoutLeadingZeros`       | ![with0](https://github.com/user-attachments/assets/1ba438f7-5a00-440f-839d-df5e1fa96406)              | ![no0](https://github.com/user-attachments/assets/7be541fa-eb56-4e96-814f-24213b0426aa)          |
| `WithPalleteMaxColorsAuto`  | ![allColors](https://github.com/user-attachments/assets/b815db57-773e-403d-b54f-7ac1a01063d4)<br>(37K) | ![pma](https://github.com/user-attachments/assets/d6546913-7601-4434-a4c9-5295c99a5051)<br>(27K) |

## cli

At `cmd/cli` there is a simple CLI app that uses the library.

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
