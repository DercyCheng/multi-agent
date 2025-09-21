"""
Pytest configuration and fixtures for LLM Service tests
"""

import asyncio
import pytest
import os
from typing import AsyncGenerator
from unittest.mock import AsyncMock, MagicMock

from fastapi.testclient import TestClient
from httpx import AsyncClient

# Set test environment
os.environ["ENVIRONMENT"] = "test"
os.environ["DEBUG"] = "true"
os.environ["POSTGRES_URL"] = "postgresql://test:test@localhost:5432/test"
os.environ["REDIS_URL"] = "redis://localhost:6379/0"
os.environ["QDRANT_URL"] = "http://localhost:6333"
os.environ["JWT_SECRET"] = "test-secret-key"
os.environ["OPENAI_API_KEY"] = "test-key"
os.environ["ANTHROPIC_API_KEY"] = "test-key"
os.environ["TRACING_ENABLED"] = "false"
os.environ["METRICS_ENABLED"] = "false"

@pytest.fixture(scope="session")
def event_loop():
    """Create an instance of the default event loop for the test session."""
    loop = asyncio.get_event_loop_policy().new_event_loop()
    yield loop
    loop.close()

@pytest.fixture
def mock_settings():
    """Mock settings for testing"""
    from src.config import Settings, ModelsConfig, ModelProviderConfig
    
    settings = Settings()
    settings.environment = "test"
    settings.debug = True
    
    # Mock model providers
    settings.models.openai.api_key = "test-key"
    settings.models.anthropic.api_key = "test-key"
    settings.models.ollama.enabled = True
    
    return settings

@pytest.fixture
def mock_model_router():
    """Mock model router for testing"""
    router = AsyncMock()
    router.providers = {}
    router.models = []
    router.circuit_breakers = {}
    router.get_model_stats = AsyncMock(return_value={})
    router.initialize = AsyncMock()
    router.shutdown = AsyncMock()
    return router

@pytest.fixture
def mock_context_engine():
    """Mock context engine for testing"""
    engine = AsyncMock()
    engine.db_pool = MagicMock()
    engine.vector_client = MagicMock()
    engine.embedding_model = MagicMock()
    engine.template_cache = {}
    engine.memory_cache = {}
    engine.initialize = AsyncMock()
    engine.shutdown = AsyncMock()
    return engine

@pytest.fixture
def mock_token_manager():
    """Mock token manager for testing"""
    manager = AsyncMock()
    manager.db_pool = MagicMock()
    manager.redis_client = AsyncMock()
    manager.redis_client.ping = AsyncMock()
    manager.budget_cache = {}
    manager.usage_cache = {}
    manager.initialize = AsyncMock()
    manager.shutdown = AsyncMock()
    return manager

@pytest.fixture
def mock_mcp_client():
    """Mock MCP client for testing"""
    client = AsyncMock()
    client.config = MagicMock()
    client.config.enabled = True
    client.config.max_connections = 100
    client.available_tools = []
    client.sessions = {}
    client.initialize = AsyncMock()
    client.shutdown = AsyncMock()
    return client

@pytest.fixture
def app_with_mocks(mock_model_router, mock_context_engine, mock_token_manager, mock_mcp_client):
    """Create FastAPI app with mocked dependencies"""
    from src.main import create_app
    from src.feature_flags.manager import FeatureFlagManager
    
    app = create_app()
    
    # Set mocked state
    app.state.model_router = mock_model_router
    app.state.context_engine = mock_context_engine
    app.state.token_manager = mock_token_manager
    app.state.mcp_client = mock_mcp_client
    app.state.feature_flags = FeatureFlagManager()
    
    return app

@pytest.fixture
def client(app_with_mocks):
    """Test client for FastAPI app"""
    return TestClient(app_with_mocks)

@pytest.fixture
async def async_client(app_with_mocks):
    """Async test client for FastAPI app"""
    async with AsyncClient(app=app_with_mocks, base_url="http://test") as ac:
        yield ac

@pytest.fixture
def feature_flag_manager():
    """Feature flag manager for testing"""
    from src.feature_flags.manager import FeatureFlagManager
    return FeatureFlagManager()

@pytest.fixture
def mock_ollama_provider():
    """Mock Ollama provider for testing"""
    from src.providers.ollama_provider import OllamaProvider
    from src.config import ModelProviderConfig
    
    config = ModelProviderConfig(
        name="ollama",
        api_key="",
        base_url="http://localhost:11434",
        models=["mistral", "llama2"],
        enabled=True
    )
    
    provider = OllamaProvider(config)
    return provider

@pytest.fixture
def sample_completion_request():
    """Sample completion request for testing"""
    from src.core.models import CompletionRequest, Message
    
    return CompletionRequest(
        model="gpt-3.5-turbo",
        messages=[
            Message(role="user", content="Hello, how are you?")
        ],
        temperature=0.7,
        max_tokens=100
    )

@pytest.fixture
def sample_health_response():
    """Sample health response for testing"""
    return {
        "status": "healthy",
        "uptime_seconds": 3600,
        "timestamp": 1640995200,
        "version": "1.0.0",
        "components": {
            "model_router": {"status": "healthy"},
            "context_engine": {"status": "healthy"},
            "token_manager": {"status": "healthy"},
            "mcp_client": {"status": "healthy"}
        }
    }