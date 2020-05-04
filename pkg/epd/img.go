package epd

import (
	"image"
	"image/color"
)

func getImageByte(j, i int, img image.RGBA, vertical bool) byte {
	var b byte
	var pixelValue int

	for x := 0; x < 8; x++ {
		xx := i*8 + x

		if vertical {
			pixelValue = getPixelValue(xx, j, img)
		} else {
			pixelValue = getPixelValue(j, xx, img)
		}

		if pixelValue > 0 {
			b = b | (1 << uint(7-x))
		}
	}

	return b
}

func getPixelValue(x, y int, img image.RGBA) int {
	c := img.At(x, y)
	if convertColorToBlackWhite(c) {
		return 1
	}
	return 0
}

func convertColorToBlackWhite(c color.Color) bool {
	r, g, b, _ := c.RGBA()
	grey := (r*299 + g*587 + b*114 + 500) / 1000
	return grey >= 0x7fff
}
