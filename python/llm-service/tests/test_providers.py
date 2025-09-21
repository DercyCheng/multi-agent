"""
Tests for LLM Providers
"""

import pytest
from unittest.mock import AsyncMock, MagicMock, patch
from src.providers.base_provider import BaseProvider
from src.providers.ollama_provider import OllamaProvider
from src.config import ModelProviderConfig
from src.core.models import CompletionRequest, Message, ModelInfo

class TestBaseProvider:
    """Test BaseProvider abstract class"""
    
    def test_provider_initialization(self):
        """Test provider initialization"""
        config = ModelProviderConfig(
            name="test",
            api_key="test-key",
            base_url="http://test.com",
            models=["model1", "model2"],
            rate_limit=50,
            timeout=20,
            max_retries=2,
            enabled=True
        )
        
        # Create a concrete implementation for testing
        class TestProvider(BaseProvider):
            async def complete(self, request):
                pass
            async def stream_complete(self, request):
                pass
            async def list_models(self):
                pass
            async def validate_api_key(self):
                return True
        
        provider = TestProvider(config)
        
        assert provider.name == "test"
        assert provider.api_key == "test-key"
        assert provider.base_url == "http://test.com"
        assert provider.models == ["model1", "model2"]
        assert provider.rate_limit == 50
        assert provider.timeout == 20
        assert provider.max_retries == 2
        assert provider.enabled is True
    
    @pytest.mark.asyncio
    async def test_health_check_success(self):
        """Test successful health check"""
        class TestProvider(BaseProvider):
            async def complete(self, request):
                pass
            async def stream_complete(self, request):
                pass
            async def list_models(self):
                pass
            async def validate_api_key(self):
                return True
        
        config = ModelProviderConfig(name="test", api_key="test-key")
        provider = TestProvider(config)
        
        result = await provider.health_check()
        assert result is True
    
    @pytest.mark.asyncio
    async def test_health_check_failure(self):
        """Test failed health check"""
        class TestProvider(BaseProvider):
            async def complete(self, request):
                pass
            async def stream_complete(self, request):
                pass
            async def list_models(self):
                pass
            async def validate_api_key(self):
                raise Exception("API key invalid")
        
        config = ModelProviderConfig(name="test", api_key="invalid-key")
        provider = TestProvider(config)
        
        result = await provider.health_check()
        assert result is False
    
    def test_calculate_tokens(self):
        """Test token calculation"""
        class TestProvider(BaseProvider):
            async def complete(self, request):
                pass
            async def stream_complete(self, request):
                pass
            async def list_models(self):
                pass
            async def validate_api_key(self):
                return True
        
        config = ModelProviderConfig(name="test", api_key="test-key")
        provider = TestProvider(config)
        
        # Test rough token calculation (1 token per 4 characters)
        tokens = provider._calculate_tokens("Hello world!")
        assert tokens == 3  # 12 characters / 4 = 3 tokens
    
    def test_format_messages(self):
        """Test message formatting"""
        class TestProvider(BaseProvider):
            async def complete(self, request):
                pass
            async def stream_complete(self, request):
                pass
            async def list_models(self):
                pass
            async def validate_api_key(self):
                return True
        
        config = ModelProviderConfig(name="test", api_key="test-key")
        provider = TestProvider(config)
        
        messages = [
            Message(role="user", content="Hello"),
            Message(role="assistant", content="Hi there!")
        ]
        
        formatted = provider._format_messages(messages)
        
        assert len(formatted) == 2
        assert formatted[0]["role"] == "user"
        assert formatted[0]["content"] == "Hello"
        assert formatted[1]["role"] == "assistant"
        assert formatted[1]["content"] == "Hi there!"

class TestOllamaProvider:
    """Test OllamaProvider implementation"""
    
    def test_ollama_provider_initialization(self):
        """Test Ollama provider initialization"""
        config = ModelProviderConfig(
            name="ollama",
            api_key="",
            base_url="http://localhost:11434",
            models=["mistral", "llama2"],
            enabled=True
        )
        
        provider = OllamaProvider(config)
        
        assert provider.name == "ollama"
        assert provider.base_url == "http://localhost:11434"
        assert provider.models == ["mistral", "llama2"]
        assert provider.enabled is True
    
    @pytest.mark.asyncio
    async def test_list_models_empty(self, mock_ollama_provider):
        """Test list_models returns empty list"""
        models = await mock_ollama_provider.list_models()
        
        assert isinstance(models, list)
        assert len(models) == 0
    
    @pytest.mark.asyncio
    async def test_validate_api_key(self, mock_ollama_provider):
        """Test API key validation"""
        result = await mock_ollama_provider.validate_api_key()
        
        assert result is True
    
    @pytest.mark.asyncio
    async def test_complete_stub(self, mock_ollama_provider, sample_completion_request):
        """Test completion stub implementation"""
        response = await mock_ollama_provider.complete(sample_completion_request)
        
        assert response is not None
        assert hasattr(response, 'id')
        assert hasattr(response, 'choices')
    
    @pytest.mark.asyncio
    async def test_stream_complete_stub(self, mock_ollama_provider, sample_completion_request):
        """Test streaming completion stub implementation"""
        stream = mock_ollama_provider.stream_complete(sample_completion_request)
        
        # Should be an async generator that yields nothing
        responses = []
        async for response in stream:
            responses.append(response)
        
        assert len(responses) == 0

class TestProviderIntegration:
    """Integration tests for providers"""
    
    @pytest.mark.integration
    @pytest.mark.asyncio
    async def test_provider_with_rate_limiting(self):
        """Test provider with rate limiting"""
        class TestProvider(BaseProvider):
            def __init__(self, config):
                super().__init__(config)
                self.call_count = 0
            
            async def complete(self, request):
                self.call_count += 1
                return f"Response {self.call_count}"
            
            async def stream_complete(self, request):
                pass
            
            async def list_models(self):
                return []
            
            async def validate_api_key(self):
                return True
        
        config = ModelProviderConfig(
            name="test",
            api_key="test-key",
            rate_limit=2  # 2 requests per minute
        )
        
        provider = TestProvider(config)
        
        # First request should work
        result1 = await provider._retry_request(provider.complete, None)
        assert "Response 1" in result1
        
        # Second request should work but be rate limited
        result2 = await provider._retry_request(provider.complete, None)
        assert "Response 2" in result2
    
    @pytest.mark.integration
    @pytest.mark.asyncio
    async def test_provider_retry_mechanism(self):
        """Test provider retry mechanism"""
        class FailingProvider(BaseProvider):
            def __init__(self, config):
                super().__init__(config)
                self.attempt_count = 0
            
            async def complete(self, request):
                self.attempt_count += 1
                if self.attempt_count < 3:
                    raise Exception(f"Attempt {self.attempt_count} failed")
                return "Success on attempt 3"
            
            async def stream_complete(self, request):
                pass
            
            async def list_models(self):
                return []
            
            async def validate_api_key(self):
                return True
        
        config = ModelProviderConfig(
            name="test",
            api_key="test-key",
            max_retries=3
        )
        
        provider = FailingProvider(config)
        
        # Should succeed after retries
        result = await provider._retry_request(provider.complete, None)
        assert result == "Success on attempt 3"
        assert provider.attempt_count == 3
    
    @pytest.mark.integration
    @pytest.mark.asyncio
    async def test_provider_max_retries_exceeded(self):
        """Test provider when max retries are exceeded"""
        class AlwaysFailingProvider(BaseProvider):
            async def complete(self, request):
                raise Exception("Always fails")
            
            async def stream_complete(self, request):
                pass
            
            async def list_models(self):
                return []
            
            async def validate_api_key(self):
                return True
        
        config = ModelProviderConfig(
            name="test",
            api_key="test-key",
            max_retries=2
        )
        
        provider = AlwaysFailingProvider(config)
        
        # Should raise exception after max retries
        with pytest.raises(Exception, match="Always fails"):
            await provider._retry_request(provider.complete, None)