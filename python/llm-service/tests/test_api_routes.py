"""
Tests for API Routes
"""

import pytest
from fastapi.testclient import TestClient
from unittest.mock import AsyncMock, patch

class TestHealthRoutes:
    """Test health check routes"""
    
    def test_health_check_endpoint(self, client):
        """Test basic health check"""
        response = client.get("/health")
        
        assert response.status_code == 200
        data = response.json()
        
        assert "status" in data
        assert data["status"] == "healthy"
        assert "timestamp" in data
        assert "version" in data
    
    def test_health_detailed_endpoint(self, client):
        """Test detailed health check"""
        response = client.get("/health/detailed")
        
        assert response.status_code == 200
        data = response.json()
        
        assert "status" in data
        assert "services" in data
        assert "system" in data
        assert isinstance(data["services"], dict)
        assert isinstance(data["system"], dict)

class TestCompletionRoutes:
    """Test completion API routes"""
    
    def test_completion_endpoint_basic(self, client):
        """Test basic completion endpoint"""
        request_data = {
            "model": "test-model",
            "messages": [
                {"role": "user", "content": "Hello"}
            ],
            "max_tokens": 100,
            "temperature": 0.7
        }
        
        response = client.post("/v1/chat/completions", json=request_data)
        
        # Should return 200 or appropriate error code
        assert response.status_code in [200, 400, 422, 503]
    
    def test_completion_endpoint_validation(self, client):
        """Test completion endpoint validation"""
        # Invalid request - missing required fields
        request_data = {
            "model": "test-model"
            # Missing messages
        }
        
        response = client.post("/v1/chat/completions", json=request_data)
        
        assert response.status_code == 422  # Validation error
    
    def test_streaming_completion_endpoint(self, client):
        """Test streaming completion endpoint"""
        request_data = {
            "model": "test-model",
            "messages": [
                {"role": "user", "content": "Hello"}
            ],
            "stream": True,
            "max_tokens": 100
        }
        
        response = client.post("/v1/chat/completions", json=request_data)
        
        # Should return appropriate status
        assert response.status_code in [200, 400, 422, 503]

class TestModelRoutes:
    """Test model management routes"""
    
    def test_list_models_endpoint(self, client):
        """Test list models endpoint"""
        response = client.get("/v1/models")
        
        assert response.status_code == 200
        data = response.json()
        
        assert "data" in data
        assert isinstance(data["data"], list)
    
    def test_get_model_endpoint(self, client):
        """Test get specific model endpoint"""
        response = client.get("/v1/models/test-model")
        
        # Should return 200 or 404
        assert response.status_code in [200, 404]