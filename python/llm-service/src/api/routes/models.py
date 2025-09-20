"""
Models API Routes for Multi-Agent LLM Service
"""

import logging
import time
from typing import List, Dict, Any
from fastapi import APIRouter, Request, HTTPException, Depends

from src.middleware.auth import get_current_user
from src.core.model_router import ModelRouter
from src.core.models import ModelInfo

logger = logging.getLogger(__name__)

router = APIRouter()

@router.get("/")
async def list_models(
    request: Request,
    current_user: dict = Depends(get_current_user)
):
    """List all available models"""
    
    try:
        if not hasattr(request.app.state, 'model_router'):
            raise HTTPException(status_code=503, detail="Model router not available")
        
        model_router = request.app.state.model_router
        
        all_models = []
        
        for provider_name, provider in model_router.providers.items():
            try:
                models = await provider.list_models()
                all_models.extend(models)
            except Exception as e:
                logger.warning(f"Failed to get models from {provider_name}: {e}")
        
        return {
            "object": "list",
            "data": [
                {
                    "id": model.id,
                    "object": "model",
                    "created": int(time.time()),
                    "owned_by": model.provider,
                    "permission": [],
                    "root": model.id,
                    "parent": None,
                    "details": {
                        "name": model.name,
                        "description": model.description,
                        "max_tokens": model.max_tokens,
                        "context_length": model.context_length,
                        "cost_per_token": model.cost_per_token,
                        "supports_streaming": model.supports_streaming,
                        "supports_tools": model.supports_tools,
                        "supports_vision": model.supports_vision,
                        "capability_score": model.capability_score
                    }
                }
                for model in all_models
            ]
        }
        
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"List models error: {e}")
        raise HTTPException(status_code=500, detail=str(e))

@router.get("/{model_id}")
async def get_model(
    model_id: str,
    request: Request,
    current_user: dict = Depends(get_current_user)
):
    """Get specific model information"""
    
    try:
        if not hasattr(request.app.state, 'model_router'):
            raise HTTPException(status_code=503, detail="Model router not available")
        
        model_router = request.app.state.model_router
        
        # Find model across all providers
        for provider_name, provider in model_router.providers.items():
            try:
                models = await provider.list_models()
                for model in models:
                    if model.id == model_id:
                        return {
                            "id": model.id,
                            "object": "model",
                            "created": int(time.time()),
                            "owned_by": model.provider,
                            "permission": [],
                            "root": model.id,
                            "parent": None,
                            "details": {
                                "name": model.name,
                                "description": model.description,
                                "max_tokens": model.max_tokens,
                                "context_length": model.context_length,
                                "cost_per_token": model.cost_per_token,
                                "supports_streaming": model.supports_streaming,
                                "supports_tools": model.supports_tools,
                                "supports_vision": model.supports_vision,
                                "capability_score": model.capability_score
                            }
                        }
            except Exception as e:
                logger.warning(f"Failed to get models from {provider_name}: {e}")
        
        raise HTTPException(status_code=404, detail=f"Model {model_id} not found")
        
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Get model error: {e}")
        raise HTTPException(status_code=500, detail=str(e))

@router.get("/{model_id}/stats")
async def get_model_stats(
    model_id: str,
    request: Request,
    current_user: dict = Depends(get_current_user)
):
    """Get model performance statistics"""
    
    try:
        if not hasattr(request.app.state, 'model_router'):
            raise HTTPException(status_code=503, detail="Model router not available")
        
        model_router = request.app.state.model_router
        
        # Get all model stats
        all_stats = await model_router.get_model_stats()
        
        # Find stats for specific model
        for model_key, stats in all_stats.items():
            if model_key.endswith(f":{model_id}") or model_key == model_id:
                return {
                    "model_id": model_id,
                    "model_key": model_key,
                    "stats": stats,
                    "timestamp": int(time.time())
                }
        
        raise HTTPException(status_code=404, detail=f"Stats for model {model_id} not found")
        
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Get model stats error: {e}")
        raise HTTPException(status_code=500, detail=str(e))

@router.get("/providers/status")
async def get_providers_status(
    request: Request,
    current_user: dict = Depends(get_current_user)
):
    """Get status of all model providers"""
    
    try:
        if not hasattr(request.app.state, 'model_router'):
            raise HTTPException(status_code=503, detail="Model router not available")
        
        model_router = request.app.state.model_router
        
        providers_status = []
        
        for provider_name, provider in model_router.providers.items():
            try:
                # Check provider health
                is_healthy = await provider.health_check()
                
                # Get models count
                models = await provider.list_models()
                models_count = len(models)
                
                status = {
                    "name": provider_name,
                    "enabled": provider.enabled,
                    "healthy": is_healthy,
                    "models_count": models_count,
                    "rate_limit": provider.rate_limit,
                    "timeout": provider.timeout,
                    "max_retries": provider.max_retries
                }
                
                providers_status.append(status)
                
            except Exception as e:
                status = {
                    "name": provider_name,
                    "enabled": provider.enabled,
                    "healthy": False,
                    "error": str(e),
                    "models_count": 0
                }
                providers_status.append(status)
        
        return {
            "providers": providers_status,
            "total_providers": len(providers_status),
            "healthy_providers": sum(1 for p in providers_status if p["healthy"]),
            "timestamp": int(time.time())
        }
        
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Get providers status error: {e}")
        raise HTTPException(status_code=500, detail=str(e))