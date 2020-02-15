package main

import (
	"fmt"
	"net/http"

	"github.com/paulcager/paraguide/sites"
	"github.com/paulcager/paraguide/templates"
)

var (
	Sites []sites.Site
)

func main() {
	Sites, _ = sites.LoadDefault()
	http.HandleFunc("/", templateHandler)
	fmt.Println("Listening on port 8080")
	http.ListenAndServe(":8080", nil)
}

func templateHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if path == "/" || path == "" {
		path = "index"
	}

	ok := templates.Execute(path, w, Sites, nil)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		fmt.Println(path, "not found")
		return
	}
}
