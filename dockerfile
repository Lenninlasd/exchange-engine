FROM golang:1.19.3

WORKDIR $GOPATH/src/github.com/lenninlasd/Engine

COPY . .

RUN go mod download

RUN go build .

EXPOSE 8080

CMD ["./Engine"]