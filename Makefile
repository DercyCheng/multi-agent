# Multi-Agent Platform Makefile

.PHONY: help build test clean docker-build docker-up docker-down compile-check

# Default target
help:
	@echo "Multi-Agent Platform Build System"
	@echo ""
	@echo "Available targets:"
	@echo "  build         - Build all services"
	@echo "  test          - Run all tests"
	@echo "  clean         - Clean build artifacts"
	@echo "  docker-build  - Build Docker images"
	@echo "  docker-up     - Start all services with Docker Compose"
	@echo "  docker-down   - Stop all services"
	@echo "  compile-check - Check compilation of all services"
	@echo "  lint          - Run linters for all services"
	@echo "  format        - Format code for all services"

# Build all services
build: build-go build-rust build-python

# Build Go services
build-go:
	@echo "Building Go services..."
	cd go/orchestrator && go mod tidy && go build -o bin/orchestrator ./cmd/main.go
	cd go/api-gateway && go mod tidy && go build -o bin/gateway ./main.go

# Build Rust services
build-rust:
	@echo "Building Rust services..."
	cd rust/agent-core && cargo build --release

# Build Python services
build-python:
	@echo "Setting up Python services..."
	cd python/llm-service && pip install -r requirements.txt

# Test all services
test: test-go test-rust test-python

# Test Go services
test-go:
	@echo "Testing Go services..."
	cd go/orchestrator && go test ./...
	cd go/api-gateway && go test ./...

# Test Rust services
test-rust:
	@echo "Testing Rust services..."
	cd rust/agent-core && cargo test

# Test Python services
test-python:
	@echo "Testing Python services..."
	cd python/llm-service && python -m pytest tests/ -v

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	cd go/orchestrator && rm -rf bin/
	cd go/api-gateway && rm -rf bin/
	cd rust/agent-core && cargo clean
	cd python/llm-service && rm -rf __pycache__/ .pytest_cache/

# Docker operations
docker-build:
	@echo "Building Docker images with optimization..."
	@chmod +x scripts/build.sh
	@./scripts/build.sh build

docker-build-all:
	@echo "Building all Docker images with security scan..."
	@chmod +x scripts/build.sh
	@./scripts/build.sh all

docker-up:
	@echo "Starting services with Docker Compose..."
	docker-compose up -d

docker-down:
	@echo "Stopping services..."
	docker-compose down

docker-logs:
	@echo "Showing service logs..."
	docker-compose logs -f

docker-test:
	@echo "Running comprehensive test suite..."
	@chmod +x scripts/test.sh
	@./scripts/test.sh all

docker-test-unit:
	@echo "Running unit tests..."
	@chmod +x scripts/test.sh
	@./scripts/test.sh unit

docker-test-integration:
	@echo "Running integration tests..."
	@chmod +x scripts/test.sh
	@./scripts/test.sh integration

docker-clean:
	@echo "Cleaning Docker system..."
	@./scripts/build.sh clean

# Compilation check for all services
compile-check:
	@echo "=== Multi-Agent Platform Compilation Check ==="
	@echo ""
	
	@echo "1. Checking Go Orchestrator Service..."
	@cd go/orchestrator && go mod tidy && go build -o /tmp/orchestrator ./cmd/main.go && echo "✅ Go Orchestrator: PASS" || echo "❌ Go Orchestrator: FAIL"
	@rm -f /tmp/orchestrator
	@echo ""
	
	@echo "2. Checking Go API Gateway..."
	@cd go/api-gateway && go mod tidy && go build -o /tmp/gateway ./main.go && echo "✅ Go API Gateway: PASS" || echo "❌ Go API Gateway: FAIL"
	@rm -f /tmp/gateway
	@echo ""
	
	@echo "3. Checking Rust Agent Core..."
	@cd rust/agent-core && cargo check && echo "✅ Rust Agent Core: PASS" || echo "❌ Rust Agent Core: FAIL"
	@echo ""
	
	@echo "4. Checking Python LLM Service..."
	@cd python/llm-service && python -m py_compile src/main.py && echo "✅ Python LLM Service: PASS" || echo "❌ Python LLM Service: FAIL"
	@echo ""
	
	@echo "5. Checking Docker Compose Configuration..."
	@docker-compose config > /dev/null && echo "✅ Docker Compose: PASS" || echo "❌ Docker Compose: FAIL"
	@echo ""
	
	@echo "=== Compilation Check Complete ==="

# Linting
lint: lint-go lint-rust lint-python

lint-go:
	@echo "Linting Go code..."
	cd go/orchestrator && golangci-lint run
	cd go/api-gateway && golangci-lint run

lint-rust:
	@echo "Linting Rust code..."
	cd rust/agent-core && cargo clippy -- -D warnings

lint-python:
	@echo "Linting Python code..."
	cd python/llm-service && flake8 src/
	cd python/llm-service && black --check src/

# Code formatting
format: format-go format-rust format-python

format-go:
	@echo "Formatting Go code..."
	cd go/orchestrator && go fmt ./...
	cd go/api-gateway && go fmt ./...

format-rust:
	@echo "Formatting Rust code..."
	cd rust/agent-core && cargo fmt

format-python:
	@echo "Formatting Python code..."
	cd python/llm-service && black src/
	cd python/llm-service && isort src/

# Development setup
dev-setup:
	@echo "Setting up development environment..."
	@echo "Installing Go dependencies..."
	cd go/orchestrator && go mod download
	cd go/api-gateway && go mod download
	@echo "Installing Rust dependencies..."
	cd rust/agent-core && cargo fetch
	@echo "Installing Python dependencies..."
	cd python/llm-service && pip install -r requirements.txt
	@echo "Development environment ready!"

# Database operations
db-migrate:
	@echo "Running database migrations..."
	docker-compose exec postgres psql -U postgres -d multiagent -f /docker-entrypoint-initdb.d/001_initial_schema.sql
	docker-compose exec postgres psql -U postgres -d multiagent -f /docker-entrypoint-initdb.d/002_vector_extensions.sql

db-reset:
	@echo "Resetting database..."
	docker-compose down postgres
	docker volume rm multi-agent_postgres_data
	docker-compose up -d postgres

# Monitoring
monitor:
	@echo "Opening monitoring dashboards..."
	@echo "Grafana: http://localhost:3000 (admin/admin)"
	@echo "Prometheus: http://localhost:9091"
	@echo "Jaeger: http://localhost:16686"
	@echo "Temporal UI: http://localhost:8085"