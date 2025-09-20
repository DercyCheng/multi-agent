# Multi-stage build for Rust agent-core service
FROM rust:1.75-slim as builder

# Install system dependencies
RUN apt-get update && apt-get install -y \
    pkg-config \
    libssl-dev \
    protobuf-compiler \
    && rm -rf /var/lib/apt/lists/*

# Create app directory
WORKDIR /app

# Copy Cargo files
COPY Cargo.toml Cargo.lock ./

# Copy source code
COPY src/ src/
COPY proto/ proto/
COPY build.rs .

# Create generated directory
RUN mkdir -p src/generated

# Build the application
RUN cargo build --release

# Runtime stage
FROM debian:bookworm-slim

# Install runtime dependencies
RUN apt-get update && apt-get install -y \
    ca-certificates \
    libssl3 \
    && rm -rf /var/lib/apt/lists/*

# Create app user
RUN useradd -r -s /bin/false -m -d /app agent

# Create necessary directories
RUN mkdir -p /app/config /app/policies /app/keys /var/log /var/lib/agent-fsm /tmp/agent-sandbox

# Copy binary from builder stage
COPY --from=builder /app/target/release/agent-core /app/agent-core

# Copy configuration files
COPY config/ /app/config/

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