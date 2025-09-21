"""
Adaptive Model Router for Multi-Agent LLM Service
Intelligently routes requests to optimal models based on cost, performance, and availability
"""

import asyncio
import logging
import time
from typing import Dict, List, Optional, Tuple, Any
from dataclasses import dataclass, field
from enum import Enum
import json

from src.config import ModelsConfig, ModelProviderConfig
# Provider imports are intentionally lazy to avoid importing heavy SDKs
# at module import time (which breaks unit tests that don't exercise
# actual provider functionality). Providers will be imported inside
# `_create_provider` when needed.
from src.core.models import CompletionRequest, CompletionResponse, ModelInfo

logger = logging.getLogger(__name__)

class OptimizationStrategy(Enum):
    """Model selection optimization strategies"""
    COST = "cost"
    PERFORMANCE = "performance"
    BALANCED = "balanced"
    AVAILABILITY = "availability"

@dataclass
class PerformanceMetrics:
    """Performance metrics for a model"""
    total_requests: int = 0
    successful_requests: int = 0
    failed_requests: int = 0
    avg_latency: float = 0.0
    avg_tokens_per_second: float = 0.0
    last_updated: float = field(default_factory=time.time)
    
    @property
    def success_rate(self) -> float:
        if self.total_requests == 0:
            return 0.0
        return self.successful_requests / self.total_requests
    
    @property
    def failure_rate(self) -> float:
        return 1.0 - self.success_rate

@dataclass
class LoadMetrics:
    """Load metrics for a model"""
    current_requests: int = 0
    max_concurrent: int = 100
    queue_size: int = 0
    last_request_time: float = 0.0
    
    @property
    def load_factor(self) -> float:
        if self.max_concurrent == 0:
            return 1.0
        return self.current_requests / self.max_concurrent
    
    @property
    def is_overloaded(self) -> bool:
        return self.load_factor > 0.9

class ModelRouter:
    """Adaptive model router with intelligent selection"""
    
    def __init__(self, config: ModelsConfig):
        self.config = config
        self.providers: Dict[str, Any] = {}
        self.models: Dict[str, ModelInfo] = {}
        self.performance_metrics: Dict[str, PerformanceMetrics] = {}
        self.load_metrics: Dict[str, LoadMetrics] = {}
        self.circuit_breakers: Dict[str, bool] = {}
        self._lock = asyncio.Lock()
        
    async def initialize(self):
        """Initialize model router and providers"""
        logger.info("Initializing model router")
        
        # Initialize providers
        await self._initialize_providers()
        
        # Load model information
        await self._load_model_info()
        
        # Start background tasks
        asyncio.create_task(self._metrics_updater())
        asyncio.create_task(self._circuit_breaker_monitor())
        
        logger.info(f"Model router initialized with {len(self.models)} models")
    
    async def _initialize_providers(self):
        """Initialize model providers"""
        provider_configs = {
            "openai": self.config.openai,
            "anthropic": self.config.anthropic,
            "cohere": self.config.cohere,
            "google": self.config.google,
            "ollama": self.config.ollama,
        }
        
        for name, config in provider_configs.items():
            if config.enabled and config.api_key:
                try:
                    provider = await self._create_provider(name, config)
                    self.providers[name] = provider
                    logger.info(f"Initialized {name} provider")
                except Exception as e:
                    logger.error(f"Failed to initialize {name} provider: {e}")
    
    async def _create_provider(self, name: str, config: ModelProviderConfig):
        """Create provider instance"""
        # Import the provider implementation on demand to avoid import-time
        # dependency requirements for unit tests that don't use them.
        if name == "openai":
            try:
                from src.providers.openai_provider import OpenAIProvider
            except Exception as e:
                raise ImportError(f"OpenAI provider not available: {e}")
            return OpenAIProvider(config)
        elif name == "anthropic":
            try:
                from src.providers.anthropic_provider import AnthropicProvider
            except Exception as e:
                raise ImportError(f"Anthropic provider not available: {e}")
            return AnthropicProvider(config)
        elif name == "cohere":
            try:
                from src.providers.cohere_provider import CohereProvider
            except Exception as e:
                raise ImportError(f"Cohere provider not available: {e}")
            return CohereProvider(config)
        elif name == "google":
            try:
                from src.providers.google_provider import GoogleProvider
            except Exception as e:
                raise ImportError(f"Google provider not available: {e}")
            return GoogleProvider(config)
        elif name == "ollama":
            try:
                from src.providers.ollama_provider import OllamaProvider
            except Exception as e:
                raise ImportError(f"Ollama provider not available: {e}")
            return OllamaProvider(config)
        else:
            raise ValueError(f"Unknown provider: {name}")
    
    async def _load_model_info(self):
        """Load model information from providers"""
        for provider_name, provider in self.providers.items():
            try:
                models = await provider.list_models()
                for model in models:
                    model_key = f"{provider_name}:{model.id}"
                    self.models[model_key] = model
                    self.performance_metrics[model_key] = PerformanceMetrics()
                    self.load_metrics[model_key] = LoadMetrics()
                    self.circuit_breakers[model_key] = False
                    
                logger.info(f"Loaded {len(models)} models from {provider_name}")
            except Exception as e:
                logger.error(f"Failed to load models from {provider_name}: {e}")
    
    async def select_optimal_model(self, request: CompletionRequest) -> Tuple[str, ModelInfo]:
        """Select optimal model for the request"""
        async with self._lock:
            # Filter available models
            candidates = await self._filter_available_models(request)
            
            if not candidates:
                raise ValueError("No available models for request")
            
            # Score models based on optimization strategy
            scored_models = []
            for model_key, model_info in candidates.items():
                score = await self._calculate_selection_score(model_key, model_info, request)
                scored_models.append((model_key, model_info, score))
            
            # Sort by score (higher is better)
            scored_models.sort(key=lambda x: x[2], reverse=True)
            
            # Select best model
            selected_key, selected_model, score = scored_models[0]
            
            logger.debug(f"Selected model {selected_key} with score {score:.3f}")
            
            return selected_key, selected_model
    
    async def _filter_available_models(self, request: CompletionRequest) -> Dict[str, ModelInfo]:
        """Filter models based on availability and requirements"""
        candidates = {}
        
        for model_key, model_info in self.models.items():
            # Check circuit breaker
            if self.circuit_breakers.get(model_key, False):
                continue
            
            # Check load
            load_metrics = self.load_metrics.get(model_key)
            if load_metrics and load_metrics.is_overloaded:
                continue
            
            # Check model capabilities
            if not self._model_supports_request(model_info, request):
                continue
            
            # Check provider availability
            provider_name = model_key.split(":")[0]
            if provider_name not in self.providers:
                continue
            
            candidates[model_key] = model_info
        
        return candidates
    
    def _model_supports_request(self, model_info: ModelInfo, request: CompletionRequest) -> bool:
        """Check if model supports the request requirements"""
        # Check max tokens
        if request.max_tokens and request.max_tokens > model_info.max_tokens:
            return False
        
        # Check context length
        estimated_context = len(request.messages) * 100  # Rough estimate
        if estimated_context > model_info.context_length:
            return False
        
        # Check model capabilities
        if request.tools and not model_info.supports_tools:
            return False
        
        if request.stream and not model_info.supports_streaming:
            return False
        
        return True
    
    async def _calculate_selection_score(
        self, 
        model_key: str, 
        model_info: ModelInfo, 
        request: CompletionRequest
    ) -> float:
        """Calculate selection score for a model"""
        
        # Base capability score
        base_score = model_info.capability_score
        
        # Performance factor (40% weight)
        performance_metrics = self.performance_metrics.get(model_key, PerformanceMetrics())
        performance_factor = self._calculate_performance_factor(performance_metrics)
        
        # Cost factor (30% weight)
        cost_factor = self._calculate_cost_factor(model_info, request)
        
        # Load factor (20% weight)
        load_metrics = self.load_metrics.get(model_key, LoadMetrics())
        load_factor = 1.0 - load_metrics.load_factor
        
        # Availability factor (10% weight)
        availability_factor = 1.0 if not self.circuit_breakers.get(model_key, False) else 0.0
        
        # Apply optimization strategy weights
        strategy = getattr(request, 'optimization_strategy', OptimizationStrategy.BALANCED)
        
        if strategy == OptimizationStrategy.COST:
            weights = [0.2, 0.1, 0.6, 0.05, 0.05]  # Prioritize cost
        elif strategy == OptimizationStrategy.PERFORMANCE:
            weights = [0.3, 0.5, 0.1, 0.05, 0.05]  # Prioritize performance
        elif strategy == OptimizationStrategy.AVAILABILITY:
            weights = [0.2, 0.2, 0.2, 0.3, 0.1]   # Prioritize availability
        else:  # BALANCED
            weights = [0.3, 0.25, 0.25, 0.15, 0.05]
        
        # Calculate weighted score
        factors = [base_score, performance_factor, cost_factor, load_factor, availability_factor]
        score = sum(w * f for w, f in zip(weights, factors))
        
        return max(0.0, min(1.0, score))  # Clamp to [0, 1]
    
    def _calculate_performance_factor(self, metrics: PerformanceMetrics) -> float:
        """Calculate performance factor from metrics"""
        if metrics.total_requests == 0:
            return 0.5  # Neutral score for new models
        
        # Success rate component (70%)
        success_component = metrics.success_rate * 0.7
        
        # Latency component (30%)
        # Lower latency is better, normalize to reasonable range
        latency_component = max(0, 1.0 - (metrics.avg_latency / 10.0)) * 0.3
        
        return success_component + latency_component
    
    def _calculate_cost_factor(self, model_info: ModelInfo, request: CompletionRequest) -> float:
        """Calculate cost factor for the model"""
        # Estimate tokens for the request
        estimated_tokens = self._estimate_tokens(request)
        
        # Calculate estimated cost
        estimated_cost = estimated_tokens * model_info.cost_per_token
        
        # Normalize cost (lower cost = higher score)
        # Assume max reasonable cost is $1.00 per request
        max_cost = 1.0
        cost_factor = max(0.0, 1.0 - (estimated_cost / max_cost))
        
        return cost_factor
    
    def _estimate_tokens(self, request: CompletionRequest) -> int:
        """Estimate token count for request"""
        # Simple estimation - would use tiktoken in production
        total_chars = sum(len(msg.content) for msg in request.messages)
        estimated_tokens = total_chars // 4  # Rough approximation
        
        # Add output tokens
        if request.max_tokens:
            estimated_tokens += request.max_tokens
        else:
            estimated_tokens += 500  # Default estimate
        
        return estimated_tokens
    
    async def execute_completion(
        self, 
        model_key: str, 
        request: CompletionRequest
    ) -> CompletionResponse:
        """Execute completion request with selected model"""
        
        # Update load metrics
        load_metrics = self.load_metrics.get(model_key, LoadMetrics())
        load_metrics.current_requests += 1
        load_metrics.last_request_time = time.time()
        
        start_time = time.time()
        
        try:
            # Get provider
            provider_name = model_key.split(":")[0]
            provider = self.providers[provider_name]
            
            # Execute request
            response = await provider.complete(request)
            
            # Update success metrics
            await self._update_success_metrics(model_key, start_time, response)
            
            return response
            
        except Exception as e:
            # Update failure metrics
            await self._update_failure_metrics(model_key, start_time, e)
            raise
            
        finally:
            # Update load metrics
            load_metrics.current_requests = max(0, load_metrics.current_requests - 1)
    
    async def _update_success_metrics(
        self, 
        model_key: str, 
        start_time: float, 
        response: CompletionResponse
    ):
        """Update metrics after successful request"""
        metrics = self.performance_metrics.get(model_key, PerformanceMetrics())
        
        duration = time.time() - start_time
        
        # Update counters
        metrics.total_requests += 1
        metrics.successful_requests += 1
        
        # Update latency (exponential moving average)
        alpha = 0.1
        if metrics.avg_latency == 0:
            metrics.avg_latency = duration
        else:
            metrics.avg_latency = alpha * duration + (1 - alpha) * metrics.avg_latency
        
        # Update tokens per second
        if response.usage and response.usage.total_tokens > 0:
            tokens_per_second = response.usage.total_tokens / duration
            if metrics.avg_tokens_per_second == 0:
                metrics.avg_tokens_per_second = tokens_per_second
            else:
                metrics.avg_tokens_per_second = (
                    alpha * tokens_per_second + (1 - alpha) * metrics.avg_tokens_per_second
                )
        
        metrics.last_updated = time.time()
        
        # Reset circuit breaker on success
        self.circuit_breakers[model_key] = False
    
    async def _update_failure_metrics(
        self, 
        model_key: str, 
        start_time: float, 
        error: Exception
    ):
        """Update metrics after failed request"""
        metrics = self.performance_metrics.get(model_key, PerformanceMetrics())
        
        duration = time.time() - start_time
        
        # Update counters
        metrics.total_requests += 1
        metrics.failed_requests += 1
        
        # Update latency for failed requests too
        alpha = 0.1
        if metrics.avg_latency == 0:
            metrics.avg_latency = duration
        else:
            metrics.avg_latency = alpha * duration + (1 - alpha) * metrics.avg_latency
        
        metrics.last_updated = time.time()
        
        # Check if circuit breaker should trip
        if metrics.failure_rate > 0.5 and metrics.total_requests >= 10:
            self.circuit_breakers[model_key] = True
            logger.warning(f"Circuit breaker tripped for model {model_key}")
    
    async def _metrics_updater(self):
        """Background task to update metrics"""
        while True:
            try:
                await asyncio.sleep(60)  # Update every minute
                
                # Clean up old metrics
                current_time = time.time()
                for model_key, metrics in self.performance_metrics.items():
                    if current_time - metrics.last_updated > 3600:  # 1 hour
                        # Reset metrics for inactive models
                        self.performance_metrics[model_key] = PerformanceMetrics()
                
            except Exception as e:
                logger.error(f"Error in metrics updater: {e}")
    
    async def _circuit_breaker_monitor(self):
        """Background task to monitor and reset circuit breakers"""
        while True:
            try:
                await asyncio.sleep(300)  # Check every 5 minutes
                
                current_time = time.time()
                for model_key, is_open in self.circuit_breakers.items():
                    if is_open:
                        metrics = self.performance_metrics.get(model_key)
                        if metrics and current_time - metrics.last_updated > 600:  # 10 minutes
                            # Reset circuit breaker after cooldown
                            self.circuit_breakers[model_key] = False
                            logger.info(f"Circuit breaker reset for model {model_key}")
                
            except Exception as e:
                logger.error(f"Error in circuit breaker monitor: {e}")
    
    async def get_model_stats(self) -> Dict[str, Any]:
        """Get model statistics"""
        stats = {}
        
        for model_key, model_info in self.models.items():
            performance = self.performance_metrics.get(model_key, PerformanceMetrics())
            load = self.load_metrics.get(model_key, LoadMetrics())
            
            stats[model_key] = {
                "model_info": {
                    "id": model_info.id,
                    "provider": model_info.provider,
                    "max_tokens": model_info.max_tokens,
                    "cost_per_token": model_info.cost_per_token,
                },
                "performance": {
                    "total_requests": performance.total_requests,
                    "success_rate": performance.success_rate,
                    "avg_latency": performance.avg_latency,
                    "avg_tokens_per_second": performance.avg_tokens_per_second,
                },
                "load": {
                    "current_requests": load.current_requests,
                    "load_factor": load.load_factor,
                    "is_overloaded": load.is_overloaded,
                },
                "circuit_breaker": self.circuit_breakers.get(model_key, False),
            }
        
        return stats
    
    async def shutdown(self):
        """Shutdown model router"""
        logger.info("Shutting down model router")
        
        # Shutdown providers
        for provider in self.providers.values():
            if hasattr(provider, 'shutdown'):
                await provider.shutdown()
        
        logger.info("Model router shutdown completed")