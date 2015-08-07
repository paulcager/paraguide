package uk.paraguide

import java.nio.file.Paths

import uk.me.jstott.jcoord.OSRef

public class Main {
	private static final String spreadsheet = "https://spreadsheets.google.com/feeds/list/13blLictRsToqT7HReMA9IcUfHp3BzUPIhmgHadMmpW8/od6/public/full";
	private static final int SMALL_ICON_SIZE = 24;
	private static final int LARGE_ICON_SIZE = 64;
	
	public static void main(String[] args) {
		println "Loading data from ${spreadsheet}"
		Map<String, Site> sites = SiteXml.load(new URL(spreadsheet)).sites;
		println "${sites.size()} sites downloaded"
		println "${sites}"
		
		println "Expanding templates"
		Map<String, Object> model = Collections.singletonMap("sites", sites.values());
		TemplateProcessor processor = new TemplateProcessor(Paths.get("src/templates"), Paths.get("tmp/out"), model);
		processor.transform();
		
		println "Generating icons"
		for (Site site : sites.values()) {
			new SiteIconGen().createImage(SMALL_ICON_SIZE, new File("tmp/out/icons/small/" + site.id + ".png"), site.name, site.wind as String[]);
			new SiteIconGen().createImage(LARGE_ICON_SIZE, new File("tmp/out/icons/large/" + site.id + ".png"), site.name, site.wind as String[]);
		}
		
		println "Success!"
	}
}
