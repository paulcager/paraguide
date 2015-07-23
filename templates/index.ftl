<html>
  <head>
    <meta name="viewport" content="initial-scale=1.0, user-scalable=no">
    <meta charset="utf-8">
    <title>Para Sites</title>
    <script src="js/jquery-1.11.3.min.js" type="text/javascript"></script>
    <script src="js/jquery.mmenu.min.js" type="text/javascript"></script>
    <script src="js/jquery.mmenu.toggles.min.js" type="text/javascript"></script>
    <script type="text/javascript">
      $(document).ready(function() {
         $("#my-menu").mmenu();
      });
    </script>
    <link href="css/jquery.mmenu.css" type="text/css" rel="stylesheet" />
    <link href="css/jquery.mmenu.toggles.css" type="text/css" rel="stylesheet" />
    <link rel='stylesheet' href='css/paraguide.css' type='text/css' media='all' />
  </head>
  <body>
  	<iframe  id="embedded-map" width="100%" height="98%" src="index-frame.html"> </iframe>
  	<#include "menu.ftl">
  </body>
</html>
