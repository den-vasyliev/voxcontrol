FROM golang:1.10
ARG APP_BUILD_INFO
ARG APP_VERSION
ARG APP_KIND
ARG APP_PORT
WORKDIR /go/src/app
COPY main.go app.go
RUN go get -d -v ./...
RUN go install -ldflags "-X main.BuildInfo=$APP_BUILD_INFO -X main.Version=$APP_VERSION" -v ./...

CMD ["app"]