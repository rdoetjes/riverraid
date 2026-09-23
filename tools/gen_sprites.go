package main

import (
	"image"
	"image/color"
	"image/png"
	"log"
	"math"
	"os"
)

func main() {
	_ = os.MkdirAll("assets/sprites", 0755)

	generatePlayerSheet()
	generateHelicopterSheet()
	generateShipSheet()
	generateDestroyerSheet()
	generateEnemyJetSheet()
	generateFuelSheet()
	generateSAMSheet()
	generateMissile()
	generateBridge()
	generateDecorations()
	generateDecorationBuilding()
	generateDecorationRock()
}

func savePNG(img image.Image, name string) {
	f, err := os.Create("assets/sprites/" + name + ".png")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		log.Fatal(err)
	}
	log.Printf("Generated %s.png", name)
}

func drawRect(img *image.RGBA, x, y, w, h int, col color.Color) {
	for i := x; i < x+w; i++ {
		for j := y; j < y+h; j++ {
			if i >= 0 && i < img.Bounds().Dx() && j >= 0 && j < img.Bounds().Dy() {
				img.Set(i, j, col)
			}
		}
	}
}

func drawCircle(img *image.RGBA, cx, cy, r int, col color.Color) {
	for i := cx - r; i <= cx+r; i++ {
		for j := cy - r; j <= cy+r; j++ {
			if (i-cx)*(i-cx)+(j-cy)*(j-cy) <= r*r {
				if i >= 0 && i < img.Bounds().Dx() && j >= 0 && j < img.Bounds().Dy() {
					img.Set(i, j, col)
				}
			}
		}
	}
}

func generatePlayerSheet() {
	w, h := 64, 64
	sheet := image.NewRGBA(image.Rect(0, 0, w*5, h))
	blue := color.RGBA{0, 120, 210, 255}
	for f := 0; f < 5; f++ {
		offsetX := f * w
		cx, cy := offsetX+w/2, h/2
		drawRect(sheet, cx-4, cy-20, 8, 40, blue)
		tilt := float64(f-2) * 6.0
		for i := -24; i <= 24; i++ {
			y := cy + int(float64(i)*tilt*0.1)
			drawRect(sheet, cx+i, y, 1, 6, blue)
		}
	}
	savePNG(sheet, "player")
}

func generateHelicopterSheet() {
	w, h := 48, 48
	sheet := image.NewRGBA(image.Rect(0, 0, w*4, h))
	olive := color.RGBA{70, 85, 65, 255}
	for f := 0; f < 4; f++ {
		offsetX := f * w
		cx, cy := offsetX+w/2, h/2
		drawCircle(sheet, cx, cy, 10, olive)
		drawRect(sheet, cx-2, cy, 4, 18, olive)
		angle := float64(f) * math.Pi / 2.0 / 4.0
		for i := 0; i < 4; i++ {
			a := angle + float64(i)*math.Pi/2
			for d := 0; d < 20; d++ {
				sheet.Set(cx+int(float64(d)*math.Cos(a)), cy+int(float64(d)*math.Sin(a)), color.White)
			}
		}
	}
	savePNG(sheet, "helicopter")
}

func generateShipSheet() {
	img := image.NewRGBA(image.Rect(0, 0, 64, 32))
	grey := color.RGBA{110, 115, 120, 255}
	drawRect(img, 10, 8, 44, 16, grey)
	drawRect(img, 54, 10, 6, 12, grey)
	savePNG(img, "ship")
}

func generateDestroyerSheet() {
	img := image.NewRGBA(image.Rect(0, 0, 32, 64))
	grey := color.RGBA{100, 105, 110, 255}
	drawRect(img, 8, 10, 16, 44, grey)
	drawRect(img, 10, 54, 12, 6, grey)
	savePNG(img, "destroyer")
}

func generateEnemyJetSheet() {
	img := image.NewRGBA(image.Rect(0, 0, 48, 48))
	blue := color.RGBA{0, 110, 200, 255}
	drawRect(img, 21, 9, 6, 30, blue)
	for i := -18; i <= 18; i++ {
		drawRect(img, 24+i, 24+int(math.Abs(float64(i))*0.5), 1, 5, blue)
	}
	savePNG(img, "enemy_jet")
}

func generateFuelSheet() {
	img := image.NewRGBA(image.Rect(0, 0, 32, 48))
	drawRect(img, 2, 2, 28, 44, color.RGBA{240, 190, 30, 255})
	drawRect(img, 6, 10, 20, 28, color.White)
	savePNG(img, "fuel")
}

func generateSAMSheet() {
	w, h := 48, 48
	sheet := image.NewRGBA(image.Rect(0, 0, w*8, h))
	for f := 0; f < 8; f++ {
		cx, cy := f*w+w/2, h/2
		drawCircle(sheet, cx, cy, 14, color.RGBA{70, 75, 80, 255})
		a := float64(f) * math.Pi * 2 / 8
		for d := 0; d < 12; d++ {
			sheet.Set(cx+int(float64(d)*math.Cos(a)), cy+int(float64(d)*math.Sin(a)), color.White)
		}
	}
	savePNG(sheet, "sam_site")
}

func generateMissile() {
	img := image.NewRGBA(image.Rect(0, 0, 16, 16))
	drawRect(img, 6, 2, 4, 12, color.White)
	savePNG(img, "missile")
}

func generateBridge() {
	img := image.NewRGBA(image.Rect(0, 0, 128, 32))
	drawRect(img, 0, 4, 128, 24, color.RGBA{45, 48, 52, 255})
	drawRect(img, 0, 2, 128, 3, color.White)
	drawRect(img, 0, 27, 128, 3, color.White)
	savePNG(img, "bridge")
}

func generateDecorations() {
	// Pine
	pine := image.NewRGBA(image.Rect(0, 0, 32, 32))
	drawCircle(pine, 16, 16, 14, color.RGBA{28, 65, 35, 255})
	savePNG(pine, "deco_pine")

	// Bush
	bush := image.NewRGBA(image.Rect(0, 0, 32, 32))
	drawCircle(bush, 16, 16, 12, color.RGBA{40, 95, 35, 255})
	savePNG(bush, "deco_bush")

	// House
	house := image.NewRGBA(image.Rect(0, 0, 32, 32))
	drawRect(house, 4, 6, 24, 20, color.RGBA{160, 70, 50, 255})
	savePNG(house, "deco_house")
}

func generateDecorationBuilding() {
	img := image.NewRGBA(image.Rect(0, 0, 64, 64))
	drawRect(img, 8, 8, 48, 48, color.RGBA{110, 115, 120, 255})
	savePNG(img, "deco_building")
}

func generateDecorationRock() {
	img := image.NewRGBA(image.Rect(0, 0, 32, 32))
	drawCircle(img, 16, 16, 12, color.RGBA{110, 115, 110, 255})
	savePNG(img, "deco_rock")
}
