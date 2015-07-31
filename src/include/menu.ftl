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
                <li><a href="http://www.xcweather.co.uk/forecast/castleton_derbyshire">XCWeather</a></li>
                <li><a href="http://www.raintoday.co.uk/mobile">Rain Today</a></li>
                <li><a href="http://rasp.inn.leedsmet.ac.uk/">RASP</a></li>
                <li><a href="https://www.wendywindblows.com/mobile/">Wendy</a></li>
                <li><a href="http://lindleyeducationaltrust.org/hollowford/weather/TSimage.jpg">MamCam</a></li>
            </ul>
          </li>
          <li><a href="#">Show Landing</a><input type="checkbox" class="Toggle" /></li>
          <li><a href="#">Show Parking</a><input type="checkbox" class="Toggle" checked="checked" /></li>
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
