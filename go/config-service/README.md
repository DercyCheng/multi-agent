# Feature Flags & Configuration Platform

This directory contains the comprehensive feature flags platform, remote configuration center, and service discovery system for the Multi-Agent platform.

## Architecture Overview

```
go/
├── config-service/           # Remote Configuration Center
│   ├── cmd/main.go          # Configuration service entry point
│   ├── internal/
│   │   ├── config/          # Configuration management
│   │   ├── discovery/       # Service discovery
│   │   ├── flags/           # Feature flags core
│   │   ├── api/             # REST API endpoints
│   │   ├── websocket/       # Real-time updates
│   │   ├── storage/         # Persistent storage
│   │   └── auth/            # Authentication & authorization
│   └── Dockerfile
├── service-registry/         # Service Discovery & Registry
│   ├── cmd/main.go          # Service registry entry point
│   ├── internal/
│   │   ├── registry/        # Service registration
│   │   ├── discovery/       # Service discovery
│   │   ├── health/          # Health checking
│   │   ├── loadbalancer/    # Load balancing
│   │   └── failover/        # Failover handling
│   └── Dockerfile
```

## Features

### Feature Flags Platform
- **Real-time toggling** via WebSocket
- **User targeting** and segmentation
- **A/B testing** support
- **Rollout strategies** (percentage, user groups)
- **Audit logging** for compliance
- **Web UI** for management
- **API endpoints** for programmatic access

### Remote Configuration Center
- **Hot reloading** without service restart
- **Environment isolation** (dev/staging/prod)
- **Version control** with rollback
- **Configuration validation**
- **Change notifications**
- **Backup and restore**

### Service Discovery
- **Automatic registration** on startup
- **Health monitoring** with custom checks
- **Load balancing** strategies
- **Failover** and circuit breaking
- **Service mesh** integration
- **DNS-based** discovery

## Quick Start

1. **Start Configuration Service**:
```bash
cd go/config-service
go run cmd/main.go
```

2. **Start Service Registry**:
```bash
cd go/service-registry
go run cmd/main.go
```

3. **Access Web UI**:
- Feature Flags: http://localhost:8080/ui/flags
- Configuration: http://localhost:8080/ui/config
- Service Discovery: http://localhost:8081/ui/registry

## Integration

Services automatically integrate with the platform through:

1. **Configuration client** for hot reloading
2. **Feature flag client** for runtime toggles
3. **Service discovery client** for finding dependencies
4. **Health check** endpoints for monitoring

See individual service README files for detailed documentation.