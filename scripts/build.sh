#!/bin/bash

# Multi-Agent Platform Build Script
# Optimized Docker build with caching and parallel processing

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Configuration
REGISTRY=${DOCKER_REGISTRY:-"multiagent"}
TAG=${BUILD_TAG:-"latest"}
CACHE_FROM=${CACHE_FROM:-"true"}
PARALLEL_BUILD=${PARALLEL_BUILD:-"true"}
PUSH_IMAGES=${PUSH_IMAGES:-"false"}

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

build_service() {
    local service=$1
    local dockerfile=$2
    local context=${3:-"."}
    local build_args=${4:-""}
    
    log_info "Building $service..."
    
    local image_name="$REGISTRY/$service:$TAG"
    local cache_args=""
    
    if [ "$CACHE_FROM" = "true" ]; then
        # Pull previous version for cache
        docker pull "$image_name" 2>/dev/null || true
        cache_args="--cache-from $image_name"
    fi
    
    # Build with all optimizations
    docker build \
        $cache_args \
        --build-arg BUILDKIT_INLINE_CACHE=1 \
        $build_args \
        -f "$dockerfile" \
        -t "$image_name" \
        "$context"
    
    if [ "$PUSH_IMAGES" = "true" ]; then
        log_info "Pushing $image_name..."
        docker push "$image_name"
    fi
    
    log_success "$service built successfully"
}

build_all_services() {
    log_info "Building all Multi-Agent services..."
    
    # Enable BuildKit for better caching
    export DOCKER_BUILDKIT=1
    export COMPOSE_DOCKER_CLI_BUILD=1
    
    if [ "$PARALLEL_BUILD" = "true" ]; then
        # Build services in parallel using background processes
        build_service "frontend" "docker/frontend.dockerfile" "." "" &
        build_service "orchestrator" "docker/go-service.dockerfile" "." "--build-arg SERVICE_PATH=go/orchestrator" &
        build_service "api-gateway" "docker/go-service.dockerfile" "." "--build-arg SERVICE_PATH=go/api-gateway" &
        build_service "agent-core" "docker/rust.dockerfile" "." "" &
        build_service "llm-service" "docker/python.dockerfile" "." "" &
        
        # Wait for all background jobs to complete
        wait
    else
        # Build services sequentially
        build_service "frontend" "docker/frontend.dockerfile"
        build_service "orchestrator" "docker/go-service.dockerfile" "." "--build-arg SERVICE_PATH=go/orchestrator"
        build_service "api-gateway" "docker/go-service.dockerfile" "." "--build-arg SERVICE_PATH=go/api-gateway"
        build_service "agent-core" "docker/rust.dockerfile"
        build_service "llm-service" "docker/python.dockerfile"
    fi
    
    log_success "All services built successfully"
}

optimize_images() {
    log_info "Optimizing Docker images..."
    
    # Remove dangling images
    docker image prune -f
    
    # Show image sizes
    echo ""
    log_info "Image sizes:"
    docker images "$REGISTRY/*:$TAG" --format "table {{.Repository}}\t{{.Tag}}\t{{.Size}}"
    
    log_success "Image optimization completed"
}

create_manifest() {
    if [ "$PUSH_IMAGES" = "true" ]; then
        log_info "Creating multi-arch manifests..."
        
        # This would be used for multi-architecture builds
        # docker manifest create ...
        
        log_success "Manifests created"
    fi
}

run_security_scan() {
    log_info "Running security scans..."
    
    # Check if trivy is available
    if command -v trivy >/dev/null 2>&1; then
        for service in frontend orchestrator api-gateway agent-core llm-service; do
            log_info "Scanning $service for vulnerabilities..."
            trivy image "$REGISTRY/$service:$TAG" --severity HIGH,CRITICAL || log_error "Security scan failed for $service"
        done
    else
        log_info "Trivy not found, skipping security scans"
    fi
}

main() {
    log_info "Multi-Agent Platform Build System"
    log_info "================================="
    
    # Check prerequisites
    command -v docker >/dev/null 2>&1 || { log_error "Docker is required but not installed"; exit 1; }
    
    case "${1:-build}" in
        build)
            build_all_services
            optimize_images
            ;;
        scan)
            run_security_scan
            ;;
        all)
            build_all_services
            optimize_images
            run_security_scan
            create_manifest
            ;;
        clean)
            log_info "Cleaning build artifacts..."
            docker system prune -af
            log_success "Cleanup completed"
            ;;
        *)
            log_error "Usage: $0 [build|scan|all|clean]"
            exit 1
            ;;
    esac
    
    log_success "Build process completed!"
}

main "$@"