var sites = {}
var landColor = "20e000";
var pinImage = new google.maps.MarkerImage(
    "http://chart.apis.google.com/chart?chst=d_map_pin_letter&chld=L|" + landColor,
    new google.maps.Size(21, 34),
    new google.maps.Point(0,0),
    new google.maps.Point(10, 34)
);

var parkingImage = new google.maps.MarkerImage(
    "img/parking.png",
    new google.maps.Size(21, 34),
    new google.maps.Point(0,0),
    new google.maps.Point(10, 34)
);

var infoWindow = new google.maps.InfoWindow({ });

function create_sites() {
    var sites = {};
    <#list sites as site>
    
    sites["${site.id}"] = create_site(
        "${site.id}",
        "${site.name}", 
        [<#list site.parking as p>{easting: "${p.easting?c}", northing: "${p.northing?c}", lat: ${p.lat?c}, lon: ${p.lon?c} }<#sep>, </#list>],
        [<#list site.takeoff as t>{easting: "${t.easting?c}", northing: "${t.northing?c}", lat: ${t.lat?c}, lon: ${t.lon?c} }<#sep>, </#list>],
        [<#list site.landing as l>{easting: "${l.easting?c}", northing: "${l.northing?c}", lat: ${l.lat?c}, lon: ${l.lon?c} }<#sep>, </#list>],  
        [<#list site.wind as w>"${w}"<#sep>, </#list>]
    );    
    </#list>
    return sites;
}

function create_site(id, name, parkings, takeoffs, landings, wind) {
      var icon = icon_url("small", name);
      site = {
        id          : id,
        name        : name,
        parkings    : create_markers(id, name, parkings, parkingImage), 
        takeoffs    : create_markers(id, name, takeoffs, {url: icon, anchor: new google.maps.Point(12, 12)}),
        landings    : create_markers(id, name, landings, pinImage),
      }
      
      site.info = site.takeoffs[0].info;
      
      return site;
  }
  
function show_url(id) {
    var url = "#" + id;
    $( '#url-dialog' ).html(
        "<p>Copy the following link to share.</p>" +
        "<a href='" + url + "'>Direct Link</a>"
    );
    $( '#url-dialog' ).dialog();
  }

function add_info_window(place) {
    var name = place.name;
    var id = place.id;
    var marker = place.marker;
    var icon = icon_url("large", name);
    var lat = place.lat;
    var lon = place.lon;
    var osGrid = place.osGrid;
    var f = function() {
        infoWindow.setContent(
            "<div>" +
              "<h1>" + name + "</h1>" +
                  "<div style=\"padding: 0.5em;\">" +
                      "<img src=\"" + icon + "\"/>" +
                  "</div>" +
                  "<p>" +
                    "<a href=\"guides/" + id + ".pdf\" target=\"guide\">Guide</a><br/>" +
                    "<a href=\"" + maps_url(lat, lon, 9, false) + "\">Directions</a><br/>" +
                    "<a href=\"" + maps_url(lat, lon, 15, true) + "\">Satellite View</a><br/>" +
                    <!-- "<a href=\"http://www.streetmap.co.uk/ids.srf?mapp=map.srf&searchp=ids&name=" + lat + "," + lon + "&type=LatLong\">OS Map</a><br/>" + -->
                    "<a href=\"http://www.streetmap.co.uk/grid/" + place.easting + "," + place.northing + "_115\">OS Map</a><br/>" +
                    "<a href=\"javascript:show_url('" + id + "')\">Bookmark Site</a><br/>" +
                  "</p>" +
            "</div>");
        infoWindow.open(map,marker);
      };
      google.maps.event.addListener(marker, 'click', f);
      place.info = f;
  }
  
  function create_markers(id, name, places, icon) {
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
        add_info_window(t); 
    }
    return places;
  }
  
  function create_landings(id, place, landings) {
    var i;
    for (i = 0; i < landings.length; i++) {
        var l = landings[i];
        var icon = icon_url("small", place);
        var marker = new google.maps.Marker({
            position: new google.maps.LatLng(l.lat, l.lon),
            map: map,
            title: place,
            icon: pinImage,
            draggable: false,
        });
        l.marker = marker;
    }
    return landings;
  }

  function create_parkings(id, place, landings) {
    var i;
    for (i = 0; i < landings.length; i++) {
        var l = landings[i];
        var icon = icon_url("small", place);
        var marker = new google.maps.Marker({
            position: new google.maps.LatLng(l.lat, l.lon),
            map: map,
            title: place,
            icon: parkingImage,
            draggable: false,
        });
        l.marker = marker;
    }
    return landings;
  }


  function icon_url(type, place) {
      return "icons/" + type + "/" + safe_name(place) + ".png";
  }
  function safe_name(place) {
      return place.replace(/'/g, "_").replace(/ /g, "_");
  }

  function maps_url(lat, lng, zoom, satellite) {
    //return "https://www.google.co.uk/maps/place/" + String(lat) + "," + String(lng) + "/@" + String(lat) + "," + String(lng) + "," + String(zoom) + "z";
    //return "https://www.google.com/maps/embed/v1/place?key=AIzaSyDYEr0NL0JlKdlNchfiRmCPJVDL9bRqsZc&q=" + String(lat) + "," + String(lng) + "&zoom=" + String(zoom) + "z&maptype=" + (satellite ? "satellite" : "roadmap");
    //return "http://maps.google.com/maps?q=" + lat + "," + lng + "&t=(satellite ? "k" : "m")
    var t = (satellite ? "k" : "m")
    var latlon = lat + "," + lng;
    return "https://maps.google.com/?q=" + latlon + "&ll=" + latlon + "&t=" + t + "&z=" + zoom;
  }

  function initialize() {
    var myLatlng = new google.maps.LatLng(53.23455851981886,-1.8269212925847522);
    var mapOptions = {
      zoom: 10,
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