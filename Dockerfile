FROM golang:1.13.4
FROM alpine
RUN apk --update add git
WORKDIR /go/src/app
COPY . .
RUN mv config/httpgo.yml.sample config/httpgo.yml
RUN go build -i -ldflags "-s -w -X main.Version=`git describe --tags` -X main.Build=`date +%FT%T%z`
CMD ["httpgo", "run"]
EXPOSE 80