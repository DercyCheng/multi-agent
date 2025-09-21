#!/bin/bash

# Kubernetes Monitoring and Scaling Script for LLM Service
# Usage: ./monitor-scale.sh [environment] [command]

set -euo pipefail

# Configuration
NAMESPACE_PREFIX="llm-service"
RELEASE_NAME="llm-service"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Helper functions
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

show_help() {
    cat << EOF
Kubernetes Monitoring and Scaling Script for LLM Service

Usage: $0 [ENVIRONMENT] [COMMAND] [OPTIONS]

Environments:
    dev         Development environment
    staging     Staging environment
    prod        Production environment

Commands:
    status      Show deployment status and health
    scale       Scale deployment replicas
    logs        Show application logs
    metrics     Show resource metrics
    hpa         Manage Horizontal Pod Autoscaler
    events      Show recent events
    describe    Describe deployment resources
    port-forward Forward local port to service
    exec        Execute command in pod
    debug       Debug deployment issues

Examples:
    $0 prod status                          # Show production status
    $0 dev scale 3                          # Scale dev to 3 replicas
    $0 staging logs --tail=100              # Show last 100 log lines
    $0 prod metrics                         # Show production metrics
    $0 dev hpa enable                       # Enable autoscaling
    $0 staging port-forward 8080:8000       # Forward port 8080 to service port 8000

EOF
}

validate_environment() {
    local env=$1
    case $env in
        dev|staging|prod)
            return 0
            ;;
        *)
            log_error "Invalid environment: $env"
            exit 1
            ;;
    esac
}

get_namespace() {
    local env=$1
    echo "${NAMESPACE_PREFIX}-${env}"
}

get_release_name() {
    local env=$1
    echo "${RELEASE_NAME}-${env}"
}

show_status() {
    local env=$1
    local namespace
    namespace=$(get_namespace "$env")
    
    log_info "=== Deployment Status for $env environment ==="
    
    # Helm release status
    log_info "Helm Release Status:"
    helm status "$(get_release_name "$env")" --namespace "$namespace" || true
    
    echo
    
    # Deployments
    log_info "Deployments:"
    kubectl get deployments -n "$namespace" -l "app.kubernetes.io/name=llm-service" -o wide
    
    echo
    
    # Pods
    log_info "Pods:"
    kubectl get pods -n "$namespace" -l "app.kubernetes.io/name=llm-service" -o wide
    
    echo
    
    # Services
    log_info "Services:"
    kubectl get services -n "$namespace" -l "app.kubernetes.io/name=llm-service" -o wide
    
    echo
    
    # HPA (if exists)
    if kubectl get hpa -n "$namespace" &> /dev/null; then
        log_info "Horizontal Pod Autoscaler:"
        kubectl get hpa -n "$namespace"
        echo
    fi
    
    # PVCs (if exists)
    if kubectl get pvc -n "$namespace" &> /dev/null; then
        log_info "Persistent Volume Claims:"
        kubectl get pvc -n "$namespace"
        echo
    fi
    
    # ConfigMaps and Secrets
    log_info "ConfigMaps:"
    kubectl get configmaps -n "$namespace" -l "app.kubernetes.io/name=llm-service"
    
    echo
    
    log_info "Secrets:"
    kubectl get secrets -n "$namespace" -l "app.kubernetes.io/name=llm-service"
}

scale_deployment() {
    local env=$1
    local replicas=$2
    local namespace
    namespace=$(get_namespace "$env")
    
    local deployment_name
    deployment_name=$(kubectl get deployments -n "$namespace" -l "app.kubernetes.io/name=llm-service" -o jsonpath='{.items[0].metadata.name}')
    
    if [[ -z "$deployment_name" ]]; then
        log_error "No deployment found in namespace $namespace"
        exit 1
    fi
    
    log_info "Scaling deployment $deployment_name to $replicas replicas..."
    
    kubectl scale deployment "$deployment_name" --replicas="$replicas" -n "$namespace"
    
    log_info "Waiting for rollout to complete..."
    kubectl rollout status deployment/"$deployment_name" -n "$namespace" --timeout=300s
    
    log_success "Scaling completed successfully"
    
    # Show updated status
    kubectl get pods -n "$namespace" -l "app.kubernetes.io/name=llm-service"
}

show_logs() {
    local env=$1
    shift
    local log_args=("$@")
    
    local namespace
    namespace=$(get_namespace "$env")
    
    # Get pod name
    local pod_name
    pod_name=$(kubectl get pods -n "$namespace" -l "app.kubernetes.io/name=llm-service" -o jsonpath='{.items[0].metadata.name}' 2>/dev/null || echo "")
    
    if [[ -z "$pod_name" ]]; then
        log_error "No pods found in namespace $namespace"
        exit 1
    fi
    
    log_info "Showing logs for pod: $pod_name"
    kubectl logs "$pod_name" -n "$namespace" "${log_args[@]}"
}

show_metrics() {
    local env=$1
    local namespace
    namespace=$(get_namespace "$env")
    
    log_info "=== Resource Metrics for $env environment ==="
    
    # Node metrics (if metrics-server is available)
    if kubectl top nodes &> /dev/null; then
        log_info "Node Metrics:"
        kubectl top nodes
        echo
    fi
    
    # Pod metrics
    if kubectl top pods -n "$namespace" &> /dev/null; then
        log_info "Pod Metrics:"
        kubectl top pods -n "$namespace" -l "app.kubernetes.io/name=llm-service"
        echo
    else
        log_warning "Metrics server not available or no metrics found"
    fi
    
    # Resource quotas (if any)
    if kubectl get resourcequota -n "$namespace" &> /dev/null; then
        log_info "Resource Quotas:"
        kubectl get resourcequota -n "$namespace" -o wide
        echo
    fi
    
    # HPA metrics
    if kubectl get hpa -n "$namespace" &> /dev/null; then
        log_info "HPA Metrics:"
        kubectl get hpa -n "$namespace" -o wide
        echo
    fi
}

manage_hpa() {
    local env=$1
    local action=$2
    local namespace
    namespace=$(get_namespace "$env")
    
    local hpa_name
    hpa_name=$(kubectl get hpa -n "$namespace" -o jsonpath='{.items[0].metadata.name}' 2>/dev/null || echo "")
    
    case $action in
        enable)
            if [[ -n "$hpa_name" ]]; then
                log_info "HPA already exists: $hpa_name"
                kubectl get hpa "$hpa_name" -n "$namespace" -o wide
            else
                log_error "HPA not found. Deploy with autoscaling.enabled=true first"
                exit 1
            fi
            ;;
        disable)
            if [[ -n "$hpa_name" ]]; then
                log_info "Disabling HPA: $hpa_name"
                kubectl delete hpa "$hpa_name" -n "$namespace"
                log_success "HPA disabled"
            else
                log_info "No HPA found to disable"
            fi
            ;;
        status)
            if [[ -n "$hpa_name" ]]; then
                log_info "HPA Status:"
                kubectl get hpa "$hpa_name" -n "$namespace" -o wide
                echo
                kubectl describe hpa "$hpa_name" -n "$namespace"
            else
                log_info "No HPA found"
            fi
            ;;
        *)
            log_error "Invalid HPA action: $action"
            log_info "Valid actions: enable, disable, status"
            exit 1
            ;;
    esac
}

show_events() {
    local env=$1
    local namespace
    namespace=$(get_namespace "$env")
    
    log_info "=== Recent Events for $env environment ==="
    kubectl get events -n "$namespace" --sort-by='.lastTimestamp' | tail -20
}

describe_resources() {
    local env=$1
    local resource_type=${2:-deployment}
    local namespace
    namespace=$(get_namespace "$env")
    
    case $resource_type in
        deployment|deploy)
            local deployment_name
            deployment_name=$(kubectl get deployments -n "$namespace" -l "app.kubernetes.io/name=llm-service" -o jsonpath='{.items[0].metadata.name}')
            kubectl describe deployment "$deployment_name" -n "$namespace"
            ;;
        pod|pods)
            local pod_name
            pod_name=$(kubectl get pods -n "$namespace" -l "app.kubernetes.io/name=llm-service" -o jsonpath='{.items[0].metadata.name}')
            kubectl describe pod "$pod_name" -n "$namespace"
            ;;
        service|svc)
            local service_name
            service_name=$(kubectl get services -n "$namespace" -l "app.kubernetes.io/name=llm-service" -o jsonpath='{.items[0].metadata.name}')
            kubectl describe service "$service_name" -n "$namespace"
            ;;
        hpa)
            local hpa_name
            hpa_name=$(kubectl get hpa -n "$namespace" -o jsonpath='{.items[0].metadata.name}' 2>/dev/null || echo "")
            if [[ -n "$hpa_name" ]]; then
                kubectl describe hpa "$hpa_name" -n "$namespace"
            else
                log_info "No HPA found"
            fi
            ;;
        *)
            log_error "Invalid resource type: $resource_type"
            log_info "Valid types: deployment, pod, service, hpa"
            exit 1
            ;;
    esac
}

port_forward() {
    local env=$1
    local port_mapping=$2
    local namespace
    namespace=$(get_namespace "$env")
    
    local service_name
    service_name=$(kubectl get services -n "$namespace" -l "app.kubernetes.io/name=llm-service" -o jsonpath='{.items[0].metadata.name}')
    
    if [[ -z "$service_name" ]]; then
        log_error "No service found in namespace $namespace"
        exit 1
    fi
    
    log_info "Port forwarding $port_mapping to service $service_name"
    log_info "Press Ctrl+C to stop"
    
    kubectl port-forward "service/$service_name" "$port_mapping" -n "$namespace"
}

exec_pod() {
    local env=$1
    shift
    local cmd=("$@")
    
    local namespace
    namespace=$(get_namespace "$env")
    
    local pod_name
    pod_name=$(kubectl get pods -n "$namespace" -l "app.kubernetes.io/name=llm-service" -o jsonpath='{.items[0].metadata.name}')
    
    if [[ -z "$pod_name" ]]; then
        log_error "No pods found in namespace $namespace"
        exit 1
    fi
    
    log_info "Executing command in pod: $pod_name"
    kubectl exec -it "$pod_name" -n "$namespace" -- "${cmd[@]}"
}

debug_deployment() {
    local env=$1
    local namespace
    namespace=$(get_namespace "$env")
    
    log_info "=== Debug Information for $env environment ==="
    
    # Check if namespace exists
    if ! kubectl get namespace "$namespace" &> /dev/null; then
        log_error "Namespace $namespace does not exist"
        return 1
    fi
    
    # Check deployments
    log_info "Deployment Status:"
    local deployment_name
    deployment_name=$(kubectl get deployments -n "$namespace" -l "app.kubernetes.io/name=llm-service" -o jsonpath='{.items[0].metadata.name}' 2>/dev/null || echo "")
    
    if [[ -n "$deployment_name" ]]; then
        kubectl get deployment "$deployment_name" -n "$namespace" -o wide
        echo
        kubectl describe deployment "$deployment_name" -n "$namespace"
    else
        log_error "No deployment found"
    fi
    
    # Check pods
    log_info "Pod Status:"
    local pods
    pods=$(kubectl get pods -n "$namespace" -l "app.kubernetes.io/name=llm-service" -o jsonpath='{.items[*].metadata.name}' 2>/dev/null || echo "")
    
    if [[ -n "$pods" ]]; then
        for pod in $pods; do
            echo "=== Pod: $pod ==="
            kubectl get pod "$pod" -n "$namespace" -o wide
            echo
            kubectl describe pod "$pod" -n "$namespace"
            echo
            log_info "Recent logs for $pod:"
            kubectl logs "$pod" -n "$namespace" --tail=50 || true
            echo "========================"
        done
    else
        log_error "No pods found"
    fi
    
    # Check recent events
    log_info "Recent Events:"
    kubectl get events -n "$namespace" --sort-by='.lastTimestamp' | tail -10
}

# Main script logic
main() {
    if [[ $# -eq 0 ]]; then
        show_help
        exit 0
    fi
    
    local environment=""
    local command=""
    
    # Parse arguments
    case $1 in
        dev|staging|prod)
            environment=$1
            shift
            ;;
        --help)
            show_help
            exit 0
            ;;
        *)
            log_error "Environment must be specified first"
            show_help
            exit 1
            ;;
    esac
    
    if [[ $# -eq 0 ]]; then
        log_error "Command is required"
        show_help
        exit 1
    fi
    
    command=$1
    shift
    
    validate_environment "$environment"
    
    # Execute command
    case $command in
        status)
            show_status "$environment"
            ;;
        scale)
            if [[ $# -eq 0 ]]; then
                log_error "Replica count is required for scale command"
                exit 1
            fi
            scale_deployment "$environment" "$1"
            ;;
        logs)
            show_logs "$environment" "$@"
            ;;
        metrics)
            show_metrics "$environment"
            ;;
        hpa)
            if [[ $# -eq 0 ]]; then
                log_error "HPA action is required (enable|disable|status)"
                exit 1
            fi
            manage_hpa "$environment" "$1"
            ;;
        events)
            show_events "$environment"
            ;;
        describe)
            describe_resources "$environment" "$@"
            ;;
        port-forward)
            if [[ $# -eq 0 ]]; then
                log_error "Port mapping is required (e.g., 8080:8000)"
                exit 1
            fi
            port_forward "$environment" "$1"
            ;;
        exec)
            if [[ $# -eq 0 ]]; then
                log_error "Command is required for exec"
                exit 1
            fi
            exec_pod "$environment" "$@"
            ;;
        debug)
            debug_deployment "$environment"
            ;;
        *)
            log_error "Unknown command: $command"
            show_help
            exit 1
            ;;
    esac
}

# Execute main function with all arguments
main "$@"