package main

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
	"gopkg.in/yaml.v2"
)

var (
	sheetsAPIKey = readKey("sheets")
	mapsAPIKey   = readKey("maps")
)

func readKey(keyType string) string {
	env := strings.ToUpper(keyType) + "_API_KEY"
	file := "/etc/paraguide/" + strings.ToLower(keyType) + "APIKey"

	if key := os.Getenv(env); key != "" {
		return key
	}

	b, err := ioutil.ReadFile(file)
	if err != nil {
		panic("Must supply either " + env + " or " + file)
	}
	return string(bytes.TrimSpace(b))
}

/*
Reminders to add to README

Go to https://console.developers.google.com/apis/dashboard?authuser=0&folder=&organizationId=&project=paraguide-uk.
Enable APIs + Services - sheets
OAuth consent screen - external    ?? Maybe
Credentials
API KEY
    /etc/paraguide_api_key
Create service account Prob not.
*/

type Club struct {
	//	ID   string
	Name string
	URL  string
}

type Loc struct {
	Lat float64
	Lon float64
}

// Direction is the range of directions (in degrees).
type Direction struct {
	Text string
	From float64
	To   float64
}

type Site struct {
	Name    string
	Clubs   []string
	Parking []Loc
	Takeoff []Loc
	Landing []Loc
	Wind    []Direction
}

type LoadError struct {
	Name string
	Err  error
}

func assertNoError(err error) {
	if err != nil {
		panic(err)
	}
}
func parseYAML() {
	str := `

sites:
  Bradwell:
    name: Bradwell
    clubs: [ DSC ]
    thing: {lat: 22,   lon: 44}

clubs:
  DSC:
    name: DSC
    url: https://derbyshiresoaringclub.org.uk/

`
	var s struct {
		Clubs map[string]Club
		Sites map[string]Site
	}

	err := yaml.Unmarshal([]byte(str), &s)
	fmt.Println(err)
	//pretty.Println(s)
	fmt.Printf("%+v\n", s.Sites["Bradwell"])
	fmt.Printf("%T\n", s.Sites["Bradwell"].Clubs)
}

func main() {
	fmt.Println(LoadDefault())
}

func LoadDefault() ([]Site, []LoadError) {
	//return Load("https://spreadsheets.google.com/feeds/list/13blLictRsToqT7HReMA9IcUfHp3BzUPIhmgHadMmpW8/od6/public/full")
	return Load("13blLictRsToqT7HReMA9IcUfHp3BzUPIhmgHadMmpW8")
}

func Load(sheet string) ([]Site, []LoadError) {
	ctx := context.Background()
	sheetsService, err := sheets.NewService(ctx,
		option.WithScopes(sheets.SpreadsheetsReadonlyScope),
		option.WithAPIKey(sheetsAPIKey),
	)
	if err != nil {
		return nil, []LoadError{{Err: err}}
	}
	resp, err := sheetsService.Spreadsheets.Values.Get(sheet, "Sites!A2:F").Do()
	if err != nil {
		return nil, []LoadError{{Err: err}}
	}

	var (
		sites  []Site
		errors []LoadError
	)

	for _, row := range resp.Values {
		parking, err1 := parseLocations(row[2].(string))
		takeoff, err2 := parseLocations(row[3].(string))
		landing, err3 := parseLocations(row[4].(string))
		wind, err4 := parseWind(row[5].(string))

		site := Site{
			Clubs:   strings.Split(row[0].(string), ","),
			Name:    row[1].(string),
			Parking: parking,
			Takeoff: takeoff,
			Landing: landing,
			Wind:    wind,
		}
		sites = append(sites, site)

		if err := anyError(err1, err2, err3, err4); err != nil {
			errors = append(errors, LoadError{
				Name: site.Name,
				Err:  err,
			})
		}
	}

	return sites, errors
}

func anyError(errors ...error) error {
	for _, err := range errors {
		if err != nil {
			return err
		}
	}
	return nil
}

var (
	spaceRegexp = regexp.MustCompile(`\s+`)
	commaRegexp = regexp.MustCompile(`,\s*`)
)

func parseLocations(s string) ([]Loc, error) {
	if s == "" {
		return nil, nil
	}
	parts := spaceRegexp.Split(s, -1)
	locs := make([]Loc, 0, len(parts))

	var lastErr error
	for _, part := range parts {
		latLon := strings.Split(part, ",")
		if len(latLon) != 2 {
			lastErr = fmt.Errorf("invalid LatLon: %s", part)
			continue
		}
		lat, err1 := strconv.ParseFloat(latLon[0], 64)
		lon, err2 := strconv.ParseFloat(latLon[1], 64)
		if err1 != nil || err2 != nil {
			lastErr = fmt.Errorf("invalid LatLon: %s", part)
			continue
		}

		locs = append(locs, Loc{
			Lat: lat,
			Lon: lon,
		})
	}

	return locs, lastErr
}

// parseWind decodes a wind direction (as a string) into a []Direction.
// The strings may look like this: "E, SE-SSW", meaning it is flyable
// when the wind is:
//	- from the East (interpreted as ENE to ESE)
//	- between Se and SSW.
func parseWind(s string) ([]Direction, error) {
	if s == "" {
		return nil, nil
	}

	parts := commaRegexp.Split(s, -1)
	dirs := make([]Direction, 0, len(parts))

	var lastErr error

	for _, part := range parts {
		fromTo := strings.Split(part, "-")

		var from, to float64
		from, err := parseDirection(fromTo[0])
		if err != nil {
			lastErr = err
			continue
		}
		if len(fromTo) == 1 {
			from = from - 22.5
			if from < 0 {
				from = from + 22.5
			}
			to = from + 45
			if to >= 360 {
				to = to - 360
			}
		} else {
			to, err = parseDirection(fromTo[1])
			if err != nil {
				lastErr = err
				continue
			}
		}
		dirs = append(dirs, Direction{
			Text: part,
			From: from,
			To:   to,
		})
	}

	return dirs, lastErr
}

var directionMap map[string]float64

func init() {
	directionMap = make(map[string]float64)
	for i, name := range []string{"N", "NNE", "NE", "ENE", "E", "ESE", "SE", "SSE", "S", "SSW", "SW", "WSW", "W", "WNW", "NW", "NNW"} {
		directionMap[name] = float64(i) * 360 / 16
	}
}

func parseDirection(s string) (float64, error) {
	if d, ok := directionMap[strings.ToUpper(s)]; ok {
		return d, nil
	} else if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f, nil
	} else {
		return 0, fmt.Errorf("invalid direction %q", s)
	}
}

//NB: to translate to OSGB, see https://www.bgs.ac.uk/data/webservices/CoordConvert_LL_BNG.cfc?method=LatLongToBNG&lat=53.191443&lon=-1.849545
