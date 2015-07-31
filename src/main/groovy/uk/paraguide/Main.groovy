package uk.paraguide

import java.nio.file.Paths;
import java.util.ArrayList;
import java.util.Collections;
import java.util.Map;

import freemarker.template.Configuration
import freemarker.template.Template
import freemarker.template.TemplateExceptionHandler

public class Main {
	private static final String spreadsheet = "https://spreadsheets.google.com/feeds/list/13blLictRsToqT7HReMA9IcUfHp3BzUPIhmgHadMmpW8/od6/public/full";
	public static void main(String[] args) {
		Map<String, Site> sites = SiteXml.load(new URL(spreadsheet)).sites;
		println "Sites: ${sites}"
		Map<String, Object> model = Collections.singletonMap("sites", sites.values());
		TemplateProcessor processor = new TemplateProcessor(Paths.get("src/templates"), Paths.get("tmp/out"), model);
		processor.transform();
	}
}
