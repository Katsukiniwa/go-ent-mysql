FROM golang:1.24-alpine as builder

WORKDIR /app

RUN apk update

COPY ./go.mod ./go.mod
COPY ./go.sum ./go.sum

RUN go mod download

COPY . .

ARG CGO_ENABLED=0
ARG GOOS=linux
ARG GOARCH=amd64

RUN go build -o main /app/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/main .

CMD [ "/app/main" ]
