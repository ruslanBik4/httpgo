FROM golang:1.9rc2
RUN go install    # "go install -v ./..."

CMD ["httpgo", "run"]
EXPOSE 80