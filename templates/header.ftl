<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="utf-8"/>
    <meta name="viewport" content="initial-scale=1.0, user-scalable=no">
    <meta name="description" content="">
    <meta name="author" content="mail@paraguide.uk">
    <meta name="keywords" content="">
    <meta name="generator" content="JBake">
    <meta charset="utf-8">

    <!-- HTML5 shim, for IE6-8 support of HTML5 elements -->
    <!--[if lt IE 9]>
      <script src="<#if (content.rootpath)??>${content.rootpath}<#else></#if>js/html5shiv.min.js"></script>
    <![endif]-->

    <title>Para Sites</title>
    <style>
      html, body, #map-canvas {
        height: 100%;
        margin: 0px;
        padding: 0px
      }
    </style>
    <script src="https://maps.googleapis.com/maps/api/js?v=3.exp&signed_in=true"></script>
    <script src="js/sites.js"></script>
    <script>
      var landColor = "20e000";
      var pinImage = new google.maps.MarkerImage(
          "http://chart.apis.google.com/chart?chst=d_map_pin_letter&chld=L|" + landColor,
          new google.maps.Size(21, 34),
          new google.maps.Point(0,0),
          new google.maps.Point(10, 34));

      var infoWindow = new google.maps.InfoWindow({ });

      function create_site(fields) {
		var i = 0;
		var place = fields[i++];
		var parking = fields[i++];
		var takeoff = fields[i++];
		var landing = fields[i++];
		var wind = fields[i++];

		// For time being, just take the first element of the list.
		// Extend later.
      	var takeoff = create_takeoff(place, takeoff.lat, takeoff.lng);
      	create_info(place, takeoff, takeoff.lat, takeoff.lng);

      	return {
          takeoff: takeoff,
          landing: create_landing(place, landing.lat, landing.lon)
      	};
      }

      function create_info(place, takeoff, takeoffLat, takeoffLon) {
    	var icon = icon_url("large", place);
    	google.maps.event.addListener(takeoff, 'click', function() {
       	 infoWindow.setContent(
        	"<div>" +
          	"<h1>" + place + "</h1>" +
                  "<div style=\"padding: 0.5em;\">" +
                      "<img src=\"" + icon + "\"/>" +
                  "</div>" +
                  "<p>" +
                    "<a href=\"guides/" + safe_name(place) + ".pdf\" target=\"guide\">Guide</a><br/>" +
            "<a href=\"" + maps_url(takeoffLat, takeoffLon, 9, false) + "\">Directions</a><br/>" +
            "<a href=\"" + maps_url(takeoffLat, takeoffLon, 15, true) + "\">Satellite View</a>" +
                  "</p>" +
        	"</div>");
         infoWindow.open(map,takeoff);
      	});
      }
      function create_takeoff(place, lat, lng) {
    	var icon = icon_url("small", place);
    	return new google.maps.Marker({
	        position: new google.maps.LatLng(lat, lng),
    	    map: map,
        	title: place,
        	icon: {url: icon}
       	});
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
    	return "https://www.google.com/maps/embed/v1/place?key=AIzaSyDYEr0NL0JlKdlNchfiRmCPJVDL9bRqsZc&q=" + String(lat) + "," + String(lng) + "&zoom=" + String(zoom) + "z&maptype=" + (satellite ? "satellite" : "roadmap");
      }

      function initialize() {
        var myLatlng = new google.maps.LatLng(53.23455851981886,-1.8269212925847522);
        var mapOptions = {
          zoom: 10,
          center: myLatlng,
        }
        map = new google.maps.Map(document.getElementById('map-canvas'), mapOptions);
        sites = create_sites();
     }
    </script>
    <script>
     google.maps.event.addDomListener(window, 'load', initialize);
    </script>
  </head>
  <body>
