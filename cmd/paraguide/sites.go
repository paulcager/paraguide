package main

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/paulcager/paraguide/scraping"
	"net/url"
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
		"lat": l.Lat,
		"lon": l.Lon,
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

func loadLookup(sheet string, ranges string) (map[string]string, error) {
	ctx := context.Background()
	sheetsService, err := sheets.NewService(ctx,
		option.WithScopes(sheets.SpreadsheetsReadonlyScope),
		option.WithAPIKey(apiKey),
	)
	if err != nil {
		return nil, err
	}
	resp, err := sheetsService.Spreadsheets.Values.Get(sheet, ranges).Do()
	if err != nil {
		return nil, err
	}

	var (
		m = make(map[string]string)
	)

	for _, row := range resp.Values {
		m[row[0].(string)] = row[1].(string)
	}

	return m, nil
}

func loadSites(clubs map[string]Club) (map[string]Site, error) {
	sites := make(map[string]Site)
	err := loadSitesFromSheet(sites, sheet, clubs)
	n := len(sites)

	if err == nil && includeKMLSites {
		err = loadSitesFromKml(sites, "NorthernSites.kml", clubs)
	}

	//if pennineSites, err := scraping.Pennine(); err == nil {
	//	addScraped(sites, pennineSites)
	//} else {
	//	// Log error but carry on
	//	fmt.Fprintf(os.Stderr, "Could not add Pennine sites: %s\n", err)
	//}
	//
	//if dalesSites, err := scraping.Dales(); err == nil {
	//	addScraped(sites, dalesSites)
	//} else {
	//	// Log error but carry on
	//	fmt.Fprintf(os.Stderr, "Could not add Dales sites: %s\n", err)
	//}

	if northWalesSites, err := scraping.NorthWales(); err == nil {
		addScraped(sites, northWalesSites)
	} else {
		// Log error but carry on
		fmt.Fprintf(os.Stderr, "Could not add Dales sites: %s\n", err)
	}


	fmt.Fprintf(os.Stderr, "Read %d + %d = %d sites\n", n, len(sites)-n, len(sites))
	return sites, err
}

func addScraped(sites map[string]Site, scraped []scraping.Site) {
	for _, s := range scraped {
		wind, err := parseWind(s.Wind)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not set wind %q for %q: %s\n", s.Wind, s.Name, err)
			wind = []WindRange{{Text: s.Wind}}
		}
		newSite := Site{
			ID:        s.Club + "-" + s.SiteID,
			Name:      s.Name,
			Club:      Club{ID: s.Club},
			Takeoff:   []Loc{s.Loc},
			Parking:   []Loc{},
			Landing:   []Loc{},
			Wind:      wind,
			SiteGuide: s.SiteURL,
		}
		sites[newSite.ID] = newSite
	}
}

func loadSitesFromSheet(sites map[string]Site, sheet string, clubs map[string]Club) error {
	ctx := context.Background()
	sheetsService, err := sheets.NewService(ctx,
		option.WithScopes(sheets.SpreadsheetsReadonlyScope),
		option.WithAPIKey(apiKey),
	)
	if err != nil {
		return err
	}
	resp, err := sheetsService.Spreadsheets.Values.Get(sheet, "Sites!A2:G").Do()
	if err != nil {
		return err
	}

	var lastErr error

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
			guideRef := row[6].(string)
			if u, err := url.Parse(guideRef); err == nil && u.IsAbs() {
				siteGuide = guideRef
			} else {
				siteGuide = fmt.Sprintf(club.SiteFormat, guideRef)
			}
		} else {
			siteGuide = club.SiteFormat
		}

		site := Site{
			ID:        id,
			Club:      club,
			Name:      name,
			Parking:   parking,
			Takeoff:   takeoff,
			Landing:   landing,
			Wind:      wind,
			SiteGuide: siteGuide,
		}
		sites[id] = site
	}

	return lastErr
}

func loadSitesFromKml(sites map[string]Site, fileName string, clubs map[string]Club) error {
	file, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	var kml KML
	err = xml.NewDecoder(file).Decode(&kml)
	if err != nil {
		return err
	}

	for _, club := range kml.Document.Folder.Folder {
		for _, place := range club.Placemark {
			var wind []WindRange
			if start := strings.LastIndexByte(place.Description, '('); start != -1 {
				windStr := place.Description[start+1:]
				if end := strings.IndexByte(windStr, ')'); end != -1 {
					wind, err = parseWind(windStr[:end])
					if err != nil {
						fmt.Fprintf(os.Stderr, "Bad wind directions for %+v\n", place)
						continue
					}
				}
			}

			if wind == nil {
				fmt.Fprintf(os.Stderr, "Wind direction unknown for %+v\n", place)
			}

			parts := strings.Split(place.Point.Coordinates, ",")
			if len(parts) != 3 {
				fmt.Fprintf(os.Stderr, "Expected 3 coordinates for %+v\n", place)
				continue
			}
			lng, err1 := strconv.ParseFloat(parts[0], 64)
			lat, err2 := strconv.ParseFloat(parts[1], 64)
			if err1 != nil || err2 != nil {
				fmt.Fprintf(os.Stderr, "Invalid coordinates for %+v\n", place)
				continue
			}

			site := Site{
				ID:   "k-" + idOf(place.Name),
				Name: place.Name,
				Club: Club{
					ID:   "k-" + club.Name,
					Name: club.Name,
				},
				Parking: []Loc{},
				Takeoff: []Loc{{Lat: lat, Lon: lng}},
				Landing: []Loc{},
				Wind:    wind,
			}

			sites[site.ID] = site
		}
	}

	return nil
}

func loadClubs() (map[string]Club, error) {
	ctx := context.Background()
	sheetsService, err := sheets.NewService(ctx,
		option.WithScopes(sheets.SpreadsheetsReadonlyScope),
		option.WithAPIKey(apiKey),
	)
	if err != nil {
		return nil, err
	}
	resp, err := sheetsService.Spreadsheets.Values.Get(sheet, "Clubs!A2:D").Do()
	if err != nil {
		return nil, err
	}

	var (
		clubs = make(map[string]Club)
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
	andRegexp = regexp.MustCompile(` and `)
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
//	- between SE and SSW.
// Other formats are also seen on club's websites, such as "NNE - NE (020-040)" (Dales)
// or "W to NW" (Pennine).
func parseWind(s string) ([]WindRange, error) {
	if s == "" {
		return nil, nil
	}

	parts := commaRegexp.Split(s, -1)
	if len(parts) == 1 {
		parts = andRegexp.Split(s, -1)
	}
	dirs := make([]WindRange, 0, len(parts))

	var lastErr error

	if strings.HasPrefix(s, "all") {
		return []WindRange{{Text: s, From: 0, To: 360}}, nil
	}

	for _, part := range parts {
		if ind := strings.IndexByte(part, '('); ind != -1 {
			// Strip out "(020 - 040)", although maybe it would be better to actually use those values.
			part = part[0:ind]
		}
		fromTo := strings.Split(part, "-")
		if len(fromTo) == 1 {
			// For the benefir of Cow Close Fell, which uses a unicode dash
			fromTo = strings.Split(part, " \xe2\x80\x93 ")
		}
		if len(fromTo) == 1 {
			fromTo = strings.Split(part, " to ")
		}
		for i := range fromTo {
			fromTo[i] = strings.TrimSpace(fromTo[i])
		}

		var from, to float64
		from, err := parseDirection(fromTo[0])
		if err != nil {
			lastErr = err
			continue
		}

		if len(fromTo) == 1 && fromTo[0] == "no wind" {
			continue
		} else if len(fromTo) == 1 {
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

		// Some sites are specified as e.g. "SW-SE" rather than "SE-SW". Use a heuristic that sites tend not to
		// have continuous range > 180 degrees.
		rang := to - from
		if to < from {
			rang = 360 - from + to
		}
		if rang > 180 {
			fmt.Println("Swapping", part)
			from, to = to, from
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
