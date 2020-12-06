package scraping

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
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
	doc.Find("section > table.linkstable > tbody > tr").Each(func(i int, s *goquery.Selection) {
		a := s.Find("td:nth-child(1) > a")
		if a.Length() == 0 {
			// Must be the heading row.
			return
		}
		href := a.AttrOr("href", "")
		id := href
		if pos := strings.LastIndexByte(href, '='); pos != -1 {
			id = href[pos+1:]
		}
		wind := s.Find("td:nth-child(3)").Text()
		site := Site{
			Club:    "CSC",
			SiteID:  id,
			Name:    a.Text(),
			SiteURL: "https://www.cumbriasoaringclub.co.uk/SiteManagement/" + href,
			Wind:    wind,
		}
		sites = append(sites, site)
	})

	var enrichedSites []Site
	for i := range sites {
		err := LakeDistrictSite(&sites[i])
		if err == nil {
			enrichedSites = append(enrichedSites, sites[i])
		} else if err, ok := err.(*url.Error); !ok || err.Err != errLeavingClub {
			fmt.Fprintf(os.Stderr, "Cannot enrich %q: %s\n", sites[i].Name, err)
		}
	}

	return enrichedSites, nil
}

func LakeDistrictSite(site *Site) error {
	r, err := openPage(site.SiteURL)
	if err != nil {
		return err
	}

	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return err
	}

	// body > section > div > div.tab1 > table > tbody > tr:nth-child(7) > td:nth-child(2)
	doc.Find("div div.tab1 table tr").Each(func(i int, s *goquery.Selection) {
		tds := s.Find("td").Map(func(i int, s *goquery.Selection) string {
			return s.Text()
		})

		if len(tds) >= 2 && strings.HasPrefix(tds[0], "Grid Ref") {
			gridRef := sanitiseGridRef(tds[1])
			lat, lon, err := osgrid.OSGridToLatLon(gridRef)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Could not decode grid ref %q\n", tds[1])
			}
			site.Loc.Lat = lat
			site.Loc.Lon = lon
		}
	})

	return nil
}

func MidWales() ([]Site, error) {
	r, err := openPage("https://www.flymidwales.org.uk/flying-sites")
	if err != nil {
		return nil, err
	}
	defer r.Close()

	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, err
	}

	var sites []Site
	doc.Find("div.view-flying-sites table tr").Each(func(i int, s *goquery.Selection) {
		a := s.Find("td:nth-child(2) > a")
		href := a.AttrOr("href", "")
		if href == "" {
			return
		}
		id := strings.ReplaceAll(href, "/", "-")
		site := Site{
			Club:    "MWHGPC",
			SiteID:  id,
			Name:    a.Text(),
			SiteURL: "https://www.flymidwales.org.uk" + href,
		}
		sites = append(sites, site)
	})

	var enrichedSites []Site
	for i := range sites {
		err := MidWalesSite(&sites[i])
		if err == nil {
			enrichedSites = append(enrichedSites, sites[i])
		} else if err, ok := err.(*url.Error); !ok || err.Err != errLeavingClub {
			fmt.Fprintf(os.Stderr, "Cannot enrich %q: %s\n", sites[i].Name, err)
		}
	}

	return enrichedSites, nil
}

func MidWalesSite(site *Site) error {
	r, err := openPage(site.SiteURL)
	if err != nil {
		return err
	}

	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return err
	}

	// #block-system-main > div > div > div.field.field-name-field-wind.field-type-computed.field-label-inline.clearfix > div.field-items > div
	site.Wind = doc.Find("div.field-name-field-wind div.field-item").First().Text()
	gridRef := sanitiseGridRef(doc.Find("div.field-name-field-ref div.field-item").First().Text())
	lat, lon, err := osgrid.OSGridToLatLon(gridRef)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not decode grid ref %q for site %q\n", gridRef, site.Name)
	}
	site.Loc.Lat = lat
	site.Loc.Lon = lon

	return nil
}

func Snowdonia() ([]Site, error) {
	r, err := openPage("https://www.snowdoniaskysports.co.uk/node/3")
	if err != nil {
		return nil, err
	}
	defer r.Close()

	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, err
	}

	var sites []Site
	doc.Find("div.view-sites-guide div.view-content table tr").Each(func(i int, s *goquery.Selection) {
		a := s.Find("td:nth-child(1) > a")
		href := a.AttrOr("href", "")
		if href == "" {
			fmt.Fprintf(os.Stderr, "Skipping Snowdonia site %q as no hyperlink. Site probably closed\n", s.Find("td:nth-child(1)").Text())
			return
		}
		id := strings.ReplaceAll(href, "/", "-")
		site := Site{
			Club:    "Snowdonia",
			SiteID:  id,
			Name:    a.Text(),
			SiteURL: "https://www.snowdoniaskysports.co.uk/" + href,
		}
		sites = append(sites, site)
	})

	var enrichedSites []Site
	for i := range sites {
		err := SnowdoniaSite(&sites[i])
		if err == nil {
			enrichedSites = append(enrichedSites, sites[i])
		} else if err, ok := err.(*url.Error); !ok || err.Err != errLeavingClub {
			fmt.Fprintf(os.Stderr, "Cannot enrich %q: %s\n", sites[i].Name, err)
		}
	}

	return enrichedSites, nil
}

func SnowdoniaSite(site *Site) error {
	r, err := openPage(site.SiteURL)
	if err != nil {
		return err
	}

	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return err
	}

	values := doc.Find("div.content div.node-flying-site")
	site.Wind = values.Find("div.field-name-field-best-wind-cardinal div.field-item").Text()
	gridRef := sanitiseGridRef(values.Find("div.field-name-field-grid-ref div.field-item").Text())
	lat, lon, err := osgrid.OSGridToLatLon(gridRef)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not decode grid ref %q for site %q\n", gridRef, site.Name)
	}
	site.Loc.Lat = lat
	site.Loc.Lon = lon

	return nil
}

func WelshBorders() ([]Site, error) {
	r, err := openPage("https://paraglidingwales.co.uk/general-information/")
	if err != nil {
		return nil, err
	}
	defer r.Close()

	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, err
	}

	var sites []Site
	doc.Find("div.entry-content table tr").Each(func(i int, s *goquery.Selection) {
		a := s.Find("td:nth-child(1) a")
		href := a.AttrOr("href", "")
		if href == "" {
			fmt.Fprintf(os.Stderr, "Skipping MWHGPC site %q as no hyperlink. Site probably closed\n", s.Find("td:nth-child(1)").Text())
			return
		}
		id := path.Base(href)
		name := a.Text()
		gridRef := sanitiseGridRef(s.Find("td:nth-child(4) span").Text())

		wind := strings.TrimSpace(s.Find("td:nth-child(3)").Text())
		lat, lon, err := osgrid.OSGridToLatLon(gridRef)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not decode grid ref %q for site %q\n", gridRef, name)
		}
		site := Site{
			Club:    "MWHGPC",
			SiteID:  id,
			Name:    name,
			SiteURL: href,
			Wind:    wind,
		}
		site.Loc.Lat = lat
		site.Loc.Lon = lon

		sites = append(sites, site)
	})

	return sites, nil
}

func Cayley() ([]Site, error) {
	r, err := openPage("https://www.cayleyparagliding.co.uk/")
	if err != nil {
		return nil, err
	}
	defer r.Close()

	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, err
	}

	var sites []Site
	doc.Find("ul.sub-menu span:contains('Sites (alphabetically)')").Parent().Siblings().Find("li a").Each(func(i int, s *goquery.Selection) {
		href := s.AttrOr("href", "")
		if href == "" {
			fmt.Fprintf(os.Stderr, "Skipping Cayley site %q as no hyperlink. Site probably closed\n", s.Text())
			return
		}
		id := path.Base(href)
		name := s.Text()
		site := Site{
			Club:    "Cayley",
			SiteID:  id,
			Name:    name,
			SiteURL: href,
		}

		sites = append(sites, site)
	})

	var enrichedSites []Site
	for i := range sites {
		err := CayleySite(&sites[i])
		if err == nil {
			enrichedSites = append(enrichedSites, sites[i])
		} else if err, ok := err.(*url.Error); !ok || err.Err != errLeavingClub {
			fmt.Fprintf(os.Stderr, "Cannot enrich %q: %s\n", sites[i].Name, err)
		}
	}

	return enrichedSites, nil
}

func CayleySite(site *Site) error {
	r, err := openPage(site.SiteURL)
	if err != nil {
		return err
	}

	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return err
	}

	site.Wind = doc.Find("h1:contains('Paragliding') ~ :contains('Wind Direction:')").Text()
	if ind := strings.LastIndexByte(site.Wind, ':'); ind != -1 {
		site.Wind = site.Wind[ind+1:]
	}

	mapRef := doc.Find("h1:contains('Paragliding') ~ :contains('Google Map Link:') a").AttrOr("href", "")
	start := strings.IndexByte(mapRef, '@')
	if start == -1 {
		fmt.Fprintf(os.Stderr, "Could not extract lat/lon for site %q: %s\n", site.Name, mapRef)
		return nil
	}
	parts := strings.Split(mapRef[start+1:], ",")
	if len(parts) < 2 {
		fmt.Fprintf(os.Stderr, "Could not extract lat/lon for site %q: %s\n", site.Name, mapRef)
		return nil
	}

	lat, err1 := strconv.ParseFloat(parts[0], 64)
	lon, err2 := strconv.ParseFloat(parts[1], 64)
	if err1 != nil || err2 != nil {
		fmt.Fprintf(os.Stderr, "Could not extract lat/lon for site %q: %s\n", site.Name, mapRef)
		return nil
	}

	site.Loc.Lat = lat
	site.Loc.Lon = lon

	return nil
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
	request, _ := http.NewRequest("GET", url, nil)
	request.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.198 Safari/537.36")
	resp, err := httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		b, _ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if len(b) > 0 {
			return nil, fmt.Errorf("HTTP status %q from %q\n%s", resp.Status, url, b)
		}
		return nil, fmt.Errorf("HTTP status %q from %q", resp.Status, url)
	}
	return resp.Body, nil
}
