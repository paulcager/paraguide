package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/kr/pretty"
	"gopkg.in/yaml.v2"
)

var _ = yaml.Marshal

type Club struct {
	//	ID   string
	Name string
	URL  string
}

type Loc struct {
	Lat float64
	Lon float64
}

type Direction struct {
	From, To string
}

type Site struct {
	Name    string
	Clubs   []string
	Thing   Loc
	Parking []Loc
	Takeoff []Loc
	Landing []Loc
	Wind    []Direction
}

func parseYAML() {
	str := `

sites:
  Bradwell:
    name: Bradwell
    clubs: [ DSC ]
    thing: {lat: 22,   lon: 44}

clubs:
  DSC:
    name: DSC
    url: https://derbyshiresoaringclub.org.uk/

`
	var s struct {
		Clubs map[string]Club
		Sites map[string]Site
	}

	err := yaml.Unmarshal([]byte(str), &s)
	fmt.Println(err)
	//pretty.Println(s)
	fmt.Printf("%+v\n", s.Sites["Bradwell"])
	fmt.Printf("%T\n", s.Sites["Bradwell"].Clubs)
}

func main() {
	loadSpreadsheet()

}

func loadSpreadsheet() ([]Site, error) {
	resp, err := http.Get("https://spreadsheets.google.com/feeds/list/13blLictRsToqT7HReMA9IcUfHp3BzUPIhmgHadMmpW8/od6/public/full")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var v struct {
		XMLName xml.Name `xml:"feed"`
		Entries []struct {
			Club    string `xml:"club"`
			Place   string `xml:"place"`
			Parking string `xml:"parking"`
			Takeoff string `xml:"takeoff"`
			Landing string `xml:"landing"`
			Wind    string `xml:"wind"`
		} `xml:"entry"`
	}
	err = xml.Unmarshal(b, &v)
	if err != nil {
		return nil, err
	}

	pretty.Println("V", v)
	return nil, nil
}
