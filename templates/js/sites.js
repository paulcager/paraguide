var airspaceOverlay;
var map;
var sites = {}
var weather = {}
var landColor = "20e000";
var pinImage = new google.maps.MarkerImage(
    "http://chart.apis.google.com/chart?chst=d_map_pin_letter&chld=L|" + landColor,
    new google.maps.Size(21, 34),
    new google.maps.Point(0, 0),
    new google.maps.Point(10, 34)
);

var parkingImage = new google.maps.MarkerImage(
    "img/parking.png",
    new google.maps.Size(21, 34),
    new google.maps.Point(0, 0),
    new google.maps.Point(10, 34)
);

function weatherImage(windSpeed, windDirection) {
    return new google.maps.MarkerImage(
        "/wind-indicator/" + windSpeed + "/" + windDirection,
        new google.maps.Size(80, 80),
        null,
        new google.maps.Point(40, 40));
}


var infoWindow = new google.maps.InfoWindow({});

function create_sites() {
    var sites = {};
    {{range .sites}}
    sites["{{id .Name}}"] = create_site(
        "{{id .Name}}",
        "{{.Name}}",
        "{{.SiteGuide}}",
        {{json .Parking}},
    {{json .Takeoff}},
    {{json .Landing}}
    );
    {{end}}

    return sites;
}

function create_site(id, name, guide, parkings, takeoffs, landings) {
    var icon = icon_url("small", name);
    site = {
        id: id,
        name: name,
        parkings: create_markers(id, name, guide, parkings, parkingImage),
        takeoffs: create_markers(id, name, guide, takeoffs, {url: icon, anchor: new google.maps.Point(12, 12)}),
        landings: create_markers(id, name, guide, landings, pinImage),
        guide: guide,
    }

    site.info = site.takeoffs[0].info;

    return site;
}

function show_url(id) {
    var url = "#" + id;
    $('#url-dialog').html(
        "<p>Copy the following link to share.</p>" +
        "<a href='" + url + "'>Direct Link</a>"
    );
    console.log("show_url(" + id + ")");
    $('#url-dialog').dialog();
}

function add_info_window(place, guide) {
    var name = place.name;
    var id = place.id;
    var marker = place.marker;
    var icon = icon_url("large", name);
    var lat = place.lat;
    var lon = place.lon;
    var osGrid = place.osGrid;

    var f = function () {
        infoWindow.setContent(
            "<div>" +
            "<h1>" + name + "</h1>" +
            "<div style='padding: 0.5em;'>" +
            "<img src='" + icon + "'/>" +
            "</div>" +
            "<p>" +
            "<a href='" + guide + "' target='guide'>Guide</a><br/>" +
            "<a href='" + maps_url(lat, lon, 9, false) + "'>Directions</a><br/>" +
            "<a href='" + maps_url(lat, lon, 15, true) + "'>Satellite View</a><br/>" +
            "<a href='http://www.streetmap.co.uk/ids.srf?mapp=map.srf&searchp=ids&name=" + lat + "," + lon + "&type=LatLong'>OS Map</a><br/>" +
            "<a href='javascript:show_url(\"" + id + "\")'>Bookmark Site</a><br/>" +
            "</p>" +
            "</div>");
        infoWindow.open(map, marker);
    };
    google.maps.event.addListener(marker, 'click', f);
    place.info = f;
}


function create_markers(id, name, guide, places, icon) {
    var i;
    for (i = 0; i < places.length; i++) {
        var t = places[i];
        t.id = id;
        t.name = name;
        var marker = new google.maps.Marker({
            position: new google.maps.LatLng(t.lat, t.lon),
            map: map,
            title: name,
            icon: icon,
            draggable: false,
        });
        t.marker = marker;
        add_info_window(t, guide);
    }
    return places;
}

function icon_url(type, place) {
    return "site-icons/" + type + "/" + safe_name(place) + ".png";
}

function safe_name(place) {
    return place.replace(/'/g, "_").replace(/ /g, "_");
}

function maps_url(lat, lng, zoom, satellite) {
    //return "https://www.google.co.uk/maps/place/" + String(lat) + "," + String(lng) + "/@" + String(lat) + "," + String(lng) + "," + String(zoom) + "z";
    //return "https://www.google.com/maps/embed/v1/place?key=XXXXXXX&q=" + String(lat) + "," + String(lng) + "&zoom=" + String(zoom) + "z&maptype=" + (satellite ? "satellite" : "roadmap");
    //return "http://maps.google.com/maps?q=" + lat + "," + lng + "&t=(satellite ? "k" : "m")
    var t = (satellite ? "k" : "m")
    var latlon = lat + "," + lng;
    return "https://maps.google.com/?q=" + latlon + "&ll=" + latlon + "&t=" + t + "&z=" + zoom;
}

function initialize() {
    var myLatlng = new google.maps.LatLng(53.23455851981886, -1.8269212925847522);
    var mapOptions = {
        zoom: 8,
        center: myLatlng,
        mapTypeControlOptions: {
            style: google.maps.MapTypeControlStyle.DEFAULT,
            position: google.maps.ControlPosition.BOTTOM_LEFT
        },
    }
    map = new google.maps.Map(document.getElementById('map-canvas'), mapOptions);
    sites = create_sites();

    if (sites[window.location.hash.substring(1)]) {
        sites[window.location.hash.substring(1)].info();
    }

    const imageBounds = {
        north: 59.0,
        south: 49.5,
        east:   2.0,
        west:  -6.5,
    };

    airspaceOverlay = new google.maps.GroundOverlay(
        "/airspace/",
        imageBounds
    );
    airspaceOverlay.setMap(map);

    // google.maps.event.addListener(map, 'bounds_changed', fetch_weather);

    toggleLanding()
    toggleParking()
}

function toggleLanding() {
    var vis = $('#toggleLanding').prop('checked');
    var s, site
    for (s in sites) {
        site = sites[s];
        var i;
        for (i = 0; i < site.landings.length; i++) {
            site.landings[i].marker.setMap(vis ? map : null);
        }
    }
}

function toggleParking() {
    var vis = $('#toggleParking').prop('checked');
    var s, site
    for (s in sites) {
        site = sites[s];
        var i;
        for (i = 0; i < site.parkings.length; i++) {
            site.parkings[i].marker.setMap(vis ? map : null);
        }
    }
}

function metTitle(obs) {
    var title = obs.Name + "(" + obs.WindSpeed + "mph";
    if (obs.WindDirection != "") {
        title += " from " + obs.WindDirection
    }
    title += ")"
    if (obs.Weather != "") {
        title += " : " + obs.Weather
    }
    return title
}

function fetch_weather() {
    var vis = $('#toggleWeather').prop('checked');
    console.log("fetch_weather: vis=" + vis);

    if (!vis) {
        for (const ID in weather) {
            weather[ID].setMap(null);
        }
        weather = {}
        return;
    }

    $.ajax({
        url: "/weather/",
        dataType: 'JSON',
        success: function(w){
            console.log("Got weather", w)
            for (const obs of w) {
                if (weather[obs.ID]) {
                    weather[obs.ID].setIcon(weatherImage(obs.WindSpeed,obs.WindDirection));
                    weather[obs.ID].setTitle(metTitle(obs));
                } else {
                    weather[obs.ID] = new google.maps.Marker({
                        position: new google.maps.LatLng(obs.Lat, obs.Lon),
                        map: map,
                        title: metTitle(obs),
                        icon: weatherImage(obs.WindSpeed,obs.WindDirection),
                        draggable: false,
                    });
                    weather[obs.ID].addListener('click', () => {
                            infoWindow.setContent(
                                "<div>" +
                                "<h1>" + metTitle(obs) + "</h1>" +
                                "<table>" +
                                "<tr><th></th><td></td></tr>" +
                                "<tr><th>Station</th><td>" + obs.ID + " - " + obs.Name + "</td></tr>" +
                                "<tr><th>Weather</th><td>" + obs.WeatherId + " - " + obs.Weather + "</td></tr>" +
                                "<tr><th>Wind Speed</th><td>" + obs.WindSpeed + "</td></tr>" +
                                "<tr><th>Wind Direction</th><td>" + obs.WindDirection + "</td></tr>" +
                                "<tr><th>Gust</th><td>" + obs.Gust + "</td></tr>" +
                                "<tr><th>Temperature</th><td>" + obs.Temperature + "</td></tr>" +
                                "<tr><th>Elevation</th><td>" + obs.Elevation + "</td></tr>" +
                                "<tr><th>Pressure</th><td>" + obs.Pressure + "</td></tr>" +
                                "<tr><th>PressureTendency</th><td>" + obs.PressureTendency + "</td></tr>" +
                                "</table>" +
                                "</div>");
                            infoWindow.open(map, weather[obs.ID]);
                        });
                }
            }

            for (const ID in weather) {
                weather[ID].setMap(map);
            }
        },
        error: function (xhr,status,err) {
            console.log("Weather failed", status, err)
        }
    });
}
