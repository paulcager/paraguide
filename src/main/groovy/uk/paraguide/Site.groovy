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
	
	public static Site fromXml(xmlEntry) {
		return new Site(
			xmlEntry.'gsx:place'.toString(),
			list(xmlEntry.'gsx:parking'),
			list(xmlEntry.'gsx:takeoff'),
			list(xmlEntry.'gsx:landing'),
			wind(xmlEntry.'gsx:wind')
		);
	}
	
	@Override
	public String toString() {
		return "Site [name=" + name + ", parking=" + parking + ", takeoff=" +
				takeoff + ", landing=" + landing + ", wind=" + wind + "]";
	}

	private static List<Location> list(value) {
		value.toString().split("  *").findAll{it.length() > 0}.collect{Location.fromString(it)}
	}
	
	private static List<String> wind(value) {
		value.toString().split(", *").findAll{it.length() > 0} as List<String>
	}

}
