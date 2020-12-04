package airspace

import (
	"fmt"
	_ "github.com/golang/geo/r2"
	"github.com/kellydunn/golang-geo"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

// Download airspace defs in yaml from https://gitlab.com/ahsparrow/airspace
// Schema is https://gitlab.com/ahsparrow/yaixm/-/blob/master/yaixm/data/schema.yaml

var (
	prohibitedAirspaceClasses = map[string]bool{
		"A": true,  // Most airways; London/Manchester TMAs.
		"C": true,  // Mostly above FL195 and some airways.
		"D": true,  // Most aerodrome CTRs and CTAs. Some TMAs and lower levels of selected airways.
		"E": false, // Scottish airways. ATC clearance not required for VFR flight, pilots encouraged to contact ATC.
		"G": false, // ‘Open FIR’, ATC clearance not required, radio not required.
	}

	// Not all are strictly prohibited, some are "avoid unless ...."
	prohibitedTypes = map[string]bool{
		"MATZ": true, // Military ATZ
		"ATZ":  true, // Air Traffic Zone
		"RAT":  true, // Temporary restricted area
		"RMZ":  true, // Radio mandatory zone
		"TMZ":  true, // Transponder mandatory zone
		"P":    true, // Prohibited area
		"R":    true, // Restricted area
		"TMA":  true, // Terminal control area
	}
)

func ClearanceRequired(f Feature) bool {
	return prohibitedAirspaceClasses[f.Class] || prohibitedTypes[f.Type]
}

//func init() {
//	a, err := Load(`https://gitlab.com/ahsparrow/airspace/-/raw/master/airspace.yaml`)
//	if err != nil {
//		panic(err)
//	}
//	pretty.Println(a)
//	os.Exit(2)
//}

// This type is used to decode YAML data from https://gitlab.com/ahsparrow/airspace/-/raw/master/airspace.yaml (and eqivalent).
type airspaceResponse struct {
	Airspace []struct {
		ID          string
		Name        string
		Type        string
		LocalType   string
		ControlType string
		Class       string
		Geometry    []struct {
			ID       string
			Name     string
			Class    string
			Seqno    int
			Boundary []struct {
				// One of:
				Circle struct {
					Radius string
					Centre string
				}
				Line []string
				Arc  struct {
					Dir    string
					Radius string
					Centre string
					To     string
				}
			}
			Lower string
			Upper string
		}
	}
}

// Airspace definitions - similar to `airspaceResponse` but sanitised.
// github.com/golang/geo/r2

type Feature struct {
	ID       string
	Name     string
	Type     string
	Class    string
	Geometry []Volume
}

type Volume struct {
	ID                string
	Name              string
	Class             string
	Sequence          int
	Lower             float64
	Upper             float64
	ClearanceRequired bool
	// The (horizontal) shape will be either a circle or a polygon.
	// One of:
	Circle  Circle
	Polygon Polygon // Don't use geo.Polygon here, as that doesn't do JSON serialisation properly.
}

type Polygon []*geo.Point

type Circle struct {
	Radius float64
	Centre *geo.Point
}

func Decode(data []byte) ([]Feature, error) {
	var a airspaceResponse
	err := yaml.Unmarshal(data, &a)
	features, err := normalise(&a)
	return features, err
}

func normalise(a *airspaceResponse) ([]Feature, error) {
	var features []Feature
	for _, f := range a.Airspace {
		t := f.Type
		if f.Type == "OTHER" {
			t = f.LocalType
		} else if f.Type == "D_OTHER" {
			t = "Danger:" + f.LocalType
		}

		feat := Feature{
			ID:    f.ID,
			Name:  f.Name,
			Type:  t,
			Class: f.Class,
		}

		for _, g := range f.Geometry {
			id := g.ID
			name := g.Name
			class := g.Class
			if id == "" {
				id = feat.ID
			}
			if name == "" {
				name = feat.Name
			}
			if class == "" {
				class = feat.Class
			}

			vol := Volume{
				ID:                id,
				Name:              name,
				Class:             class,
				Sequence:          g.Seqno,
				Lower:             decodeHeight(g.Lower),
				Upper:             decodeHeight(g.Upper),
				ClearanceRequired: ClearanceRequired(feat),
			}

			var currentPos *geo.Point
			for _, b := range g.Boundary {
				if b.Circle.Radius != "" {
					var err error
					vol.Circle.Radius = decodeDistance(b.Circle.Radius)
					vol.Circle.Centre, err = parseLatLng(b.Circle.Centre)
					if err != nil {
						return nil, fmt.Errorf("bad circle %v: %s", b, err)
					}
				}
				for i := range b.Line {
					p, err := parseLatLng(b.Line[i])
					if err != nil {
						return nil, fmt.Errorf("bad line %v: %s", b, err)
					}
					vol.Polygon = append(vol.Polygon, p)
					currentPos = p
				}
				if b.Arc.Radius != "" {
					to, err := parseLatLng(b.Arc.To)
					if err != nil {
						return nil, fmt.Errorf("bad arc %v: %s", b, err)
					}
					radius := decodeDistance(b.Arc.Radius)
					centre, _ := parseLatLng(b.Arc.Centre)
					dir := +1.0
					if b.Arc.Dir == "ccw" {
						dir = -1.0
					}

					arc := arcToPolygon(centre, radius, currentPos, to, dir)
					vol.Polygon = append(vol.Polygon, arc...)
				}
			}

			feat.Geometry = append(feat.Geometry, vol)
		}

		features = append(features, feat)
	}

	return features, nil
}

func arcToPolygon(centre *geo.Point, radius float64, initialPoint *geo.Point, to *geo.Point, dir float64) []*geo.Point {
	initialAngleDeg := centre.BearingTo(initialPoint)
	finalAngleDeg := centre.BearingTo(to)

	if dir > 0 {
		// Clockwise
		if finalAngleDeg < initialAngleDeg {
			finalAngleDeg += 360
		}
	} else {
		if finalAngleDeg > initialAngleDeg {
			initialAngleDeg += 360
		}
	}

	var poly []*geo.Point
	for a := initialAngleDeg; dir*a < dir*finalAngleDeg; a += dir * 10 {
		point := centre.PointAtDistanceAndBearing(radius/1000, a)
		poly = append(poly, point)
	}
	poly = append(poly, to)

	return poly
}

func parseLatLng(str string) (*geo.Point, error) {
	returnedError := fmt.Errorf("bad point: %#q, must be in format %q (degrees,minutes,seconds)", str, "502257N 0033739W")

	if len(str) != 16 || str[7] != ' ' {
		return nil, returnedError
	}

	deg, err1 := strconv.ParseUint(str[0:2], 10, 64)
	mm, err2 := strconv.ParseUint(str[2:4], 10, 64)
	ss, err3 := strconv.ParseUint(str[4:6], 10, 64)
	if err1 != nil || err2 != nil || err3 != nil {
		return nil, returnedError
	}

	lat := float64(deg) + float64(mm)/60.0 + float64(ss)/2600.0
	if str[6] == 'S' {
		lat = -lat
	} else if str[6] != 'N' {
		return nil, returnedError
	}

	deg, err1 = strconv.ParseUint(str[8:11], 10, 64)
	mm, err2 = strconv.ParseUint(str[11:13], 10, 64)
	ss, err3 = strconv.ParseUint(str[13:15], 10, 64)
	if err1 != nil || err2 != nil || err3 != nil {
		return nil, returnedError
	}

	lng := float64(deg) + float64(mm)/60.0 + float64(ss)/2600.0
	if str[15] == 'W' {
		lng = -lng
	} else if str[15] != 'E' {
		return nil, returnedError
	}

	return geo.NewPoint(lat, lng), nil
}

func decodeHeight(h string) float64 {
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

func decodeDistance(d string) float64 {
	f, err := strconv.ParseFloat(strings.TrimSuffix(d, " nm"), 64)
	if err != nil {
		log.Printf("Invalid distance %#q: %s\n", d, err)
	}
	return nautMilesToMeters(f)
}

func nautMilesToMeters(nm float64) float64 {
	return nm * 1852
}

func metersToDegreesOfLat(m float64) float64 {
	return m / 1852 / degToNautMileY
}
func degreesOfLatToMeters(d float64) float64 {
	return d * 1852 * degToNautMileY
}
func degreesOfLngToMeters(d float64) float64 {
	return d * 1852 * degToNautMileX
}
func metersToDegreesOfLng(m float64) float64 {
	return m / 1852 / degToNautMileX
}

func Load(url string) ([]Feature, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return Decode(b)
}

func LoadFile(fileName string) ([]Feature, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	b, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}
	return Decode(b)
}

// https://developers.google.com/maps/documentation/javascript/overlays
// https://www.w3.org/Graphics/SVG/IG/resources/svgprimer.html#scale
// https://www.doc-developpement-durable.org/file/Projets-informatiques/cours-&-manuels-informatiques/htm-html-xml-ccs/Building%20Web%20Applications%20with%20SVG.pdf
// See https://eloquentjavascript.net/17_canvas.html
// http://jsfiddle.net/w1t1j2a1/
// https://en.wikipedia.org/wiki/Quadtree
