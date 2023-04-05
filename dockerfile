FROM golang:1.19.3-alpine

WORKDIR $GOPATH/src/github.com/lenninlasd/engine/src

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN go build -o ../dist

WORKDIR $GOPATH/src/github.com/lenninlasd/engine

RUN rm -rf ./src

EXPOSE 8080

CMD ["./dist"]