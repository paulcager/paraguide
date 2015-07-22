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


public class SiteIconGen {
	public static void main(String[] args) throws Exception {
		for (String line : places) {
			String[] fields = line.split("\t");
			String place = fields[0];
			place = place.replaceAll("'", "_").replaceAll(" ", "_");
			String[] winds = fields[7].split(",");
			createImage(24, new File("/tmp/icons/small/" + place + ".png"), place, directions);
			createImage(64, new File("/tmp/icons/large/" + place + ".png"), place, directions);
		}
	}
	
	private void createImage(int size, File file, String place, String[] winds) throws IOException { 
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

	private static final String[] places = {
		//"Place	ParkingOSGrid	Parking	TakeoffOSGrid	Takeoff	LandingOSGrid	Landing	Wind",
		"Broadlee Bank	SK 117 854		SK 117 854	53.365791,-1.824888	SK 117 854	53.365791,-1.824888	SE-SSE",
		"Cat's Tor			SJ 994 759	53.280526,-2.009709	SJ 994 766	53.286818,-2.009710	W-NW",
		"Cocking Tor			SK 347 607	53.142771,-1.481949	SK 347 606	53.141872,-1.481960	NE",
		"Curbar Edge			SK 260 748	53.270005,-1.610863			SW-WSW",
		"Dale Head			SK 096 838	53.351451,-1.856494	SK 104 846	53.358627,-1.844450	NE-ESE",
		"Eyam Edge			SK 203 777	53.296320,-1.696143			S-SW",
		"Lord's Seat			SK 119 835	53.348708,-1.821954	SK 112 841	53.354116,-1.832448	NNW-NE",
		"Mam Tor NW			SK 127 836	53.349588,-1.809932	SK 139 850	53.362142,-1.791845	WNW-NNW",
		"Mam Tor E			SK 128 837	53.350485,-1.808426	SK 140 831	53.345062,-1.790426	NE-SE",
		"Treak Cliff			SK 134 829	53.343279,-1.799446	SK 140 831	53.345062,-1.790426	NNE-ENE",
		"Long Cliff			SK 138 825	53.339674,-1.793456	SK 140 831	53.345062,-1.790426	NNE-ENE",
		"Stanage Edge			SK 251 828	53.341956,-1.623726			SW",
		"Bradwell			SK 180 804	53.320672,-1.730498	SK 175 806	53.322486,-1.737992	WSW-NW",
		"Bunster			SK 140 515	53.061016,-1.791805			E, SE-SSW, W",
		"Chelmorton			SK 116 708	53.234559,-1.826921			NNW-N",
		"Back of Ecton			SK 101 581	53.120430,-1.849792			NE",
		"Edge Top			SK 055 657	53.188813,-1.918396			SSW-WSW",
		"High Edge			SK 063 689	53.217568,-1.906361			NE-E",
		"High Wheeldon			SK 101 660	53.191443,-1.849545			SW-WSW",
		"Shining Tor			SJ 995 737	53.260751,-2.008205			SW-WNW",
		"Wetton Hill			SK 114 563	53.104224,-1.830431			W-N"
	};
	
	private static final Map<String, Double> compass;
	static {
		compass = new HashMap<>();
		String[] points = "NNE,NE,ENE,E,ESE,SE,SSE,S,SSW,SW,WSW,W,WNW,NW,NNW,N".split(",");
		for (int i = 0; i < points.length; i++) {
			compass.put(points[i], 11.25 + i*22.5);
		}
	}
}