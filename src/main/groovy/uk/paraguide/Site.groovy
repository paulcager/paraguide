package uk.paraguide;

import java.util.List;


public class Site {
	public final String name;
	public final String id;
	public final String club;
	public final List<Location> parking;
	public final List<Location> takeoff;
	public final List<Location> landing;
	public final List<String> wind;
	
	public Site(String club, String name, List<Location> parking, List<Location> takeoff,
			List<Location> landing, List<String> wind) {
		this.club = club;
		this.name = name;
		this.id = name.replaceAll("[^A-Za-z0-9]", "_");
		this.parking = parking;
		this.takeoff = takeoff;
		this.landing = landing;
		this.wind = wind;
	}
	
	public String getClub() {
		return club;
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
	
	public String getId() {
		return id;
	}

	@Override
	public String toString() {
		return "Site [name=" + name + ", parking=" + parking + ", takeoff=" +
				takeoff + ", landing=" + landing + ", wind=" + wind + "]";
	}


}
