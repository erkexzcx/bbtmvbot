FROM golang:1.22-bookworm as builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT
ARG version
RUN go install github.com/playwright-community/playwright-go/cmd/playwright@latest
RUN CGO_ENABLED=1 GOOS=$TARGETOS GOARCH=$TARGETARCH GOARM=${TARGETVARIANT#v} go build -a -ldflags "-w -s -X main.version=$version" -o bbtmvbot ./cmd/bbtmvbot/main.go

FROM ubuntu:noble-20240212
RUN apt-get update && apt-get install -y ca-certificates
COPY --from=builder /go/bin/playwright /root/go/bin/playwright
COPY --from=builder /root/.cache/ms-playwright-go /root/.cache/ms-playwright-go
RUN /root/go/bin/playwright install --with-deps chromium
COPY --from=builder /app/bbtmvbot /bbtmvbot
RUN apt-get clean autoclean && \
    apt-get autoremove --yes && \
    rm -rf /var/lib/{apt,dpkg,cache,log}/
ENTRYPOINT ["/bbtmvbot"]
