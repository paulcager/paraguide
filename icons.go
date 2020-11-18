package main

import (
	"image"
	"image/color"
	"math"

	"github.com/llgcode/draw2d/draw2dimg"
	"github.com/llgcode/draw2d/draw2dkit"
)

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
