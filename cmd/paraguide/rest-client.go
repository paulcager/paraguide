package main

import (
	"encoding/json"
	"fmt"
	airspace "github.com/paulcager/gb-airspace"
	"github.com/paulcager/osgridref"
	"net/http"
	"time"
)

var (
	httpClient     = &http.Client{Timeout: 4 * time.Second}
	heightServer   = "http://osheight-server:9091"
	airspaceServer = "http://airspace-server:9092"
)

type LocationInfo struct {
	GridRef  string  `json:"gridRef"`
	Easting  int     `json:"easting"`
	Northing int     `json:"northing"`
	Lat      float64 `json:"lat"`
	Lon      float64 `json:"lon"`
	Height   float64 `json:"height"`
	Airspace []airspace.Volume
}

func GetLocationInfo(gridRef osgridref.OsGridRef) (LocationInfo, error) {
	// TODO change to concurrent queries.
	height, err1 := getHeight(gridRef)
	lat, lon := gridRef.ToLatLon()
	volumes, err2 := getAirspace(lat, lon)

	err := err1
	if err == nil {
		err = err2
	}

	return LocationInfo{
		GridRef:  gridRef.StringN(10),
		Easting:  gridRef.Easting,
		Northing: gridRef.Northing,
		Lat:      lat,
		Lon:      lon,
		Height:   height,
		Airspace: volumes,
	}, err
}

func getHeight(gridRef osgridref.OsGridRef) (float64, error) {
	resp, err := httpClient.Get(heightServer + "/v4/height/" + gridRef.StringNCompact(8))
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var reply struct {
		Height float64 `json:"height"`
	}
	err = json.NewDecoder(resp.Body).Decode(&reply)
	return reply.Height, err
}

func getAirspace(lat, lon float64) ([]airspace.Volume, error) {
	resp, err := httpClient.Get(airspaceServer + "/v4/airspace/" + fmt.Sprintf("%f,%f", lat, lon))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var reply []airspace.Volume
	err = json.NewDecoder(resp.Body).Decode(&reply)
	return reply, err
}
