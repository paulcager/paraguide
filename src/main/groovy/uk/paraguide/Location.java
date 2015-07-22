package uk.paraguide;

public class Location {
	public double lat;
	public double lon;
	//public String osGrid;
	
	
	
	public static Location fromString(String s) {
		if (s == null || s.length() == 0) {
			return null;
		}
		String[] parts = s.split(",");
		if (parts.length != 2) {
			throw new IllegalArgumentException(s);
		}

		return new Location(Double.parseDouble(parts[0]), Double.parseDouble(parts[1]));
	}
	
	public Location(double lat, double lon) {
		this.lat = lat;
		this.lon = lon;
	}

	@Override
	public String toString() {
		return lat + "," + lon;
	}
	
	
}