"""
Rate Limiting Middleware for Multi-Agent LLM Service
"""

import asyncio
import logging
import time
from typing import Dict, Optional
from fastapi import HTTPException, Request, Response
import redis.asyncio as redis

logger = logging.getLogger(__name__)

class RateLimitMiddleware:
    """Rate limiting middleware with Redis backend"""
    
    def __init__(self):
        self.redis_client: Optional[redis.Redis] = None
        self.fallback_limiter = {}  # In-memory fallback
        self.default_limits = {
            "requests_per_minute": 60,
            "requests_per_hour": 1000,
            "requests_per_day": 10000
        }
    
    async def __call__(self, request: Request, call_next):
        """Process rate limiting for each request"""
        
        # Skip rate limiting for health endpoints
        if request.url.path.startswith("/health"):
            return await call_next(request)
        
        try:
            # Initialize Redis if not done
            if not self.redis_client:
                await self._init_redis()
            
            # Get user identifier
            user_id = self._get_user_identifier(request)
            
            # Check rate limits
            await self._check_rate_limits(user_id, request)
            
            # Process request
            start_time = time.time()
            response = await call_next(request)
            processing_time = time.time() - start_time
            
            # Add rate limit headers
            self._add_rate_limit_headers(response, user_id)
            
            # Record request
            await self._record_request(user_id, processing_time)
            
            return response
            
        except HTTPException:
            raise
        except Exception as e:
            logger.error(f"Rate limit middleware error: {e}")
            # Continue without rate limiting on error
            return await call_next(request)
    
    async def _init_redis(self):
        """Initialize Redis connection"""
        try:
            self.redis_client = redis.from_url(
                "redis://localhost:6379/3",  # Use database 3 for rate limiting
                decode_responses=True
            )
            await self.redis_client.ping()
            logger.info("Rate limiter Redis connection established")
        except Exception as e:
            logger.warning(f"Failed to connect to Redis for rate limiting: {e}")
            self.redis_client = None
    
    def _get_user_identifier(self, request: Request) -> str:
        """Get user identifier for rate limiting"""
        
        # Try to get user from auth middleware
        if hasattr(request.state, 'user'):
            user = request.state.user
            return f"user:{user.get('tenant_id', 'unknown')}:{user.get('user_id', 'unknown')}"
        
        # Fallback to IP address
        client_ip = request.client.host if request.client else "unknown"
        forwarded_for = request.headers.get("X-Forwarded-For")
        if forwarded_for:
            client_ip = forwarded_for.split(",")[0].strip()
        
        return f"ip:{client_ip}"
    
    async def _check_rate_limits(self, user_id: str, request: Request):
        """Check if user is within rate limits"""
        
        # Get user-specific limits
        limits = await self._get_user_limits(user_id, request)
        
        # Check each time window
        for window, limit in limits.items():
            if not await self._is_within_limit(user_id, window, limit):
                # Get remaining time until reset
                reset_time = await self._get_reset_time(user_id, window)
                
                raise HTTPException(
                    status_code=429,
                    detail=f"Rate limit exceeded for {window}",
                    headers={
                        "Retry-After": str(int(reset_time)),
                        "X-RateLimit-Limit": str(limit),
                        "X-RateLimit-Remaining": "0",
                        "X-RateLimit-Reset": str(int(time.time() + reset_time))
                    }
                )
    
    async def _get_user_limits(self, user_id: str, request: Request) -> Dict[str, int]:
        """Get rate limits for user"""
        
        # Default limits
        limits = self.default_limits.copy()
        
        # Get user-specific limits from auth info
        if hasattr(request.state, 'user'):
            user = request.state.user
            user_rate_limit = user.get("rate_limit")
            
            if user_rate_limit:
                # Scale limits based on user tier
                multiplier = user_rate_limit / 100  # Base rate is 100
                limits = {
                    "requests_per_minute": int(60 * multiplier),
                    "requests_per_hour": int(1000 * multiplier),
                    "requests_per_day": int(10000 * multiplier)
                }
        
        return limits
    
    async def _is_within_limit(self, user_id: str, window: str, limit: int) -> bool:
        """Check if user is within limit for given window"""
        
        if self.redis_client:
            return await self._redis_check_limit(user_id, window, limit)
        else:
            return await self._memory_check_limit(user_id, window, limit)
    
    async def _redis_check_limit(self, user_id: str, window: str, limit: int) -> bool:
        """Check limit using Redis"""
        
        try:
            # Get window duration in seconds
            window_seconds = self._get_window_seconds(window)
            
            # Redis key for this user and window
            key = f"rate_limit:{user_id}:{window}"
            
            # Use Redis pipeline for atomic operations
            pipe = self.redis_client.pipeline()
            
            # Increment counter
            pipe.incr(key)
            
            # Set expiration if key is new
            pipe.expire(key, window_seconds)
            
            # Execute pipeline
            results = await pipe.execute()
            
            current_count = results[0]
            
            return current_count <= limit
            
        except Exception as e:
            logger.error(f"Redis rate limit check error: {e}")
            return True  # Allow request on error
    
    async def _memory_check_limit(self, user_id: str, window: str, limit: int) -> bool:
        """Check limit using in-memory storage (fallback)"""
        
        current_time = time.time()
        window_seconds = self._get_window_seconds(window)
        
        # Initialize user data if not exists
        if user_id not in self.fallback_limiter:
            self.fallback_limiter[user_id] = {}
        
        user_data = self.fallback_limiter[user_id]
        
        # Initialize window data if not exists
        if window not in user_data:
            user_data[window] = {"count": 0, "reset_time": current_time + window_seconds}
        
        window_data = user_data[window]
        
        # Reset counter if window expired
        if current_time >= window_data["reset_time"]:
            window_data["count"] = 0
            window_data["reset_time"] = current_time + window_seconds
        
        # Check limit
        if window_data["count"] >= limit:
            return False
        
        # Increment counter
        window_data["count"] += 1
        
        return True
    
    def _get_window_seconds(self, window: str) -> int:
        """Get window duration in seconds"""
        
        if window == "requests_per_minute":
            return 60
        elif window == "requests_per_hour":
            return 3600
        elif window == "requests_per_day":
            return 86400
        else:
            return 60  # Default to 1 minute
    
    async def _get_reset_time(self, user_id: str, window: str) -> int:
        """Get time until rate limit reset"""
        
        if self.redis_client:
            try:
                key = f"rate_limit:{user_id}:{window}"
                ttl = await self.redis_client.ttl(key)
                return max(0, ttl)
            except Exception:
                pass
        
        # Fallback to memory storage
        if user_id in self.fallback_limiter and window in self.fallback_limiter[user_id]:
            reset_time = self.fallback_limiter[user_id][window]["reset_time"]
            return max(0, int(reset_time - time.time()))
        
        return 0
    
    def _add_rate_limit_headers(self, response: Response, user_id: str):
        """Add rate limit headers to response"""
        
        # This would typically get current usage from Redis
        # For now, we'll add basic headers
        response.headers["X-RateLimit-Limit"] = "1000"
        response.headers["X-RateLimit-Remaining"] = "999"
        response.headers["X-RateLimit-Reset"] = str(int(time.time() + 3600))
    
    async def _record_request(self, user_id: str, processing_time: float):
        """Record request for analytics"""
        
        if self.redis_client:
            try:
                # Record request metrics
                metrics_key = f"metrics:{user_id}:requests"
                
                pipe = self.redis_client.pipeline()
                
                # Increment request count
                pipe.incr(f"{metrics_key}:count")
                
                # Add processing time to list (keep last 100)
                pipe.lpush(f"{metrics_key}:times", processing_time)
                pipe.ltrim(f"{metrics_key}:times", 0, 99)
                
                # Set expiration
                pipe.expire(f"{metrics_key}:count", 86400)  # 24 hours
                pipe.expire(f"{metrics_key}:times", 86400)
                
                await pipe.execute()
                
            except Exception as e:
                logger.error(f"Failed to record request metrics: {e}")
    
    async def get_user_stats(self, user_id: str) -> Dict[str, any]:
        """Get rate limiting stats for user"""
        
        stats = {
            "current_limits": self.default_limits,
            "current_usage": {},
            "reset_times": {}
        }
        
        if self.redis_client:
            try:
                for window in self.default_limits.keys():
                    key = f"rate_limit:{user_id}:{window}"
                    
                    # Get current count
                    count = await self.redis_client.get(key)
                    stats["current_usage"][window] = int(count) if count else 0
                    
                    # Get reset time
                    ttl = await self.redis_client.ttl(key)
                    stats["reset_times"][window] = int(time.time() + ttl) if ttl > 0 else 0
                    
            except Exception as e:
                logger.error(f"Failed to get user stats: {e}")
        
        return stats