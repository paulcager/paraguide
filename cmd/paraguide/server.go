package main

import (
	"encoding/json"
	"fmt"
	"github.com/paulcager/osgridref"
	"image/png"
	"log"
	"net"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/paulcager/gb-airspace"
	"github.com/paulcager/go-http-middleware"
	flag "github.com/spf13/pflag"
)

const (
	apiVersion = "v4"
)

var (
	model           = make(map[string]interface{})
	Airspace        map[string]airspace.Feature
	fs              = http.FileServer(http.Dir("static"))
	imageCache      time.Duration
	staticCache     time.Duration
	listenPort      string
	includeKMLSites bool
	clubCacheMaxAge time.Duration
	clubCacheDir    = "club-cache"
	heightServer    = "http://osheight-server:9091"
	airspaceServer  = "http://airspace-server:9092"
)

func main() {
	flag.StringVar(&listenPort, "port", ":8080", "Port to listen on")
	flag.DurationVar(&imageCache, "image-cache-max-age", 7*24*time.Hour, "If not zero, the max-age property to set in Cache-Control for images")
	flag.DurationVar(&staticCache, "static-cache-max-age", 1*time.Hour, "If not zero, the max-age property to set in Cache-Control for static/template files")
	flag.BoolVar(&includeKMLSites, "include-kml-sites", false, "Include sites read from KML file")
	flag.DurationVar(&clubCacheMaxAge, "club-cache-max-age", 24*time.Hour, "Ignore cached scrapes of sites if older tna this.")
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

	model["airspaceServer"] = airspaceServer
	model["heightServer"] = heightServer

	Airspace, err = GetAirspace()
	if err != nil {
		log.Printf("Could not get airspace from server: %s\n", err)
	} else {
		model["airspace"] = Airspace
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
	http.Handle("/"+apiVersion+"/site-icons/", middleware.MakeCachingHandler(imageCache, http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			iconHandler(sites, w, r)
		})))

	http.Handle("/"+apiVersion+"/wind-indicator/", middleware.MakeCachingHandler(imageCache, http.HandlerFunc(windHandler)))

	http.HandleFunc("/"+apiVersion+"/airspace/", getAirspaceHandler)

	http.HandleFunc("/"+apiVersion+"/location", locationInfoHandler)

	// Encourage Google to drop the cached sites by returning "Gone"
	http.Handle("/sites/", middleware.MakeCachingHandler(24*time.Hour, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "410 page gone", http.StatusGone)
	})))

	http.HandleFunc("/headers", headersHandler)
	http.HandleFunc("/about", aboutHandler)

	http.Handle("/", middleware.MakeCachingHandler(staticCache, http.HandlerFunc(rootHandler)))

	if !strings.Contains(listenPort, ":") {
		listenPort = ":" + listenPort
	}

	log.Println("Starting HTTP server on " + listenPort)
	s := &http.Server{
		ReadHeaderTimeout: 20 * time.Second,
		WriteTimeout:      2 * time.Minute,
		IdleTimeout:       10 * time.Minute,
		Handler:           middleware.MakeLoggingHandler(http.DefaultServeMux),
		Addr:              listenPort,
	}

	return s
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

func locationInfoHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	var (
		gridRef osgridref.OsGridRef
		err     error
	)
	if s := strings.TrimSpace(q.Get("gridref")); s != "" {
		gridRef, err = osgridref.ParseOsGridRef(s)
	} else if s := strings.TrimSpace(q.Get("latlon")); s != "" {
		var latLon osgridref.LatLonEllipsoidalDatum
		latLon, err = osgridref.ParseLatLon(s, 0, osgridref.WGS84)
		if err == nil {
			gridRef = latLon.ToOsGridRef()
		}
	} else {
		err = fmt.Errorf("missing gridref or latlon parameters")
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	info, err := GetLocationInfo(gridRef)
	if err != nil {
		log.Printf("Error getting location info for %q: %s\n", gridRef, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//json.NewEncoder(os.Stderr).Encode(info)

	w.Header().Add("Content-Type", "text/html")
	t := templates["/loc-info.html"]
	err = t.Execute(w, info)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", r.URL, err)
		// In case nothing has yet been sent
		w.WriteHeader(http.StatusBadGateway)
		fmt.Fprintf(w, "%s: %s\n", r.URL, err)
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

func getAirspaceHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/"+apiVersion+"/airspace/")
	feature, ok := Airspace[id]
	if !ok {
		http.NotFound(w, r)
		return
	}

	b, err := json.Marshal(&feature)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	_, err = w.Write(b)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func airspaceSVGHandler(w http.ResponseWriter, r *http.Request) {
	a, err := airspace.Load(`https://gitlab.com/ahsparrow/airspace/-/raw/master/airspace.yaml`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.Header().Add("Content-Type", "image/svg+xml")
	if err := airspace.ToSVG(a, w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func headersHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain")
	r.Header.Write(w)
}

func aboutHandler(w http.ResponseWriter, r *http.Request) {
	var originIP string
	for _, ff := range r.Header.Values("X-Forwarded-For") {
		ip := net.ParseIP(ff)
		if ip == nil {
			continue
		}

		if ip.IsPrivate() {
			continue
		}

		originIP = ff
	}

	if originIP == "" {
		// rick-roll instead.
		w.Header().Add("Location", "https://www.youtube.com/watch?v=dQw4w9WgXcQ")
		w.WriteHeader(http.StatusTemporaryRedirect)
		return
	}

	var geo struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	}

	resp, err := http.Get("http://ipwhois.app/json/" + originIP)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&geo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	fmt.Printf("%s geolocated to %v\n", r.Header.Values("X-Forwarded-For"), geo)
	url := fmt.Sprintf("https://nuclearsecrecy.com/nukemap/?&kt=50000&lat=%f&lng=%f&hob_psi=5&hob_ft=37743&ff=3&psi=20,5,1&zm=9", geo.Latitude, geo.Longitude)
	w.Header().Add("Location", url)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
