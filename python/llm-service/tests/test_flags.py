import asyncio

import pytest

from src.feature_flags.manager import FeatureFlagManager


@pytest.mark.asyncio
async def test_set_and_get_flag():
    manager = FeatureFlagManager()
    await manager.set_flag("cron.enabled", True)
    val = await manager.get_flag("cron.enabled")
    assert val is True

    await manager.set_flag("cron.enabled", False)
    val2 = await manager.get_flag("cron.enabled")
    assert val2 is False


def test_flags_api_toggle_and_list():
    from fastapi.testclient import TestClient
    from src.main import create_app

    app = create_app()
    # Create event loop for manager
    loop = asyncio.new_event_loop()
    asyncio.set_event_loop(loop)
    mgr = FeatureFlagManager()
    loop.run_until_complete(mgr.set_flag("cron.enabled", False))
    app.state.feature_flags = mgr

    client = TestClient(app)

    # Check listing
    resp = client.get("/flags/")
    assert resp.status_code == 200
    assert isinstance(resp.json(), dict)

    # Toggle flag
    resp = client.post("/flags/toggle", json={"key": "cron.enabled", "enabled": True})
    assert resp.status_code == 200
    assert resp.json().get("enabled") is True

    # Verify via manager
    val = loop.run_until_complete(mgr.get_flag("cron.enabled"))
    assert val is True
