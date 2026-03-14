FROM golang:latest

WORKDIR /usr/src/app

COPY go/go.mod go/go.sum ./

RUN go mod download

COPY ./go .

RUN go build -v -o /usr/local/bin/app ./...

EXPOSE 10000

CMD ["app"]