"""
Metrics Collection Middleware for Multi-Agent LLM Service
"""

import logging
import time
from typing import Dict, Any
from fastapi import Request, Response
from prometheus_client import Counter, Histogram, Gauge, generate_latest, CONTENT_TYPE_LATEST

logger = logging.getLogger(__name__)

# Prometheus metrics
REQUEST_COUNT = Counter(
    'llm_service_requests_total',
    'Total number of requests',
    ['method', 'endpoint', 'status_code', 'user_id', 'tenant_id']
)

REQUEST_DURATION = Histogram(
    'llm_service_request_duration_seconds',
    'Request duration in seconds',
    ['method', 'endpoint', 'user_id', 'tenant_id']
)

ACTIVE_REQUESTS = Gauge(
    'llm_service_active_requests',
    'Number of active requests',
    ['endpoint']
)

TOKEN_USAGE = Counter(
    'llm_service_tokens_total',
    'Total tokens used',
    ['model', 'provider', 'user_id', 'tenant_id', 'token_type']
)

COST_TRACKING = Counter(
    'llm_service_cost_usd_total',
    'Total cost in USD',
    ['model', 'provider', 'user_id', 'tenant_id']
)

MODEL_REQUESTS = Counter(
    'llm_service_model_requests_total',
    'Total requests per model',
    ['model', 'provider', 'status']
)

CACHE_HITS = Counter(
    'llm_service_cache_hits_total',
    'Cache hits',
    ['cache_type']
)

CACHE_MISSES = Counter(
    'llm_service_cache_misses_total',
    'Cache misses',
    ['cache_type']
)

ERROR_COUNT = Counter(
    'llm_service_errors_total',
    'Total number of errors',
    ['error_type', 'endpoint']
)

class MetricsMiddleware:
    """Metrics collection middleware"""
    
    def __init__(self):
        self.active_requests_count = {}
    
    async def __call__(self, request: Request, call_next):
        """Collect metrics for each request"""
        
        start_time = time.time()
        endpoint = self._get_endpoint_label(request.url.path)
        method = request.method
        
        # Track active requests
        ACTIVE_REQUESTS.labels(endpoint=endpoint).inc()
        
        try:
            # Process request
            response = await call_next(request)
            
            # Calculate duration
            duration = time.time() - start_time
            
            # Get user info
            user_id, tenant_id = self._get_user_info(request)
            
            # Record metrics
            REQUEST_COUNT.labels(
                method=method,
                endpoint=endpoint,
                status_code=response.status_code,
                user_id=user_id,
                tenant_id=tenant_id
            ).inc()
            
            REQUEST_DURATION.labels(
                method=method,
                endpoint=endpoint,
                user_id=user_id,
                tenant_id=tenant_id
            ).observe(duration)
            
            # Add custom headers
            response.headers["X-Response-Time"] = f"{duration:.3f}s"
            
            return response
            
        except Exception as e:
            # Record error
            ERROR_COUNT.labels(
                error_type=type(e).__name__,
                endpoint=endpoint
            ).inc()
            
            raise
            
        finally:
            # Decrease active requests
            ACTIVE_REQUESTS.labels(endpoint=endpoint).dec()
    
    def _get_endpoint_label(self, path: str) -> str:
        """Get normalized endpoint label"""
        
        # Normalize paths for better grouping
        if path.startswith("/v1/chat/completions"):
            return "/v1/chat/completions"
        elif path.startswith("/v1/models"):
            return "/v1/models"
        elif path.startswith("/health"):
            return "/health"
        elif path.startswith("/metrics"):
            return "/metrics"
        else:
            return "other"
    
    def _get_user_info(self, request: Request) -> tuple[str, str]:
        """Get user and tenant ID from request"""
        
        if hasattr(request.state, 'user'):
            user = request.state.user
            return user.get('user_id', 'unknown'), user.get('tenant_id', 'unknown')
        
        return 'anonymous', 'default'

class MetricsCollector:
    """Additional metrics collection utilities"""
    
    @staticmethod
    def record_token_usage(
        model: str,
        provider: str,
        user_id: str,
        tenant_id: str,
        prompt_tokens: int,
        completion_tokens: int
    ):
        """Record token usage metrics"""
        
        TOKEN_USAGE.labels(
            model=model,
            provider=provider,
            user_id=user_id,
            tenant_id=tenant_id,
            token_type='prompt'
        ).inc(prompt_tokens)
        
        TOKEN_USAGE.labels(
            model=model,
            provider=provider,
            user_id=user_id,
            tenant_id=tenant_id,
            token_type='completion'
        ).inc(completion_tokens)
    
    @staticmethod
    def record_cost(
        model: str,
        provider: str,
        user_id: str,
        tenant_id: str,
        cost_usd: float
    ):
        """Record cost metrics"""
        
        COST_TRACKING.labels(
            model=model,
            provider=provider,
            user_id=user_id,
            tenant_id=tenant_id
        ).inc(cost_usd)
    
    @staticmethod
    def record_model_request(model: str, provider: str, status: str):
        """Record model request metrics"""
        
        MODEL_REQUESTS.labels(
            model=model,
            provider=provider,
            status=status
        ).inc()
    
    @staticmethod
    def record_cache_hit(cache_type: str):
        """Record cache hit"""
        
        CACHE_HITS.labels(cache_type=cache_type).inc()
    
    @staticmethod
    def record_cache_miss(cache_type: str):
        """Record cache miss"""
        
        CACHE_MISSES.labels(cache_type=cache_type).inc()
    
    @staticmethod
    def record_error(error_type: str, endpoint: str):
        """Record error"""
        
        ERROR_COUNT.labels(
            error_type=error_type,
            endpoint=endpoint
        ).inc()

def get_metrics() -> str:
    """Get Prometheus metrics"""
    return generate_latest()

def get_metrics_content_type() -> str:
    """Get metrics content type"""
    return CONTENT_TYPE_LATEST