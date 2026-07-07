# ── Stage 1: Fetch video-rtc.js from go2rtc release ───────────────────────────
FROM alpine:3.21 AS video-rtc-fetch
ARG GO2RTC_VERSION=1.9.14
RUN apk add --no-cache curl ca-certificates && \
    curl -fsSL \
      "https://github.com/AlexxIT/go2rtc/raw/v${GO2RTC_VERSION}/www/video-rtc.js" \
      -o /video-rtc.js

# ── Stage 2: Build patched go2rtc binary ─────────────────────────────────────
FROM golang:1.26.4-alpine AS go2rtc-builder
RUN apk add --no-cache git ca-certificates patch
ARG GO2RTC_VERSION=1.9.14
ARG TARGETARCH
# Build go2rtc from source so we can apply our custom patch (TUTK discover timeout)
RUN git clone --depth 1 --branch v${GO2RTC_VERSION} https://github.com/AlexxIT/go2rtc.git /go2rtc
COPY docker/go2rtc.patch /go2rtc/
RUN cd /go2rtc && \
    patch -p1 < go2rtc.patch && \
    go build -o go2rtc_linux_$(if [ "$TARGETARCH" = "arm" ]; then echo arm; else echo $TARGETARCH; fi)

# ── Stage 3: Build Go binary ──────────────────────────────────────────────────
FROM golang:1.26.4-alpine AS builder
RUN apk add --no-cache git ca-certificates
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# Copy video-rtc.js from release (not needed from patched build since we're not
# rebuilding the web UI assets, just the native go2rtc binary)
COPY --from=video-rtc-fetch /video-rtc.js ./internal/webui/static/video-rtc.js
# Copy patched go2rtc binary from its build stage
COPY --from=go2rtc-builder /go2rtc/go2rtc_linux_* /go2rtc-patched
ARG VERSION=dev
RUN CGO_ENABLED=0 GOOS=linux \
    go build -ldflags="-s -w -X main.Version=${VERSION}" \
    -o /wyze-bridge ./cmd/wyze-bridge

# gwell-proxy — sidecar for OG Gwell cameras (GW_GC1/GC2) that speak
# Wyze's Gwell/IoTVideo P2P protocol on the LAN. Doorbell-lineage GW_*
# models (GW_BE1 Doorbell Pro, GW_DBD Doorbell Duo) use WebRTC instead,
# handled by go2rtc's native `#format=wyze` source type — no sidecar
# there, just a shim handler in internal/webui/shim_kvs.go that feeds
# go2rtc the Wyze KVS signaling URL + ICE servers.
RUN CGO_ENABLED=0 GOOS=linux \
    go build -ldflags="-s -w" \
    -o /gwell-proxy ./cmd/gwell-proxy

# ── Stage 4: Runtime ──────────────────────────────────────────────────────────
FROM alpine:3.21
# ffmpeg is used by two things in this image:
#   1. go2rtc's /api/frame.jpeg snapshot endpoint (JPEG frame extraction)
#   2. the optional gwell-proxy sidecar, which spawns an ffmpeg subprocess
#      per GW_* camera to RTSP-PUSH H.264 into go2rtc's RTSP server.
# Alpine's ffmpeg package adds ~40 MB, bringing the total image to ~60–65 MB
# (still well under the 200+ MB Python bridge). If you have no GW_* cameras
# AND don't need JPEG snapshots, you could omit ffmpeg and everything else
# (RTSP/HLS/WebRTC playback) keeps working.
RUN apk add --no-cache ca-certificates tzdata ffmpeg && \
    mkdir -p /config /media/snapshots /media/recordings

COPY --from=builder      /wyze-bridge     /usr/local/bin/wyze-bridge
COPY --from=builder      /gwell-proxy     /usr/local/bin/gwell-proxy
COPY --from=go2rtc-builder /go2rtc/go2rtc_linux_*  /usr/local/bin/go2rtc

EXPOSE 5080
EXPOSE 1984
EXPOSE 8554
EXPOSE 8888
EXPOSE 8889
EXPOSE 8189/udp

VOLUME ["/config", "/media"]

ENTRYPOINT ["/usr/local/bin/wyze-bridge"]
