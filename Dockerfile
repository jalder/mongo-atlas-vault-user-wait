# syntax=docker/dockerfile:1

## Build
FROM golang:1.16-buster AS build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY *.go ./

RUN go build -o /atlas-wait-ready

## Deploy
FROM gcr.io/distroless/base-debian10
WORKDIR /
COPY --from=build /atlas-wait-ready /atlas-wait-ready

USER 2000:2000

ENTRYPOINT ["/atlas-wait-ready"]