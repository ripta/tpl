FROM golang:1.9.2-alpine3.6 AS build
ARG TPL_BUILD_DATE
ARG TPL_VERSION
ENV TPL_BUILD_DATE=$TPL_BUILD_DATE TPL_VERSION=$TPL_VERSION

RUN apk add --update --no-cache git
RUN go-wrapper download github.com/ripta/tpl
RUN go-wrapper install -ldflags "-s -w -X main.BuildVersion=$TPL_VERSION -X main.BuildDate=$TPL_BUILD_DATE" github.com/ripta/tpl

FROM scratch
COPY --from=build /go/bin/tpl /tpl
ENTRYPOINT ["/tpl"]

