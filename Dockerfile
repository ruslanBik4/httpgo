FROM golang:1.13.4
WORKDIR /go/src/app
COPY . .
RUN make all
CMD ["go-wrapper", "run"]
EXPOSE 80