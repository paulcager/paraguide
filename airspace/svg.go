package airspace

import (
	"bytes"
	"fmt"
	"github.com/kr/pretty"
	"io"
	"log"
	"math"
	"strconv"
	"strings"
	"text/template"
)

const (
	minLat               = 49.5
	maxLat               = 59
	minLon               = -6.5
	maxLon               = 2
	peakLat              = 53.35  // The latitude of the Peak District, used to centre the projection's distortion.
	maxInterestingHeight = 10_000 // We are not interested in restrictions > 10,000 feet
)

var (
	// One minute of latitude is always one nautical mile.
	degToNautMileY = 60.0
	// One minute of longitude depends on your latitude. This will be different at the N and S of the map,
	// so use the Peak District as the logical centre of the map.
	//
	// This does not affect where Google Maps will place the objects we draw, but *does* affect how circles
	// etc appear once Maps has stretched our image to cover the map's canvas.
	degToNautMileX = 60.0 * math.Cos(math.Pi/180.0*peakLat)

	heightNautMiles = (maxLat - minLat) * degToNautMileY
	widthNautMiles  = (maxLon - minLon) * degToNautMileX
)

func ToSVG(a Airspace, w io.Writer) error {
	// The image has origin (0,0) in NW corner. All dimensions are in nautical miles.

	params := map[string]interface{}{
		"minLat":   minLat,
		"maxLat":   maxLat,
		"minLon":   minLon,
		"maxLon":   maxLon,
		"height":   heightNautMiles,
		"width":    widthNautMiles,
		"features": a.Airspace,
	}

	t := template.Must(template.New("airspace").Funcs(funcMap).Parse(tmplt))
	return t.Execute(w, params)
}

var funcMap = template.FuncMap{
	// x converts a longitude to nautical miles from the origin
	"x": xPos,
	// y converts a longitude to nautical miles from the origin
	"y":             yPos,
	"d":             distance,
	"pretty":        func(obj interface{}) string { return pretty.Sprint(obj) },
	"height":        height,
	"colourise":     colourise,
	"isInteresting": func(h string) bool { return height(h) <= maxInterestingHeight },
	"polygon":       polygon,
}

func xPos(x float64) float64 { return (x - minLon) * degToNautMileX }
func yPos(y float64) float64 { return (maxLat - y) * degToNautMileY }

func height(h string) float64 {
	h = strings.ToUpper(strings.TrimSpace(h))
	if h == "" || h == "SFC" {
		return 0
	}

	if strings.HasPrefix(h, "FL") {
		// Flight level.
		f, err := strconv.ParseFloat(h[2:], 64)
		if err != nil {
			log.Printf("Could not parse flight levele %#q: %s\n", h, err)
		}
		return f * 100 // Standard pressure and so on.
	}

	h = strings.TrimSpace(strings.TrimSuffix(h, "FT"))
	f, err := strconv.ParseFloat(h, 64)
	if err != nil {
		log.Printf("Could not parse height %#q: %s\n", h, err)
	}
	return f
}

func chooseColour(h float64) (string, float64) {
	switch {
	case h == 0:
		return "red", 0.25
	case h < 1000:
		return "red", 0.1
	case h < 3000:
		return "green", 0.1
	case h < 5000:
		return "blue", 0.1
	default:
		return "#000000", 0.1
	}
}

func colourise(hStr string) string {
	colour, opacity := chooseColour(height(hStr))
	return fmt.Sprintf(`fill="%s" opacity="%f"`, colour, opacity)
}

func polygon(h string, bounds []Boundary) string {
	colour, opacity := chooseColour(height(h))
	b := new(bytes.Buffer)
	b.WriteString(`<path d="`)
	segs := 0
	for _, bound := range bounds {
		if len(bound.Line) > 0 {
			for _, p := range bound.Line {
				if segs == 0 {
					fmt.Fprintf(b, "M %f,%f ", xPos(p.X()), yPos(p.Y()))
				} else {
					fmt.Fprintf(b, "L %f,%f ", xPos(p.X()), yPos(p.Y()))
				}
				segs++
			}
		}
		arc := bound.Arc
		if arc.Radius != "" {
			radius := distance(arc.Radius)
			fmt.Sprintf(`A %f %f 0 0 0 0 %f %f`, radius, radius, xPos(arc.To.X()), yPos(arc.To.Y()))
		}
	}
	fmt.Fprintf(b, `Z" fill="none" stroke="%s" stroke-opacity="%f" stroke-width="0.25" />`, colour, opacity)
	return b.String()
}

func distance(d string) float64 {
	f, _ := strconv.ParseFloat(strings.TrimSuffix(d, " nm"), 64)
	return f
}

const tmplt = `
{{ $ := . }}
<svg viewBox="0 0 {{.width}} {{.height}}" preserveAspectRatio="none" xmlns="http://www.w3.org/2000/svg">
{{range .features -}}
	{{range .Geometry -}}
		{{ $volume := . -}}
		{{- if isInteresting .Lower -}}
			<!-- {{.ID}} {{.Name}} {{.Class}} {{.Lower}} -->
			{{range .Boundary -}}
				{{if ne "" .Circle.Centre.Text -}}
					<circle cx="{{x .Circle.Centre.X}}" cy="{{y .Circle.Centre.Y}}" r="{{d .Circle.Radius}}" {{colourise $volume.Lower}}/>
				{{- end -}}
			{{- end }}
			{{- polygon .Lower .Boundary -}}
		{{- end -}}
	{{end -}}
{{end}}
<!-- <rect x="0" y="0" height="{{.height}}" width="{{.width}}" fill="none" stroke="#000000" stroke-width="5"/> -->
</svg>
`
