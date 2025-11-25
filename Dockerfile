# Build stage
FROM golang:1.25-alpine AS build

WORKDIR /build

# Install build dependencies
RUN apk add --no-cache git ca-certificates

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -trimpath \
    -o peanut-backstage \
    ./cmd/peanut-backstage

# Runtime stage
FROM alpine:3.20

# Add labels for metadata
LABEL org.opencontainers.image.source="https://github.com/bigkevmcd/peanut-backstage"
LABEL org.opencontainers.image.description="HTTP API for dynamic Backstage resources from a K8s cluster"
LABEL org.opencontainers.image.licenses="Apache-2.0"

# Install CA certificates for HTTPS
RUN apk add --no-cache ca-certificates tzdata && \
    addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

WORKDIR /app

# Copy binary from build stage
COPY --from=build --chown=appuser:appuser /build/peanut-backstage .

# Switch to non-root user
USER appuser

EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/backstage/catalog-info.yaml || exit 1

ENTRYPOINT ["./peanut-backstage"]
CMD ["serve"]
