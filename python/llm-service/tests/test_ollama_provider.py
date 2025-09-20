import asyncio

import pytest

from src.providers.ollama_provider import OllamaProvider
from src.config import ModelsConfig


@pytest.mark.asyncio
async def test_list_models_basic():
    cfg = ModelsConfig()
    cfg.ollama.api_key = ""
    cfg.ollama.base_url = "http://localhost:11434"
    provider = OllamaProvider(cfg.ollama)

    models = await provider.list_models()
    assert isinstance(models, list)
