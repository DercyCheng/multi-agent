"""
Simple in-memory Feature Flag manager
"""

import asyncio
from typing import Dict

class FeatureFlagManager:
    def __init__(self):
        self._flags: Dict[str, bool] = {}
        self._lock = asyncio.Lock()

    async def set_flag(self, key: str, value: bool):
        async with self._lock:
            self._flags[key] = value

    async def get_flag(self, key: str) -> bool:
        async with self._lock:
            return self._flags.get(key, False)

    async def list_flags(self) -> Dict[str, bool]:
        async with self._lock:
            return dict(self._flags)
