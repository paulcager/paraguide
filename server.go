package main

import (
	"encoding/json"
	"fmt"
	flag "github.com/spf13/pflag"
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
)

var (
	model       = make(map[string]interface{})
	fs          = http.FileServer(http.Dir("static"))
	imageCache  time.Duration
	staticCache time.Duration
	metRefresh  time.Duration
	noWeather   bool
	listenPort  string
)

func main() {
	flag.StringVar(&listenPort, "port", ":8080", "Port to listen on")
	flag.DurationVar(&imageCache, "image-cache-max-age", 7*24*time.Hour, "If not zero, the max-age property to set in Cache-Control for images")
	flag.DurationVar(&staticCache, "static-cache-max-age", 1*time.Hour, "If not zero, the max-age property to set in Cache-Control for static/template files")
	flag.DurationVar(&metRefresh, "met-refresh", 10*time.Minute, "How often to refresh weather data from metoffice")
	flag.BoolVar(&noWeather, "no-weather", false, "Prevent querying metoffice for weather.")
	flag.Parse()

	clubs, err := loadClubs(sheet)
	if err != nil {
		panic(err)
	}
	model["clubs"] = clubs

	sites, err := loadSites(sheet, clubs)
	if err != nil {
		panic(err)
	}
	model["sites"] = sites

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

	//queryMetSites()
	if !noWeather {
		startMetofficeRefresh(metRefresh)
	}

	s := makeHTTPServer(sites, listenPort)
	log.Fatal(s.ListenAndServe())
}

func makeHTTPServer(sites map[string]Site, listenPort string) *http.Server {
	http.Handle("/site-icons/", makeCachingHandler(imageCache, http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			iconHandler(sites, w, r)
		})))

	http.Handle("/wind-indicator/", makeCachingHandler(imageCache, http.HandlerFunc(windHandler)))

	http.Handle("/weather/", makeCachingHandler(metRefresh, http.HandlerFunc(weatherHandler)))

	http.Handle("/airspace/", makeCachingHandler(imageCache, http.HandlerFunc(airspaceHandler)))

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
	path := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/site-icons/"), ".png")
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
	path := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/wind-indicator/"), ".png")
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
	if err:= airspace.ToSVG(a, w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
