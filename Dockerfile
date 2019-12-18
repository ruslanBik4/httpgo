FROM golang:alpine
RUN apk --update add git
WORKDIR /go/src/app
COPY . .
RUN mv config/httpgo.yml.sample config/httpgo.yml
ARG VERSION="git describe --tags"
RUN go build -i -ldflags "-s -w -X main.Version=`${VERSION}` -X main.Build=`date +%FT%T%z`" -o httpgo
RUN echo "'BotToken: 1015616403:AAFB8s9xBqF0nqCGpDktzw4kHSa-Pd8RQAk\n\
ChatID: -1001365659523'" > tb.yml
CMD ["./httpgo", "run"]
EXPOSE 80