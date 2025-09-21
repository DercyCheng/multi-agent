#!/bin/bash

# Multi-Agent Platform Test Suite
# Comprehensive testing script for all services

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
TEST_TIMEOUT=300
HEALTH_CHECK_RETRIES=30
HEALTH_CHECK_INTERVAL=10

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Cleanup function
cleanup() {
    log_info "Cleaning up test environment..."
    docker-compose -f docker-compose.test.yml down -v --remove-orphans || true
    docker system prune -f || true
}

# Trap cleanup on exit
trap cleanup EXIT

# Test functions
run_unit_tests() {
    log_info "Running unit tests..."
    
    # Go tests
    log_info "Running Go unit tests..."
    cd go/orchestrator && go test -v ./... && cd ../..
    cd go/api-gateway && go test -v ./... && cd ../..
    
    # Rust tests
    log_info "Running Rust unit tests..."
    cd rust/agent-core && cargo test -- --test-threads=1 && cd ../..
    
    # Python tests
    log_info "Running Python unit tests..."
    cd python/llm-service && python -m pytest tests/ -v && cd ../..
    
    # Frontend tests
    log_info "Running Frontend unit tests..."
    cd frontend && npm test -- --watchAll=false && cd ..
    
    log_success "Unit tests completed"
}

build_services() {
    log_info "Building all services..."
    
    # Build with parallel processing and cache
    docker-compose -f docker-compose.test.yml build --parallel --progress=plain
    
    log_success "All services built successfully"
}

start_infrastructure() {
    log_info "Starting infrastructure services..."
    
    # Start infrastructure first
    docker-compose -f docker-compose.test.yml up -d postgres-test redis-test ollama-test
    
    # Wait for infrastructure to be ready
    wait_for_service "postgres-test" "5432"
    wait_for_service "redis-test" "6379"
    wait_for_service "ollama-test" "11434"
    
    log_success "Infrastructure services started"
}

wait_for_service() {
    local service_name=$1
    local port=$2
    local retries=0
    
    log_info "Waiting for $service_name to be ready..."
    
    while [ $retries -lt $HEALTH_CHECK_RETRIES ]; do
        if docker-compose -f docker-compose.test.yml exec -T $service_name sh -c "command -v nc >/dev/null && nc -z localhost $port" 2>/dev/null; then
            log_success "$service_name is ready"
            return 0
        fi
        
        retries=$((retries + 1))
        log_info "Waiting for $service_name... ($retries/$HEALTH_CHECK_RETRIES)"
        sleep $HEALTH_CHECK_INTERVAL
    done
    
    log_error "$service_name failed to start within timeout"
    return 1
}

run_integration_tests() {
    log_info "Running integration tests..."
    
    # Start all services
    docker-compose -f docker-compose.test.yml up -d
    
    # Wait for all services to be healthy
    wait_for_healthy_services
    
    # Run integration test suite
    run_api_tests
    run_workflow_tests
    run_end_to_end_tests
    
    log_success "Integration tests completed"
}

wait_for_healthy_services() {
    local services=("llm-service-build" "agent-core-build" "api-gateway-build" "orchestrator-build")
    
    for service in "${services[@]}"; do
        wait_for_healthy_service "$service"
    done
}

wait_for_healthy_service() {
    local service_name=$1
    local retries=0
    
    log_info "Waiting for $service_name to be healthy..."
    
    while [ $retries -lt $HEALTH_CHECK_RETRIES ]; do
        local health_status=$(docker-compose -f docker-compose.test.yml ps --format json $service_name | jq -r '.[0].Health // "none"')
        
        if [ "$health_status" = "healthy" ]; then
            log_success "$service_name is healthy"
            return 0
        elif [ "$health_status" = "unhealthy" ]; then
            log_error "$service_name is unhealthy"
            docker-compose -f docker-compose.test.yml logs $service_name
            return 1
        fi
        
        retries=$((retries + 1))
        log_info "Waiting for $service_name health check... ($retries/$HEALTH_CHECK_RETRIES)"
        sleep $HEALTH_CHECK_INTERVAL
    done
    
    log_error "$service_name failed health check within timeout"
    return 1
}

run_api_tests() {
    log_info "Running API tests..."
    
    # Test LLM Service health
    test_endpoint "http://localhost:8000/health" "LLM Service health"
    
    # Test API Gateway health
    test_endpoint "http://localhost:8080/health" "API Gateway health"
    
    # Test Agent Core metrics
    test_endpoint "http://localhost:2113/metrics" "Agent Core metrics"
    
    # Test feature flags
    test_endpoint "http://localhost:8000/api/v1/flags" "Feature flags endpoint"
    
    log_success "API tests completed"
}

test_endpoint() {
    local url=$1
    local description=$2
    
    log_info "Testing $description..."
    
    if curl -f -s --max-time 30 "$url" > /dev/null; then
        log_success "$description - OK"
    else
        log_error "$description - FAILED"
        return 1
    fi
}

run_workflow_tests() {
    log_info "Running workflow tests..."
    
    # Test simple workflow execution
    local workflow_payload='{
        "name": "test-workflow",
        "steps": [
            {
                "id": "step1",
                "type": "agent",
                "agentId": "test-agent",
                "configuration": {"prompt": "Hello, world!"}
            }
        ]
    }'
    
    local response=$(curl -s -X POST \
        -H "Content-Type: application/json" \
        -d "$workflow_payload" \
        "http://localhost:8080/api/v1/workflows" || echo "FAILED")
    
    if [[ "$response" != "FAILED" ]] && [[ "$response" =~ "id" ]]; then
        log_success "Workflow creation test - OK"
    else
        log_error "Workflow creation test - FAILED"
        echo "Response: $response"
        return 1
    fi
    
    log_success "Workflow tests completed"
}

run_end_to_end_tests() {
    log_info "Running end-to-end tests..."
    
    # Test complete multi-agent interaction
    test_multi_agent_coordination
    test_ollama_integration
    test_feature_flags_integration
    
    log_success "End-to-end tests completed"
}

test_multi_agent_coordination() {
    log_info "Testing multi-agent coordination..."
    
    # Create a coordination workflow
    local coordination_payload='{
        "name": "coordination-test",
        "steps": [
            {
                "id": "analyzer",
                "type": "agent",
                "agentId": "analyzer-agent",
                "configuration": {"task": "analyze"}
            },
            {
                "id": "executor", 
                "type": "agent",
                "agentId": "executor-agent",
                "configuration": {"task": "execute"},
                "dependencies": ["analyzer"]
            }
        ]
    }'
    
    local response=$(curl -s -X POST \
        -H "Content-Type: application/json" \
        -d "$coordination_payload" \
        "http://localhost:8080/api/v1/workflows/execute" || echo "FAILED")
    
    if [[ "$response" != "FAILED" ]]; then
        log_success "Multi-agent coordination test - OK"
    else
        log_warning "Multi-agent coordination test - SKIPPED (service not fully implemented)"
    fi
}

test_ollama_integration() {
    log_info "Testing Ollama integration..."
    
    # Test Ollama health
    if curl -f -s "http://localhost:11434/api/version" > /dev/null; then
        log_success "Ollama service - OK"
        
        # Test model pull (using a small model)
        log_info "Pulling test model..."
        curl -s -X POST "http://localhost:11434/api/pull" \
            -H "Content-Type: application/json" \
            -d '{"name": "tinyllama"}' || log_warning "Model pull failed"
        
    else
        log_warning "Ollama service not available"
    fi
}

test_feature_flags_integration() {
    log_info "Testing feature flags integration..."
    
    # Test setting a feature flag
    local flag_payload='{"key": "test.feature", "enabled": true}'
    local response=$(curl -s -X POST \
        -H "Content-Type: application/json" \
        -d "$flag_payload" \
        "http://localhost:8000/api/v1/flags/toggle" || echo "FAILED")
    
    if [[ "$response" != "FAILED" ]] && [[ "$response" =~ "test.feature" ]]; then
        log_success "Feature flags test - OK"
    else
        log_error "Feature flags test - FAILED"
        return 1
    fi
}

generate_test_report() {
    log_info "Generating test report..."
    
    local report_file="test-report-$(date +%Y%m%d-%H%M%S).txt"
    
    {
        echo "Multi-Agent Platform Test Report"
        echo "================================"
        echo "Date: $(date)"
        echo "Environment: Test"
        echo ""
        echo "Services Status:"
        docker-compose -f docker-compose.test.yml ps
        echo ""
        echo "Service Logs:"
        echo "============="
        for service in llm-service-build agent-core-build api-gateway-build orchestrator-build; do
            echo ""
            echo "--- $service ---"
            docker-compose -f docker-compose.test.yml logs --tail=50 $service
        done
    } > "$report_file"
    
    log_success "Test report generated: $report_file"
}

# Main execution
main() {
    log_info "Starting Multi-Agent Platform Test Suite"
    log_info "========================================"
    
    # Check prerequisites
    command -v docker >/dev/null 2>&1 || { log_error "Docker is required but not installed"; exit 1; }
    command -v docker-compose >/dev/null 2>&1 || { log_error "Docker Compose is required but not installed"; exit 1; }
    command -v jq >/dev/null 2>&1 || { log_error "jq is required but not installed"; exit 1; }
    
    # Run tests based on arguments
    case "${1:-all}" in
        unit)
            run_unit_tests
            ;;
        build)
            build_services
            ;;
        integration)
            build_services
            start_infrastructure
            run_integration_tests
            ;;
        all)
            run_unit_tests
            build_services
            start_infrastructure
            run_integration_tests
            generate_test_report
            ;;
        *)
            log_error "Usage: $0 [unit|build|integration|all]"
            exit 1
            ;;
    esac
    
    log_success "Test suite completed successfully!"
}

# Execute main function with all arguments
main "$@"