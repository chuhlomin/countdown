# gif-experiment

Tiny Go app that generates a GIF image with timer countdown.

Example:

![output.gif](https://github.com/user-attachments/assets/3866f8c6-e035-4d2c-bc85-d696b80ca139)

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
    	start time
  -h int
     	image height (default 400)
  -max int
    	max frames
  -o string
    	output file (default "output.gif")
  -s float
    	font size (default 48)
  -w int
   	image width (default 600)
```

If `-f` flag is not provided, the app will use the default Face7x13 font.

If `-max` flag is not provided, the app will generate all frames until the end of the countdown.

If `-ca` flag is provided, `-cy` flag will be ignored.

Example:

```
go run . \
  -f fonts/Gorton\ Digital\ Regular.otf \
  -s 120 \
  -bg "#8af" \
  -c "white" \
  -from "2h30s" \
  -ca \
  -max 100
```
