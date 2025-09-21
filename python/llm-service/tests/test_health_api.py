"""
Tests for Health Check API
"""

import pytest
from unittest.mock import AsyncMock, MagicMock
from fastapi.testclient import TestClient

class TestHealthAPI:
    """Test Health Check API endpoints"""
    
    def test_basic_health_check(self, client):
        """Test basic health check endpoint"""
        response = client.get("/health/")
        
        assert response.status_code == 200
        data = response.json()
        
        assert data["status"] == "healthy"
        assert "uptime_seconds" in data
        assert isinstance(data["uptime_seconds"], int)
    
    def test_liveness_check(self, client):
        """Test liveness check endpoint"""
        response = client.get("/health/live")
        
        assert response.status_code == 200
        data = response.json()
        
        assert data["status"] == "alive"
        assert "timestamp" in data
    
    def test_readiness_check_ready(self, client):
        """Test readiness check when services are ready"""
        response = client.get("/health/ready")
        
        assert response.status_code == 200
        data = response.json()
        
        assert data["status"] == "ready"
    
    def test_readiness_check_not_ready(self, app_with_mocks):
        """Test readiness check when services are not ready"""
        # Remove a required service
        delattr(app_with_mocks.state, 'model_router')
        
        client = TestClient(app_with_mocks)
        response = client.get("/health/ready")
        
        assert response.status_code == 503
        data = response.json()
        
        assert data["status"] == "not_ready"
        assert data["missing_service"] == "model_router"
    
    def test_detailed_health_check_healthy(self, client):
        """Test detailed health check when all components are healthy"""
        response = client.get("/health/detailed")
        
        assert response.status_code == 200
        data = response.json()
        
        assert data["status"] == "healthy"
        assert "timestamp" in data
        assert "uptime_seconds" in data
        assert "version" in data
        assert "components" in data
        
        components = data["components"]
        assert "model_router" in components
        assert "context_engine" in components
        assert "token_manager" in components
        assert "mcp_client" in components
    
    def test_detailed_health_check_missing_component(self, app_with_mocks):
        """Test detailed health check with missing component"""
        # Remove a component
        delattr(app_with_mocks.state, 'context_engine')
        
        client = TestClient(app_with_mocks)
        response = client.get("/health/detailed")
        
        assert response.status_code == 503
        data = response.json()
        
        assert data["status"] == "degraded"
        assert data["components"]["context_engine"]["status"] == "missing"
    
    def test_providers_health_check(self, client, mock_model_router):
        """Test providers health check"""
        # Mock provider
        mock_provider = AsyncMock()
        mock_provider.enabled = True
        mock_provider.rate_limit = 100
        mock_provider.health_check = AsyncMock(return_value=True)
        mock_provider.list_models = AsyncMock(return_value=[])
        
        mock_model_router.providers = {"openai": mock_provider}
        
        response = client.get("/health/providers")
        
        assert response.status_code == 200
        data = response.json()
        
        assert data["status"] == "healthy"
        assert data["healthy_providers"] == 1
        assert data["total_providers"] == 1
        assert len(data["providers"]) == 1
        
        provider_status = data["providers"][0]
        assert provider_status["name"] == "openai"
        assert provider_status["healthy"] is True
        assert provider_status["enabled"] is True
    
    def test_providers_health_check_no_router(self, app_with_mocks):
        """Test providers health check without model router"""
        # Remove model router
        delattr(app_with_mocks.state, 'model_router')
        
        client = TestClient(app_with_mocks)
        response = client.get("/health/providers")
        
        assert response.status_code == 503
        data = response.json()
        
        assert "error" in data
        assert "Model router not available" in data["error"]

class TestHealthHelpers:
    """Test health check helper functions"""
    
    @pytest.mark.asyncio
    async def test_check_model_router_health_healthy(self, mock_model_router):
        """Test model router health check when healthy"""
        from src.api.routes.health import _check_model_router_health
        
        mock_model_router.providers = {"openai": MagicMock()}
        mock_model_router.models = ["gpt-3.5-turbo"]
        mock_model_router.circuit_breakers = {"openai": False}
        mock_model_router.get_model_stats = AsyncMock(return_value={})
        
        result = await _check_model_router_health(mock_model_router)
        
        assert result["status"] == "healthy"
        assert result["providers_count"] == 1
        assert result["models_count"] == 1
        assert result["circuit_breakers_open"] == 0
    
    @pytest.mark.asyncio
    async def test_check_model_router_health_no_providers(self, mock_model_router):
        """Test model router health check with no providers"""
        from src.api.routes.health import _check_model_router_health
        
        mock_model_router.providers = {}
        
        result = await _check_model_router_health(mock_model_router)
        
        assert result["status"] == "unhealthy"
        assert "No providers available" in result["error"]
    
    @pytest.mark.asyncio
    async def test_check_context_engine_health_healthy(self, mock_context_engine):
        """Test context engine health check when healthy"""
        from src.api.routes.health import _check_context_engine_health
        
        result = await _check_context_engine_health(mock_context_engine)
        
        assert result["status"] == "healthy"
        assert result["database"] is True
        assert result["vector_db"] is True
        assert result["embedding_model"] is True
    
    @pytest.mark.asyncio
    async def test_check_token_manager_health_healthy(self, mock_token_manager):
        """Test token manager health check when healthy"""
        from src.api.routes.health import _check_token_manager_health
        
        result = await _check_token_manager_health(mock_token_manager)
        
        assert result["status"] == "healthy"
        assert result["database"] is True
        assert result["redis"] is True
    
    @pytest.mark.asyncio
    async def test_check_mcp_client_health_enabled(self, mock_mcp_client):
        """Test MCP client health check when enabled"""
        from src.api.routes.health import _check_mcp_client_health
        
        result = await _check_mcp_client_health(mock_mcp_client)
        
        assert result["status"] == "healthy"
        assert result["enabled"] is True
        assert result["tools_count"] == 0
        assert result["active_sessions"] == 0
    
    @pytest.mark.asyncio
    async def test_check_mcp_client_health_disabled(self, mock_mcp_client):
        """Test MCP client health check when disabled"""
        from src.api.routes.health import _check_mcp_client_health
        
        mock_mcp_client.config.enabled = False
        
        result = await _check_mcp_client_health(mock_mcp_client)
        
        assert result["status"] == "disabled"
        assert result["enabled"] is False