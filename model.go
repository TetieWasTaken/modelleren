package main

import (
	"fmt"
	"image"
	"image/color"
	imagedraw "image/draw"
	"image/png"
	"log"
	"math"
	"os"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/font"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	plotdraw "gonum.org/v1/plot/vg/draw"
)

const (
	G          = 6.674e-11
	M          = 5.972e24
	R          = 6.371e6
	m          = 100
	r          = 4
	Cw         = 0.47
	dt         = 1
	iterations = 2000000
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	log.Println("[INFO] Start simulatie")

	t := 0.0
	x := 8e7
	y := 0.0
	vx := 0.0
	vy := 8.63e2

	orbitData := make(plotter.XYs, 0, iterations)
	xData := make(plotter.XYs, 0, iterations)
	yData := make(plotter.XYs, 0, iterations)

out:

	for i := 0; i < iterations; i++ {
		orbitData = append(orbitData, plotter.XY{X: x, Y: y})
		xData = append(xData, plotter.XY{X: t, Y: x})
		yData = append(yData, plotter.XY{X: t, Y: y})

		d := math.Sqrt(x*x + y*y)
		if d < R {
			break out
		}

		agx := -G * M * x / (d * d * d)
		agy := -G * M * y / (d * d * d)

		h := d - R
		if h <= 100000 {
			log.Printf("Entered atmosphere")
		}
		rho := 0.0
		if h < 0 {
			rho = 1.225
		} else if h < 20000 {
			rho = 1.225 * math.Exp(-1.0e-4*h)
		} else if h <= 60000 {
			rho = 0.08803 * math.Exp(-1.4e-4*(h-20000))
		} else if h <= 100000 {
			rho = 3.1e-4 * math.Exp(-1.9e-4*(h-60000))
		} else {
			rho = 0.0
		}

		A := math.Pi * r * r
		v := math.Hypot(vx, vy)

		awx := -(0.5 * rho * Cw * A / m) * v * vx
		awy := -(0.5 * rho * Cw * A / m) * v * vy

		ax := agx + awx
		ay := agy + awy

		vx += ax * dt
		vy += ay * dt
		x += vx * dt
		y += vy * dt
		t += dt

		if i%1000 == 0 {
			log.Printf("[DEBUG] i=%d t=%.0f x=%.3e y=%.3e h=%.0f", i, t, x, y, h)
		}
	}

	// baan
	pBaan := plot.New()
	pBaan.Title.Text = "Baan van ruimtepuin (x tegen y)"
	pBaan.X.Label.Text = "x-positie (m)"
	pBaan.Y.Label.Text = "y-positie (m)"
	pBaan.Add(plotter.NewGrid())
	applyHugeStyle(pBaan)

	err := addFadedOrbit(pBaan, orbitData, color.NRGBA{R: 220, G: 20, B: 60, A: 255})
	must(err)

	earthCircle := makeEarthCircle(R, 720)
	earthLine, err := plotter.NewLine(earthCircle)
	must(err)
	earthLine.Color = color.RGBA{R: 30, G: 144, B: 255, A: 255}
	earthLine.Width = vg.Points(2.5)
	pBaan.Add(earthLine)
	pBaan.Legend.Add("Aarde (R = 6.371e6 m)", earthLine)

	center, err := plotter.NewScatter(plotter.XYs{{X: 0, Y: 0}})
	must(err)
	center.Color = color.RGBA{R: 0, G: 0, B: 120, A: 255}
	center.Shape = plotdraw.CircleGlyph{}
	center.Radius = vg.Points(4)
	pBaan.Add(center)

	lim := symmetricLimitAroundZero(orbitData, 1.05)
	pBaan.X.Min, pBaan.X.Max = -lim, lim
	pBaan.Y.Min, pBaan.Y.Max = -lim, lim

	if err := pBaan.Save(14*vg.Inch, 14*vg.Inch, "baan.png"); err != nil {
		log.Fatalf("[ERROR] baan.png opslaan mislukt: %v", err)
	}
	log.Println("[INFO] baan.png geschreven")

	// x(t)
	pX := plot.New()
	pX.Title.Text = "x-positie als functie van de tijd"
	pX.X.Label.Text = "Tijd (s)"
	pX.Y.Label.Text = "x-positie (m)"
	pX.Add(plotter.NewGrid())
	applyHugeStyle(pX)

	lijnX, err := plotter.NewLine(xData)
	must(err)
	lijnX.Color = color.RGBA{R: 34, G: 139, B: 34, A: 255}
	lijnX.Width = vg.Points(3.0)
	pX.Add(lijnX)
	pX.Legend.Add("x(t)", lijnX)

	if err := pX.Save(14*vg.Inch, 7*vg.Inch, "x_tijd.png"); err != nil {
		log.Fatalf("[ERROR] x_tijd.png opslaan mislukt: %v", err)
	}
	log.Println("[INFO] x_tijd.png geschreven")

	// y(t)
	pY := plot.New()
	pY.Title.Text = "y-positie als functie van de tijd"
	pY.X.Label.Text = "Tijd (s)"
	pY.Y.Label.Text = "y-positie (m)"
	pY.Add(plotter.NewGrid())
	applyHugeStyle(pY)

	lijnY, err := plotter.NewLine(yData)
	must(err)
	lijnY.Color = color.RGBA{R: 65, G: 105, B: 225, A: 255}
	lijnY.Width = vg.Points(3.0)
	pY.Add(lijnY)
	pY.Legend.Add("y(t)", lijnY)

	if err := pY.Save(14*vg.Inch, 7*vg.Inch, "y_tijd.png"); err != nil {
		log.Fatalf("[ERROR] y_tijd.png opslaan mislukt: %v", err)
	}
	log.Println("[INFO] y_tijd.png geschreven")

	// combineren
	if err := stitchPNGsVertical("gecombineerd.png", "baan.png", "x_tijd.png", "y_tijd.png"); err != nil {
		log.Fatalf("[ERROR] Samenvoegen mislukt: %v", err)
	}
	log.Println("[INFO] gecombineerd.png geschreven")
	log.Println("[INFO] Klaar")
}

func applyHugeStyle(p *plot.Plot) {
	p.Title.TextStyle.Font.Size = font.Length(40)
	p.X.Label.TextStyle.Font.Size = font.Length(32)
	p.Y.Label.TextStyle.Font.Size = font.Length(32)

	p.X.Tick.Label.Font.Size = font.Length(24)
	p.Y.Tick.Label.Font.Size = font.Length(24)

	p.Legend.TextStyle.Font.Size = font.Length(22)
	p.Legend.Top = true
	p.Legend.Left = true
	p.Legend.Padding = vg.Points(6)
	p.Legend.ThumbnailWidth = vg.Points(28)

	p.X.Padding = vg.Points(10)
	p.Y.Padding = vg.Points(10)

	p.X.Tick.Length = vg.Points(7)
	p.Y.Tick.Length = vg.Points(7)
	p.X.LineStyle.Width = vg.Points(1.2)
	p.Y.LineStyle.Width = vg.Points(1.2)
}

func stitchPNGsVertical(out string, files ...string) error {
	if len(files) == 0 {
		return fmt.Errorf("geen bestanden om samen te voegen")
	}
	log.Printf("[INFO] Samenvoegen van %d afbeeldingen naar %s", len(files), out)

	images := make([]image.Image, 0, len(files))
	maxW := 0
	totalH := 0

	for _, f := range files {
		img, err := readPNG(f)
		if err != nil {
			return fmt.Errorf("lezen %s: %w", f, err)
		}
		b := img.Bounds()
		w, h := b.Dx(), b.Dy()
		log.Printf("[DEBUG] %s: %dx%d", f, w, h)
		if w > maxW {
			maxW = w
		}
		totalH += h
		images = append(images, img)
	}

	canvas := image.NewRGBA(image.Rect(0, 0, maxW, totalH))
	imagedraw.Draw(canvas, canvas.Bounds(), &image.Uniform{C: color.White}, image.Point{}, imagedraw.Src)

	yOff := 0
	for i, img := range images {
		b := img.Bounds()
		w, h := b.Dx(), b.Dy()
		xOff := (maxW - w) / 2
		dst := image.Rect(xOff, yOff, xOff+w, yOff+h)
		imagedraw.Draw(canvas, dst, img, b.Min, imagedraw.Over)
		log.Printf("[DEBUG] afbeelding %d geplaatst op x=%d y=%d", i, xOff, yOff)
		yOff += h
	}

	outFile, err := os.Create(out)
	if err != nil {
		return err
	}
	defer outFile.Close()

	if err := png.Encode(outFile, canvas); err != nil {
		return err
	}
	log.Printf("[INFO] Eindafmeting gecombineerd: %dx%d", maxW, totalH)
	return nil
}

func readPNG(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	img, err := png.Decode(f)
	if err != nil {
		return nil, err
	}
	return img, nil
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func symmetricLimitAroundZero(data plotter.XYs, marginFactor float64) float64 {
	maxAbs := 0.0
	for _, p := range data {
		ax := math.Abs(p.X)
		ay := math.Abs(p.Y)
		if ax > maxAbs {
			maxAbs = ax
		}
		if ay > maxAbs {
			maxAbs = ay
		}
	}
	if maxAbs == 0 {
		maxAbs = 1
	}
	return maxAbs * marginFactor
}

func makeEarthCircle(radius float64, n int) plotter.XYs {
	if n < 3 {
		n = 3
	}
	pts := make(plotter.XYs, n+1)
	for i := 0; i <= n; i++ {
		theta := 2 * math.Pi * float64(i) / float64(n)
		pts[i].X = radius * math.Cos(theta)
		pts[i].Y = radius * math.Sin(theta)
	}
	return pts
}

func addFadedOrbit(p *plot.Plot, data plotter.XYs, base color.NRGBA) error {
	const segments = 200
	if len(data) < 2 {
		return nil
	}

	step := len(data) / segments
	if step < 2 {
		step = 2
	}

	for i := 0; i < segments; i++ {
		start := i * step
		end := start + step + 1
		if start >= len(data)-1 {
			break
		}
		if end > len(data) {
			end = len(data)
		}

		alpha := uint8(30 + (225*i)/(segments-1))
		line, err := plotter.NewLine(data[start:end])
		if err != nil {
			return err
		}
		line.Color = color.NRGBA{R: base.R, G: base.G, B: base.B, A: alpha}
		line.Width = vg.Points(3.0)
		p.Add(line)
	}
	return nil
}
