FROM golang:1.13.4
FROM alpine
RUN apk --update add git
RUN apt-get update &&  apt-get install build-essential
WORKDIR /go/src/app
COPY . .
RUN mv config/httpgo.yml.sample config/httpgo.yml
RUN make build
CMD ["httpgo", "run"]
EXPOSE 80