package main

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"iter"
	"log"
	"math/cmplx"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"

	"golang.org/x/sync/errgroup"
)

type Pair[T any] [2]T

func main() {
	if len(os.Args) != 5 {
		fmt.Println("Порядок вызова: mandelbrot FILE PIXELS UPPERLEFT LOWERRIGHT")
		fmt.Printf("Пример: %s mandel.png 1000x750 -1.20,0.35 -1,0.20\n", os.Args[0])
		os.Exit(1)
	}

	var bounds Pair[int]
	if x, y, ok := parsePair[int](os.Args[2], "x"); ok {
		bounds = Pair[int]{x, y}
	} else {
		log.Fatal("ошибка при разборе параметров изображения")
	}

	var upperLeft complex128
	if z, ok := parseComplex(os.Args[3]); ok {
		upperLeft = z
	} else {
		log.Fatal("ошибка при разборе координат верхнего левого угла")
	}

	var lowerRight complex128
	if z, ok := parseComplex(os.Args[4]); ok {
		lowerRight = z
	} else {
		log.Fatal("ошибка при разборе координат правого нижнего угла")
	}

	pixels := make([]byte, bounds[0]*bounds[1])
	// single threaded rendering
	// render(pixels, bounds, upperLeft, lowerRight)

	// multi-threaded rendering
	threads := runtime.GOMAXPROCS(0)
	rowsPerBand := bounds[1]/threads + 1

	rctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	g, gctx := errgroup.WithContext(rctx)
	for i, band := range makeBands(pixels, bounds[0], rowsPerBand) {
		g.Go(func() error {
			if err := gctx.Err(); err != nil {
				return err
			}

			top := rowsPerBand * i
			height := len(band) / bounds[0]
			bandBounds := Pair[int]{bounds[0], height}
			bandUpperLeft := pixelToPoint(bounds, Pair[int]{0, top}, upperLeft, lowerRight)
			bandLowerRight := pixelToPoint(bounds, Pair[int]{bounds[0], top + height}, upperLeft, lowerRight)
			render(band, bandBounds, bandUpperLeft, bandLowerRight)
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		log.Println("ERROR:", err)
	}

	err := writeImage(os.Args[1], pixels, bounds)
	if err != nil {
		log.Println("ошибка при записи PNG-файла")
	}
}

func makeBands(buf []byte, width, step int) iter.Seq2[int, []byte] {
	return func(yield func(int, []byte) bool) {
		if len(buf) == 0 || width <= 0 || step <= 0 {
			return
		}

		height := len(buf) / width
		for top := 0; top < height; top += step {
			h := step
			if top+h > height {
				h = height - top
			}

			head := top * width
			tail := head + h*width
			band := buf[head:tail]
			if !yield(top/step, band) {
				return
			}
		}
	}
}

func escapeTime(c complex128, limit int) (int, bool) {
	z := 0 + 0i
	for i := range limit {
		z = z*z + c
		if cmplx.Abs(z) > 4 {
			return i, true
		}
	}
	return 0, false
}

type Number interface {
	int | float64
}

func parsePair[T Number](s, sep string) (T, T, bool) {
	var t T

	split := strings.SplitN(s, sep, 2)
	if len(split) != 2 {
		return t, t, false
	}

	switch any(t).(type) {
	case int:
		x, err := strconv.Atoi(split[0])
		if err != nil {
			return t, t, false
		}
		y, err := strconv.Atoi(split[1])
		if err != nil {
			return t, t, false
		}
		return T(x), T(y), true
	case float64:
		x, err := strconv.ParseFloat(split[0], 64)
		if err != nil {
			return t, t, false
		}
		y, err := strconv.ParseFloat(split[1], 64)
		if err != nil {
			return t, t, false
		}
		return T(x), T(y), true
	default:
		return t, t, false
	}
}

func parseComplex(s string) (complex128, bool) {
	re, im, ok := parsePair[float64](s, ",")
	if !ok {
		return 0, false
	}
	return complex(re, im), true
}

func pixelToPoint(bounds Pair[int], pixel Pair[int], upperLeft complex128, lowerRight complex128) complex128 {
	width, height := real(lowerRight)-real(upperLeft), imag(upperLeft)-imag(lowerRight)
	return complex(
		real(upperLeft)+float64(pixel[0])*width/float64(bounds[0]),
		imag(upperLeft)-float64(pixel[1])*height/float64(bounds[1]))
}

func render(pixels []byte, bounds Pair[int], upperLeft complex128, lowerRight complex128) {
	if len(pixels) != bounds[0]*bounds[1] {
		panic("wrong size")
	}

	for row := range bounds[1] {
		for column := range bounds[0] {
			point := pixelToPoint(bounds, Pair[int]{column, row}, upperLeft, lowerRight)
			if v, ok := escapeTime(point, 255); ok {
				pixels[row*bounds[0]+column] = 255 - byte(v)
			}
		}
	}
}

func writeImage(filename string, pixels []byte, bounds Pair[int]) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	rect := image.Rect(0, 0, bounds[0], bounds[1])
	img := image.NewGray(rect)
	for y := range bounds[1] {
		offset := y * bounds[0]
		for x := range bounds[0] {
			img.Set(x, y, color.Gray{Y: pixels[offset+x]})
		}
	}

	return png.Encode(file, img)
}
