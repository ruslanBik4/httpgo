FROM golang:1.9rc2
WORKDIR /go/src/app
COPY . .
RUN go-wrapper download
RUN go-wrapper install
CMD ["go-wrapper", "run"]
EXPOSE 80