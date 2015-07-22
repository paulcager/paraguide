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
  <div id="menucontainer">
  <nav id="my-menu">
        <ul>
          <li><a href="/">Home</a></li>
          <li><span>Sites</span>
		<ul>
			<li><a href="#">Test1</a></li>
			<li><a href="#">Test2</a></li>
		</ul>
	  </li>
          <li><a href="#">Show Landing</a><input type="checkbox" class="Toggle" /></li>
          <li><a href="#">Show Parking</a><input type="checkbox" class="Toggle" checked="checked" /></li>
          <li><a href="#">Contact</a></li>
        </ul>
  </nav>
  <a href="#my-menu"><img src="img/menu.png" alt="Menu"/></a>
  </div>
 
  </body>
</html>