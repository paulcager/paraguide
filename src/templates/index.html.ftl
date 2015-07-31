<html>
  <head>
    <meta name="viewport" content="initial-scale=1.0, user-scalable=no">
    <meta charset="utf-8">
    <title>Para Sites</title>
    <script src="js/jquery-1.11.3.min.js" type="text/javascript"></script>
    <script src="js/jquery-ui.min.js" type="text/javascript"></script>
    <script src="js/jquery.mmenu.min.js" type="text/javascript"></script>
    <script src="js/jquery.mmenu.toggles.min.js" type="text/javascript"></script>
    <script type="text/javascript">
      $(document).ready(function() {
         $("#my-menu").mmenu();
      });
    </script>
    <script src="https://maps.googleapis.com/maps/api/js?v=3.exp&signed_in=true"></script>
    <script src="js/sites.js"></script>
    <script>
     google.maps.event.addDomListener(window, 'load', initialize);
    </script>
    
    <link href="css/jquery.mmenu.css" type="text/css" rel="stylesheet" />
    <link href="css/jquery.mmenu.toggles.css" type="text/css" rel="stylesheet" />
    <link rel="stylesheet" href="//code.jquery.com/ui/1.11.4/themes/smoothness/jquery-ui.css">
    <link rel='stylesheet' href='css/paraguide.css' type='text/css' media='all' />
  </head>
  <body>
    <div id="map-canvas"></div>
  	<#include "menu.ftl">
  	<div id="url-dialog" title="Bookmark">
       <p></p>
    </div>
  </body>
</html>