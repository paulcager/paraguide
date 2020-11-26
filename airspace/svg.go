package airspace

import (
	"fmt"
	"github.com/kr/pretty"
	"io"
	"math"
	"strconv"
	"strings"
	"text/template"
)

const (
	minLat = 49.5
	maxLat = 59
	minLon = -6.5
	maxLon = 2
	//height  = maxLat - minLat
	//width   = 1.657 * (maxLon - minLon)
	peakLat = 53.35 // The latitude of the Peak District, used to centre the projection's distortion.
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

func init() {
	fmt.Println("One degree of lat is", degToNautMileX, "one minute is", degToNautMileX/60.0)
	fmt.Println("width", widthNautMiles, "height", heightNautMiles)
}

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
	"x": func(x float64) float64 { return (x - minLon) * degToNautMileX },
	// y converts a longitude to nautical miles from the origin
	"y":         func(y float64) float64 { return (maxLat - y) * degToNautMileY },
	"d":         distance,
	"pretty":    func(obj interface{}) string { return pretty.Sprint(obj) },
	"colourise": colourise,
}

func colourise(h string) string {
	h = strings.ToUpper(strings.TrimSpace(h))
	if h == "" || h == "SFC" {
		return `fill="red" opacity="0.25"`
	}
	return `fill="blue" opacity="0.1"`

}

func distance(d string) float64 {
	f, _ := strconv.ParseFloat(strings.TrimSuffix(d, " nm"), 64)
	return f
}

const tmplt = `
<svg viewBox="0 0 {{.width}} {{.height}}" preserveAspectRatio="none" xmlns="http://www.w3.org/2000/svg">
{{range .features -}}
	{{range .Geometry -}}
		{{ $volume := . -}}
		<!-- {{.ID}} {{.Name}} {{.Class}} {{.Lower}} -->
		{{range .Boundary -}}
			{{if ne "" .Circle.Centre.Text -}}
				{{- if eq "SFC" $volume.Lower }}
				<circle cx="{{x .Circle.Centre.X}}" cy="{{y .Circle.Centre.Y}}" r="{{d .Circle.Radius}}" {{colourise $volume.Lower}}/>
				{{- else }}
				<circle cx="{{x .Circle.Centre.X}}" cy="{{y .Circle.Centre.Y}}" r="{{d .Circle.Radius}}" fill="#0000ff" opacity="0.25"/>
				{{- end -}}
			{{- end -}}
		{{ end -}}
	{{end -}}
{{end}}
<!-- <rect x="0" y="0" height="{{.height}}" width="{{.width}}" fill="none" stroke="#000000" stroke-width="5"/> -->
</svg>
`
