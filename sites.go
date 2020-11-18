package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type Club struct {
	ID         string
	Name       string
	URL        string
	SiteFormat string // E.g. http://www.longmynd.org/?page_id=%s
}

type Loc struct {
	Lat float64
	Lon float64
}

func (l Loc) String() string {
	return fmt.Sprintf("%.6f,%.6f", l.Lat, l.Lon)
}
func (l Loc) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{
		"lat":      l.Lat,
		"lon":      l.Lon,
	}
	return json.Marshal(m)
}

// WindRange is the range of directions (in degrees).
type WindRange struct {
	Text string
	From float64
	To   float64
}

type Site struct {
	ID        string
	Name      string
	Club      Club
	Parking   []Loc
	Takeoff   []Loc
	Landing   []Loc
	Wind      []WindRange
	SiteGuide string
}

var sheetsAPIKey = "AIzaSyCtdAkZbd3K1nBsDUf-Isj2t49lVD4xvVY"

func loadLookup(sheet string, ranges string) (map[string]string, error) {
	ctx := context.Background()
	sheetsService, err := sheets.NewService(ctx,
		option.WithScopes(sheets.SpreadsheetsReadonlyScope),
		option.WithAPIKey(sheetsAPIKey),
	)
	if err != nil {
		return nil, err
	}
	resp, err := sheetsService.Spreadsheets.Values.Get(sheet, ranges).Do()
	if err != nil {
		return nil, err
	}

	var (
		m   = make(map[string]string)
	)

	for _, row := range resp.Values {
		m[row[0].(string)] = row[1].(string)
	}

	return m, nil
}

func loadSites(sheet string, clubs map[string]Club) (map[string]Site, error) {
	ctx := context.Background()
	sheetsService, err := sheets.NewService(ctx,
		option.WithScopes(sheets.SpreadsheetsReadonlyScope),
		option.WithAPIKey(sheetsAPIKey),
	)
	if err != nil {
		return nil, err
	}
	resp, err := sheetsService.Spreadsheets.Values.Get(sheet, "Sites!A2:G").Do()
	if err != nil {
		return nil, err
	}

	var (
		sites   = make(map[string]Site)
		lastErr error
	)

	for _, row := range resp.Values {
		parking, err1 := parseLocations(row[2].(string))
		takeoff, err2 := parseLocations(row[3].(string))
		landing, err3 := parseLocations(row[4].(string))
		wind, err4 := parseWind(row[5].(string))

		if err := anyError(err1, err2, err3, err4); err != nil {
			lastErr = err
			fmt.Fprintf(os.Stderr,
				"Error in record %#q\n\t%s",
				row,
				err)
			continue
		}

		club := clubs[row[0].(string)]
		name := row[1].(string)
		id := idOf(name)

		siteGuide := ""
		if len(row) > 6 {
			siteGuide = fmt.Sprintf(club.SiteFormat, row[6].(string))
		} else {
			siteGuide = club.SiteFormat
		}

		site := Site{
			ID:      id,
			Club:    club,
			Name:    name,
			Parking: parking,
			Takeoff: takeoff,
			Landing: landing,
			Wind:    wind,
			SiteGuide: siteGuide,
		}
		sites[id] = site
	}

	return sites, lastErr
}

func loadClubs(sheet string) (map[string]Club, error) {
	ctx := context.Background()
	sheetsService, err := sheets.NewService(ctx,
		option.WithScopes(sheets.SpreadsheetsReadonlyScope),
		option.WithAPIKey(sheetsAPIKey),
	)
	if err != nil {
		return nil, err
	}
	resp, err := sheetsService.Spreadsheets.Values.Get(sheet, "Clubs!A2:D").Do()
	if err != nil {
		return nil, err
	}

	var (
		clubs   = make(map[string]Club)
	)

	for _, row := range resp.Values {
		c := Club{
			ID:         row[0].(string),
			Name:       row[1].(string),
			URL:        row[2].(string),
			SiteFormat: row[3].(string),
		}

		clubs[c.ID] = c
	}

	return clubs, nil
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
		return []Loc{}, nil
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

// parseWind decodes a wind direction (as a string) into a []WindRange.
// The strings may look like this: "E, SE-SSW", meaning it is flyable
// when the wind is:
//	- from the East (interpreted as ENE to ESE)
//	- between Se and SSW.
func parseWind(s string) ([]WindRange, error) {
	if s == "" {
		return nil, nil
	}

	parts := commaRegexp.Split(s, -1)
	dirs := make([]WindRange, 0, len(parts))

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
		dirs = append(dirs, WindRange{
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
