<!DOCTYPE html>
<html>
  <head>
    <meta name="viewport" content="initial-scale=1.0, user-scalable=no">
    <meta charset="utf-8">
    <title>Para Sites</title>
    
    
    <style type="text/css">
    </style>
    
    <script src="https://cdnjs.cloudflare.com/ajax/libs/jquery/1.11.3/jquery.min.js"></script>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/jqueryui/1.12.1/jquery-ui.min.js"></script>
    <script src="js/jquery.mmenu.min.all.js" type="text/javascript"></script>
    
    <script type="text/javascript">
      $(document).ready(function() {
         $("#my-menu").mmenu({
            offCanvas : {
                position : "left",
                zposition : "front"
            },
            navbars: {
                content: [ "searchfield" ]
            },
            searchfield: {
                add: true,
                addTo: "panels",
                showSubPanels: true
            }
         })
      });
    </script>
    <script src="https://maps.googleapis.com/maps/api/js?sensor=false&libraries=geometry,places&ext=.js"></script>
    <script src="js/sites.js"></script>
    <script>
     google.maps.event.addDomListener(window, 'load', initialize);
    </script>
    
    <link href="css/jquery.mmenu.all.css" type="text/css" rel="stylesheet" />
    <link rel="stylesheet" href="//code.jquery.com/ui/1.11.4/themes/smoothness/jquery-ui.css">
    <link rel='stylesheet' href='css/paraguide.css' type='text/css' media='all' />
  </head>
  <body>
    <div id="outermost">
        <div id="map-canvas"></div>
      	<#include "menu.ftl">
      	<div id="url-dialog" title="Bookmark">
           <p></p>
        </div>
    </div>
  </body>
</html>
