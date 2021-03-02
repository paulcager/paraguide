module github.com/paulcager/paraguide

go 1.15

require (
	github.com/PuerkitoBio/goquery v1.6.0
	github.com/kr/pretty v0.2.1
	github.com/llgcode/draw2d v0.0.0-20200930101115-bfaf5d914d1e
	github.com/paulcager/gb-airspace v1.0.1
	github.com/paulcager/go-http-middleware v0.0.2
	github.com/paulcager/osgridref v1.2.0
	github.com/spf13/pflag v1.0.5
	google.golang.org/api v0.35.0
)

replace (
	github.com/paulcager/gb-airspace => ../gb-airspace
)