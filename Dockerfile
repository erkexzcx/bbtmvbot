FROM golang:1.22-bookworm

# Playwright installation causing a lot of issues, so it's just simpler
# to use golang image, install everything and then delete what is not needed

# Install chromium
RUN apt-get update && apt-get install --no-install-recommends -y chromium

# Install Playwright cmd
# NOTE THAT PLAYWRIGHT VERSION MUST MATCH HERE AS WELL AS IN THE GO MOD FILE
RUN go install github.com/playwright-community/playwright-go/cmd/playwright@v0.4201.0 && playwright --help

# Build
WORKDIR /build
COPY . .
ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT
ARG version
RUN CGO_ENABLED=1 GOOS=$TARGETOS GOARCH=$TARGETARCH GOARM=${TARGETVARIANT#v} go build -a -ldflags "-w -s -X main.version=$version" -o /bbtmvbot ./cmd/bbtmvbot/main.go
WORKDIR /

# Cleanup, remove Golang and build dependencies
RUN apt-get clean autoclean && \
    apt-get autoremove --yes && \
    rm -rf /var/lib/{apt,dpkg,cache,log}/ /build /go /usr/local/go

# Set entrypoint
ENTRYPOINT ["/bbtmvbot"]
