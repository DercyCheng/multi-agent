"""
Base Provider Interface for Multi-Agent LLM Service
Abstract base class for all LLM providers
"""

from abc import ABC, abstractmethod
from typing import List, Optional, AsyncGenerator
import asyncio
import logging

from src.config import ModelProviderConfig
from src.core.models import CompletionRequest, CompletionResponse, StreamResponse, ModelInfo

logger = logging.getLogger(__name__)

class BaseProvider(ABC):
    """Abstract base class for LLM providers"""
    
    def __init__(self, config: ModelProviderConfig):
        self.config = config
        self.name = config.name
        self.api_key = config.api_key
        self.base_url = config.base_url
        self.models = config.models
        self.rate_limit = config.rate_limit
        self.timeout = config.timeout
        self.max_retries = config.max_retries
        self.enabled = config.enabled
        
        # Rate limiting
        self._request_semaphore = asyncio.Semaphore(rate_limit)
        self._last_request_time = 0.0
        
    @abstractmethod
    async def complete(self, request: CompletionRequest) -> CompletionResponse:
        """Complete a chat completion request"""
        pass
    
    @abstractmethod
    async def stream_complete(self, request: CompletionRequest) -> AsyncGenerator[StreamResponse, None]:
        """Stream a chat completion request"""
        pass
    
    @abstractmethod
    async def list_models(self) -> List[ModelInfo]:
        """List available models"""
        pass
    
    @abstractmethod
    async def validate_api_key(self) -> bool:
        """Validate API key"""
        pass
    
    async def health_check(self) -> bool:
        """Check provider health"""
        try:
            return await self.validate_api_key()
        except Exception as e:
            logger.error(f"Health check failed for {self.name}: {e}")
            return False
    
    async def _rate_limit(self):
        """Apply rate limiting"""
        async with self._request_semaphore:
            import time
            current_time = time.time()
            time_since_last = current_time - self._last_request_time
            min_interval = 60.0 / self.rate_limit  # seconds between requests
            
            if time_since_last < min_interval:
                await asyncio.sleep(min_interval - time_since_last)
            
            self._last_request_time = time.time()
    
    async def _retry_request(self, func, *args, **kwargs):
        """Retry request with exponential backoff"""
        last_exception = None
        
        for attempt in range(self.max_retries + 1):
            try:
                await self._rate_limit()
                return await func(*args, **kwargs)
            except Exception as e:
                last_exception = e
                
                if attempt < self.max_retries:
                    # Exponential backoff
                    delay = 2 ** attempt
                    logger.warning(f"Request failed (attempt {attempt + 1}), retrying in {delay}s: {e}")
                    await asyncio.sleep(delay)
                else:
                    logger.error(f"Request failed after {self.max_retries + 1} attempts: {e}")
        
        raise last_exception
    
    def _format_messages(self, messages: List) -> List:
        """Format messages for provider-specific format"""
        # Default implementation - override in subclasses if needed
        return [{"role": msg.role, "content": msg.content} for msg in messages]
    
    def _calculate_tokens(self, text: str) -> int:
        """Rough token calculation - override in subclasses for accuracy"""
        return len(text) // 4
    
    async def shutdown(self):
        """Shutdown provider"""
        logger.info(f"Shutting down {self.name} provider")
        # Override in subclasses if cleanup needed