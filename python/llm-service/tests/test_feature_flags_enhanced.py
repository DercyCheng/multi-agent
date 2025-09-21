"""
Enhanced tests for Feature Flags functionality
"""

import pytest
import asyncio
from unittest.mock import AsyncMock, patch
from fastapi.testclient import TestClient

from src.feature_flags.manager import FeatureFlagManager

class TestFeatureFlagManager:
    """Test FeatureFlagManager class"""
    
    @pytest.mark.asyncio
    async def test_set_and_get_flag(self):
        """Test setting and getting feature flags"""
        manager = FeatureFlagManager()
        
        # Test setting a flag
        await manager.set_flag("test.feature", True)
        result = await manager.get_flag("test.feature")
        assert result is True
        
        # Test updating a flag
        await manager.set_flag("test.feature", False)
        result = await manager.get_flag("test.feature")
        assert result is False
    
    @pytest.mark.asyncio
    async def test_get_nonexistent_flag(self):
        """Test getting a non-existent flag returns False"""
        manager = FeatureFlagManager()
        
        result = await manager.get_flag("nonexistent.flag")
        assert result is False
    
    @pytest.mark.asyncio
    async def test_list_flags(self):
        """Test listing all flags"""
        manager = FeatureFlagManager()
        
        # Set multiple flags
        await manager.set_flag("flag1", True)
        await manager.set_flag("flag2", False)
        await manager.set_flag("flag3", True)
        
        flags = await manager.list_flags()
        
        assert len(flags) == 3
        assert flags["flag1"] is True
        assert flags["flag2"] is False
        assert flags["flag3"] is True

class TestFeatureFlagsAPI:
    """Test Feature Flags API endpoints"""
    
    def test_list_flags_endpoint(self, client, feature_flag_manager):
        """Test listing flags via API"""
        # Setup flags in app state
        client.app.state.feature_flags = feature_flag_manager
        
        response = client.get("/flags/")
        
        assert response.status_code == 200
        data = response.json()
        assert isinstance(data, dict)
    
    def test_toggle_flag_endpoint(self, client, feature_flag_manager):
        """Test toggling flags via API"""
        # Setup flags in app state
        client.app.state.feature_flags = feature_flag_manager
        
        # Toggle flag to True
        response = client.post("/flags/toggle", json={
            "key": "toggle.test",
            "enabled": True
        })
        
        assert response.status_code == 200
        data = response.json()
        
        assert data["key"] == "toggle.test"
        assert data["enabled"] is True