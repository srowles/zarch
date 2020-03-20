package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"math"
	"math/rand"
	"os"

	"github.com/ojrac/opensimplex-go"
)

func main() {
	noise := opensimplex.NewNormalized(rand.Int63())

	w, h := 1000, 600
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	octave1 := 12.5
	octave2 := 5.6
	octave3 := 2.5
	minval1 := math.MaxFloat64
	maxval1 := 0.0
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			xFloat := float64(x) / float64(w)
			yFloat := float64(y) / float64(h)
			val1 := noise.Eval2(octave1*xFloat, octave1*yFloat)
			val2 := noise.Eval2(octave2*xFloat, octave2*yFloat)
			val3 := noise.Eval2(octave3*xFloat, octave3*yFloat)
			v := val1 + (val2 * 0.5) + (val3 * 0.25)
			v = v / 1.75
			if v < minval1 {
				minval1 = v
			}
			if v > maxval1 {
				maxval1 = v
			}
			img.Set(x, y, getColor(v))
		}
	}
	fmt.Println("min", minval1, "max", maxval1)

	f, err := os.Create("image.png")
	if err != nil {
		log.Fatal("Failed to create image:", err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		log.Fatal("Failed to encode image:", err)
	}
}

func getColor(height float64) color.RGBA {
	switch {
	case height < 0.25:
		return blue(height)
	case height >= 0.25 && height <= 0.3:
		return sand(height)
	case height > 0.75:
		return white(height)
	}
	val := uint8(height * 255)

	return color.RGBA{
		A: 255,
		R: val,
		G: val,
		B: val,
	}
}

func sand(height float64) color.RGBA {
	height = height - 0.25
	height = height * 20
	height = 1 - height
	return color.RGBA{
		A: 255,
		R: uint8(242 * height),
		G: uint8(209 * height),
		B: uint8(107 * height),
	}
}
func white(height float64) color.RGBA {
	v := uint8(height * 255)
	return color.RGBA{
		A: 255,
		R: v,
		G: v,
		B: v,
	}
}
func blue(height float64) color.RGBA {
	height = height * 4
	fmt.Println(height)
	return color.RGBA{
		A: 255,
		R: 0,
		G: 0,
		B: uint8(height * 255),
	}
}
