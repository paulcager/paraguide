package uk.paraguide;

import uk.me.jstott.jcoord.LatLng;
import uk.me.jstott.jcoord.OSRef;

public class Location {
	public final double lat;
	public final double lon;
	public final int easting;
	public final int northing;
	public final String osGrid;
	public final String latLon;
	
	public static Location fromString(String latLon) {
		if (latLon == null || latLon.length() == 0) {
			return null;
		}
		String[] parts = latLon.split(",");
		if (parts.length != 2) {
			throw new IllegalArgumentException(latLon);
		}

		double lat = Double.parseDouble(parts[0]);
		double lon = Double.parseDouble(parts[1]);
		
		LatLng osLatLon = new LatLng(lat, lon);
		osLatLon.toOSGB36();
		OSRef osRef = osLatLon.toOSRef();
		return new Location(osRef.toSixFigureString(), (int)osRef.getEasting(), (int)osRef.getNorthing(), latLon, lat, lon);
	}
	
	public Location(String osGrid, int easting, int northing, String latLon, double lat, double lon) {
		this.osGrid = osGrid;
		this.latLon = latLon;
		this.easting = easting;
		this.northing = northing;
		this.lat = lat;
		this.lon = lon;
	}

	public double getLat() {
		return lat;
	}

	public double getLon() {
		return lon;
	}

	public String getOsGrid() {
		return osGrid;
	}

	public String getLatLon() {
		return latLon;
	}
	
	public double getEasting() {
		return easting;
	}
	
	public double getNorthing() {
		return northing;
	}

	@Override
	public String toString() {
		return lat + "," + lon;
	}
}