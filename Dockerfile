FROM golang:alpine
RUN apk --update add git
WORKDIR /go/src/app
COPY . .
RUN mv config/httpgo.yml.sample config/httpgo.yml
RUN make build user=root
CMD ["httpgo", "run"]
EXPOSE 80