package main

type KML struct {
	Document struct {
		Name   string `xml:"name"`
		Folder struct {
			Name        string `xml:"name"`
			Open        string `xml:"open"`
			Description string `xml:"description"`
			Folder      []struct {
				Name        string `xml:"name"`
				Description string `xml:"description"`
				Placemark   []struct {
					Name        string `xml:"name"`
					Description string `xml:"description"`
					Point       struct {
						Coordinates string `xml:"coordinates"`
					} `xml:"Point"`
				} `xml:"Placemark"`
			} `xml:"Folder"`
		} `xml:"Folder"`
	} `xml:"Document"`
}
