package main

import (
	"github.com/llgcode/draw2d/draw2dimg"
	"github.com/llgcode/draw2d/draw2dkit"
	"image"
	"image/color"
	"image/png"
	"io"
	"math"
)

func windIcon(w io.Writer, size int, windRange []WindRange) error {
	radius := float64(size) / 2
	dest := image.NewRGBA(image.Rect(0, 0, size, size))
	gc := draw2dimg.NewGraphicContext(dest)
	gc.SetStrokeColor(color.Black)
	gc.SetLineWidth(0)

	gc.SetFillColor(color.RGBA{0xff, 0x00, 0x00, 0xff})
	draw2dkit.Circle(gc, radius, radius, radius)
	gc.FillStroke()

	gc.SetFillColor(color.RGBA{0x00, 0xff, 0x00, 0xff})
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
		gc.FillStroke()
	}

	return png.Encode(w, dest)
}
