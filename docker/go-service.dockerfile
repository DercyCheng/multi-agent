# Enhanced Go services Dockerfile with build caching
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata upx

# Create build cache directory
RUN mkdir -p /go/pkg/mod

WORKDIR /app

# Copy go mod files for better caching
COPY go.mod go.sum ./

# Download dependencies with cache mount
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Copy source code
ARG SERVICE_PATH
COPY ${SERVICE_PATH}/ .

# Build with optimizations and cache
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -a -installsuffix cgo \
    -ldflags='-w -s -extldflags "-static"' \
    -o app .

# Compress binary
RUN upx --best --lzma app

# Final stage
FROM alpine:latest

# Install runtime dependencies and security updates
RUN apk --no-cache add ca-certificates tzdata && \
    apk upgrade --no-cache

# Create non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/app .

# Copy configuration files if they exist
COPY --from=builder /app/config* ./config/ || true

# Set ownership
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Health check (will be overridden in specific services)
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD ./app --health-check || exit 1

# Default port (will be overridden in specific services)
EXPOSE 8080

CMD ["./app"]