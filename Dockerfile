FROM paulcager/go-base:latest as build
WORKDIR /go/src/

RUN go get -v \
    github.com/llgcode/draw2d/draw2dimg \
    github.com/arl/go-rquad \
    github.com/golang/geo/r2 \
    github.com/kellydunn/golang-geo

COPY . /go/src/github.com/paulcager/paraguide
RUN cd /go/src/github.com/paulcager/paraguide && go build -v -o /go/bin/paraguide && go test ./...

####################################################################################################


FROM debian:stable-slim
RUN apt-get update && apt-get -y upgrade && apt-get install -y ca-certificates
WORKDIR /app
COPY --from=build /go/bin/paraguide .
COPY templates/ templates/
COPY static/    static/
COPY fonts/     fonts/
EXPOSE 80
CMD ["/app/paraguide", "--port", ":80" ]

