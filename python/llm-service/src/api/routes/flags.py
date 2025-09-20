"""
Feature Flags API
"""

from fastapi import APIRouter, Request, HTTPException
from pydantic import BaseModel

router = APIRouter()

class FlagToggle(BaseModel):
    key: str
    enabled: bool


@router.get("/")
async def list_flags(request: Request):
    manager = getattr(request.app.state, 'feature_flags', None)
    if manager is None:
        raise HTTPException(status_code=503, detail="Feature flag manager not available")
    return await manager.list_flags()


@router.post("/toggle")
async def toggle_flag(payload: FlagToggle, request: Request):
    manager = getattr(request.app.state, 'feature_flags', None)
    if manager is None:
        raise HTTPException(status_code=503, detail="Feature flag manager not available")
    await manager.set_flag(payload.key, payload.enabled)
    return {"key": payload.key, "enabled": payload.enabled}
