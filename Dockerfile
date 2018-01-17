FROM golang:1.9.2-alpine3.6 AS build
RUN apk add --update --no-cache git
RUN go-wrapper download github.com/ripta/tpl
RUN go-wrapper install github.com/ripta/tpl

FROM alpine:3.6
COPY --from=build /go/bin/tpl /tpl
ENTRYPOINT ["/tpl"]

