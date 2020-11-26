package airspace

import "gopkg.in/yaml.v2"

// Download airspace defs in yaml from https://gitlab.com/ahsparrow/airspace
// Schema is https://gitlab.com/ahsparrow/yaixm/-/blob/master/yaixm/data/schema.yaml
// yaml decoder is https://github.com/go-yaml/yaml

type Airspace struct{
	Airspace []Feature
}

type Feature struct {
	ID string
	Name string
	Type string
	LocalType string
	ControlType string
	Class string
	Geometry []Volume
}


type Volume struct {
	ID string
	Name string
	Class string
	Boundary []Boundary
	Lower string
	Upper string
}

// A Boundary is either a (single) circle, or a line followed by zero or more lines and arcs.
type Boundary struct {
	// One of:
	Circle [1]Circle
	Line []Line
	Arc Arc
}

type Circle struct {
	Radius string
	Centre Point
}

type Point string // E.g. 513239N 0010957W
type Line Point
type Arc struct {
	Dir string
	Radius string
	Centre Point
	To Point
}

func Decode(data []byte) (Airspace, error){
	var a Airspace
	err := yaml.Unmarshal(data, &a)
	return a, err
}

// https://developers.google.com/maps/documentation/javascript/overlays
// https://www.w3.org/Graphics/SVG/IG/resources/svgprimer.html#scale
// https://www.doc-developpement-durable.org/file/Projets-informatiques/cours-&-manuels-informatiques/htm-html-xml-ccs/Building%20Web%20Applications%20with%20SVG.pdf
// See https://eloquentjavascript.net/17_canvas.html
// http://jsfiddle.net/w1t1j2a1/
// https://en.wikipedia.org/wiki/Quadtree
