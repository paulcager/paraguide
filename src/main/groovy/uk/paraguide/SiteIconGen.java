package uk.paraguide;
import java.awt.Color;
import java.awt.Graphics2D;
import java.awt.geom.AffineTransform;
import java.awt.geom.Arc2D;
import java.awt.geom.Ellipse2D;
import java.awt.image.BufferedImage;
import java.io.File;
import java.io.IOException;
import java.util.Arrays;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.stream.Collectors;

import javax.imageio.ImageIO;


class SiteIconGen {
	
	void createImage(int size, File file, String place, String[] winds) throws IOException { 
		List<Double[]> directions = Arrays.stream(winds)
				.map(x -> x.trim())
				.map(x -> x.split("-"))
				.map(x -> windDirectionToCompass(x))
				.collect(Collectors.toList());
		BufferedImage bi = new BufferedImage(size, size, BufferedImage.TYPE_INT_ARGB);
		Graphics2D g = bi.createGraphics();

		AffineTransform transform = AffineTransform.getQuadrantRotateInstance(-1, size / 2, size / 2);
		transform.scale(1, -1);
		transform.translate(0, -size);
		g.setTransform(transform);

		g.setBackground(new Color(0, 0, 0, 0));
		g.clearRect(0, 0, size, size);

		g.setPaint(Color.RED);
		g.fill(new Ellipse2D.Double(0, 0, size, size));

		g.setPaint(Color.GREEN);
		for (Double[] coords : directions) {
			Arc2D arc = new Arc2D.Double(0, 0, size, size, coords[0], coords[1] - coords[0], Arc2D.PIE);
			g.fill(arc);
		}

		g.dispose();
		file.getParentFile().mkdirs();
		ImageIO.write(bi, "png", file);
	}
	
	private final Double[] windDirectionToCompass(String[] wind) {
		double start;
		double end;
		if (wind.length == 1) {
			start = compass.get(wind[0]);
			end = start + 22.5;
		} else {
			start = compass.get(wind[0]);
			end = compass.get(wind[1]);
		}
		if (end < start) end += 360; 
		return new Double[] {start, end};
	}

	private static final Map<String, Double> compass;
	static {
		compass = new HashMap<>();
		String[] points = "NNE,NE,ENE,E,ESE,SE,SSE,S,SSW,SW,WSW,W,WNW,NW,NNW,N".split(",");
		for (int i = 0; i < points.length; i++) {
			compass.put(points[i], 11.25 + i*22.5);
		}
	}
}