package scraping

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/paulcager/gosdata/osgrid"
)

type Site struct {
	Club    string
	SiteID  string
	Name    string
	SiteURL string
	Loc     struct {
		Lat float64
		Lon float64
	}
	Wind string
}

var errLeavingClub = errors.New("HTTP redirect away from club's site")

var httpClient = &http.Client{
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) >= 10 {
			return errors.New("stopped after 10 redirects")
		}
		if req.URL.Host != via[0].URL.Host {
			return errLeavingClub
		}
		return nil
	},
	Timeout: 30 * time.Second,
}

func Pennine() ([]Site, error) {
	r, err := openPage("http://www.penninesoaringclub.org.uk/sites")
	if err != nil {
		return nil, err
	}
	defer r.Close()

	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, err
	}

	var sites []Site
	doc.Find("a[href^='/sites/']").Each(func(i int, s *goquery.Selection) {
		href := s.AttrOr("href", "??")
		if href == "/sites/" {
			return
		}
		id := strings.TrimPrefix(strings.TrimSuffix(href, "/"), "/sites/")
		site := Site{
			Club:    "PSC",
			SiteID:  id,
			Name:    s.Text(),
			SiteURL: "http://www.penninesoaringclub.org.uk" + href,
		}
		sites = append(sites, site)
	})

	for i := range sites {
		err := pennineSite(&sites[i])
		if err != nil {
			panic(err) // TODO
		}
	}

	return sites, nil
}

func pennineSite(s *Site) error {
	r, err := openPage(s.SiteURL)
	if err != nil {
		return err
	}

	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return err
	}

	items := doc.Find("table.table td").Map(func(i int, s *goquery.Selection) string { return s.Text() })
	s.Wind = items[0]
	gridRef := sanitiseGridRef(items[1])
	lat, lon, err := osgrid.OSGridToLatLon(gridRef)
	if err != nil {
		return fmt.Errorf("invalid gridref %q: %s", gridRef, err)
	}
	s.Loc.Lat = lat
	s.Loc.Lon = lon

	return nil
}

func Dales() ([]Site, error) {
	r, err := openPage("https://www.dhpc.org.uk/sites")
	if err != nil {
		return nil, err
	}
	defer r.Close()

	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, err
	}

	var sites []Site
	doc.Find("aside#sites-list h4 a").Each(func(i int, s *goquery.Selection) {
		href := s.AttrOr("href", "??")
		id := strings.TrimPrefix(href, "site-guide/")
		site := Site{
			Club:    "DHPC",
			SiteID:  id,
			Name:    s.Text(),
			SiteURL: "https://www.dhpc.org.uk/" + href,
		}
		sites = append(sites, site)
	})

	var enrichedSites []Site
	for i := range sites {
		err := dalesSite(&sites[i])
		if err == nil {
			enrichedSites = append(enrichedSites, sites[i])
		} else if err, ok := err.(*url.Error); !ok || err.Err != errLeavingClub {
			fmt.Fprintf(os.Stderr, "Cannot enrich %q: %s\n", sites[i].Name, err)
		}
	}

	return enrichedSites, nil
}

func dalesSite(s *Site) error {
	r, err := openPage(s.SiteURL)
	if err != nil {
		return err
	}

	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return err
	}

	foundLatLon := false
	foundWind := false

	items := doc.Find("#main-content > div.left-col > p").Map(func(i int, s *goquery.Selection) string { return s.Text() })
	for i, str := range items {
		if strings.HasPrefix(str, "Lat, long:") {
			str = strings.TrimPrefix(str, "Lat, long:")
			parts := strings.Split(str, ",")
			if len(parts) != 2 {
				return fmt.Errorf("invalid lat/lon: %q", str)
			}
			parts[0] = strings.TrimSpace(parts[0])
			parts[1] = strings.TrimSpace(parts[1])
			lat, err0 := strconv.ParseFloat(parts[0], 64)
			lon, err1 := strconv.ParseFloat(parts[1], 64)
			if err0 != nil || err1 != nil {
				return fmt.Errorf("invalid lat/lon: %q", str)
			}
			s.Loc.Lat = lat
			s.Loc.Lon = lon

			foundLatLon = true
			continue
		}
		if str == "Wind direction" {
			if i < len(items)-1 {
				s.Wind = items[i+1]
				foundWind = true
			} else {
				return fmt.Errorf("invalid wind: %q", str)
			}
		}
	}

	if !foundWind || !foundLatLon || strings.TrimSpace(s.Wind) == "" {
		fmt.Println("Not worked")
	}

	return nil
}

func NorthWales() ([]Site, error) {
	r, err := openPage("https://www.nwhgpc.org/sites.html")
	if err != nil {
		return nil, err
	}
	defer r.Close()

	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, err
	}

	var sites []Site
	doc.Find("div.sitescontent li a").Each(func(i int, s *goquery.Selection) {
		href := s.AttrOr("href", "??")
		id := href
		site := Site{
			Club:    "NWHGPC",
			SiteID:  strings.TrimPrefix(id, "#"),
			Name:    s.Text(),
			SiteURL: "https://www.nwhgpc.org/sites.html" + href,
		}

		doc.Find("div" + href + " table.infotable tr").Each(func(i int, s *goquery.Selection) {
			if s.Find("th").Text() == "Map Location" {
				gridRef := sanitiseGridRef(s.Find("td").Text())
				lat, lon, err := osgrid.OSGridToLatLon(gridRef)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Could not decode grid ref %q\n", s.Find("td").Text())
				}
				site.Loc.Lat = lat
				site.Loc.Lon = lon
			}

			if strings.HasPrefix(s.Find("th").Text(), "Wind dir") {
				site.Wind = s.Find("td").Text()
			}
		})
		sites = append(sites, site)
	})

	return sites, nil
}

func LakeDistrict() ([]Site, error) {
	r, err := openPage("https://www.cumbriasoaringclub.co.uk/SiteManagement/CSC_SiteIndex.php")
	if err != nil {
		return nil, err
	}
	defer r.Close()

	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, err
	}

	var sites []Site
	doc.Find("div.sitescontent li a").Each(func(i int, s *goquery.Selection) {
		href := s.AttrOr("href", "??")
		id := href
		site := Site{
			Club:    "NWHGPC",
			SiteID:  strings.TrimPrefix(id, "#"),
			Name:    s.Text(),
			SiteURL: "https://www.nwhgpc.org/sites.html" + href,
		}

		doc.Find("div" + href + " table.infotable tr").Each(func(i int, s *goquery.Selection) {
			if s.Find("th").Text() == "Map Location" {
				gridRef := sanitiseGridRef(s.Find("td").Text())
				lat, lon, err := osgrid.OSGridToLatLon(gridRef)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Could not decode grid ref %q\n", s.Find("td").Text())
				}
				site.Loc.Lat = lat
				site.Loc.Lon = lon
			}

			if strings.HasPrefix(s.Find("th").Text(), "Wind dir") {
				site.Wind = s.Find("td").Text()
			}
		})
		sites = append(sites, site)
	})

	return sites, nil
}

func sanitiseGridRef(gridRef string) string {
	ind := strings.Index(gridRef, "(")
	if ind == -1 {
		ind = strings.Index(gridRef, " Sheet")
	}
	if ind != -1 {
		// E.g. "SD 782 403 (Sheet 103)"
		gridRef = gridRef[0:ind]
	}

	return gridRef
}

func openPage(url string) (io.ReadCloser, error) {
	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("HTTP status %q from %q", resp.Status, url)
	}
	return resp.Body, nil
}
