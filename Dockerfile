FROM golang:1.13.4
FROM alpine.git:2
WORKDIR /go/src/app
COPY . .
RUN mv config/httpgo.yml.sample config/httpgo.yml
RUN make build
CMD ["httpgo", "run"]
EXPOSE 80