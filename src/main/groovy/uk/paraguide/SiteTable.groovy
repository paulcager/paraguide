package uk.paraguide

import java.util.Collection;

class SiteTable {
	public static void main(String[] args) {
		SiteTable sites = fromUrl(new URL("https://spreadsheets.google.com/feeds/list/13blLictRsToqT7HReMA9IcUfHp3BzUPIhmgHadMmpW8/od6/public/full"));
	}
	
	public static SiteTable fromUrl(URL url) {
		String text = url.getText();
		def xml = new XmlSlurper().parseText(text).declareNamespace(gsx:'http://schemas.google.com/spreadsheets/2006/extended')
//		println xml.author.name;
		List<Site> sites = new ArrayList<Site>();
		for (def entry : xml.entry) {
			sites.add(Site.fromXml(entry));
		}
		return new SiteTable(sites);
	}
	
	private final Collection<Site> sites;
	
	public SiteTable(Collection<Site> sites) {
		this.sites = sites;
	}

	public Collection<Site> getSites() {
		return sites;
	}
}
