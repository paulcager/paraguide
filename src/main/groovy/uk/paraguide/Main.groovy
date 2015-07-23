package uk.paraguide

public class Main {
	private static final String spreadsheet = "https://spreadsheets.google.com/feeds/list/13blLictRsToqT7HReMA9IcUfHp3BzUPIhmgHadMmpW8/od6/public/full";
	public static void main(String[] args) {
		SiteTable sites = SiteTable.fromUrl(new URL(spreadsheet));
//		try (PrintWriter pw = new PrintWriter()) {
		PrintWriter pw = new PrintWriter(new FileWriter("assets/js/sites.js"));
			createSitesJs(sites, pw);
			pw.close();
//		}
	}
	
	public static void createSitesJs(SiteTable sites, PrintWriter pw) {
		pw.println("function create_sites() {");
		pw.println("    var sites = {};");
		pw.println();

		for (Site s : sites.getSites()) {
			// sites["Broadlee Bank"] = create_site(["Broadlee Bank", "SK 117 854", 53.365791,-1.824888, "SK 117 854", 53.365791,-1.824888, "SK 117 854", 53.365791,-1.824888, "SE"]);
			String parkings = "[" + s.parking.collect{it -> """{ osGrid: "${it.osGrid}", latLng: "${it.latLon}"}"""}.join(", ") + "]";
			String takeoffs = "[" + s.takeoff.collect{it -> """{ osGrid: "${it.osGrid}", latLng: "${it.latLon}"}"""}.join(", ") + "]";
			String landings = "[" + s.landing.collect{it -> """{ osGrid: "${it.osGrid}", latLng: "${it.latLon}"}"""}.join(", ") + "]";
			String winds =    "[" + s.wind.collect{it -> "\"$it\""}.join(", ") + "]";
			pw.println """    sites["${s.name}"] = create_site(["${s.name}", ${parkings}, ${takeoffs}, ${landings}, ${winds}]);"""
		}
		
		pw.println();
		pw.println "    return sites;"
		pw.println("}");
		pw.flush();
	}
}
