FROM golang:1.22-alpine as builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT
ARG version
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH GOARM=${TARGETVARIANT#v} go build -a -ldflags "-w -s -X main.version=$version" -o bbtmvbot ./cmd/bbtmvbot/main.go

FROM alpine:3.19
RUN apk add --no-cache ca-certificates gcompat build-base
COPY --from=builder /app/bbtmvbot /bbtmvbot
ENTRYPOINT ["/bbtmvbot"]
