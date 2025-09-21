# Kubernetes Multi-Instance Deployment Guide

This guide covers the enhanced Kubernetes and Helm deployment for the Multi-Agent LLM Service, featuring advanced multi-instance deployment capabilities, horizontal scaling, service discovery, load balancing, and comprehensive monitoring.

## 🚀 Features

### Core Deployment Features
- **Multi-Instance Deployment**: Horizontal scaling with Pod Anti-Affinity
- **Autoscaling**: Horizontal Pod Autoscaler (HPA) with CPU and memory metrics
- **Load Balancing**: Intelligent load balancing with health checks
- **Service Discovery**: Automatic service registration and discovery
- **High Availability**: Pod Disruption Budgets and rolling updates

### Advanced Configuration
- **Multi-Environment Support**: Development, Staging, and Production configurations
- **Security**: Network policies, security contexts, and RBAC
- **Storage**: Persistent volumes with storage class support
- **Monitoring**: Prometheus metrics, Grafana dashboards, and Jaeger tracing
- **Dependency Management**: Redis caching and PostgreSQL database

## 📁 Directory Structure

```
deploy/helm/llm-service/
├── Chart.yaml                     # Helm chart metadata
├── values.yaml                    # Default values
├── values-dev.yaml                # Development environment
├── values-staging.yaml            # Staging environment  
├── values-prod.yaml               # Production environment
└── templates/
    ├── _helpers.tpl               # Template helpers
    ├── deployment.yaml            # Main deployment
    ├── service.yaml               # Service definition
    ├── hpa.yaml                   # Horizontal Pod Autoscaler
    ├── pvc.yaml                   # Persistent Volume Claims
    ├── configmap.yaml             # Configuration
    ├── secret.yaml                # Secrets
    ├── serviceaccount.yaml        # Service Account
    ├── poddisruptionbudget.yaml   # Pod Disruption Budget
    ├── servicemonitor.yaml        # Prometheus ServiceMonitor
    ├── ingress.yaml               # Ingress configuration
    ├── networkpolicy.yaml         # Network policies
    └── cronjob.yaml               # CronJob definition
```

## 🛠️ Prerequisites

### Required Tools
- **Kubernetes cluster** (v1.20+)
- **Helm** (v3.0+)
- **kubectl** configured with cluster access

### Optional Dependencies
- **Prometheus Operator** (for metrics)
- **cert-manager** (for TLS certificates)
- **nginx-ingress** or similar ingress controller

## 🚀 Quick Start

### 1. Install Dependencies

```bash
# Add Bitnami repository for Redis and PostgreSQL
helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo update
```

### 2. Deploy to Development

```bash
# Using the deployment script
./scripts/deploy-helm.sh dev

# Or manually with Helm
helm install llm-service-dev ./deploy/helm/llm-service \
  --namespace llm-service-dev \
  --create-namespace \
  --values ./deploy/helm/llm-service/values-dev.yaml
```

### 3. Deploy to Staging

```bash
# With automatic upgrade and wait
./scripts/deploy-helm.sh staging --upgrade --wait

# Check deployment status
./scripts/monitor-scale.sh staging status
```

### 4. Deploy to Production

```bash
# Production deployment with custom timeout
./scripts/deploy-helm.sh prod --upgrade --wait --timeout 900s

# Enable autoscaling
./scripts/monitor-scale.sh prod hpa status
```

## 📊 Environment Configurations

### Development Environment
- **Replicas**: 1 (no autoscaling)
- **Resources**: 500m CPU, 1Gi memory
- **Storage**: Disabled (ephemeral)
- **Monitoring**: Basic Prometheus only
- **Security**: Minimal (for development ease)

### Staging Environment
- **Replicas**: 2-5 (autoscaling enabled)
- **Resources**: 750m CPU, 1.5Gi memory
- **Storage**: 5Gi persistent volume
- **Monitoring**: Full monitoring with Grafana
- **Security**: Network policies enabled

### Production Environment
- **Replicas**: 3-20 (advanced autoscaling)
- **Resources**: 2000m CPU, 4Gi memory
- **Storage**: 50Gi premium SSD
- **Monitoring**: Complete observability stack
- **Security**: Full security hardening

## 🔧 Configuration Options

### Core Service Configuration

```yaml
# Replica and scaling configuration
replicaCount: 3
autoscaling:
  enabled: true
  minReplicas: 2
  maxReplicas: 10
  targetCPUUtilizationPercentage: 70
  targetMemoryUtilizationPercentage: 80

# Resource limits
resources:
  limits:
    cpu: 1000m
    memory: 2Gi
  requests:
    cpu: 500m
    memory: 1Gi
```

### Multi-Model Ollama Configuration

```yaml
ollama:
  enabled: true
  baseUrl: http://ollama:11434
  models:
    - mistral
    - codellama
    - llama2
    - mixtral
  healthCheck:
    enabled: true
    interval: 60s
    timeout: 10s
  loadBalancing:
    enabled: true
    strategy: least_connections
  multiServer:
    enabled: true
    servers:
      - url: http://ollama-1:11434
        weight: 1
      - url: http://ollama-2:11434
        weight: 1
```

### Database and Caching

```yaml
# PostgreSQL database
database:
  enabled: true
  type: postgresql
  host: postgresql
  port: 5432
  database: llm_service
  username: llm_user
  ssl: true
  maxConnections: 100

# Redis caching
redis:
  enabled: true
  auth:
    enabled: true
  cluster:
    enabled: true
  master:
    persistence:
      enabled: true
      size: 20Gi
```

### Monitoring and Observability

```yaml
monitoring:
  enabled: true
  prometheus:
    enabled: true
    path: /metrics
    port: 9090
    interval: 30s
  grafana:
    enabled: true
    dashboards:
      enabled: true
  jaeger:
    enabled: true
    agent:
      host: jaeger-agent
      port: 6831
```

## 📈 Scaling and Management

### Manual Scaling

```bash
# Scale to specific replica count
./scripts/monitor-scale.sh prod scale 5

# Check current status
./scripts/monitor-scale.sh prod status
```

### Autoscaling Management

```bash
# Check HPA status
./scripts/monitor-scale.sh prod hpa status

# View metrics
./scripts/monitor-scale.sh prod metrics
```

### Rolling Updates

```bash
# Update with new image tag
./scripts/deploy-helm.sh prod --upgrade --set image.tag=v2.0.0

# Rollback if needed
./scripts/deploy-helm.sh prod --rollback
```

## 🔍 Monitoring and Debugging

### View Logs

```bash
# Real-time logs
./scripts/monitor-scale.sh prod logs --follow

# Last 100 lines
./scripts/monitor-scale.sh prod logs --tail=100
```

### Resource Monitoring

```bash
# Resource usage
./scripts/monitor-scale.sh prod metrics

# Detailed pod information
./scripts/monitor-scale.sh prod describe pod
```

### Debug Issues

```bash
# Comprehensive debugging
./scripts/monitor-scale.sh prod debug

# View recent events
./scripts/monitor-scale.sh prod events
```

### Port Forwarding for Testing

```bash
# Forward local port 8080 to service port 8000
./scripts/monitor-scale.sh prod port-forward 8080:8000

# Test the service
curl http://localhost:8080/health
```

## 🔒 Security Features

### Network Policies
- Ingress controls for external access
- Egress controls for external services
- Pod-to-pod communication restrictions

### Security Contexts
- Non-root user execution
- Read-only root filesystem
- Dropped capabilities
- Seccomp profiles

### RBAC and Service Accounts
- Minimal service account permissions
- Role-based access control
- Secret management

## 🗄️ Storage Management

### Persistent Volumes

```yaml
persistence:
  enabled: true
  storageClass: premium-ssd
  accessMode: ReadWriteOnce
  size: 50Gi
  annotations:
    volume.beta.kubernetes.io/storage-class: "premium-ssd"
```

### Backup and Recovery

```bash
# Create volume snapshot (if supported)
kubectl create volumesnapshot llm-service-backup \
  --source-pvc llm-service-prod-pvc \
  --namespace llm-service-prod
```

## 🌐 Load Balancing and Service Discovery

### Load Balancer Configuration

```yaml
loadBalancer:
  enabled: true
  strategy: round_robin  # round_robin, least_connections, ip_hash
  healthCheck:
    enabled: true
    path: /health
    interval: 30s
    timeout: 5s
    failureThreshold: 3
```

### Service Discovery

```yaml
serviceDiscovery:
  enabled: true
  consul:
    enabled: false
    address: consul:8500
  eureka:
    enabled: false
    serviceUrl: http://eureka:8761/eureka
```

## 📋 Production Checklist

### Pre-Deployment
- [ ] Cluster resources sufficient
- [ ] Storage classes configured
- [ ] Monitoring stack deployed
- [ ] SSL certificates ready
- [ ] Database migrations completed
- [ ] Secrets configured

### Post-Deployment
- [ ] Health checks passing
- [ ] Metrics collecting
- [ ] Logs flowing
- [ ] Autoscaling functioning
- [ ] Load balancing working
- [ ] Backup configured

## 🆘 Troubleshooting

### Common Issues

**Pods not starting:**
```bash
# Check pod status and events
./scripts/monitor-scale.sh prod describe pod
./scripts/monitor-scale.sh prod events
```

**Resource constraints:**
```bash
# Check resource usage
./scripts/monitor-scale.sh prod metrics
kubectl describe nodes
```

**Storage issues:**
```bash
# Check PVC status
kubectl get pvc -n llm-service-prod
kubectl describe pvc llm-service-prod-pvc -n llm-service-prod
```

**Network connectivity:**
```bash
# Test service connectivity
kubectl run debug --rm -it --image=busybox -- /bin/sh
nslookup llm-service-prod.llm-service-prod.svc.cluster.local
```

### Performance Tuning

**CPU/Memory optimization:**
```yaml
resources:
  limits:
    cpu: 2000m      # Adjust based on workload
    memory: 4Gi     # Monitor actual usage
  requests:
    cpu: 1000m      # Start with 50% of limits
    memory: 2Gi
```

**Autoscaling tuning:**
```yaml
autoscaling:
  targetCPUUtilizationPercentage: 70    # Lower for faster scaling
  targetMemoryUtilizationPercentage: 80
  behavior:
    scaleUp:
      stabilizationWindowSeconds: 60    # Faster scale-up
    scaleDown:
      stabilizationWindowSeconds: 300   # Slower scale-down
```

## 📚 Additional Resources

### Helm Commands Reference
```bash
# List releases
helm list --all-namespaces

# Get values
helm get values llm-service-prod -n llm-service-prod

# Upgrade with dry-run
helm upgrade llm-service-prod ./deploy/helm/llm-service \
  --values values-prod.yaml --dry-run

# Rollback
helm rollback llm-service-prod 1 -n llm-service-prod
```

### Kubectl Commands Reference
```bash
# Resource monitoring
kubectl top pods -n llm-service-prod
kubectl top nodes

# Log aggregation
kubectl logs -l app.kubernetes.io/name=llm-service -n llm-service-prod

# Port forwarding
kubectl port-forward svc/llm-service-prod 8080:8000 -n llm-service-prod
```

## 🤝 Contributing

For improvements to the Kubernetes deployment:

1. Test changes in development environment
2. Validate with staging deployment
3. Update documentation
4. Submit pull request with deployment notes

## 📞 Support

For deployment issues:
- Check troubleshooting section
- Use debug scripts provided
- Review cluster logs and events
- Contact platform team for cluster-level issues