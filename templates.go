package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"text/template"
)

var (
	templates = make(map[string]*template.Template)
	funcMap   = template.FuncMap{
		"id":      idOf,
		"json":    toJSON,
		"api_key": func() string { return apiKey },
	}
)

func idOf(s string) string {
	s = strings.ReplaceAll(s, " ", "_")
	s = strings.ReplaceAll(s, "'", "_")
	return s
}

func toJSON(x interface{}) (string, error) {
	if x == nil && reflect.ValueOf(x).Kind() == reflect.Slice {
		return "[]", nil
	}
	var buff bytes.Buffer
	encoder := json.NewEncoder(&buff)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(x)
	b := bytes.TrimRight(buff.Bytes(), "\n")
	return string(b), err
}

func init() {
	filepath.Walk("templates",
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}

			name := strings.TrimPrefix(path, "templates")
			b, err := ioutil.ReadFile(path)
			if err != nil {
				panic(err)
			}
			t := template.Must(template.New(name).Funcs(funcMap).Parse(string(b)))
			templates[name] = t
			if name == "/index.html" {
				templates["/"] = templates[name]
			}
			return nil
		})
}
