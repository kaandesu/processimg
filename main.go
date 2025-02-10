package main

import (
	"image"
	"image/color"
	"image/png"
	"log/slog"
	"os"
)

type ImageKernel struct {
	inputImage  image.Image
	outputImage image.Image
	path        string
	width       int
	height      int
	kernel      [3][3]float32
}

func NewImageKernel(path string) *ImageKernel {
	ik := new(ImageKernel)
	ik.path = path

	ik.UpdateInputImage(path)

	return ik
}

func (ik *ImageKernel) UpdateInputImage(path string) {
	file, err := os.Open(path)
	if err != nil {
		slog.Error("could not open the file", "err", err)
	}
	defer file.Close()
	ik.inputImage, err = png.Decode(file)
	if err != nil {
		slog.Error("could not decode the file", "err", err)
	}
	bounds := ik.inputImage.Bounds()
	ik.width, ik.height = bounds.Dx(), bounds.Dy()
}

func (ik *ImageKernel) Convolve(kernel [3][3]float32) {
	ik.kernel = kernel

	img := ik.inputImage
	temp := [3][3]color.Color{}

	testImage := image.NewRGBA(image.Rect(0, 0, ik.width, ik.height))
	black := color.RGBA{0, 0, 0, 255}
	for y := 0; y < ik.height; y++ {
		for x := 0; x < ik.width; x++ {
			testImage.Set(x, y, black)
		}
	}

	for y := 1; y < ik.width-1; y++ {
		for x := 1; x < ik.height-1; x++ {
			temp = [3][3]color.Color{
				{img.At(y-1, x-1), img.At(y-1, x), img.At(y-1, x+1)},
				{img.At(y, x-1), img.At(y, x), img.At(y, x+1)},
				{img.At(y+1, x-1), img.At(y+1, x), img.At(y+1, x+1)},
			}
			newColor := ik.MultiplyMatrices(temp)
			testImage.Set(y, x, newColor)

		}
	}

	_ = temp

	f, err := os.Create("output.png")
	if err != nil {
		slog.Error("Could not create output.png", "err", err)
	}
	defer f.Close()
	if err := png.Encode(f, testImage); err != nil {
		slog.Error("Could not encode the input image", "err", err)
	}
}

func (ik *ImageKernel) MultiplyMatrices(imgMatrix [3][3]color.Color) color.Color {
	var r2, g2, b2, a2 float32
	for i, row := range imgMatrix {
		r2, g2, b2, a2 = 0, 0, 0, 0
		for j, val := range row {
			r, g, b, a := val.RGBA()
			r8, g8, b8 := float32(r>>8), float32(g>>8), float32(b>>8)
			weight := ik.kernel[i][j]
			r2 += r8 * weight
			g2 += g8 * weight
			b2 += b8 * weight
			a2 = float32(a)
		}
	}

	clamp := func(val float32) uint8 {
		if val < 0 {
			return 0
		} else if val > 255 {
			return 255
		}
		return uint8(val)
	}

	return color.RGBA{
		R: clamp(r2),
		G: clamp(g2),
		B: clamp(b2),
		A: clamp(a2),
	}
}

func main() {
	kernel := [3][3]float32{{-1, -2, -1}, {0, 0, 0}, {1, 2, 1}}
	kernel2 := [3][3]float32{{-1, 0, 1}, {-2, 0, 2}, {-1, 0, 1}}
	blur := [3][3]float32{{1, 1, 1}, {1, 1, 1}, {1, 1, 1}}
	gauss := [3][3]float32{{1 / 16.0, 2 / 16.0, 1 / 16.0}, {2 / 16.0, 4 / 16.0, 2 / 16.0}, {1 / 16.0, 2 / 16.0, 1 / 16.0}}
	_ = kernel
	_ = kernel2
	_ = blur
	_ = gauss
	imageKernel := NewImageKernel("./images/car.png")
	imageKernel.Convolve(kernel)
	imageKernel.UpdateInputImage("./output.png")
	imageKernel.Convolve(kernel2)
}
