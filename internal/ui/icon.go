package ui

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"log"

	"fyne.io/fyne/v2"
)

// createAppIcon создает иконку приложения с улицей (проспектом) и буквой P
func createAppIcon() fyne.Resource {
	// Создаем изображение 64x64 пикселя для лучшего качества
	img := image.NewRGBA(image.Rect(0, 0, 64, 64))

	// Фон - небо (светло-голубой)
	skyColor := color.RGBA{R: 135, G: 206, B: 250, A: 255} // LightSkyBlue
	for y := 0; y < 64; y++ {
		for x := 0; x < 64; x++ {
			img.Set(x, y, skyColor)
		}
	}

	// Рисуем дорогу/улицу (темно-серый асфальт)
	roadColor := color.RGBA{R: 64, G: 64, B: 64, A: 255} // DarkGray
	roadY := 40                                          // Позиция дороги снизу
	roadHeight := 24
	for y := roadY; y < 64; y++ {
		for x := 0; x < 64; x++ {
			img.Set(x, y, roadColor)
		}
	}

	// Рисуем разметку дороги (желтые линии)
	lineColor := color.RGBA{R: 255, G: 215, B: 0, A: 255} // Gold
	centerY := roadY + roadHeight/2
	// Центральная прерывистая линия
	for x := 8; x < 56; x += 8 {
		for i := 0; i < 4; i++ {
			if x+i < 64 {
				img.Set(x+i, centerY, lineColor)
				img.Set(x+i, centerY+1, lineColor)
			}
		}
	}

	// Боковые линии дороги
	for x := 0; x < 64; x++ {
		img.Set(x, roadY+2, lineColor)
		img.Set(x, roadY+3, lineColor)
		img.Set(x, 62, lineColor)
		img.Set(x, 63, lineColor)
	}

	// Рисуем большую букву "P" в центре (белая, с тенью)
	// Тень буквы P (темно-серый)
	shadowColor := color.RGBA{R: 100, G: 100, B: 100, A: 255}
	// Вертикальная линия P (тень) - увеличиваем размер
	for y := 8; y < 36; y++ {
		for x := 20; x < 25; x++ {
			img.Set(x+1, y+1, shadowColor)
		}
	}
	// Верхняя горизонтальная линия P (тень) - увеличиваем
	for x := 20; x < 40; x++ {
		for y := 8; y < 13; y++ {
			img.Set(x+1, y+1, shadowColor)
		}
	}
	// Средняя горизонтальная линия P (тень) - увеличиваем
	for x := 20; x < 36; x++ {
		for y := 20; y < 25; y++ {
			img.Set(x+1, y+1, shadowColor)
		}
	}
	// Правая вертикальная линия P (тень, верхняя часть) - увеличиваем
	for y := 8; y < 25; y++ {
		for x := 36; x < 41; x++ {
			img.Set(x+1, y+1, shadowColor)
		}
	}

	// Сама буква P (белая) - увеличенная версия
	letterColor := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	// Вертикальная линия P - толще и выше
	for y := 8; y < 36; y++ {
		for x := 20; x < 25; x++ {
			img.Set(x, y, letterColor)
		}
	}
	// Верхняя горизонтальная линия P - шире
	for x := 20; x < 40; x++ {
		for y := 8; y < 13; y++ {
			img.Set(x, y, letterColor)
		}
	}
	// Средняя горизонтальная линия P - шире
	for x := 20; x < 36; x++ {
		for y := 20; y < 25; y++ {
			img.Set(x, y, letterColor)
		}
	}
	// Правая вертикальная линия P (верхняя часть) - выше
	for y := 8; y < 25; y++ {
		for x := 36; x < 41; x++ {
			img.Set(x, y, letterColor)
		}
	}

	// Конвертируем в ресурс Fyne
	return fyne.NewStaticResource("icon.png", encodePNG(img))
}

// encodePNG кодирует изображение в PNG формат
func encodePNG(img *image.RGBA) []byte {
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		log.Printf("Error encoding PNG: %v", err)
		return nil
	}
	return buf.Bytes()
}
