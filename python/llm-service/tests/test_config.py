"""
Tests for configuration management
"""

import pytest
import os
from unittest.mock import patch

from src.config import Settings, ModelsConfig, ServerConfig, DatabaseConfig

class TestSettings:
    """Test Settings configuration"""
    
    def test_default_settings(self):
        """Test default settings values"""
        settings = Settings()
        
        assert settings.environment == "development"
        assert settings.debug is False
        assert settings.server.host == "0.0.0.0"
        assert settings.server.port == 8000
    
    def test_environment_validation(self):
        """Test environment validation"""
        with patch.dict(os.environ, {"ENVIRONMENT": "invalid"}):
            with pytest.raises(ValueError):
                Settings()
    
    def test_production_check(self):
        """Test production environment check"""
        with patch.dict(os.environ, {"ENVIRONMENT": "production"}):
            settings = Settings()
            assert settings.is_production() is True
            assert settings.is_development() is False
    
    def test_development_check(self):
        """Test development environment check"""
        settings = Settings()
        assert settings.is_development() is True
        assert settings.is_production() is False

class TestModelsConfig:
    """Test Models configuration"""
    
    def test_default_models_config(self):
        """Test default models configuration"""
        config = ModelsConfig()
        
        assert config.default_provider == "openai"
        assert config.default_model == "gpt-3.5-turbo"
        assert config.max_tokens == 4096
        assert config.temperature == 0.7
    
    def test_provider_configs(self):
        """Test provider configurations"""
        config = ModelsConfig()
        
        # Test OpenAI config
        assert config.openai.name == "openai"
        assert "gpt-3.5-turbo" in config.openai.models
        assert "gpt-4" in config.openai.models
        
        # Test Anthropic config
        assert config.anthropic.name == "anthropic"
        assert "claude-3-sonnet-20240229" in config.anthropic.models
        
        # Test Ollama config (should be disabled by default)
        assert config.ollama.name == "ollama"
        assert config.ollama.enabled is False
    
    def test_get_model_config(self):
        """Test getting model provider configuration"""
        settings = Settings()
        
        openai_config = settings.get_model_config("openai")
        assert openai_config is not None
        assert openai_config.name == "openai"
        
        invalid_config = settings.get_model_config("invalid")
        assert invalid_config is None
    
    def test_get_enabled_providers(self):
        """Test getting enabled providers"""
        with patch.dict(os.environ, {
            "OPENAI_API_KEY": "test-key",
            "ANTHROPIC_API_KEY": "test-key"
        }):
            settings = Settings()
            enabled = settings.get_enabled_providers()
            
            assert "openai" in enabled
            assert "anthropic" in enabled

class TestServerConfig:
    """Test Server configuration"""
    
    def test_default_server_config(self):
        """Test default server configuration"""
        config = ServerConfig()
        
        assert config.host == "0.0.0.0"
        assert config.port == 8000
        assert config.workers == 1
        assert config.reload is False
    
    def test_environment_override(self):
        """Test environment variable override"""
        with patch.dict(os.environ, {
            "HOST": "127.0.0.1",
            "PORT": "9000",
            "WORKERS": "4"
        }):
            config = ServerConfig()
            
            assert config.host == "127.0.0.1"
            assert config.port == 9000
            assert config.workers == 4

class TestDatabaseConfig:
    """Test Database configuration"""
    
    def test_database_config_from_env(self):
        """Test database configuration from environment"""
        with patch.dict(os.environ, {
            "POSTGRES_URL": "postgresql://user:pass@localhost:5432/db",
            "REDIS_URL": "redis://localhost:6379/0",
            "QDRANT_URL": "http://localhost:6333"
        }):
            config = DatabaseConfig()
            
            assert config.postgres_url == "postgresql://user:pass@localhost:5432/db"
            assert config.redis_url == "redis://localhost:6379/0"
            assert config.qdrant_url == "http://localhost:6333"