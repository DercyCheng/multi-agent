"""
Ollama Provider for Multi-Agent LLM Service

Basic implementation using Ollama HTTP API (local Ollama server).
"""

import asyncio
import logging
import uuid
from typing import List, Optional, AsyncGenerator

import httpx

from src.providers.base_provider import BaseProvider
from src.core.models import (
    CompletionRequest, CompletionResponse, StreamResponse, ModelInfo,
    Usage, Choice, StreamChoice, Message, MessageRole
)

logger = logging.getLogger(__name__)


class OllamaProvider(BaseProvider):
    """Ollama provider implementation using REST API"""

    def __init__(self, config):
        super().__init__(config)
        self.client = httpx.AsyncClient(base_url=self.base_url, timeout=self.timeout)

        # Minimal model info mapping — in production, query Ollama API for available models
        self.model_info = {
            "mistral": ModelInfo(
                id="mistral",
                provider="ollama",
                name="Mistral (ollama)",
                description="Local Mistral model via Ollama",
                max_tokens=8192,
                context_length=32768,
                cost_per_token=0.0,
                supports_streaming=True,
                supports_tools=False,
                supports_vision=False,
                capability_score=0.8
            )
        }

    async def complete(self, request: CompletionRequest) -> CompletionResponse:
        async def _call():
            payload = {
                "model": request.model or "mistral",
                "messages": self._format_messages(request.messages),
                "max_tokens": request.max_tokens,
                "temperature": request.temperature
            }

            # Remove None
            payload = {k: v for k, v in payload.items() if v is not None}

            resp = await self.client.post("/api/generate", json=payload)
            resp.raise_for_status()
            data = resp.json()

            # Convert to our CompletionResponse
            choices = []
            text = data.get("text") or data.get("output") or ""
            message = Message(role=MessageRole.ASSISTANT, content=text)
            choices.append(Choice(index=0, message=message, finish_reason=None))

            usage = None
            # Ollama may not provide token usage; leave None

            return CompletionResponse(
                id=str(uuid.uuid4()),
                model=payload.get("model"),
                choices=choices,
                usage=usage,
                execution_id=str(uuid.uuid4()),
                provider="ollama"
            )

        return await self._retry_request(_call)

    async def stream_complete(self, request: CompletionRequest) -> AsyncGenerator[StreamResponse, None]:
        # Ollama supports streaming; use event stream if available
        await self._rate_limit()

        params = {
            "model": request.model or "mistral",
            "messages": self._format_messages(request.messages),
            "max_tokens": request.max_tokens,
            "temperature": request.temperature,
            "stream": True
        }

        # Clean
        params = {k: v for k, v in params.items() if v is not None}

        async with self.client.stream("POST", "/api/generate", json=params) as resp:
            resp.raise_for_status()
            async for line in resp.aiter_lines():
                if not line:
                    continue
                try:
                    # Each line may be a JSON chunk
                    chunk = httpx.Response(200, content=line).json()
                except Exception:
                    # Fallback: wrap raw text
                    chunk = {"id": str(uuid.uuid4()), "model": params.get("model"), "text": line}

                text = chunk.get("text") or chunk.get("output") or ""
                message = Message(role=MessageRole.ASSISTANT, content=text)
                stream_choice = StreamChoice(index=0, delta=message, finish_reason=None)

                yield StreamResponse(id=chunk.get("id"), model=params.get("model"), choices=[stream_choice])

    async def list_models(self) -> List[ModelInfo]:
        return list(self.model_info.values())

    async def validate_api_key(self) -> bool:
        # Ollama local server may not need API key; do a simple health check
        try:
            resp = await self.client.get("/api/models")
            resp.raise_for_status()
            return True
        except Exception as e:
            logger.error(f"Ollama health check failed: {e}")
            return False

    def _format_messages(self, messages: List) -> List[dict]:
        # Ollama expects a single prompt string; join messages
        if not messages:
            return []

        # Simplify: return list of dicts as used elsewhere
        return [{"role": m.role, "content": m.content} for m in messages]

    async def shutdown(self):
        await self.client.aclose()
