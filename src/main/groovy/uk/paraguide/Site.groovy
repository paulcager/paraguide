package uk.paraguide;

import java.util.List;


public class Site {
	public final String name;
	public final List<Location> parking;
	public final List<Location> takeoff;
	public final List<Location> landing;
	public final List<String> wind;
	
	public Site(String name, List<Location> parking, List<Location> takeoff,
			List<Location> landing, List<String> wind) {
		this.name = name;
		this.parking = parking;
		this.takeoff = takeoff;
		this.landing = landing;
		this.wind = wind;
	}
	
	public String getName() {
		return name;
	}

	public List<Location> getParking() {
		return parking;
	}

	public List<Location> getTakeoff() {
		return takeoff;
	}

	public List<Location> getLanding() {
		return landing;
	}

	public List<String> getWind() {
		return wind;
	}

	public static Site fromXml(xmlEntry) {
		return new Site(
			xmlEntry.'gsx:place'.toString(),
			location(xmlEntry.'gsx:parkingosgrid', xmlEntry.'gsx:parking'),
			location(xmlEntry.'gsx:takeoffosgrid', xmlEntry.'gsx:takeoff'),
			location(xmlEntry.'gsx:landingosgrid', xmlEntry.'gsx:landing'),
			wind(xmlEntry.'gsx:wind')
		);
	}
	
	@Override
	public String toString() {
		return "Site [name=" + name + ", parking=" + parking + ", takeoff=" +
				takeoff + ", landing=" + landing + ", wind=" + wind + "]";
	}

	private static List<Location> location(osGrid, latLngs) {
		latLngs.toString().split("  *").findAll{it.length() > 0}.collect{Location.fromStrings(osGrid.toString(), it)}
	}
	
	private static List<String> wind(value) {
		value.toString().split(", *").findAll{it.length() > 0} as List<String>
	}

}
