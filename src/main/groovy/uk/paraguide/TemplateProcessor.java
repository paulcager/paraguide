package uk.paraguide;

import java.io.IOException;
import java.io.PrintWriter;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.nio.file.StandardCopyOption;
import java.util.ArrayList;
import java.util.Collections;
import java.util.Map;

import freemarker.cache.FileTemplateLoader;
import freemarker.cache.MultiTemplateLoader;
import freemarker.cache.TemplateLoader;
import freemarker.template.Configuration;
import freemarker.template.Template;
import freemarker.template.TemplateException;
import freemarker.template.TemplateExceptionHandler;

class TemplateProcessor {
	public static void main(String[] args) throws Exception {
		Map<String, Object> model = Collections.singletonMap("sites", new ArrayList<Object>());
		new TemplateProcessor(Paths.get("src/templates"), Paths.get("tmp/out"), model).transform();
	}
	
	private final Path templateDirectory;
	private final Path includeDirectory;
	private final Path outputDirectory;
	private final Object model;
	
	private final Configuration cfg;
	
	public TemplateProcessor(Path templateDirectory, Path outputDirectory, Object model) throws IOException {
		this.templateDirectory = templateDirectory;
		this.outputDirectory = outputDirectory;
		this.model = model;
		this.includeDirectory = templateDirectory.resolveSibling("include");
		this.cfg = new Configuration(Configuration.VERSION_2_3_22);
		//cfg.setDirectoryForTemplateLoading(templateDirectory.toFile());
		cfg.setTemplateLoader(new MultiTemplateLoader(new TemplateLoader[] {
				new FileTemplateLoader(templateDirectory.toFile()),
				new FileTemplateLoader(includeDirectory.toFile())
		}));
		cfg.setDefaultEncoding("UTF-8");
		cfg.setTemplateExceptionHandler(TemplateExceptionHandler.RETHROW_HANDLER);
		
	}

	public void transform() throws IOException, TemplateException {
		for (Path p : Files.walk(templateDirectory).toArray(Path[]::new)) {
			expandFile(p);
		}
	}
	
	private void expandFile(Path p) throws IOException, TemplateException {
		Path output = outputDirectory.resolve(templateDirectory.relativize(p));
		if (p.toFile().isDirectory()) {
			output.toFile().mkdirs();
		} else if (p.toString().endsWith(".ftl")) {
			Path renamedOutput = Paths.get(output.toString().replaceFirst("\\.ftl$", "")); 
			Template temp = cfg.getTemplate(templateDirectory.relativize(p).toString());
			try (PrintWriter pw = new PrintWriter(Files.newBufferedWriter(renamedOutput))) {
				temp.process(model, new PrintWriter(pw));
			}
		} else {
			Files.copy(p, output, StandardCopyOption.REPLACE_EXISTING, StandardCopyOption.COPY_ATTRIBUTES);
		}
	}
}
