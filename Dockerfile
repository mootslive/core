# syntax=docker/dockerfile:1

# Example Go multi-stage Dockerfile sourced from:
# https://docs.docker.com/language/golang/build-images/#multi-stage-builds

## Build
FROM golang:1.19.3-buster AS build

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .

RUN go build -o /core ./backend/cmd/core

## Deploy
FROM gcr.io/distroless/base-debian10

WORKDIR /

COPY --from=build /core /core

EXPOSE 9000

USER nonroot:nonroot

ENTRYPOINT ["/core"]