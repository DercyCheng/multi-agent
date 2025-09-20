"""
Health Check API Routes for Multi-Agent LLM Service
"""

import asyncio
import logging
import time
from typing import Dict, Any

from fastapi import APIRouter, Request, Depends
from fastapi.responses import JSONResponse

from src.core.models import HealthStatus, ProviderStatus
from src.core.model_router import ModelRouter
from src.core.context_engine import ContextEngine
from src.core.token_manager import TokenManager
from src.integrations.mcp_client import MCPClient

logger = logging.getLogger(__name__)

router = APIRouter()

# Service start time for uptime calculation
SERVICE_START_TIME = time.time()

@router.get("/", response_model=HealthStatus)
async def health_check(request: Request):
    """Basic health check"""
    
    uptime = int(time.time() - SERVICE_START_TIME)
    
    return HealthStatus(
        status="healthy",
        uptime_seconds=uptime
    )

@router.get("/ready")
async def readiness_check(request: Request):
    """Readiness check for Kubernetes"""
    
    try:
        # Check if all required services are initialized
        required_services = ["model_router", "context_engine", "token_manager", "mcp_client"]
        
        for service_name in required_services:
            if not hasattr(request.app.state, service_name):
                return JSONResponse(
                    status_code=503,
                    content={"status": "not_ready", "missing_service": service_name}
                )
        
        return {"status": "ready"}
        
    except Exception as e:
        logger.error(f"Readiness check failed: {e}")
        return JSONResponse(
            status_code=503,
            content={"status": "not_ready", "error": str(e)}
        )

@router.get("/live")
async def liveness_check():
    """Liveness check for Kubernetes"""
    return {"status": "alive", "timestamp": int(time.time())}

@router.get("/detailed")
async def detailed_health_check(request: Request):
    """Detailed health check with component status"""
    
    try:
        health_data = {
            "status": "healthy",
            "timestamp": int(time.time()),
            "uptime_seconds": int(time.time() - SERVICE_START_TIME),
            "version": "1.0.0",
            "components": {}
        }
        
        overall_healthy = True
        
        # Check model router
        if hasattr(request.app.state, 'model_router'):
            model_router = request.app.state.model_router
            router_health = await _check_model_router_health(model_router)
            health_data["components"]["model_router"] = router_health
            if router_health["status"] != "healthy":
                overall_healthy = False
        else:
            health_data["components"]["model_router"] = {"status": "missing"}
            overall_healthy = False
        
        # Check context engine
        if hasattr(request.app.state, 'context_engine'):
            context_engine = request.app.state.context_engine
            context_health = await _check_context_engine_health(context_engine)
            health_data["components"]["context_engine"] = context_health
            if context_health["status"] != "healthy":
                overall_healthy = False
        else:
            health_data["components"]["context_engine"] = {"status": "missing"}
            overall_healthy = False
        
        # Check token manager
        if hasattr(request.app.state, 'token_manager'):
            token_manager = request.app.state.token_manager
            token_health = await _check_token_manager_health(token_manager)
            health_data["components"]["token_manager"] = token_health
            if token_health["status"] != "healthy":
                overall_healthy = False
        else:
            health_data["components"]["token_manager"] = {"status": "missing"}
            overall_healthy = False
        
        # Check MCP client
        if hasattr(request.app.state, 'mcp_client'):
            mcp_client = request.app.state.mcp_client
            mcp_health = await _check_mcp_client_health(mcp_client)
            health_data["components"]["mcp_client"] = mcp_health
            if mcp_health["status"] != "healthy":
                overall_healthy = False
        else:
            health_data["components"]["mcp_client"] = {"status": "missing"}
            overall_healthy = False
        
        # Set overall status
        if not overall_healthy:
            health_data["status"] = "degraded"
        
        status_code = 200 if overall_healthy else 503
        
        return JSONResponse(
            status_code=status_code,
            content=health_data
        )
        
    except Exception as e:
        logger.error(f"Detailed health check failed: {e}")
        return JSONResponse(
            status_code=500,
            content={
                "status": "unhealthy",
                "error": str(e),
                "timestamp": int(time.time())
            }
        )

@router.get("/providers")
async def providers_health_check(request: Request):
    """Check health of all LLM providers"""
    
    try:
        if not hasattr(request.app.state, 'model_router'):
            return JSONResponse(
                status_code=503,
                content={"error": "Model router not available"}
            )
        
        model_router = request.app.state.model_router
        providers_status = []
        
        # Check each provider
        for provider_name, provider in model_router.providers.items():
            try:
                # Run health check with timeout
                is_healthy = await asyncio.wait_for(
                    provider.health_check(),
                    timeout=10.0
                )
                
                # Get provider stats
                models = await provider.list_models()
                
                provider_status = ProviderStatus(
                    name=provider_name,
                    enabled=provider.enabled,
                    healthy=is_healthy,
                    models_count=len(models),
                    requests_per_minute=provider.rate_limit
                )
                
                providers_status.append(provider_status.dict())
                
            except asyncio.TimeoutError:
                provider_status = ProviderStatus(
                    name=provider_name,
                    enabled=provider.enabled,
                    healthy=False,
                    error_message="Health check timeout",
                    models_count=0
                )
                providers_status.append(provider_status.dict())
                
            except Exception as e:
                provider_status = ProviderStatus(
                    name=provider_name,
                    enabled=provider.enabled,
                    healthy=False,
                    error_message=str(e),
                    models_count=0
                )
                providers_status.append(provider_status.dict())
        
        # Calculate overall provider health
        healthy_providers = sum(1 for p in providers_status if p["healthy"])
        total_providers = len(providers_status)
        
        overall_status = "healthy" if healthy_providers == total_providers else "degraded" if healthy_providers > 0 else "unhealthy"
        
        return {
            "status": overall_status,
            "healthy_providers": healthy_providers,
            "total_providers": total_providers,
            "providers": providers_status,
            "timestamp": int(time.time())
        }
        
    except Exception as e:
        logger.error(f"Providers health check failed: {e}")
        return JSONResponse(
            status_code=500,
            content={"error": str(e)}
        )

async def _check_model_router_health(model_router: ModelRouter) -> Dict[str, Any]:
    """Check model router health"""
    
    try:
        # Check if providers are available
        if not model_router.providers:
            return {
                "status": "unhealthy",
                "error": "No providers available"
            }
        
        # Check if models are loaded
        if not model_router.models:
            return {
                "status": "unhealthy",
                "error": "No models loaded"
            }
        
        # Get basic stats
        stats = await model_router.get_model_stats()
        
        return {
            "status": "healthy",
            "providers_count": len(model_router.providers),
            "models_count": len(model_router.models),
            "circuit_breakers_open": sum(1 for cb in model_router.circuit_breakers.values() if cb)
        }
        
    except Exception as e:
        return {
            "status": "unhealthy",
            "error": str(e)
        }

async def _check_context_engine_health(context_engine: ContextEngine) -> Dict[str, Any]:
    """Check context engine health"""
    
    try:
        # Check database connection
        db_healthy = context_engine.db_pool is not None
        
        # Check vector database connection
        vector_healthy = context_engine.vector_client is not None
        
        # Check embedding model
        embedding_healthy = context_engine.embedding_model is not None
        
        if not all([db_healthy, vector_healthy, embedding_healthy]):
            return {
                "status": "degraded",
                "database": db_healthy,
                "vector_db": vector_healthy,
                "embedding_model": embedding_healthy
            }
        
        return {
            "status": "healthy",
            "database": db_healthy,
            "vector_db": vector_healthy,
            "embedding_model": embedding_healthy,
            "template_cache_size": len(context_engine.template_cache),
            "memory_cache_size": len(context_engine.memory_cache)
        }
        
    except Exception as e:
        return {
            "status": "unhealthy",
            "error": str(e)
        }

async def _check_token_manager_health(token_manager: TokenManager) -> Dict[str, Any]:
    """Check token manager health"""
    
    try:
        # Check database connection
        db_healthy = token_manager.db_pool is not None
        
        # Check Redis connection
        redis_healthy = token_manager.redis_client is not None
        
        if redis_healthy:
            try:
                await token_manager.redis_client.ping()
            except Exception:
                redis_healthy = False
        
        if not all([db_healthy, redis_healthy]):
            return {
                "status": "degraded",
                "database": db_healthy,
                "redis": redis_healthy
            }
        
        return {
            "status": "healthy",
            "database": db_healthy,
            "redis": redis_healthy,
            "budget_cache_size": len(token_manager.budget_cache),
            "usage_cache_size": len(token_manager.usage_cache)
        }
        
    except Exception as e:
        return {
            "status": "unhealthy",
            "error": str(e)
        }

async def _check_mcp_client_health(mcp_client: MCPClient) -> Dict[str, Any]:
    """Check MCP client health"""
    
    try:
        if not mcp_client.config.enabled:
            return {
                "status": "disabled",
                "enabled": False
            }
        
        # Check available tools
        tools_count = len(mcp_client.available_tools)
        
        # Check active sessions
        active_sessions = len([s for s in mcp_client.sessions.values() if s.is_active])
        
        return {
            "status": "healthy",
            "enabled": True,
            "tools_count": tools_count,
            "active_sessions": active_sessions,
            "max_connections": mcp_client.config.max_connections
        }
        
    except Exception as e:
        return {
            "status": "unhealthy",
            "error": str(e)
        }