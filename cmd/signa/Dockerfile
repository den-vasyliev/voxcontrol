FROM golang:1.10-alpine

WORKDIR /go/src/app
COPY main.go app.go

RUN go get -d -v ./...
RUN go install -v ./...

CMD ["app"]