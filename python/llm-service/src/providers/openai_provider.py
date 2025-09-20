"""
OpenAI Provider for Multi-Agent LLM Service
"""

import asyncio
import logging
import time
from typing import List, Optional, AsyncGenerator
import json
import uuid

import openai
from openai import AsyncOpenAI

from src.providers.base_provider import BaseProvider
from src.core.models import (
    CompletionRequest, CompletionResponse, StreamResponse, ModelInfo,
    Usage, Choice, StreamChoice, Message, MessageRole
)

logger = logging.getLogger(__name__)

class OpenAIProvider(BaseProvider):
    """OpenAI provider implementation"""
    
    def __init__(self, config):
        super().__init__(config)
        self.client = AsyncOpenAI(
            api_key=self.api_key,
            base_url=self.base_url,
            timeout=self.timeout
        )
        
        # Model information
        self.model_info = {
            "gpt-3.5-turbo": ModelInfo(
                id="gpt-3.5-turbo",
                provider="openai",
                name="GPT-3.5 Turbo",
                description="Fast, capable model for most tasks",
                max_tokens=4096,
                context_length=16385,
                cost_per_token=0.002,
                supports_streaming=True,
                supports_tools=True,
                supports_vision=False,
                capability_score=0.7
            ),
            "gpt-4": ModelInfo(
                id="gpt-4",
                provider="openai",
                name="GPT-4",
                description="Most capable model for complex tasks",
                max_tokens=8192,
                context_length=8192,
                cost_per_token=0.03,
                supports_streaming=True,
                supports_tools=True,
                supports_vision=False,
                capability_score=0.9
            ),
            "gpt-4-turbo-preview": ModelInfo(
                id="gpt-4-turbo-preview",
                provider="openai",
                name="GPT-4 Turbo",
                description="Latest GPT-4 model with improved performance",
                max_tokens=4096,
                context_length=128000,
                cost_per_token=0.01,
                supports_streaming=True,
                supports_tools=True,
                supports_vision=True,
                capability_score=0.95
            ),
            "gpt-4-vision-preview": ModelInfo(
                id="gpt-4-vision-preview",
                provider="openai",
                name="GPT-4 Vision",
                description="GPT-4 with vision capabilities",
                max_tokens=4096,
                context_length=128000,
                cost_per_token=0.01,
                supports_streaming=True,
                supports_tools=True,
                supports_vision=True,
                capability_score=0.9
            )
        }
    
    async def complete(self, request: CompletionRequest) -> CompletionResponse:
        """Complete a chat completion request"""
        
        async def _make_request():
            # Prepare request parameters
            params = {
                "model": request.model or "gpt-3.5-turbo",
                "messages": self._format_messages(request.messages),
                "max_tokens": request.max_tokens,
                "temperature": request.temperature,
                "top_p": request.top_p,
                "frequency_penalty": request.frequency_penalty,
                "presence_penalty": request.presence_penalty,
                "stop": request.stop,
                "stream": False
            }
            
            # Add tools if provided
            if request.tools:
                params["tools"] = [self._format_tool(tool) for tool in request.tools]
                if request.tool_choice:
                    params["tool_choice"] = request.tool_choice
            
            # Remove None values
            params = {k: v for k, v in params.items() if v is not None}
            
            # Make API call
            response = await self.client.chat.completions.create(**params)
            
            # Convert to our format
            return self._convert_response(response, request)
        
        return await self._retry_request(_make_request)
    
    async def stream_complete(self, request: CompletionRequest) -> AsyncGenerator[StreamResponse, None]:
        """Stream a chat completion request"""
        
        # Prepare request parameters
        params = {
            "model": request.model or "gpt-3.5-turbo",
            "messages": self._format_messages(request.messages),
            "max_tokens": request.max_tokens,
            "temperature": request.temperature,
            "top_p": request.top_p,
            "frequency_penalty": request.frequency_penalty,
            "presence_penalty": request.presence_penalty,
            "stop": request.stop,
            "stream": True
        }
        
        # Add tools if provided
        if request.tools:
            params["tools"] = [self._format_tool(tool) for tool in request.tools]
            if request.tool_choice:
                params["tool_choice"] = request.tool_choice
        
        # Remove None values
        params = {k: v for k, v in params.items() if v is not None}
        
        await self._rate_limit()
        
        try:
            stream = await self.client.chat.completions.create(**params)
            
            async for chunk in stream:
                if chunk.choices:
                    choice = chunk.choices[0]
                    
                    # Convert delta to message
                    delta_dict = choice.delta.dict() if hasattr(choice.delta, 'dict') else choice.delta
                    delta_message = Message(
                        role=MessageRole(delta_dict.get("role", "assistant")),
                        content=delta_dict.get("content", ""),
                        tool_calls=delta_dict.get("tool_calls")
                    )
                    
                    stream_choice = StreamChoice(
                        index=choice.index,
                        delta=delta_message,
                        finish_reason=choice.finish_reason
                    )
                    
                    yield StreamResponse(
                        id=chunk.id,
                        model=chunk.model,
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
            await self.client.models.list()
            return True
        except Exception as e:
            logger.error(f"API key validation failed: {e}")
            return False
    
    def _format_messages(self, messages: List) -> List[dict]:
        """Format messages for OpenAI API"""
        formatted = []
        
        for msg in messages:
            formatted_msg = {
                "role": msg.role,
                "content": msg.content
            }
            
            if msg.name:
                formatted_msg["name"] = msg.name
            
            if msg.tool_calls:
                formatted_msg["tool_calls"] = msg.tool_calls
            
            if msg.tool_call_id:
                formatted_msg["tool_call_id"] = msg.tool_call_id
            
            formatted.append(formatted_msg)
        
        return formatted
    
    def _format_tool(self, tool) -> dict:
        """Format tool for OpenAI API"""
        return {
            "type": tool.type,
            "function": {
                "name": tool.function.name,
                "description": tool.function.description,
                "parameters": tool.function.parameters
            }
        }
    
    def _convert_response(self, response, request: CompletionRequest) -> CompletionResponse:
        """Convert OpenAI response to our format"""
        
        # Convert choices
        choices = []
        for choice in response.choices:
            message = Message(
                role=MessageRole(choice.message.role),
                content=choice.message.content or "",
                tool_calls=getattr(choice.message, 'tool_calls', None)
            )
            
            choices.append(Choice(
                index=choice.index,
                message=message,
                finish_reason=choice.finish_reason
            ))
        
        # Convert usage
        usage = None
        if response.usage:
            usage = Usage(
                prompt_tokens=response.usage.prompt_tokens,
                completion_tokens=response.usage.completion_tokens,
                total_tokens=response.usage.total_tokens
            )
        
        return CompletionResponse(
            id=response.id,
            model=response.model,
            choices=choices,
            usage=usage,
            execution_id=str(uuid.uuid4()),
            provider="openai"
        )