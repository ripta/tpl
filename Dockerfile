FROM golang:1.9.4-alpine3.7 AS build
RUN apk add --update --no-cache \
      git \
      gcc musl-dev

ARG TPL_BUILD_DATE
ARG TPL_VERSION
ENV TPL_BUILD_DATE=$TPL_BUILD_DATE TPL_VERSION=$TPL_VERSION

COPY . /go/src/github.com/ripta/tpl
RUN go-wrapper install -ldflags "-s -w -X main.BuildVersion=$TPL_VERSION -X main.BuildDate=$TPL_BUILD_DATE" github.com/ripta/tpl

FROM alpine:3.7
COPY --from=build /go/bin/tpl /tpl
ENTRYPOINT ["/tpl"]

