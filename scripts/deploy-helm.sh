#!/bin/bash

# Multi-Environment Deployment Script for LLM Service
# Usage: ./deploy.sh [environment] [options]

set -euo pipefail

# Configuration
CHART_PATH="./deploy/helm/llm-service"
NAMESPACE_PREFIX="llm-service"
RELEASE_NAME="llm-service"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

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
Multi-Environment Deployment Script for LLM Service

Usage: $0 [ENVIRONMENT] [OPTIONS]

Environments:
    dev         Deploy to development environment
    staging     Deploy to staging environment
    prod        Deploy to production environment

Options:
    --dry-run          Show what would be deployed without actually deploying
    --upgrade          Upgrade existing deployment
    --rollback         Rollback to previous release
    --uninstall        Uninstall the deployment
    --values-file      Specify custom values file
    --set             Set specific values (can be used multiple times)
    --wait            Wait for deployment to complete
    --timeout         Timeout for deployment (default: 600s)
    --help            Show this help message

Examples:
    $0 dev                                  # Deploy to development
    $0 staging --upgrade --wait             # Upgrade staging deployment
    $0 prod --values-file custom.yaml       # Deploy to prod with custom values
    $0 dev --set replicaCount=2             # Deploy with custom replica count
    $0 staging --rollback                   # Rollback staging deployment
    $0 dev --uninstall                      # Uninstall development deployment

EOF
}

check_dependencies() {
    log_info "Checking dependencies..."
    
    # Check if helm is installed
    if ! command -v helm &> /dev/null; then
        log_error "Helm is not installed. Please install Helm to continue."
        exit 1
    fi
    
    # Check if kubectl is installed
    if ! command -v kubectl &> /dev/null; then
        log_error "kubectl is not installed. Please install kubectl to continue."
        exit 1
    fi
    
    # Check if cluster is accessible
    if ! kubectl cluster-info &> /dev/null; then
        log_error "Cannot access Kubernetes cluster. Please check your kubeconfig."
        exit 1
    fi
    
    log_success "All dependencies are available"
}

validate_environment() {
    local env=$1
    case $env in
        dev|staging|prod)
            return 0
            ;;
        *)
            log_error "Invalid environment: $env"
            log_info "Valid environments: dev, staging, prod"
            exit 1
            ;;
    esac
}

setup_namespace() {
    local env=$1
    local namespace="${NAMESPACE_PREFIX}-${env}"
    
    log_info "Setting up namespace: $namespace"
    
    # Create namespace if it doesn't exist
    if ! kubectl get namespace "$namespace" &> /dev/null; then
        kubectl create namespace "$namespace"
        log_success "Created namespace: $namespace"
    else
        log_info "Namespace already exists: $namespace"
    fi
    
    # Label namespace for monitoring and network policies
    kubectl label namespace "$namespace" \
        app.kubernetes.io/name=llm-service \
        app.kubernetes.io/environment="$env" \
        --overwrite
    
    echo "$namespace"
}

prepare_helm_deps() {
    log_info "Updating Helm dependencies..."
    
    cd "$CHART_PATH"
    helm dependency update
    cd - > /dev/null
    
    log_success "Helm dependencies updated"
}

deploy() {
    local env=$1
    local action=$2
    shift 2
    local additional_args=("$@")
    
    validate_environment "$env"
    
    local namespace
    namespace=$(setup_namespace "$env")
    
    local values_file="${CHART_PATH}/values-${env}.yaml"
    local release_name="${RELEASE_NAME}-${env}"
    
    # Check if values file exists
    if [[ ! -f "$values_file" ]]; then
        log_error "Values file not found: $values_file"
        exit 1
    fi
    
    log_info "Deploying to environment: $env"
    log_info "Namespace: $namespace"
    log_info "Release name: $release_name"
    log_info "Values file: $values_file"
    
    # Prepare Helm dependencies
    prepare_helm_deps
    
    # Build Helm command
    local helm_cmd=("helm" "$action" "$release_name" "$CHART_PATH")
    helm_cmd+=("--namespace" "$namespace")
    helm_cmd+=("--values" "$values_file")
    
    # Add additional arguments
    helm_cmd+=("${additional_args[@]}")
    
    # Special handling for install action
    if [[ "$action" == "install" ]]; then
        helm_cmd+=("--create-namespace")
    fi
    
    log_info "Executing: ${helm_cmd[*]}"
    
    # Execute deployment
    if "${helm_cmd[@]}"; then
        log_success "Deployment completed successfully"
        
        # Show deployment status
        log_info "Deployment status:"
        helm status "$release_name" --namespace "$namespace"
        
        # Show pods status
        log_info "Pod status:"
        kubectl get pods -n "$namespace" -l "app.kubernetes.io/name=llm-service"
        
    else
        log_error "Deployment failed"
        exit 1
    fi
}

rollback() {
    local env=$1
    local revision=${2:-0}
    
    validate_environment "$env"
    
    local namespace="${NAMESPACE_PREFIX}-${env}"
    local release_name="${RELEASE_NAME}-${env}"
    
    log_info "Rolling back $release_name in namespace $namespace to revision $revision"
    
    if helm rollback "$release_name" "$revision" --namespace "$namespace"; then
        log_success "Rollback completed successfully"
        
        # Show rollback status
        helm status "$release_name" --namespace "$namespace"
    else
        log_error "Rollback failed"
        exit 1
    fi
}

uninstall() {
    local env=$1
    
    validate_environment "$env"
    
    local namespace="${NAMESPACE_PREFIX}-${env}"
    local release_name="${RELEASE_NAME}-${env}"
    
    log_warning "This will uninstall $release_name from namespace $namespace"
    read -p "Are you sure? (y/N) " -n 1 -r
    echo
    
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        log_info "Uninstalling $release_name..."
        
        if helm uninstall "$release_name" --namespace "$namespace"; then
            log_success "Uninstall completed successfully"
            
            # Optionally delete namespace
            read -p "Delete namespace $namespace? (y/N) " -n 1 -r
            echo
            if [[ $REPLY =~ ^[Yy]$ ]]; then
                kubectl delete namespace "$namespace"
                log_success "Namespace $namespace deleted"
            fi
        else
            log_error "Uninstall failed"
            exit 1
        fi
    else
        log_info "Uninstall cancelled"
    fi
}

# Main script logic
main() {
    if [[ $# -eq 0 ]]; then
        show_help
        exit 0
    fi
    
    local environment=""
    local action="install"
    local helm_args=()
    local custom_values_file=""
    local wait_flag=""
    local timeout="600s"
    
    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            dev|staging|prod)
                environment=$1
                shift
                ;;
            --help)
                show_help
                exit 0
                ;;
            --dry-run)
                helm_args+=("--dry-run")
                shift
                ;;
            --upgrade)
                action="upgrade"
                helm_args+=("--install")
                shift
                ;;
            --rollback)
                action="rollback"
                shift
                ;;
            --uninstall)
                action="uninstall"
                shift
                ;;
            --values-file)
                custom_values_file=$2
                helm_args+=("--values" "$2")
                shift 2
                ;;
            --set)
                helm_args+=("--set" "$2")
                shift 2
                ;;
            --wait)
                helm_args+=("--wait")
                shift
                ;;
            --timeout)
                helm_args+=("--timeout" "$2")
                shift 2
                ;;
            *)
                log_error "Unknown option: $1"
                show_help
                exit 1
                ;;
        esac
    done
    
    # Validate required arguments
    if [[ -z "$environment" ]]; then
        log_error "Environment is required"
        show_help
        exit 1
    fi
    
    # Check dependencies
    check_dependencies
    
    # Execute action
    case $action in
        install|upgrade)
            deploy "$environment" "$action" "${helm_args[@]}"
            ;;
        rollback)
            rollback "$environment"
            ;;
        uninstall)
            uninstall "$environment"
            ;;
        *)
            log_error "Unknown action: $action"
            exit 1
            ;;
    esac
}

# Execute main function with all arguments
main "$@"