package airspace

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"
	"strconv"
)

// Download airspace defs in yaml from https://gitlab.com/ahsparrow/airspace
// Schema is https://gitlab.com/ahsparrow/yaixm/-/blob/master/yaixm/data/schema.yaml
// yaml decoder is https://github.com/go-yaml/yaml

type Airspace struct {
	Airspace []Feature
}

type Feature struct {
	ID          string
	Name        string
	Type        string
	LocalType   string
	ControlType string
	Class       string
	Geometry    []Volume
}

type Volume struct {
	ID       string
	Name     string
	Class    string
	Boundary []Boundary
	Lower    string
	Upper    string
}

// A Boundary is either a (single) circle, or a line followed by zero or more lines and arcs.
type Boundary struct {
	// One of:
	Circle Circle
	Line   []Point
	Arc    Arc
}

type Circle struct {
	Radius string
	Centre Point
}

type Point struct {
	valid bool
	Text  string
	x,y  float64
}

func NewPoint(str string) Point {
	var p Point
	if err := (&p).Set(str); err != nil {
		panic(err)
	}

	return p
}

func (p Point) X() float64 {
	if !p.valid {
		panic("Must create Points from NewPoint or Point.Set: " + p.Text)
	}
	return p.x
}
func (p Point) Y() float64 {
	if !p.valid {
		panic("Must create Points from NewPoint or Point.Set: " + p.Text)
	}
	return p.y
}

func (p *Point) Set(str string) error {
	p.valid = false
	p.Text = str

	returnedError := fmt.Errorf("bad point: %#q, must be in format %q (degrees,minutes,seconds)", str, "502257N 0033739W")

	if len(str) != 16 || str[7] != ' ' {
		return returnedError
	}

	deg, err1 := strconv.ParseUint(str[0:2], 10, 64)
	mm, err2 := strconv.ParseUint(str[2:4], 10, 64)
	ss, err3 := strconv.ParseUint(str[4:6], 10, 64)
	if err1 != nil || err2 != nil || err3 != nil {
		return returnedError
	}

	p.y = float64(deg) + float64(mm)/60.0 + float64(ss)/2600.0
	if str[6] == 'S' {
		p.y = -p.y
	} else if str[6] != 'N' {
		return returnedError
	}

	deg, err1 = strconv.ParseUint(str[8:11], 10, 64)
	mm, err2 = strconv.ParseUint(str[11:13], 10, 64)
	ss, err3 = strconv.ParseUint(str[13:15], 10, 64)
	if err1 != nil || err2 != nil || err3 != nil {
		return returnedError
	}

	p.x = float64(deg) + float64(mm)/60.0 + float64(ss)/2600.0
	if str[15] == 'W' {
		p.x = -p.x
	} else if str[15] != 'E' {
		return returnedError
	}

	p.valid = true
	return nil
}

func (p *Point) UnmarshalJSON(data []byte) error {
	panic("Not implemented - expecting YAML")
}

func (p *Point) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var str string
	err := unmarshal(&str)
	if err != nil {
		return err
	}

	return p.Set(str)
}

func (p Point) String() string {
	if !p.valid {
		return "Unverified:" + p.Text
	}
	return p.Text
}

type Arc struct {
	Dir    string
	Radius string
	Centre Point
	To     Point
}



func Decode(data []byte) (Airspace, error) {
	var a Airspace
	err := yaml.Unmarshal(data, &a)
	return a, err
}

func Load(url string) (Airspace, error) {
	resp, err := http.Get(url)
	if err != nil {
		return Airspace{}, err
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Airspace{}, err
	}
	return Decode(b)
}

// https://developers.google.com/maps/documentation/javascript/overlays
// https://www.w3.org/Graphics/SVG/IG/resources/svgprimer.html#scale
// https://www.doc-developpement-durable.org/file/Projets-informatiques/cours-&-manuels-informatiques/htm-html-xml-ccs/Building%20Web%20Applications%20with%20SVG.pdf
// See https://eloquentjavascript.net/17_canvas.html
// http://jsfiddle.net/w1t1j2a1/
// https://en.wikipedia.org/wiki/Quadtree
