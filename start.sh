#!/bin/bash

# Multi-Agent Platform Quick Start Script
# 多智能体平台快速启动脚本

set -e

echo "🚀 Multi-Agent Platform Quick Start"
echo "=================================="

# 检查 Docker 是否安装
if ! command -v docker &> /dev/null; then
    echo "❌ Docker is not installed. Please install Docker first."
    echo "   Visit: https://docs.docker.com/get-docker/"
    exit 1
fi

# 检查 Docker Compose 是否安装
if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
    echo "❌ Docker Compose is not installed. Please install Docker Compose first."
    echo "   Visit: https://docs.docker.com/compose/install/"
    exit 1
fi

# 检查环境变量文件
if [ ! -f ".env" ]; then
    echo "⚠️  .env file not found. Creating from template..."
    if [ -f ".env.example" ]; then
        cp .env.example .env
        echo "✅ Created .env file from template"
        echo "📝 Please edit .env file and add your API keys"
    else
        echo "❌ .env.example template not found"
        exit 1
    fi
fi

# 获取启动模式
MODE=${1:-dev}

echo ""
echo "🔧 Starting Multi-Agent Platform in $MODE mode..."

case $MODE in
    "dev"|"development")
        echo "🛠️  Development mode - includes hot reload and debug logs"
        docker-compose up -d
        ;;
    "prod"|"production")
        echo "🏭 Production mode - optimized for performance"
        docker-compose -f docker-compose.prod.yml up -d
        ;;
    *)
        echo "❌ Invalid mode: $MODE"
        echo "Usage: $0 [dev|prod]"
        exit 1
        ;;
esac

echo ""
echo "⏳ Waiting for services to start..."
sleep 10

# 检查服务状态
echo ""
echo "🔍 Checking service health..."

# 检查前端
if curl -s http://localhost:3000 > /dev/null; then
    echo "✅ Frontend (React): http://localhost:3000"
else
    echo "❌ Frontend service not responding"
fi

# 检查API网关
if curl -s http://localhost:8080/health > /dev/null; then
    echo "✅ API Gateway: http://localhost:8080"
else
    echo "❌ API Gateway not responding"
fi

# 检查监控服务
if curl -s http://localhost:3001 > /dev/null; then
    echo "✅ Grafana Monitoring: http://localhost:3001"
else
    echo "⚠️  Grafana not responding (optional service)"
fi

if curl -s http://localhost:9090 > /dev/null; then
    echo "✅ Prometheus Metrics: http://localhost:9090"
else
    echo "⚠️  Prometheus not responding (optional service)"
fi

echo ""
echo "🎉 Multi-Agent Platform Started Successfully!"
echo ""
echo "📱 Access Points:"
echo "   🌐 Web Interface:    http://localhost:3000"
echo "   🔌 API Gateway:      http://localhost:8080"
echo "   📊 Grafana:          http://localhost:3001 (admin/admin)"
echo "   📈 Prometheus:       http://localhost:9090"
echo ""
echo "📚 Next Steps:"
echo "   1. Open http://localhost:3000 in your browser"
echo "   2. Register a new account or login"
echo "   3. Explore the dashboard and create your first agent"
echo "   4. Check docs/ folder for detailed documentation"
echo ""
echo "🛑 To stop all services:"
echo "   docker-compose down"
echo ""
echo "📖 For detailed documentation:"
echo "   - Frontend Guide: frontend/README.md"
echo "   - API Documentation: docs/"
echo "   - Architecture Overview: docs/SHANNON-PLATFORM-ARCHITECTURE.md"