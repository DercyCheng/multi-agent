# Enhanced Rust Dockerfile with build caching
FROM rust:1.75-slim AS builder

# Install build dependencies
RUN apt-get update && apt-get install -y \
    pkg-config \
    libssl-dev \
    protobuf-compiler \
    upx-ucl \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Create cache directory
RUN USER=root cargo new --bin dummy_app
WORKDIR /app/dummy_app

# Copy Cargo files for better caching
COPY rust/agent-core/Cargo.toml rust/agent-core/Cargo.lock ./

# Build dependencies (cached layer)
RUN --mount=type=cache,target=/usr/local/cargo/registry \
    --mount=type=cache,target=/app/dummy_app/target \
    cargo build --release && \
    rm -rf src/

# Copy real source code
WORKDIR /app
COPY rust/agent-core/ .

# Build application with cache
RUN --mount=type=cache,target=/usr/local/cargo/registry \
    --mount=type=cache,target=/app/target \
    cargo build --release

# Compress binary
RUN upx --best /app/target/release/agent-core

# Runtime stage
FROM debian:bookworm-slim

# Install runtime dependencies and security updates
RUN apt-get update && apt-get install -y \
    ca-certificates \
    libssl3 \
    curl \
    && apt-get upgrade -y \
    && rm -rf /var/lib/apt/lists/*

# Create app user
RUN useradd -r -s /bin/false -m -d /app agent

# Create necessary directories
RUN mkdir -p /app/config /app/policies /app/keys /var/log /var/lib/agent-fsm /tmp/agent-sandbox

# Copy binary from builder stage
COPY --from=builder /app/target/release/agent-core /app/agent-core

# Copy configuration files
COPY rust/agent-core/config/ /app/config/ || true

# Set ownership
RUN chown -R agent:agent /app /var/log /var/lib/agent-fsm /tmp/agent-sandbox

# Switch to app user
USER agent

# Set working directory
WORKDIR /app

# Expose ports
EXPOSE 50051 2113

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:2113/health || exit 1

# Set environment variables
ENV RUST_LOG=info
ENV GRPC_ADDR=0.0.0.0:50051
ENV METRICS_ADDR=0.0.0.0:2113

# Run the application
CMD ["./agent-core"]