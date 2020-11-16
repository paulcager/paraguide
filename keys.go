package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"strings"
)

var (
	sheet  = getKey("sheet")
	apiKey = getKey("google_api")
	metOffice = getKey("metoffice")
)

func getKey(name string) string {
	if v := os.Getenv(strings.ToUpper(name)); v != "" {
		return v
	}

	if b, err := ioutil.ReadFile("secrets/" + name); err == nil {
		return string(bytes.TrimSpace(b))
	}

	if b, err := ioutil.ReadFile("/etc/" + name); err == nil {
		return string(bytes.TrimSpace(b))
	}

	panic("No value for key: " + name)
}
