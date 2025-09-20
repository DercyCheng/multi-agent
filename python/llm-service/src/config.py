"""
Configuration management for Multi-Agent LLM Service
"""

from typing import Dict, List, Optional, Any
from pydantic import BaseSettings, Field, validator
from pydantic_settings import BaseSettings as PydanticBaseSettings
import os

class ServerConfig(BaseSettings):
    """Server configuration"""
    host: str = Field(default="0.0.0.0", env="HOST")
    port: int = Field(default=8000, env="PORT")
    workers: int = Field(default=1, env="WORKERS")
    reload: bool = Field(default=False, env="RELOAD")
    max_connections: int = Field(default=1000, env="MAX_CONNECTIONS")
    keepalive_timeout: int = Field(default=5, env="KEEPALIVE_TIMEOUT")

class DatabaseConfig(BaseSettings):
    """Database configuration"""
    postgres_url: str = Field(env="POSTGRES_URL")
    redis_url: str = Field(env="REDIS_URL")
    qdrant_url: str = Field(env="QDRANT_URL")
    pool_size: int = Field(default=10, env="DB_POOL_SIZE")
    max_overflow: int = Field(default=20, env="DB_MAX_OVERFLOW")

class ModelProviderConfig(BaseSettings):
    """Model provider configuration"""
    name: str
    api_key: str
    base_url: Optional[str] = None
    models: List[str] = []
    rate_limit: int = 100  # requests per minute
    timeout: int = 30  # seconds
    max_retries: int = 3
    enabled: bool = True

class ModelsConfig(BaseSettings):
    """Models configuration"""
    default_provider: str = Field(default="openai", env="DEFAULT_PROVIDER")
    default_model: str = Field(default="gpt-3.5-turbo", env="DEFAULT_MODEL")
    max_tokens: int = Field(default=4096, env="MAX_TOKENS")
    temperature: float = Field(default=0.7, env="TEMPERATURE")
    
    # Provider configurations
    openai: ModelProviderConfig = ModelProviderConfig(
        name="openai",
        api_key=os.getenv("OPENAI_API_KEY", ""),
        models=["gpt-3.5-turbo", "gpt-4", "gpt-4-turbo-preview"]
    )
    
    anthropic: ModelProviderConfig = ModelProviderConfig(
        name="anthropic",
        api_key=os.getenv("ANTHROPIC_API_KEY", ""),
        models=["claude-3-sonnet-20240229", "claude-3-opus-20240229"]
    )
    
    cohere: ModelProviderConfig = ModelProviderConfig(
        name="cohere",
        api_key=os.getenv("COHERE_API_KEY", ""),
        models=["command", "command-nightly"]
    )
    
    google: ModelProviderConfig = ModelProviderConfig(
        name="google",
        api_key=os.getenv("GOOGLE_API_KEY", ""),
        models=["gemini-pro", "gemini-pro-vision"]
    )

    ollama: ModelProviderConfig = ModelProviderConfig(
        name="ollama",
        api_key=os.getenv("OLLAMA_API_KEY", ""),
        base_url=os.getenv("OLLAMA_BASE_URL", "http://localhost:11434"),
        models=["mistral", "llama2", "vicuna"],
        enabled=bool(os.getenv("OLLAMA_ENABLED", "false").lower() in ["1", "true", "yes"])  # default disabled
    )

class ContextConfig(BaseSettings):
    """Context engineering configuration"""
    max_context_length: int = Field(default=32000, env="MAX_CONTEXT_LENGTH")
    compression_threshold: float = Field(default=0.8, env="COMPRESSION_THRESHOLD")
    knowledge_injection_enabled: bool = Field(default=True, env="KNOWLEDGE_INJECTION_ENABLED")
    memory_retrieval_enabled: bool = Field(default=True, env="MEMORY_RETRIEVAL_ENABLED")
    template_cache_size: int = Field(default=1000, env="TEMPLATE_CACHE_SIZE")

class TokenConfig(BaseSettings):
    """Token management configuration"""
    cost_tracking_enabled: bool = Field(default=True, env="COST_TRACKING_ENABLED")
    budget_enforcement_enabled: bool = Field(default=True, env="BUDGET_ENFORCEMENT_ENABLED")
    default_budget: float = Field(default=10.0, env="DEFAULT_BUDGET")  # USD
    cost_per_token: Dict[str, float] = {
        "gpt-3.5-turbo": 0.002,
        "gpt-4": 0.03,
        "claude-3-sonnet": 0.003,
        "claude-3-opus": 0.015,
    }

class MCPConfig(BaseSettings):
    """MCP (Model Context Protocol) configuration"""
    enabled: bool = Field(default=True, env="MCP_ENABLED")
    server_host: str = Field(default="localhost", env="MCP_SERVER_HOST")
    server_port: int = Field(default=8001, env="MCP_SERVER_PORT")
    max_connections: int = Field(default=100, env="MCP_MAX_CONNECTIONS")
    timeout: int = Field(default=30, env="MCP_TIMEOUT")
    tools_registry_url: str = Field(default="", env="MCP_TOOLS_REGISTRY_URL")

class SecurityConfig(BaseSettings):
    """Security configuration"""
    jwt_secret: str = Field(env="JWT_SECRET")
    jwt_algorithm: str = Field(default="HS256", env="JWT_ALGORITHM")
    jwt_expiration: int = Field(default=3600, env="JWT_EXPIRATION")  # seconds
    api_key_header: str = Field(default="X-API-Key", env="API_KEY_HEADER")
    rate_limit_per_minute: int = Field(default=100, env="RATE_LIMIT_PER_MINUTE")
    max_request_size: int = Field(default=10485760, env="MAX_REQUEST_SIZE")  # 10MB

class LoggingConfig(BaseSettings):
    """Logging configuration"""
    level: str = Field(default="INFO", env="LOG_LEVEL")
    format: str = Field(default="json", env="LOG_FORMAT")
    access_log: bool = Field(default=True, env="ACCESS_LOG")
    log_file: Optional[str] = Field(default=None, env="LOG_FILE")

class TracingConfig(BaseSettings):
    """Distributed tracing configuration"""
    enabled: bool = Field(default=True, env="TRACING_ENABLED")
    service_name: str = Field(default="llm-service", env="TRACING_SERVICE_NAME")
    jaeger_host: str = Field(default="localhost", env="JAEGER_HOST")
    jaeger_port: int = Field(default=14268, env="JAEGER_PORT")
    sample_rate: float = Field(default=0.1, env="TRACING_SAMPLE_RATE")

class MetricsConfig(BaseSettings):
    """Metrics configuration"""
    enabled: bool = Field(default=True, env="METRICS_ENABLED")
    port: int = Field(default=8001, env="METRICS_PORT")
    path: str = Field(default="/metrics", env="METRICS_PATH")
    collection_interval: int = Field(default=15, env="METRICS_COLLECTION_INTERVAL")

class Settings(PydanticBaseSettings):
    """Main application settings"""
    
    # Environment
    environment: str = Field(default="development", env="ENVIRONMENT")
    debug: bool = Field(default=False, env="DEBUG")
    
    # Component configurations
    server: ServerConfig = ServerConfig()
    database: DatabaseConfig = DatabaseConfig()
    models: ModelsConfig = ModelsConfig()
    context: ContextConfig = ContextConfig()
    tokens: TokenConfig = TokenConfig()
    mcp: MCPConfig = MCPConfig()
    security: SecurityConfig = SecurityConfig()
    logging: LoggingConfig = LoggingConfig()
    tracing: TracingConfig = TracingConfig()
    metrics: MetricsConfig = MetricsConfig()
    
    class Config:
        env_file = ".env"
        env_file_encoding = "utf-8"
        case_sensitive = False
    
    @validator("environment")
    def validate_environment(cls, v):
        allowed = ["development", "staging", "production"]
        if v not in allowed:
            raise ValueError(f"Environment must be one of {allowed}")
        return v
    
    def is_production(self) -> bool:
        """Check if running in production"""
        return self.environment == "production"
    
    def is_development(self) -> bool:
        """Check if running in development"""
        return self.environment == "development"
    
    def get_model_config(self, provider: str) -> Optional[ModelProviderConfig]:
        """Get model provider configuration"""
        return getattr(self.models, provider, None)
    
    def get_enabled_providers(self) -> List[str]:
        """Get list of enabled model providers"""
        providers = []
        for provider_name in ["openai", "anthropic", "cohere", "google"]:
            config = self.get_model_config(provider_name)
            if config and config.enabled and config.api_key:
                providers.append(provider_name)
        return providers