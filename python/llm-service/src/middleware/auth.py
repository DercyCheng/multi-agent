"""
Authentication Middleware for Multi-Agent LLM Service
"""

import logging
import time
from typing import Optional, Dict, Any
import jwt
from fastapi import HTTPException, Request, Depends
from fastapi.security import HTTPBearer, HTTPAuthorizationCredentials

from src.config import Settings

logger = logging.getLogger(__name__)

security = HTTPBearer(auto_error=False)

class AuthMiddleware:
    """Authentication middleware"""
    
    def __init__(self):
        self.settings = Settings()
    
    async def __call__(self, request: Request, call_next):
        """Process authentication for each request"""
        
        # Skip auth for health endpoints
        if request.url.path.startswith("/health"):
            return await call_next(request)
        
        try:
            # Extract and validate token
            user_info = await self._validate_request(request)
            
            # Add user info to request state
            request.state.user = user_info
            
            # Process request
            response = await call_next(request)
            
            return response
            
        except HTTPException:
            raise
        except Exception as e:
            logger.error(f"Auth middleware error: {e}")
            raise HTTPException(status_code=500, detail="Authentication error")
    
    async def _validate_request(self, request: Request) -> Dict[str, Any]:
        """Validate request authentication"""
        
        # Check for API key in header
        api_key = request.headers.get(self.settings.security.api_key_header)
        if api_key:
            return await self._validate_api_key(api_key)
        
        # Check for JWT token
        auth_header = request.headers.get("Authorization")
        if auth_header and auth_header.startswith("Bearer "):
            token = auth_header[7:]  # Remove "Bearer " prefix
            return await self._validate_jwt_token(token)
        
        # No authentication provided
        raise HTTPException(
            status_code=401,
            detail="Authentication required"
        )
    
    async def _validate_api_key(self, api_key: str) -> Dict[str, Any]:
        """Validate API key"""
        
        # In production, this would check against a database
        # For now, we'll use a simple validation
        
        if not api_key or len(api_key) < 20:
            raise HTTPException(
                status_code=401,
                detail="Invalid API key format"
            )
        
        # Mock validation - replace with actual API key validation
        return {
            "user_id": "api_user_" + api_key[:8],
            "tenant_id": "tenant_" + api_key[8:16],
            "auth_type": "api_key",
            "permissions": ["read", "write"],
            "rate_limit": 1000
        }
    
    async def _validate_jwt_token(self, token: str) -> Dict[str, Any]:
        """Validate JWT token"""
        
        try:
            # Decode JWT token
            payload = jwt.decode(
                token,
                self.settings.security.jwt_secret,
                algorithms=[self.settings.security.jwt_algorithm]
            )
            
            # Check expiration
            if payload.get("exp", 0) < time.time():
                raise HTTPException(
                    status_code=401,
                    detail="Token expired"
                )
            
            # Extract user information
            return {
                "user_id": payload.get("user_id"),
                "tenant_id": payload.get("tenant_id"),
                "auth_type": "jwt",
                "permissions": payload.get("permissions", []),
                "rate_limit": payload.get("rate_limit", 100)
            }
            
        except jwt.InvalidTokenError as e:
            raise HTTPException(
                status_code=401,
                detail=f"Invalid token: {str(e)}"
            )
        except Exception as e:
            logger.error(f"JWT validation error: {e}")
            raise HTTPException(
                status_code=401,
                detail="Token validation failed"
            )

async def get_current_user(
    request: Request,
    credentials: Optional[HTTPAuthorizationCredentials] = Depends(security)
) -> Dict[str, Any]:
    """Get current authenticated user"""
    
    if hasattr(request.state, 'user'):
        return request.state.user
    
    # If no user in state, try to authenticate
    auth_middleware = AuthMiddleware()
    
    try:
        user_info = await auth_middleware._validate_request(request)
        request.state.user = user_info
        return user_info
    except Exception:
        raise HTTPException(
            status_code=401,
            detail="Authentication required"
        )

def require_permissions(required_permissions: list):
    """Decorator to require specific permissions"""
    
    def decorator(func):
        async def wrapper(*args, **kwargs):
            # Get user from request context
            request = kwargs.get('request') or args[0] if args else None
            
            if not request or not hasattr(request.state, 'user'):
                raise HTTPException(
                    status_code=401,
                    detail="Authentication required"
                )
            
            user = request.state.user
            user_permissions = user.get("permissions", [])
            
            # Check if user has required permissions
            if not all(perm in user_permissions for perm in required_permissions):
                raise HTTPException(
                    status_code=403,
                    detail="Insufficient permissions"
                )
            
            return await func(*args, **kwargs)
        
        return wrapper
    return decorator