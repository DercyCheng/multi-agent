"""
Core data models for Multi-Agent LLM Service
"""

from typing import Dict, List, Optional, Any, Union
from pydantic import BaseModel, Field, validator
from enum import Enum
import time

class MessageRole(str, Enum):
    """Message roles in conversation"""
    SYSTEM = "system"
    USER = "user"
    ASSISTANT = "assistant"
    TOOL = "tool"

class Message(BaseModel):
    """Chat message"""
    role: MessageRole
    content: str
    name: Optional[str] = None
    tool_calls: Optional[List[Dict[str, Any]]] = None
    tool_call_id: Optional[str] = None

class ToolFunction(BaseModel):
    """Tool function definition"""
    name: str
    description: str
    parameters: Dict[str, Any]

class Tool(BaseModel):
    """Tool definition"""
    type: str = "function"
    function: ToolFunction

class CompletionRequest(BaseModel):
    """Completion request"""
    messages: List[Message]
    model: Optional[str] = None
    max_tokens: Optional[int] = None
    temperature: Optional[float] = 0.7
    top_p: Optional[float] = 1.0
    frequency_penalty: Optional[float] = 0.0
    presence_penalty: Optional[float] = 0.0
    stop: Optional[Union[str, List[str]]] = None
    stream: bool = False
    tools: Optional[List[Tool]] = None
    tool_choice: Optional[Union[str, Dict[str, Any]]] = None
    
    # Multi-Agent specific fields
    user_id: str
    tenant_id: str
    session_id: str
    context_id: Optional[str] = None
    optimization_strategy: Optional[str] = "balanced"
    budget_limit: Optional[float] = None
    
    @validator("temperature")
    def validate_temperature(cls, v):
        if v is not None and not (0.0 <= v <= 2.0):
            raise ValueError("Temperature must be between 0.0 and 2.0")
        return v
    
    @validator("top_p")
    def validate_top_p(cls, v):
        if v is not None and not (0.0 <= v <= 1.0):
            raise ValueError("Top_p must be between 0.0 and 1.0")
        return v

class Usage(BaseModel):
    """Token usage information"""
    prompt_tokens: int
    completion_tokens: int
    total_tokens: int
    
    @property
    def efficiency_ratio(self) -> float:
        """Ratio of completion to prompt tokens"""
        if self.prompt_tokens == 0:
            return 0.0
        return self.completion_tokens / self.prompt_tokens

class Choice(BaseModel):
    """Completion choice"""
    index: int
    message: Message
    finish_reason: Optional[str] = None
    logprobs: Optional[Dict[str, Any]] = None

class CompletionResponse(BaseModel):
    """Completion response"""
    id: str
    object: str = "chat.completion"
    created: int = Field(default_factory=lambda: int(time.time()))
    model: str
    choices: List[Choice]
    usage: Optional[Usage] = None
    
    # Multi-Agent specific fields
    execution_id: str
    provider: str
    cost_usd: Optional[float] = None
    latency_ms: Optional[int] = None
    cached: bool = False

class StreamChoice(BaseModel):
    """Streaming completion choice"""
    index: int
    delta: Message
    finish_reason: Optional[str] = None

class StreamResponse(BaseModel):
    """Streaming completion response"""
    id: str
    object: str = "chat.completion.chunk"
    created: int = Field(default_factory=lambda: int(time.time()))
    model: str
    choices: List[StreamChoice]

class ModelInfo(BaseModel):
    """Model information"""
    id: str
    provider: str
    name: str
    description: Optional[str] = None
    max_tokens: int
    context_length: int
    cost_per_token: float
    supports_streaming: bool = True
    supports_tools: bool = False
    supports_vision: bool = False
    capability_score: float = 0.5  # 0.0 to 1.0
    
    class Config:
        frozen = True

class ProviderStatus(BaseModel):
    """Provider status information"""
    name: str
    enabled: bool
    healthy: bool
    last_check: int = Field(default_factory=lambda: int(time.time()))
    error_message: Optional[str] = None
    models_count: int = 0
    requests_per_minute: int = 0
    avg_latency_ms: float = 0.0

class ContextRequest(BaseModel):
    """Context engineering request"""
    query: str
    context: Dict[str, Any] = {}
    user_id: str
    session_id: str
    task_type: str = "general"
    knowledge_budget: int = 1000  # tokens
    available_tools: List[str] = []
    user_preferences: Dict[str, Any] = {}

class EngineeredContext(BaseModel):
    """Engineered context result"""
    system_instructions: str
    knowledge: str
    tools: List[Tool] = []
    memory: Dict[str, Any] = {}
    metadata: Dict[str, Any] = {}
    token_count: int = 0
    compression_ratio: Optional[float] = None

class TokenBudget(BaseModel):
    """Token budget information"""
    user_id: str
    tenant_id: str
    total_budget: float  # USD
    used_budget: float = 0.0
    remaining_budget: float = 0.0
    daily_limit: Optional[float] = None
    monthly_limit: Optional[float] = None
    last_reset: int = Field(default_factory=lambda: int(time.time()))
    
    @property
    def budget_utilization(self) -> float:
        """Budget utilization percentage"""
        if self.total_budget == 0:
            return 0.0
        return (self.used_budget / self.total_budget) * 100

class CacheEntry(BaseModel):
    """Cache entry for responses"""
    key: str
    response: CompletionResponse
    created_at: int = Field(default_factory=lambda: int(time.time()))
    access_count: int = 0
    last_accessed: int = Field(default_factory=lambda: int(time.time()))
    ttl: int = 3600  # seconds
    
    @property
    def is_expired(self) -> bool:
        """Check if cache entry is expired"""
        return (int(time.time()) - self.created_at) > self.ttl

class MCPToolRequest(BaseModel):
    """MCP tool execution request"""
    tool_name: str
    parameters: Dict[str, Any]
    user_id: str
    session_id: str
    timeout: int = 30

class MCPToolResponse(BaseModel):
    """MCP tool execution response"""
    tool_name: str
    result: Any
    success: bool
    error_message: Optional[str] = None
    execution_time_ms: int
    
class SecurityValidation(BaseModel):
    """Security validation result"""
    is_safe: bool
    risk_score: float  # 0.0 to 1.0
    violations: List[str] = []
    recommendations: List[str] = []
    
class HealthStatus(BaseModel):
    """Health status"""
    status: str  # healthy, degraded, unhealthy
    timestamp: int = Field(default_factory=lambda: int(time.time()))
    version: str = "1.0.0"
    uptime_seconds: int = 0
    
class MetricsSummary(BaseModel):
    """Metrics summary"""
    total_requests: int = 0
    successful_requests: int = 0
    failed_requests: int = 0
    avg_latency_ms: float = 0.0
    total_cost_usd: float = 0.0
    active_sessions: int = 0
    cache_hit_rate: float = 0.0
    
    @property
    def success_rate(self) -> float:
        """Success rate percentage"""
        if self.total_requests == 0:
            return 0.0
        return (self.successful_requests / self.total_requests) * 100

# Error models
class ErrorResponse(BaseModel):
    """Error response"""
    error: str
    message: str
    code: Optional[str] = None
    details: Optional[Dict[str, Any]] = None

class ValidationError(BaseModel):
    """Validation error"""
    field: str
    message: str
    value: Any = None