package uk.paraguide;

import java.io.IOException;
import java.net.URL;
import java.util.Map;

import javax.xml.parsers.ParserConfigurationException;

import org.xml.sax.SAXException;

@Deprecated
class SiteTable {
//	public static void main(String[] args) throws Exception {
//		SiteTable sites = fromUrl(new URL("https://spreadsheets.google.com/feeds/list/13blLictRsToqT7HReMA9IcUfHp3BzUPIhmgHadMmpW8/od6/public/full"));
//	}
	
	public static SiteTable fromUrl(URL url) throws IOException, SAXException, ParserConfigurationException {
		return new SiteTable(SiteXml.load(url).getSites());
	}
	
	private final Map<String, Site> sites;
	
	public SiteTable(Map<String, Site> sites) {
		this.sites = sites;
	}

	public Map<String, Site> getSites() {
		return sites;
	}
}
