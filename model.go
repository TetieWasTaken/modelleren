package main

import (
	"fmt"
	"image/color"
	"math"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
)

const (
	G          = 6.674e-11
	M          = 5.972e24
	dt         = 10
	iterations = 10000
)

func main() {
	t := 0.0
	x := 7e6
	y := 0.0
	vx := 0.0
	vy := 9.0e3

	orbitData := make(plotter.XYs, 0)
	xData := make(plotter.XYs, 0)
	yData := make(plotter.XYs, 0)

	for i := 0; i < iterations; i++ {
		if i%2 == 0 {
			fmt.Printf("Writing [%d]/[%d]\n", i, iterations)
			orbitData = append(orbitData, plotter.XY{X: x, Y: y})
			xData = append(xData, plotter.XY{X: t, Y: x})
			yData = append(yData, plotter.XY{X: t, Y: y})
		}

		r := math.Sqrt(x*x + y*y)
		ax := -G * M * x / (r * r * r)
		ay := -G * M * y / (r * r * r)

		vx += ax * dt
		vy += ay * dt
		x += vx * dt
		y += vy * dt
		t += dt
	}

	pBaan := plot.New()
	pBaan.Title.Text = "Baan"
	pBaan.X.Label.Text = "X (meters)"
	pBaan.Y.Label.Text = "Y (meters)"

	lijnBaan, _ := plotter.NewLine(orbitData)
	lijnBaan.Color = color.RGBA{R: 255, A: 255} // Rood
	pBaan.Add(lijnBaan)

	aardeData := plotter.XYs{{X: 0, Y: 0}}
	aarde, _ := plotter.NewScatter(aardeData)
	aarde.GlyphStyle.Color = color.RGBA{B: 255, A: 255}
	aarde.GlyphStyle.Shape = draw.CircleGlyph{}
	pBaan.Add(aarde)

	pBaan.Save(6*vg.Inch, 6*vg.Inch, "baan.png")

	pX := plot.New()
	pX.Title.Text = "X-positie over tijd"
	pX.X.Label.Text = "Tijd (s)"
	pX.Y.Label.Text = "X positie (m)"

	lijnX, _ := plotter.NewLine(xData)
	lijnX.Color = color.RGBA{R: 0, G: 150, B: 0, A: 255}
	pX.Add(lijnX)

	pX.Save(6*vg.Inch, 4*vg.Inch, "x_tijd.png")

	pY := plot.New()
	pY.Title.Text = "Y-positie over tijd"
	pY.X.Label.Text = "Tijd (s)"
	pY.Y.Label.Text = "Y positie (m)"

	lijnY, _ := plotter.NewLine(yData)
	lijnY.Color = color.RGBA{R: 0, G: 0, B: 255, A: 255}
	pY.Add(lijnY)

	pY.Save(6*vg.Inch, 4*vg.Inch, "y_tijd.png")
}
