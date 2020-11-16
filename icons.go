package main

import (
	"image"
	"image/color"
	"math"
	"strconv"

	"github.com/llgcode/draw2d"
	"github.com/llgcode/draw2d/draw2dimg"
	"github.com/llgcode/draw2d/draw2dkit"
)

func init() {
	draw2d.SetFontFolder("fonts")
}

func windIcon(size int, windRange []WindRange) image.Image {
	good := color.RGBA{R: 0x00, G: 0xff, B: 0x00, A: 0xff}
	bad := color.RGBA{0xff, 0x00, 0x00, 0xff}

	radius := float64(size) / 2
	dest := image.NewRGBA(image.Rect(0, 0, size, size))
	gc := draw2dimg.NewGraphicContext(dest)
	gc.SetStrokeColor(color.Black)
	gc.SetLineWidth(0)

	draw2dkit.Circle(gc, radius, radius, radius)
	gc.SetFillColor(bad)
	gc.FillStroke()

	for _, w := range windRange {
		// We measure angles from North; the X axis is 90 degrees clockwise from that.
		// To convert degrees to radians, multiply by π/180
		start := math.Pi / 180 * (w.From - 90)
		angle := math.Pi / 180 * (w.To - w.From)
		if angle < 0 {
			angle += (2 * math.Pi)
		}

		gc.MoveTo(radius, radius)
		gc.ArcTo(radius, radius, radius, radius, start, angle)
		gc.LineTo(radius, radius)
	}
	gc.SetFillColor(good)
	gc.FillStroke()

	return dest
}

func windIndicator(speed float64, direction float64) image.Image {
	const size = 80
	const circleRadius = 16

	red := color.RGBA{0xff, 0x00, 0x00, 0xff}
	blue := color.RGBA{0x00, 0x00, 0xff, 0xff}
	gray := color.RGBA{0x00, 0x00, 0x00, 0x80}
	_, _ = red, blue

	dest := image.NewRGBA(image.Rect(0, 0, size, size))
	gc := draw2dimg.NewGraphicContext(dest)

	// Move origin to centre.
	gc.Translate(size/2, size/2)

	gc.SetStrokeColor(gray)
	gc.SetLineWidth(2)
	gc.SetFillColor(color.Transparent)
	draw2dkit.Circle(gc, 0, 0, circleRadius)
	gc.FillStroke()

	s := strconv.FormatInt(int64(speed), 10)
	gc.SetFillColor(color.Black)
	gc.SetFontSize(12)
	l, t, r, b := gc.GetStringBounds(s)
	gc.FillStringAt(s, -(r-l)/2, -(t-b)/2)

	if speed > 0 {
		// Direction is where the wind is blowing from; we want where it is blowing to.
		angle := direction + 180
		rotation := -90 + angle		// Subtract 9 to move X axis from E to N.
		gc.Rotate(rotation / 180.0 * math.Pi)

		gc.SetStrokeColor(red)
		gc.SetLineWidth(4)

		gc.MoveTo( circleRadius, 0)
		gc.LineTo(size/2, 0)
		gc.FillStroke()
	}

	return dest
}
