package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

var (
	model = make(map[string]interface{})
)

func main() {
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

	http.HandleFunc("/site-icons/", func(w http.ResponseWriter, r *http.Request){
		iconHandler(sites, w, r)
	})
	http.HandleFunc("/", rootHandler)
	log.Println("Starting HTTP server on 18080")
	log.Fatal(http.ListenAndServe(":18080", nil))
}

var (
	static = http.Dir("static")
	fs     = http.FileServer(static)
)

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

func iconHandler(sites map[string]Site, w http.ResponseWriter, r *http.Request) {
	path := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/site-icons/"), ".png")
	parts := strings.Split(path, "/")
	if len(parts) != 2 {
		http.Error(w, path + " invalid", http.StatusBadRequest)
		return
	}

	var size int
	switch parts[0] {
	case "small":
		size = 24
	case "large":
		size = 64
	default:
		if i, e := strconv.ParseUint(parts[0], 10, 32); e != nil {
			http.Error(w, parts[0] + " invalid", http.StatusBadRequest)
			return
		} else {
			size = int(i)
		}
	}

	if s, ok := sites[parts[1]]; ok {
		windIcon(w, size, s.Wind)
	} else {
		http.NotFound(w, r)
	}
}
