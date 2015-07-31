function create_sites() {
    var sites = {};
    <#list sites as site>
    
    sites["${site.id}"] = create_site(
        "${site.id}",
        "${site.name}", 
        [<#list site.parking as p>{osGrid: "${p.osGrid}", lat: ${p.lat?c}, lon: ${p.lon?c} }<#sep>, </#list>],
        [<#list site.takeoff as t>{osGrid: "${t.osGrid}", lat: ${t.lat?c}, lon: ${t.lon?c} }<#sep>, </#list>],
        [<#list site.landing as l>{osGrid: "${l.osGrid}", lat: ${l.lat?c}, lon: ${l.lon?c} }<#sep>, </#list>],  
        [<#list site.wind as w>"${w}"<#sep>, </#list>]
    );    
    </#list>
    return sites;
}

  var landColor = "20e000";
  var pinImage = new google.maps.MarkerImage(
      "http://chart.apis.google.com/chart?chst=d_map_pin_letter&chld=L|" + landColor,
      new google.maps.Size(21, 34),
      new google.maps.Point(0,0),
      new google.maps.Point(10, 34));

  var infoWindow = new google.maps.InfoWindow({ });

  function create_site(id, place, parking, takeoff, landing, wind) {

    // For time being, just take the first element of the list.
    // Extend later.
      var markers = create_takeoff(place, takeoff);
      var t = takeoff[0];
      var info = create_info(id, place, markers[0], t.lat, t.lon);
      return {
          takeoff: markers,
          landing: create_landing(place, landing.lat, landing.lon),
          info: info
      };
  }
  
  function show_url(id) {
    var url = "#" + id;
    $( '#url-dialog' ).html(
        "<p>Copy the following link to share.</p>" +
        "<a href='" + url + "'>Direct Link</a>"
    );
    $( '#url-dialog' ).dialog();
  }

  function create_info(id, place, marker, takeoffLat, takeoffLon) {
    var icon = icon_url("large", place);
    var f = function() {
        infoWindow.setContent(
            "<div>" +
              "<h1>" + place + "</h1>" +
                  "<div style=\"padding: 0.5em;\">" +
                      "<img src=\"" + icon + "\"/>" +
                  "</div>" +
                  "<p>" +
                    "<a href=\"guides/" + safe_name(place) + ".pdf\" target=\"guide\">Guide</a><br/>" +
                    "<a href=\"" + maps_url(takeoffLat, takeoffLon, 9, false) + "\">Directions</a><br/>" +
                    "<a href=\"" + maps_url(takeoffLat, takeoffLon, 15, true) + "\">Satellite View</a><br/>" +
                    "<a href=\"javascript:show_url('" + id + "')\">Url</a><br/>" +
                  "</p>" +
            "</div>");
        infoWindow.open(map,marker);
      };
      google.maps.event.addListener(marker, 'click', f);
      return f;
  }
  
  function create_takeoff(place, takeoff) {
    var markers = [];
    var i;
    for (i = 0; i < takeoff.length; i++) {
        var t = takeoff[i];
        var icon = icon_url("small", place);
        var marker = new google.maps.Marker({
            position: new google.maps.LatLng(t.lat, t.lon),
            map: map,
            title: place,
            icon: {url: icon}
        });
        markers.push(marker);
    }
    return markers;
  }
  
  function create_landing(place, lat, lng) {
    return new google.maps.Marker({
        position: new google.maps.LatLng(lat, lng),
        map: map,
        icon: pinImage,
        title: place
      });
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
    console.log("Frag: " + window.location.hash.substring(1));
    if (sites[window.location.hash.substring(1)]) {
        sites[window.location.hash.substring(1)].info();
    }
 }
