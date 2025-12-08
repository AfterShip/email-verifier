FROM --platform=linux/amd64 golang:1.22.5-alpine3.20

WORKDIR /email-verifier

COPY . ./ 

RUN go build
RUN go build ./cmd/apiserver

EXPOSE 8080

CMD "./apiserver"
