FROM paulcager/go-base:latest as build
WORKDIR /go/src/

COPY . /go/src/github.com/paulcager/paraguide
RUN cd /go/src/github.com/paulcager/paraguide && CGO_ENABLED=0 go build -v -o /go/bin/paraguide ./cmd/... && go test ./...

####################################################################################################


FROM scratch
# To include diagnostic tools, use:
	#FROM debian:stable-slim
	#RUN apt-get update && apt-get -y upgrade && apt-get install -y ca-certificates
WORKDIR /app
COPY --from=build /go/bin/paraguide .
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY templates/ templates/
COPY static/    static/
COPY fonts/     fonts/
EXPOSE 80
CMD [ "/app/paraguide", "--port", ":80" ]

