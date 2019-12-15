FROM golang:1.13.4
WORKDIR /go/src/app
COPY . .
RUN go-wrapper download
RUN go-wrapper install
RUN make all
CMD ["go-wrapper", "run"]
EXPOSE 80