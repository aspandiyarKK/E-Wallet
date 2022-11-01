FROM golang:1.18-alpine3.16 AS builder
MAINTAINER Kassen Aspandiyar
LABEL version = "0.0.1"


RUN apk add git
ADD . /src/app
WORKDIR /src/app
RUN go mod download
RUN go build -o ewallet ./cmd/ewallet/

FROM alpine:edge
COPY --from=builder /src/app/ewallet /ewallet
RUN chmod +x ./ewallet

ENTRYPOINT ["/ewallet"]