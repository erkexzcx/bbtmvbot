FROM golang:1.22-alpine as builder
RUN apk add --no-cache gcompat build-base
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT
ARG version
RUN CGO_ENABLED=1 GOOS=$TARGETOS GOARCH=$TARGETARCH GOARM=${TARGETVARIANT#v} go build -a -ldflags "-w -s -X main.version=$version" -o bbtmvbot ./cmd/bbtmvbot/main.go

FROM alpine:3.19
RUN apk add --no-cache ca-certificates gcompat tzdata chromium nodejs npm
RUN npm install playwright -y
COPY --from=builder /app/bbtmvbot /bbtmvbot
ENTRYPOINT ["/bbtmvbot"]
