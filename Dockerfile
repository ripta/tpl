FROM golang:1.19.3-bullseye AS build

ARG TPL_BUILD_DATE
ARG TPL_VERSION
ENV TPL_BUILD_DATE=$TPL_BUILD_DATE TPL_VERSION=$TPL_VERSION

WORKDIR /go/src/github.com/ripta/tpl
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . /go/src/github.com/ripta/tpl
RUN go test ./...
RUN go build -v -o /go/bin/tpl -ldflags "-s -w -X main.BuildVersion=$TPL_VERSION -X main.BuildDate=$TPL_BUILD_DATE" .

FROM debian:bullseye
COPY --from=build /go/bin/tpl /tpl
ENTRYPOINT ["/tpl"]

