"""
Metrics API Routes for Multi-Agent LLM Service
"""

import logging
from typing import Dict, Any
from fastapi import APIRouter, Request, Depends
from fastapi.responses import PlainTextResponse

from src.middleware.auth import get_current_user
from src.middleware.metrics import get_metrics, get_metrics_content_type
from src.core.model_router import ModelRouter
from src.core.token_manager import TokenManager

logger = logging.getLogger(__name__)

router = APIRouter()

@router.get("/prometheus", response_class=PlainTextResponse)
async def prometheus_metrics():
    """Prometheus metrics endpoint"""
    
    metrics_data = get_metrics()
    return PlainTextResponse(
        content=metrics_data,
        media_type=get_metrics_content_type()
    )

@router.get("/summary")
async def metrics_summary(
    request: Request,
    current_user: dict = Depends(get_current_user)
):
    """Get metrics summary for user"""
    
    try:
        # Get model router stats
        model_stats = {}
        if hasattr(request.app.state, 'model_router'):
            model_router = request.app.state.model_router
            model_stats = await model_router.get_model_stats()
        
        # Get token usage stats
        usage_stats = {}
        if hasattr(request.app.state, 'token_manager'):
            token_manager = request.app.state.token_manager
            usage_stats = await token_manager.get_usage_statistics(
                current_user['user_id'],
                current_user['tenant_id']
            )
        
        return {
            "user_id": current_user['user_id'],
            "tenant_id": current_user['tenant_id'],
            "model_stats": model_stats,
            "usage_stats": usage_stats,
            "timestamp": int(time.time())
        }
        
    except Exception as e:
        logger.error(f"Metrics summary error: {e}")
        raise HTTPException(status_code=500, detail=str(e))

@router.get("/usage/{period}")
async def usage_metrics(
    period: str,
    request: Request,
    current_user: dict = Depends(get_current_user)
):
    """Get usage metrics for specific period"""
    
    try:
        if period not in ["day", "week", "month"]:
            raise HTTPException(status_code=400, detail="Invalid period. Use: day, week, month")
        
        if not hasattr(request.app.state, 'token_manager'):
            raise HTTPException(status_code=503, detail="Token manager not available")
        
        token_manager = request.app.state.token_manager
        
        stats = await token_manager.get_usage_statistics(
            current_user['user_id'],
            current_user['tenant_id'],
            period
        )
        
        return {
            "period": period,
            "user_id": current_user['user_id'],
            "tenant_id": current_user['tenant_id'],
            "stats": stats,
            "timestamp": int(time.time())
        }
        
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Usage metrics error: {e}")
        raise HTTPException(status_code=500, detail=str(e))