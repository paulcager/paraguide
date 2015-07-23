package uk.paraguide;

public class Location {
	public final double lat;
	public final double lon;
	public final String osGrid;
	public final String latLng;
	
	public static Location fromString(String latLng) {
		return fromStrings("", latLng);
	}
	
	public static Location fromStrings(String osGrid, String latLng) {
		if (latLng == null || latLng.length() == 0) {
			return null;
		}
		String[] parts = latLng.split(",");
		if (parts.length != 2) {
			throw new IllegalArgumentException(latLng);
		}

		return new Location(osGrid, latLng, Double.parseDouble(parts[0]), Double.parseDouble(parts[1]));
	}
	
	public Location(String osGrid, String latLng, double lat, double lon) {
		this.osGrid = osGrid;
		this.latLng = latLng;
		this.lat = lat;
		this.lon = lon;
	}

	@Override
	public String toString() {
		return lat + "," + lon;
	}
	
	
}