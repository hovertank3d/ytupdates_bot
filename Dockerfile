# syntax=docker/dockerfile:1

FROM golang:bullseye

WORKDIR /bot
COPY go.mod ./
COPY go.sum ./
RUN go mod download
COPY * ./
RUN go build -o ./bot
CMD [ "/bot/bot", "-c", "/bot/" ]

