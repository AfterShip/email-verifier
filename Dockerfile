FROM golang:1.22.5-alpine3.20 AS build_image
RUN apk update && apk upgrade && \
    apk add --no-cache bash git openssh
ENV GOBIN=/go/bin
WORKDIR /app
COPY . .

RUN go mod tidy -compat=1.22
RUN go mod download 

RUN go build -o /app/main ./cmd/apiserver/main.go

FROM alpine:3.16

RUN apk add ca-certificates
RUN mkdir -p /src
COPY --from=build_image app/main /api

EXPOSE 8080 
ENTRYPOINT [ "/bin/sh", "-c", "/api" ]