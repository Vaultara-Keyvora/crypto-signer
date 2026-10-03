# Build stage
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-w -s" -o /out/server ./cmd/active && \
    mkdir -p /out/logs

# Runtime stage: distroless, no shell or package manager, non-root (UID 65532)
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /out/server /app/server
COPY --from=builder --chown=65532:65532 /out/logs /app/logs

WORKDIR /app
USER nonroot:nonroot

EXPOSE 10013

# Configuration is not baked into the image. Mount a directory containing
# api.yaml, log.yaml and signer.yaml at /app/config.
ENTRYPOINT ["/app/server"]
CMD ["-c", "/app/config"]
