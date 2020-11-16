FROM golang:1.15.5 as build
WORKDIR /go/src/app

RUN go get -v \
    github.com/llgcode/draw2d/draw2dimg \
    google.golang.org/api/option \
    google.golang.org/api/sheets/v4

COPY . .
RUN go build -v -o /go/bin/paraguide

####################################################################################################


FROM debian:stable-slim
RUN apt-get update && apt-get -y upgrade && apt-get install -y ca-certificates
WORKDIR /app
COPY --from=build /go/bin/paraguide .
COPY templates/ templates/
COPY static/    static/
CMD ["/app/paraguide"]

