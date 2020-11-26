package airspace

import (
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)
import "github.com/stretchr/testify/assert"

var data = `
airspace:
- name: ABERDEEN CTA
  id: aberdeen-cta
  type: CTA
  class: D
  geometry:
  - seqno: 1
    upper: FL115
    lower: 1500 ft
    boundary:
    - line:
      - 572153N 0015835W
      - 572100N 0015802W
      - 572100N 0023356W
    - arc:
        dir: cw
        radius: 10 nm
        centre: 571834N 0021602W
        to: 572153N 0015835W
  - seqno: 2
    upper: FL115
    lower: 1500 ft
    boundary:
    - line:
      - 571522N 0015428W
      - 570845N 0015019W
    - arc:
        dir: cw
        radius: 10 nm
        centre: 570531N 0020740W
        to: 570214N 0022458W
    - line:
      - 570850N 0022913W
    - arc:
        dir: ccw
        radius: 10 nm
        centre: 571207N 0021152W
        to: 571522N 0015428W
  - seqno: 3
    upper: FL115
    lower: 3000 ft
    boundary:
    - line:
      - 572100N 0023356W
      - 570015N 0025056W
      - 565433N 0023557W
      - 565533N 0020635W
    - arc:
        dir: cw
        radius: 10 nm
        centre: 570531N 0020740W
        to: 570214N 0022458W
    - line:
      - 571520N 0023326W
    - arc:
        dir: cw
        radius: 10 nm
        centre: 571834N 0021602W
        to: 572100N 0023356W


`
func TestDecode(t *testing.T) {
	a, err := Decode([]byte(data))
	require.NoError(t, err)
	assert.Equal(t, "aberdeen-cta", a.Airspace[0].ID)
	assert.Equal(t, "D", a.Airspace[0].Class)
	assert.Equal(t, 3, len(a.Airspace[0].Geometry))
	assert.Equal(t, "FL115", a.Airspace[0].Geometry[0].Upper)
	assert.Equal(t, "1500 ft", a.Airspace[0].Geometry[0].Lower)
	assert.Equal(t, 2, len(a.Airspace[0].Geometry[0].Boundary))
	assert.Equal(t, Arc{}, a.Airspace[0].Geometry[0].Boundary[0].Arc)
	assert.Equal(t, Circle{}, a.Airspace[0].Geometry[0].Boundary[0].Circle)
	assert.Equal(t, []Point([]Point{NewPoint("572153N 0015835W"), NewPoint("572100N 0015802W"), NewPoint("572100N 0023356W")}), a.Airspace[0].Geometry[0].Boundary[0].Line)
	assert.Equal(t, Arc(Arc{Dir: "cw", Radius: "10 nm", Centre: NewPoint("571834N 0021602W"), To: NewPoint("572153N 0015835W")}), a.Airspace[0].Geometry[0].Boundary[1].Arc)
}

func TestDownload(t *testing.T) {
	// Verify real-life data exists and can be parsed correctly.
	url := `https://gitlab.com/ahsparrow/airspace/-/raw/master/airspace.yaml`
	a, err := Load(url)
	require.NoError(t, err)

	assert.Greater(t, len(a.Airspace), 600)
}

func TestSVG(t *testing.T) {
	url := `https://gitlab.com/ahsparrow/airspace/-/raw/master/airspace.yaml`
	a, err := Load(url)
	require.NoError(t, err)

	err = ToSVG(a, os.Stdout)
	require.NoError(t, err)
}