package uk.paraguide

import uk.me.jstott.jcoord.LatLng;
import groovy.util.slurpersupport.GPathResult

class SiteXml {
	private final GPathResult xml;
	
	private SiteXml(GPathResult xml) {
		this.xml = xml;
	}

	public static SiteXml load(URL url) {
		String text = url.text;
		GPathResult xml = new XmlSlurper().parseText(text).declareNamespace(
				"gsx": "http://schemas.google.com/spreadsheets/2006/extended"
		);
		return new SiteXml(xml);
	}
	
	public Map<String, Site> getSites() {
		Map<String, Site> sites = new TreeMap<>();
		for (def entry : xml.entry) {
			Site site = decodeSiteEntry(entry);
			sites.put(site.id, site);
		}
		return sites;
	} 
	
	private Site decodeSiteEntry(xmlEntry) {
//		println xmlEntry.'gsx:place'.toString()
		return new Site(
			xmlEntry.'gsx:place'.toString(),
			location(xmlEntry.'gsx:parking'),
			location(xmlEntry.'gsx:takeoff'),
			location(xmlEntry.'gsx:landing'),
			wind(xmlEntry.'gsx:wind')
		);
	}
	
	private static List<Location> location(latLngs) {
		latLngs.toString().split("\\s+").findAll{it.length() > 0}.collect{Location.fromString(it)}
	}
	
	private static List<String> wind(value) {
		value.toString().split("[\\s,]+").findAll{it.length() > 0} as List<String>
	}

}
