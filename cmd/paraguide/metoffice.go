package main

import (
	"encoding/json"
	"fmt"
	"github.com/kr/pretty"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"sync/atomic"
	"time"
)

type MetofficeSite struct {
	ID        string
	Name      string
	Elevation float64
	Lat, Lon  float64
}

type MetofficeReport struct {
	MetofficeSite
	Time time.Time
	// TODO Period is time.Duration daily, 3-hourly etc.
	Weather          string
	WeatherId        int
	Gust             float64
	Temperature      float64
	Visibility       string
	WindDirection    string
	WindSpeed        float64
	Pressure         float64
	PressureTendency string
}

type LatLon struct {
	Lat, Lon float64
}

type Rectangle struct {
	Min, Max LatLon
}

type etag struct {
	eTag string
	date string
}

var (
	LatestReports atomic.Value // of map[string] MetofficeReport
)

func init() {
	LatestReports.Store(make(map[string]MetofficeReport))
}

func getMatchingWeather(rect Rectangle) []MetofficeReport {
	m := LatestReports.Load().(map[string]MetofficeReport)
	var ret []MetofficeReport
	for _, report := range m {
		if report.Lat >= rect.Min.Lat && report.Lat <= rect.Max.Lat &&
			report.Lon >= rect.Min.Lon && report.Lon <= rect.Max.Lon {
			ret = append(ret, report)
		}
	}

	return ret
}

func getAllWeather() []MetofficeReport {
	m := LatestReports.Load().(map[string]MetofficeReport)
	var ret []MetofficeReport
	for _, report := range m {
		ret = append(ret, report)
	}

	return ret
}

func startMetofficeRefresh(interval time.Duration) error {
	// Load it synchronously the first time, then kick off a background job to refresh it.
	// Keep track of the ETags to reduce traffic.

	var eTag etag
	m, err := queryCurrentMetWeather(&eTag)
	if err != nil {
		return err
	}
	LatestReports.Store(m)

	go func() {
		for {
			time.Sleep(interval)
			if m, err := queryCurrentMetWeather(&eTag); err == nil {
				if m != nil {
					log.Println("Updated current weather report,", len(m), "items")
					LatestReports.Store(m)
				} else {
					log.Println("No updates to the weather report")
				}
			} else {
				log.Println("Could not refresh metoffice data: ", err)
			}
		}
	}()

	return nil
}

func queryMetSites() (map[string]MetofficeSite, error) {
	type metSite struct {
		Locations struct {
			Location []struct {
				ID              string
				Elevation       string
				Latitude        string
				Longitude       string
				Name            string
				Region          string
				UnitaryAuthArea string
			}
		}
	}

	var metResponse metSite
	_, err := readURL("http://datapoint.metoffice.gov.uk/public/data/val/wxobs/all/json/sitelist?key="+metOffice, nil, &metResponse)
	if err != nil {
		return nil, err
	}

	m := make(map[string]MetofficeSite)
	for _, loc := range metResponse.Locations.Location {
		// Silently ignore any invalid / missing values
		elevation, _ := strconv.ParseFloat(loc.Elevation, 64)
		lat, _ := strconv.ParseFloat(loc.Latitude, 64)
		lon, _ := strconv.ParseFloat(loc.Longitude, 64)

		m[loc.ID] = MetofficeSite{
			ID:        loc.ID,
			Name:      loc.Name,
			Elevation: elevation,
			Lat:       lat,
			Lon:       lon,
		}
	}

	pretty.Println(m)
	pretty.Println(metWeatherLookup)
	return m, err
}

var metWeatherLookup = map[string]string{
	// See https://www.metoffice.gov.uk/services/data/datapoint/code-definitions
	"NA": "Not available",
	"0":  "Clear night",
	"1":  "Sunny day",
	"2":  "Partly cloudy (night)",
	"3":  "Partly cloudy (day)",
	"4":  "Not used",
	"5":  "Mist",
	"6":  "Fog",
	"7":  "Cloudy",
	"8":  "Overcast",
	"9":  "Light rain shower (night)",
	"10": "Light rain shower (day)",
	"11": "Drizzle",
	"12": "Light rain",
	"13": "Heavy rain shower (night)",
	"14": "Heavy rain shower (day)",
	"15": "Heavy rain",
	"16": "Sleet shower (night)",
	"17": "Sleet shower (day)",
	"18": "Sleet",
	"19": "Hail shower (night)",
	"20": "Hail shower (day)",
	"21": "Hail",
	"22": "Light snow shower (night)",
	"23": "Light snow shower (day)",
	"24": "Light snow",
	"25": "Heavy snow shower (night)",
	"26": "Heavy snow shower (day)",
	"27": "Heavy snow",
	"28": "Thunder shower (night)",
	"29": "Thunder shower (day)",
	"30": "Thunder",
}

func queryCurrentMetWeather(eTag *etag) (map[string]MetofficeReport, error) {
	type report struct {
		SiteRep struct {
			DV struct {
				DataDate string
				Type     string
				Location []struct {
					I         string
					Lat       string
					Lon       string
					Name      string
					Elevation string
					Period    struct {
						Type  string // "Day"
						Value string // "YYYY-MM-DDZ"
						Rep   struct {
							G, T, V, D, S, W, P, Pt string
						}
					}
				}
			}
		}
	}

	url := fmt.Sprintf("http://datapoint.metoffice.gov.uk/public/data/val/wxobs/all/json/all?&res=hourly&time=%d&key=%s",
		time.Now().UTC().Hour(),
		metOffice)
	var metResponse report
	found, err := readURL(url, eTag, &metResponse)
	if !found || err != nil {
		return nil, err
	}

	m := make(map[string]MetofficeReport)
	for _, loc := range metResponse.SiteRep.DV.Location {
		// Silently ignore any invalid / missing values
		elevation, _ := strconv.ParseFloat(loc.Elevation, 64)
		lat, _ := strconv.ParseFloat(loc.Lat, 64)
		lon, _ := strconv.ParseFloat(loc.Lon, 64)

		timestamp, _ := time.Parse("2006-01-02Z", loc.Period.Value)

		rep := loc.Period.Rep
		gust, _ := strconv.ParseFloat(rep.G, 64)
		temp, _ := strconv.ParseFloat(rep.T, 64)
		windSpeed, _ := strconv.ParseFloat(rep.S, 64)
		press, _ := strconv.ParseFloat(rep.P, 64)
		weather, err := strconv.ParseInt(rep.W, 10, 32)
		if err != nil {
			weather = -1
		}

		// See also https://erikflowers.github.io/weather-icons/
		// and https://openweathermap.org/current

		m[loc.I] = MetofficeReport{
			MetofficeSite: MetofficeSite{
				ID:        loc.I,
				Name:      loc.Name,
				Elevation: elevation,
				Lat:       lat,
				Lon:       lon,
			},
			Time:             timestamp,
			Weather:          metWeatherLookup[rep.W],
			WeatherId:        int(weather),
			Gust:             gust,
			Temperature:      temp,
			Visibility:       rep.V,
			WindDirection:    rep.D,
			WindSpeed:        windSpeed,
			Pressure:         press,
			PressureTendency: rep.Pt,
		}
		//pretty.Println(m[loc.I])
	}

	return m, err
}

func readURL(url string, eTag *etag, obj interface{}) (bool, error) {
	if t := reflect.TypeOf(obj); t.Kind() != reflect.Ptr {
		panic("Expecting pointer, got " + t.String())
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return false, err
	}

	if eTag != nil && eTag.date != "" && eTag.eTag != "" {
		req.Header.Add("If-None-Match", eTag.eTag)
		req.Header.Add("If-Modified-Since", eTag.date)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	log.Println("Status: ", resp.Status, "Old Etag: ", eTag, "new", resp.Header.Get("ETag"), resp.Header.Get("Date"))

	if resp.StatusCode == http.StatusNotModified {
		return false, nil
	}

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("HTTP error: %s", resp.Status)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	err = json.Unmarshal(b, obj)
	if err != nil {
		return false, fmt.Errorf("could not unmarshal %#q: %s", b, err)
	}

	if eTag != nil {
		eTag.eTag = resp.Header.Get("ETag")
		eTag.date = resp.Header.Get("Date")
	}
	return true, nil
}
