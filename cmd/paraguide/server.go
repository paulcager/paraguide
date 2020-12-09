package main

import (
	"encoding/json"
	"fmt"
	"image/png"
	"log"
	"math"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/paulcager/paraguide/airspace"
	"github.com/paulcager/paraguide/scraping"
	flag "github.com/spf13/pflag"
)

const (
	apiVersion = "v4"
)

var (
	model           = make(map[string]interface{})
	fs              = http.FileServer(http.Dir("static"))
	imageCache      time.Duration
	staticCache     time.Duration
	metRefresh      time.Duration
	osgridServer string
	noWeather       bool
	listenPort      string
	includeKMLSites bool
)

func main() {
	flag.StringVar(&listenPort, "port", ":8080", "Port to listen on")
	flag.DurationVar(&imageCache, "image-cache-max-age", 7*24*time.Hour, "If not zero, the max-age property to set in Cache-Control for images")
	flag.DurationVar(&staticCache, "static-cache-max-age", 1*time.Hour, "If not zero, the max-age property to set in Cache-Control for static/template files")
	flag.DurationVar(&metRefresh, "met-refresh", 10*time.Minute, "How often to refresh weather data from metoffice")
	flag.BoolVar(&noWeather, "no-weather", false, "Prevent querying metoffice for weather.")
	flag.BoolVar(&includeKMLSites, "include-kml-sites", false, "Include sites read from KML file")
	flag.StringVar(&scraping.OSGridServer, "osgrid-server", "http://localhost:9090/", "REST server address for OS grid ref translation")
	flag.Parse()

	http.DefaultClient.Timeout = time.Minute

	model["apiVersion"] = apiVersion

	clubs, err := loadClubs()
	if err != nil {
		panic(err)
	}
	model["clubs"] = clubs

	sites, err := loadSites(clubs)
	if err != nil {
		panic(err)
	}
	model["sites"] = sites
	if err := saveSites(sites); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not save sites file: %s\n", err)
	}

	siteIDs := sortSites(sites)
	model["siteIDs"] = siteIDs

	forecasts, err := loadLookup(sheet, "Forecasts!A:B")
	if err != nil {
		panic(err)
	}
	model["forecasts"] = forecasts

	webcams, err := loadLookup(sheet, "Webcams!A:B")
	if err != nil {
		panic(err)
	}
	model["webcams"] = webcams

	air, err := airspace.Load(`https://gitlab.com/ahsparrow/airspace/-/raw/master/airspace.yaml`)
	//air, err := airspace.LoadFile(`donc.yml`)
	if err != nil {
		panic(err)
	}
	model["airspace"] = air

	//queryMetSites()
	if !noWeather {
		startMetofficeRefresh(metRefresh)
	}

	s := makeHTTPServer(sites, listenPort)
	log.Fatal(s.ListenAndServe())
}

func sortSites(sites map[string]Site) []string {
	// Add a sorted list of sites, to display in menus etc. Sorted on club name, then site name.
	siteIDs := make([]string, 0, len(sites))
	for id := range sites {
		siteIDs = append(siteIDs, id)
	}
	sort.Slice(siteIDs, func(i, j int) bool {
		siteI := sites[siteIDs[i]]
		siteJ := sites[siteIDs[j]]
		if siteI.Club.ID != siteJ.Club.ID {
			return siteI.Club.ID < siteJ.Club.ID
		}
		return siteI.Name < siteJ.Name
	})
	return siteIDs
}

func makeHTTPServer(sites map[string]Site, listenPort string) *http.Server {
	http.Handle("/"+apiVersion+"/site-icons/", makeCachingHandler(imageCache, http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			iconHandler(sites, w, r)
		})))

	http.Handle("/"+apiVersion+"/wind-indicator/", makeCachingHandler(imageCache, http.HandlerFunc(windHandler)))

	http.Handle("/"+apiVersion+"/weather/", makeCachingHandler(metRefresh, http.HandlerFunc(weatherHandler)))

	//http.Handle("/airspace/", makeCachingHandler(imageCache, http.HandlerFunc(airspaceHandler)))
	//features, _ := airspace.Load(`https://gitlab.com/ahsparrow/airspace/-/raw/master/airspace.yaml`)
	http.HandleFunc("/airspace/debug/", func(w http.ResponseWriter, r *http.Request) {
		features, _ := airspace.LoadFile("donc.yml")
		w.Header().Add("Content-Type", "text/plain")
		json.NewEncoder(w).Encode(features)
	})

	http.Handle("/", makeCachingHandler(staticCache, http.HandlerFunc(rootHandler)))

	if !strings.Contains(listenPort, ":") {
		listenPort = ":" + listenPort
	}

	log.Println("Starting HTTP server on " + listenPort)
	s := &http.Server{
		ReadHeaderTimeout: 20 * time.Second,
		WriteTimeout:      2 * time.Minute,
		IdleTimeout:       10 * time.Minute,
		Handler:           makeLoggingHandler(http.DefaultServeMux),
		Addr:              listenPort,
	}

	return s
}

func makeLoggingHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			h.ServeHTTP(w, r)
			end := time.Now()

			uri := r.URL.String()
			method := r.Method
			fmt.Printf("%s %s %s %d\n", method, uri, r.RemoteAddr, end.Sub(start).Milliseconds())
		})
}

func makeCachingHandler(age time.Duration, h http.Handler) http.Handler {
	ageSeconds := int64(math.Round(age.Seconds()))
	if ageSeconds <= 0 {
		return h
	}

	header := fmt.Sprintf("public,max-age=%d", ageSeconds)
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Cache-Control", header)
			h.ServeHTTP(w, r)
		})
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	if t, ok := templates[r.URL.Path]; ok {
		switch {
		case strings.HasSuffix(r.URL.Path, ".js"):
			w.Header().Add("Content-Type", "text/javascript")
		case strings.HasSuffix(r.URL.Path, ".html") || r.URL.Path == "/":
			w.Header().Add("Content-Type", "text/html")
		}
		err := t.Execute(w, model)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %s\n", r.URL, err)
			// In case nothing has yet been sent
			w.WriteHeader(http.StatusBadGateway)
			fmt.Fprintf(w, "%s: %s\n", r.URL, err)
		}
	} else {
		fs.ServeHTTP(w, r)
	}
}

func weatherHandler(w http.ResponseWriter, r *http.Request) {
	var reports []MetofficeReport
	if r.URL.RawQuery == "" {
		reports = getAllWeather()
	} else {
		lat1, err1 := floatParam(r, "south")
		lon1, err2 := floatParam(r, "west")
		lat2, err3 := floatParam(r, "north")
		lon2, err4 := floatParam(r, "east")

		if anyError(err1, err2, err3, err4) != nil {
			http.Error(w, "Invalid params: "+r.URL.String(), http.StatusBadRequest)
			return
		}

		rect := Rectangle{
			Min: LatLon{
				Lat: lat1,
				Lon: lon1,
			},
			Max: LatLon{
				Lat: lat2,
				Lon: lon2,
			},
		}

		reports = getMatchingWeather(rect)
	}

	if reports == nil {
		reports = []MetofficeReport{}
	}
	sort.Slice(reports, func(i, j int) bool { return reports[i].ID < reports[j].ID })

	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	err := enc.Encode(reports)
	if err != nil {
		http.Error(w, "Marshal failed "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func floatParam(r *http.Request, name string) (float64, error) {
	return strconv.ParseFloat(r.URL.Query().Get(name), 64)
}

func iconHandler(sites map[string]Site, w http.ResponseWriter, r *http.Request) {
	path := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/"+apiVersion+"/site-icons/"), ".png")
	parts := strings.Split(path, "/")
	if len(parts) != 2 {
		http.NotFound(w, r)
		return
	}

	var size int
	switch parts[0] {
	case "small":
		size = 24
	case "large":
		size = 64
	case "massive":
		size = 256
	default:
		// The following is good for testing, but it would enable DoS attacks.
		/*if i, e := strconv.ParseUint(parts[0], 10, 32); e != nil {
			http.Error(w, parts[0]+" invalid", http.StatusBadRequest)
			return
		} else {
			size = int(i)
		}*/
		http.NotFound(w, r)
		return
	}

	s, ok := sites[parts[1]]
	if !ok {
		fmt.Fprintf(os.Stderr, "No site %#q\n", parts[1])
		http.NotFound(w, r)
		return
	}

	// Note that generating these icons is somewhat expensive. We rely on caching in the reverse proxy and at the
	// Cloudflare edge.
	img := windIcon(size, s.Wind)
	if err := png.Encode(w, img); err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
	}
}

func windHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/"+apiVersion+"/wind-indicator/"), ".png")
	parts := strings.Split(path, "/")
	if len(parts) != 2 {
		http.NotFound(w, r)
		return
	}

	speed, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	direction, err := parseDirection(parts[1])
	if err != nil && speed > 0 {
		http.NotFound(w, r)
		return
	}

	img := windIndicator(speed, direction)
	if err := png.Encode(w, img); err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
	}
}

func airspaceHandler(w http.ResponseWriter, r *http.Request) {
	a, err := airspace.Load(`https://gitlab.com/ahsparrow/airspace/-/raw/master/airspace.yaml`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.Header().Add("Content-Type", "image/svg+xml")
	if err := airspace.ToSVG(a, w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
