package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	_ "image/png"
	"log"
	"os"
	"runtime"
)

var black = color.Gray{0}
var white = color.Gray{255}

func alive(n color.Color) bool {
	r, _, _, _ := n.RGBA()
	return r == 0
}

func countAlive(ns []color.Color) int {
	m := 0

	for _, n := range ns {
		if alive(n) {
			m++
		}
	}

	return m
}

func calc(c color.Color, ns []color.Color) color.Color {
	alive_n := countAlive(ns)

	if alive_n == 3 || (alive(c) && alive_n == 2) {
		return black
	}

	return white
}

func neighbours(p image.Point) []image.Point {
	pts := make([]image.Point, 0, 8)

	pts = append(pts, image.Point{p.X - 1, p.Y - 1})
	pts = append(pts, image.Point{p.X + 1, p.Y - 1})
	pts = append(pts, image.Point{p.X - 1, p.Y + 1})
	pts = append(pts, image.Point{p.X + 1, p.Y + 1})

	pts = append(pts, image.Point{p.X, p.Y + 1})
	pts = append(pts, image.Point{p.X, p.Y - 1})
	pts = append(pts, image.Point{p.X - 1, p.Y})
	pts = append(pts, image.Point{p.X + 1, p.Y})

	return pts
}

func evolve(i, j int, at func(int, int) color.Color, set func(int, int, color.Color), done chan bool) {
	ns := make([]color.Color, 0)
	for _, n := range neighbours(image.Point{i, j}) {
		ns = append(ns, at(n.X, n.Y))
	}

	set(i, j, calc(at(i, j), ns))
	done <- true
}

func incarcerate(i image.Image, bounds image.Rectangle) func(x, y int) color.Color {
	return func(x, y int) color.Color {
		if image.Pt(x, y).In(bounds) {
			return i.At(x, y)
		} else {
			return white
		}
	}
}

func tick(in image.Image) image.Image {
	var (
		out   = image.NewGray(in.Bounds())
		bound = out.Bounds()
		count = 0
		joins = 0
		good  = make(chan bool, 4096)
	)

	for i := bound.Min.X; i < bound.Max.X; i++ {
		for j := bound.Min.Y; j < bound.Max.Y; j++ {
			count++
			go evolve(i, j, incarcerate(in, bound), out.Set, good)
		}
	}

	for ; joins < count; joins++ {
		<-good
	}

	return out
}

func save(i image.Image, nth int) {
	name := "evolution.t" + fmt.Sprintf("%03d", nth) + ".png"

	file, err := os.Create(name)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	fmt.Printf("Creating: %s\n", name)
	if err := png.Encode(file, i); err != nil {
		log.Fatal(err)
	}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	target, _, err := image.Decode(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}

	for i := 0; i < 20; i++ {
		save(target, i)
		target = tick(target)
	}
}
