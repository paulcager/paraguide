package uk.paraguide;

public class Location {
	public final double lat;
	public final double lon;
	public final String osGrid;
	public final String latLon;
	
	public static Location fromString(String latLon) {
		return fromStrings("", latLon);
	}
	
	public static Location fromStrings(String osGrid, String latLon) {
		if (latLon == null || latLon.length() == 0) {
			return null;
		}
		String[] parts = latLon.split(",");
		if (parts.length != 2) {
			throw new IllegalArgumentException(latLon);
		}

		return new Location(osGrid, latLon, Double.parseDouble(parts[0]), Double.parseDouble(parts[1]));
	}
	
	public Location(String osGrid, String latLon, double lat, double lon) {
		this.osGrid = osGrid;
		this.latLon = latLon;
		this.lat = lat;
		this.lon = lon;
	}

	@Override
	public String toString() {
		return lat + "," + lon;
	}
}