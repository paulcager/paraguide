  <div id="menucontainer">
      <nav id="my-menu">
            <ul>
              <li><a href="/">Home</a></li>
              <li><span>Go to site...</span>
                <ul>
                    <#include "menu_sites.ftl">
                </ul>
              </li>
              <li><span>Guides</span>
    			<ul>
    				<#include "menu_guides.ftl">
    			</ul>
    		  </li>
              <li><span>Weather</span>
                <ul>
                    <span>Forecasts</span>
                    <li><a href="http://www.xcweather.co.uk/forecast/castleton_derbyshire">XCWeather</a></li>
                    <li><a href="http://www.raintoday.co.uk/mobile">Rain Today</a></li>
                    <li><a href="http://rasp.inn.leedsmet.ac.uk/">RASP</a></li>
                    <li><a href="https://www.wendywindblows.com/mobile/">Wendy</a></li>
                    <li><a href="http://www.windfinder.com/forecast/the_roaches">Windfinder</a></li>
                    <li><a href="http://earth.nullschool.net/#current/wind/surface/level/orthographic=-4.35,54.90,3000">nullschool</a></li>
                    <span>Webcams</span>
                    <li><a href="http://lindleyeducationaltrust.org/hollowford/weather/TSimage.jpg">MamCam</a></li>
                    <li><a href="http://xnet.hsl.gov.uk/siteweather/image00002.jpg">Buxton, Harpur Hill</a></li>
                    <li><a href="http://buxton.viewnetcam.com/CgiStart?page=Single&Language=0">Buxton</a></li>
                    <li><a href="http://www.maccinfo.com/cat/">Cat and Fiddle</a></li>
                    
                    
                </ul>
              </li>
              <li><span>Show Landing</span><input id="toggleLanding" type="checkbox" class="Toggle" checked="checked" onClick="javascript: toggleLanding()"/></li>
              <li><span>Show Parking</span><input id="toggleParking" type="checkbox" class="Toggle" checked="checked" onClick="javascript: toggleParking()"/></li>
              <li><span>Clubs</span>
                <ul>
                    <li><a href="http://derbyshiresoaringclub.com/smf/">Derbyshire Soaring Club</a></li>
                    <li><a href="http://www.airways-airsports.com/">Airways</a></li>
                    <li><a href="http://www.peaksoaring.co.uk//">Peak Soaring Association</a></li>
                </ul>
              </li>
    
              <li><a href="#">Contact</a></li>
            </ul>
      </nav>
      <a href="#my-menu"><img src="img/menu.png" alt="Menu"/></a>
  </div>
