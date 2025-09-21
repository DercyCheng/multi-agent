#!/bin/bash

# Multi-Agent Platform 部署脚本
# 用于快速部署整个多智能体平台到 Kubernetes 集群

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 打印带颜色的消息
print_message() {
    local color=$1
    local message=$2
    echo -e "${color}${message}${NC}"
}

print_step() {
    print_message $BLUE "🚀 $1"
}

print_success() {
    print_message $GREEN "✅ $1"
}

print_warning() {
    print_message $YELLOW "⚠️  $1"
}

print_error() {
    print_message $RED "❌ $1"
}

# 检查必要的工具
check_prerequisites() {
    print_step "检查部署环境..."
    
    # 检查 kubectl
    if ! command -v kubectl &> /dev/null; then
        print_error "kubectl 未安装，请先安装 kubectl"
        exit 1
    fi
    
    # 检查 docker
    if ! command -v docker &> /dev/null; then
        print_error "Docker 未安装，请先安装 Docker"
        exit 1
    fi
    
    # 检查 Kubernetes 集群连接
    if ! kubectl cluster-info &> /dev/null; then
        print_error "无法连接到 Kubernetes 集群，请检查 kubeconfig"
        exit 1
    fi
    
    print_success "环境检查通过"
}

# 构建 Docker 镜像
build_images() {
    print_step "构建 Docker 镜像..."
    
    # 构建 config-service
    print_message $YELLOW "构建 config-service..."
    cd go/config-service
    docker build -t multi-agent/config-service:latest .
    cd ../..
    
    # 构建 service-registry
    print_message $YELLOW "构建 service-registry..."
    cd go/service-registry
    docker build -t multi-agent/service-registry:latest .
    cd ../..
    
    # 构建 frontend
    print_message $YELLOW "构建 frontend..."
    cd frontend
    docker build -t multi-agent/frontend:latest .
    cd ..
    
    print_success "所有镜像构建完成"
}

# 部署到 Kubernetes
deploy_to_kubernetes() {
    print_step "部署到 Kubernetes 集群..."
    
    # 创建命名空间和基础资源
    print_message $YELLOW "创建命名空间和配置..."
    kubectl apply -f deploy/kubernetes/multi-agent.yaml
    
    # 等待 PostgreSQL 就绪
    print_message $YELLOW "等待 PostgreSQL 启动..."
    kubectl wait --for=condition=ready pod -l app=postgres -n multi-agent --timeout=300s
    
    # 运行数据库迁移
    print_message $YELLOW "运行数据库迁移..."
    kubectl exec -n multi-agent $(kubectl get pod -l app=postgres -n multi-agent -o jsonpath='{.items[0].metadata.name}') -- psql -U postgres -d multiagent -c "$(cat migrations/postgres/001_initial_schema.sql)"
    kubectl exec -n multi-agent $(kubectl get pod -l app=postgres -n multi-agent -o jsonpath='{.items[0].metadata.name}') -- psql -U postgres -d multiagent -c "$(cat migrations/postgres/002_vector_extensions.sql)"
    
    # 等待所有服务就绪
    print_message $YELLOW "等待服务启动..."
    kubectl wait --for=condition=ready pod -l app=config-service -n multi-agent --timeout=300s
    kubectl wait --for=condition=ready pod -l app=service-registry -n multi-agent --timeout=300s
    kubectl wait --for=condition=ready pod -l app=frontend -n multi-agent --timeout=300s
    
    print_success "部署完成"
}

# 检查部署状态
check_deployment() {
    print_step "检查部署状态..."
    
    echo ""
    print_message $BLUE "=== Pod 状态 ==="
    kubectl get pods -n multi-agent
    
    echo ""
    print_message $BLUE "=== Service 状态 ==="
    kubectl get services -n multi-agent
    
    echo ""
    print_message $BLUE "=== Ingress 状态 ==="
    kubectl get ingress -n multi-agent
    
    # 获取访问地址
    FRONTEND_URL=$(kubectl get service frontend -n multi-agent -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
    if [ -z "$FRONTEND_URL" ]; then
        FRONTEND_URL=$(kubectl get service frontend -n multi-agent -o jsonpath='{.status.loadBalancer.ingress[0].hostname}')
    fi
    
    if [ ! -z "$FRONTEND_URL" ]; then
        echo ""
        print_success "🎉 部署成功！"
        print_message $GREEN "前端访问地址: http://$FRONTEND_URL"
        print_message $GREEN "Config Service API: http://$FRONTEND_URL/api/config"
        print_message $GREEN "Service Registry API: http://$FRONTEND_URL/api/registry"
    else
        print_warning "LoadBalancer IP 还未分配，请稍后检查"
        print_message $YELLOW "使用以下命令检查: kubectl get service frontend -n multi-agent"
    fi
}

# 端口转发（用于本地测试）
port_forward() {
    print_step "设置端口转发（本地测试）..."
    
    print_message $YELLOW "启动端口转发..."
    print_message $GREEN "前端访问地址: http://localhost:3000"
    print_message $GREEN "Config Service API: http://localhost:8080"
    print_message $GREEN "Service Registry API: http://localhost:8081"
    
    # 后台运行端口转发
    kubectl port-forward service/frontend 3000:80 -n multi-agent &
    kubectl port-forward service/config-service 8080:80 -n multi-agent &
    kubectl port-forward service/service-registry 8081:80 -n multi-agent &
    
    print_success "端口转发已启动"
    print_warning "按 Ctrl+C 停止端口转发"
    
    # 等待用户中断
    trap 'kill $(jobs -p); exit' INT
    wait
}

# 清理部署
cleanup() {
    print_step "清理部署..."
    
    kubectl delete namespace multi-agent --ignore-not-found=true
    
    print_success "清理完成"
}

# 显示帮助信息
show_help() {
    echo "Multi-Agent Platform 部署脚本"
    echo ""
    echo "用法: $0 [选项]"
    echo ""
    echo "选项:"
    echo "  deploy      完整部署（构建镜像 + 部署到 K8s）"
    echo "  build       仅构建 Docker 镜像"
    echo "  k8s         仅部署到 Kubernetes"
    echo "  status      检查部署状态"
    echo "  port-forward 设置端口转发（本地测试）"
    echo "  cleanup     清理部署"
    echo "  help        显示此帮助信息"
    echo ""
    echo "示例:"
    echo "  $0 deploy           # 完整部署"
    echo "  $0 build            # 仅构建镜像"
    echo "  $0 port-forward     # 本地测试"
    echo "  $0 cleanup          # 清理环境"
}

# 主函数
main() {
    case "${1:-help}" in
        "deploy")
            check_prerequisites
            build_images
            deploy_to_kubernetes
            check_deployment
            ;;
        "build")
            check_prerequisites
            build_images
            ;;
        "k8s"|"kubernetes")
            check_prerequisites
            deploy_to_kubernetes
            check_deployment
            ;;
        "status")
            check_deployment
            ;;
        "port-forward"|"pf")
            port_forward
            ;;
        "cleanup"|"clean")
            cleanup
            ;;
        "help"|"-h"|"--help")
            show_help
            ;;
        *)
            print_error "未知选项: $1"
            show_help
            exit 1
            ;;
    esac
}

# 执行主函数
main "$@"