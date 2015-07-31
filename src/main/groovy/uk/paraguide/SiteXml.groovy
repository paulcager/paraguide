package uk.paraguide

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
		return new Site(
			xmlEntry.'gsx:place'.toString(),
			location(xmlEntry.'gsx:parkingosgrid', xmlEntry.'gsx:parking'),
			location(xmlEntry.'gsx:takeoffosgrid', xmlEntry.'gsx:takeoff'),
			location(xmlEntry.'gsx:landingosgrid', xmlEntry.'gsx:landing'),
			wind(xmlEntry.'gsx:wind')
		);
	}
	
	private static List<Location> location(osGrid, latLngs) {
		latLngs.toString().split("  *").findAll{it.length() > 0}.collect{Location.fromStrings(osGrid.toString(), it)}
	}
	
	private static List<String> wind(value) {
		value.toString().split(", *").findAll{it.length() > 0} as List<String>
	}

}
