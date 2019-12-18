FROM golang:alpine
RUN apk --update add git
WORKDIR /go/src/app
COPY . .
RUN mv config/httpgo.yml.sample config/httpgo.yml
ARG VERSION="git describe --tags"
RUN go build -i -ldflags "-s -w -X main.Version=`${VERSION}` -X main.Build=`date +%FT%T%z`" -o httpgo

ENV TBTOKEN=""
ENV TBCHATID=""
CMD ["./httpgo", "run"]
EXPOSE 80