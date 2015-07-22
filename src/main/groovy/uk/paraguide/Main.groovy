package uk.paraguide

public class Main {
	private static final String spreadsheet = "https://spreadsheets.google.com/feeds/list/13blLictRsToqT7HReMA9IcUfHp3BzUPIhmgHadMmpW8/od6/public/full";
	public static void main(String[] args) {
		SiteTable sites = SiteTable.fromUrl(new URL(spreadsheet));
	}

}
