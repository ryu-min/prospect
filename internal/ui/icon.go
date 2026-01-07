package ui

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"log"

	"fyne.io/fyne/v2"
)

func createAppIcon() fyne.Resource {
	img := image.NewRGBA(image.Rect(0, 0, 64, 64))

	skyColor := color.RGBA{R: 135, G: 206, B: 250, A: 255}
	for y := 0; y < 64; y++ {
		for x := 0; x < 64; x++ {
			img.Set(x, y, skyColor)
		}
	}

	roadColor := color.RGBA{R: 64, G: 64, B: 64, A: 255}
	roadY := 40
	roadHeight := 24
	for y := roadY; y < 64; y++ {
		for x := 0; x < 64; x++ {
			img.Set(x, y, roadColor)
		}
	}

	lineColor := color.RGBA{R: 255, G: 215, B: 0, A: 255}
	centerY := roadY + roadHeight/2
	for x := 8; x < 56; x += 8 {
		for i := 0; i < 4; i++ {
			if x+i < 64 {
				img.Set(x+i, centerY, lineColor)
				img.Set(x+i, centerY+1, lineColor)
			}
		}
	}

	for x := 0; x < 64; x++ {
		img.Set(x, roadY+2, lineColor)
		img.Set(x, roadY+3, lineColor)
		img.Set(x, 62, lineColor)
		img.Set(x, 63, lineColor)
	}

	shadowColor := color.RGBA{R: 100, G: 100, B: 100, A: 255}
	for y := 8; y < 36; y++ {
		for x := 20; x < 25; x++ {
			img.Set(x+1, y+1, shadowColor)
		}
	}
	for x := 20; x < 40; x++ {
		for y := 8; y < 13; y++ {
			img.Set(x+1, y+1, shadowColor)
		}
	}
	for x := 20; x < 36; x++ {
		for y := 20; y < 25; y++ {
			img.Set(x+1, y+1, shadowColor)
		}
	}
	for y := 8; y < 25; y++ {
		for x := 36; x < 41; x++ {
			img.Set(x+1, y+1, shadowColor)
		}
	}

	letterColor := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	for y := 8; y < 36; y++ {
		for x := 20; x < 25; x++ {
			img.Set(x, y, letterColor)
		}
	}
	for x := 20; x < 40; x++ {
		for y := 8; y < 13; y++ {
			img.Set(x, y, letterColor)
		}
	}
	for x := 20; x < 36; x++ {
		for y := 20; y < 25; y++ {
			img.Set(x, y, letterColor)
		}
	}
	for y := 8; y < 25; y++ {
		for x := 36; x < 41; x++ {
			img.Set(x, y, letterColor)
		}
	}

	return fyne.NewStaticResource("icon.png", encodePNG(img))
}

func encodePNG(img *image.RGBA) []byte {
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		log.Printf("Error encoding PNG: %v", err)
		return nil
	}
	return buf.Bytes()
}
