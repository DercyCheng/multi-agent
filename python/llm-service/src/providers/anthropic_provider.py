"""
Anthropic Provider for Multi-Agent LLM Service
"""

import asyncio
import logging
import time
from typing import List, Optional, AsyncGenerator
import json
import uuid

import anthropic
from anthropic import AsyncAnthropic

from src.providers.base_provider import BaseProvider
from src.core.models import (
    CompletionRequest, CompletionResponse, StreamResponse, ModelInfo,
    Usage, Choice, StreamChoice, Message, MessageRole
)

logger = logging.getLogger(__name__)

class AnthropicProvider(BaseProvider):
    """Anthropic Claude provider implementation"""
    
    def __init__(self, config):
        super().__init__(config)
        self.client = AsyncAnthropic(
            api_key=self.api_key,
            timeout=self.timeout
        )
        
        # Model information
        self.model_info = {
            "claude-3-sonnet-20240229": ModelInfo(
                id="claude-3-sonnet-20240229",
                provider="anthropic",
                name="Claude 3 Sonnet",
                description="Balanced model for various tasks",
                max_tokens=4096,
                context_length=200000,
                cost_per_token=0.003,
                supports_streaming=True,
                supports_tools=True,
                supports_vision=True,
                capability_score=0.85
            ),
            "claude-3-opus-20240229": ModelInfo(
                id="claude-3-opus-20240229",
                provider="anthropic",
                name="Claude 3 Opus",
                description="Most capable Claude model",
                max_tokens=4096,
                context_length=200000,
                cost_per_token=0.015,
                supports_streaming=True,
                supports_tools=True,
                supports_vision=True,
                capability_score=0.95
            ),
            "claude-3-haiku-20240307": ModelInfo(
                id="claude-3-haiku-20240307",
                provider="anthropic",
                name="Claude 3 Haiku",
                description="Fast and efficient Claude model",
                max_tokens=4096,
                context_length=200000,
                cost_per_token=0.00025,
                supports_streaming=True,
                supports_tools=True,
                supports_vision=True,
                capability_score=0.7
            )
        }
    
    async def complete(self, request: CompletionRequest) -> CompletionResponse:
        """Complete a chat completion request"""
        
        async def _make_request():
            # Convert messages to Anthropic format
            system_message, messages = self._format_messages(request.messages)
            
            # Prepare request parameters
            params = {
                "model": request.model or "claude-3-sonnet-20240229",
                "messages": messages,
                "max_tokens": request.max_tokens or 1024,
                "temperature": request.temperature,
                "top_p": request.top_p,
                "stop_sequences": request.stop if isinstance(request.stop, list) else [request.stop] if request.stop else None,
                "stream": False
            }
            
            if system_message:
                params["system"] = system_message
            
            # Add tools if provided
            if request.tools:
                params["tools"] = [self._format_tool(tool) for tool in request.tools]
            
            # Remove None values
            params = {k: v for k, v in params.items() if v is not None}
            
            # Make API call
            response = await self.client.messages.create(**params)
            
            # Convert to our format
            return self._convert_response(response, request)
        
        return await self._retry_request(_make_request)
    
    async def stream_complete(self, request: CompletionRequest) -> AsyncGenerator[StreamResponse, None]:
        """Stream a chat completion request"""
        
        # Convert messages to Anthropic format
        system_message, messages = self._format_messages(request.messages)
        
        # Prepare request parameters
        params = {
            "model": request.model or "claude-3-sonnet-20240229",
            "messages": messages,
            "max_tokens": request.max_tokens or 1024,
            "temperature": request.temperature,
            "top_p": request.top_p,
            "stop_sequences": request.stop if isinstance(request.stop, list) else [request.stop] if request.stop else None,
            "stream": True
        }
        
        if system_message:
            params["system"] = system_message
        
        # Add tools if provided
        if request.tools:
            params["tools"] = [self._format_tool(tool) for tool in request.tools]
        
        # Remove None values
        params = {k: v for k, v in params.items() if v is not None}
        
        await self._rate_limit()
        
        try:
            stream = await self.client.messages.create(**params)
            
            response_id = str(uuid.uuid4())
            
            async for event in stream:
                if event.type == "content_block_delta":
                    delta_message = Message(
                        role=MessageRole.ASSISTANT,
                        content=event.delta.text if hasattr(event.delta, 'text') else "",
                    )
                    
                    stream_choice = StreamChoice(
                        index=0,
                        delta=delta_message,
                        finish_reason=None
                    )
                    
                    yield StreamResponse(
                        id=response_id,
                        model=params["model"],
                        choices=[stream_choice]
                    )
                
                elif event.type == "message_stop":
                    # Send final chunk with finish reason
                    delta_message = Message(
                        role=MessageRole.ASSISTANT,
                        content="",
                    )
                    
                    stream_choice = StreamChoice(
                        index=0,
                        delta=delta_message,
                        finish_reason="stop"
                    )
                    
                    yield StreamResponse(
                        id=response_id,
                        model=params["model"],
                        choices=[stream_choice]
                    )
                    
        except Exception as e:
            logger.error(f"Streaming error: {e}")
            raise
    
    async def list_models(self) -> List[ModelInfo]:
        """List available models"""
        available_models = []
        
        for model_id in self.models:
            if model_id in self.model_info:
                available_models.append(self.model_info[model_id])
        
        return available_models
    
    async def validate_api_key(self) -> bool:
        """Validate API key"""
        try:
            # Try a simple request to validate the key
            await self.client.messages.create(
                model="claude-3-haiku-20240307",
                max_tokens=1,
                messages=[{"role": "user", "content": "Hi"}]
            )
            return True
        except Exception as e:
            logger.error(f"API key validation failed: {e}")
            return False
    
    def _format_messages(self, messages: List) -> tuple[Optional[str], List[dict]]:
        """Format messages for Anthropic API"""
        system_message = None
        formatted_messages = []
        
        for msg in messages:
            if msg.role == MessageRole.SYSTEM:
                # Anthropic uses separate system parameter
                system_message = msg.content
            else:
                formatted_msg = {
                    "role": "user" if msg.role == MessageRole.USER else "assistant",
                    "content": msg.content
                }
                formatted_messages.append(formatted_msg)
        
        return system_message, formatted_messages
    
    def _format_tool(self, tool) -> dict:
        """Format tool for Anthropic API"""
        return {
            "name": tool.function.name,
            "description": tool.function.description,
            "input_schema": tool.function.parameters
        }
    
    def _convert_response(self, response, request: CompletionRequest) -> CompletionResponse:
        """Convert Anthropic response to our format"""
        
        # Extract content
        content = ""
        if response.content:
            for block in response.content:
                if hasattr(block, 'text'):
                    content += block.text
        
        # Create message
        message = Message(
            role=MessageRole.ASSISTANT,
            content=content
        )
        
        # Create choice
        choice = Choice(
            index=0,
            message=message,
            finish_reason=response.stop_reason
        )
        
        # Create usage
        usage = None
        if hasattr(response, 'usage'):
            usage = Usage(
                prompt_tokens=response.usage.input_tokens,
                completion_tokens=response.usage.output_tokens,
                total_tokens=response.usage.input_tokens + response.usage.output_tokens
            )
        
        return CompletionResponse(
            id=response.id,
            model=response.model,
            choices=[choice],
            usage=usage,
            execution_id=str(uuid.uuid4()),
            provider="anthropic"
        )