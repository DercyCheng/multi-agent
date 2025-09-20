"""
Completion API Routes for Multi-Agent LLM Service
"""

import asyncio
import logging
import time
import uuid
from typing import Optional
from decimal import Decimal

from fastapi import APIRouter, Request, HTTPException, Depends
from fastapi.responses import StreamingResponse
from sse_starlette.sse import EventSourceResponse

from src.core.models import (
    CompletionRequest, CompletionResponse, StreamResponse,
    ErrorResponse, Usage
)
from src.core.model_router import ModelRouter
from src.core.context_engine import ContextEngine
from src.core.token_manager import TokenManager
from src.integrations.mcp_client import MCPClient
from src.middleware.auth import get_current_user
from src.utils.security import SecurityValidator

logger = logging.getLogger(__name__)

router = APIRouter()

async def get_model_router(request: Request) -> ModelRouter:
    """Get model router from app state"""
    return request.app.state.model_router

async def get_context_engine(request: Request) -> ContextEngine:
    """Get context engine from app state"""
    return request.app.state.context_engine

async def get_token_manager(request: Request) -> TokenManager:
    """Get token manager from app state"""
    return request.app.state.token_manager

async def get_mcp_client(request: Request) -> MCPClient:
    """Get MCP client from app state"""
    return request.app.state.mcp_client

@router.post("/chat/completions", response_model=CompletionResponse)
async def create_completion(
    request: CompletionRequest,
    model_router: ModelRouter = Depends(get_model_router),
    context_engine: ContextEngine = Depends(get_context_engine),
    token_manager: TokenManager = Depends(get_token_manager),
    mcp_client: MCPClient = Depends(get_mcp_client),
    current_user: dict = Depends(get_current_user)
):
    """Create a chat completion"""
    
    start_time = time.time()
    request_id = str(uuid.uuid4())
    
    try:
        # Validate request
        await _validate_completion_request(request, current_user)
        
        # Select optimal model
        model_key, model_info = await model_router.select_optimal_model(request)
        
        # Estimate cost
        estimated_cost = await token_manager.estimate_cost(
            request, model_info.id, model_info.provider
        )
        
        # Reserve budget
        budget_reserved = await token_manager.reserve_budget(
            request.user_id, request.tenant_id, estimated_cost, request_id
        )
        
        if not budget_reserved:
            raise HTTPException(
                status_code=402,
                detail="Insufficient budget for request"
            )
        
        try:
            # Engineer context if needed
            if request.context_id:
                from src.core.models import ContextRequest
                context_request = ContextRequest(
                    query=request.messages[-1].content if request.messages else "",
                    user_id=request.user_id,
                    session_id=request.session_id,
                    task_type="general",
                    available_tools=[tool.function.name for tool in request.tools] if request.tools else []
                )
                
                engineered_context = await context_engine.engineer_context(context_request)
                
                # Inject context into request
                if engineered_context.system_instructions:
                    # Add or update system message
                    system_msg_found = False
                    for msg in request.messages:
                        if msg.role == "system":
                            msg.content = engineered_context.system_instructions
                            system_msg_found = True
                            break
                    
                    if not system_msg_found:
                        from src.core.models import Message, MessageRole
                        system_msg = Message(
                            role=MessageRole.SYSTEM,
                            content=engineered_context.system_instructions
                        )
                        request.messages.insert(0, system_msg)
                
                # Add knowledge to last user message
                if engineered_context.knowledge:
                    last_user_msg = None
                    for msg in reversed(request.messages):
                        if msg.role == "user":
                            last_user_msg = msg
                            break
                    
                    if last_user_msg:
                        last_user_msg.content += f"\n\nRelevant context:\n{engineered_context.knowledge}"
                
                # Add tools
                if engineered_context.tools and not request.tools:
                    request.tools = engineered_context.tools
            
            # Handle streaming
            if request.stream:
                return StreamingResponse(
                    _stream_completion(
                        request, model_key, model_router, token_manager, 
                        context_engine, request_id, start_time
                    ),
                    media_type="text/event-stream"
                )
            
            # Execute completion
            response = await model_router.execute_completion(model_key, request)
            
            # Calculate actual cost
            if response.usage:
                cost_calc = await token_manager.calculate_cost(
                    model_info.id, model_info.provider, response.usage
                )
                
                # Consume budget
                await token_manager.consume_budget(
                    request.user_id, request.tenant_id, cost_calc.cost_usd,
                    request_id, response.usage, model_info.id, model_info.provider
                )
                
                # Add cost to response
                response.cost_usd = float(cost_calc.cost_usd)
            
            # Add execution metadata
            response.execution_id = request_id
            response.latency_ms = int((time.time() - start_time) * 1000)
            
            # Store conversation memory
            if context_engine:
                await context_engine.store_conversation_memory(
                    request.user_id, request.session_id, request.messages + [response.choices[0].message]
                )
            
            return response
            
        except Exception as e:
            # Release reserved budget on error
            await token_manager.release_budget(request.user_id, request.tenant_id, request_id)
            raise
            
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Completion error: {e}", exc_info=True)
        raise HTTPException(status_code=500, detail=str(e))

async def _stream_completion(
    request: CompletionRequest,
    model_key: str,
    model_router: ModelRouter,
    token_manager: TokenManager,
    context_engine: ContextEngine,
    request_id: str,
    start_time: float
):
    """Stream completion response"""
    
    try:
        # Get model info for cost calculation
        model_info = model_router.models[model_key]
        
        # Track tokens for streaming
        total_tokens = 0
        completion_tokens = 0
        
        async for chunk in model_router.providers[model_key.split(":")[0]].stream_complete(request):
            # Convert to SSE format
            chunk_data = chunk.dict()
            
            # Track tokens (rough estimation for streaming)
            if chunk.choices and chunk.choices[0].delta.content:
                content_tokens = len(chunk.choices[0].delta.content) // 4
                completion_tokens += content_tokens
                total_tokens += content_tokens
            
            yield f"data: {chunk.json()}\n\n"
            
            # Check if stream is complete
            if chunk.choices and chunk.choices[0].finish_reason:
                break
        
        # Send final usage and cost information
        if total_tokens > 0:
            # Estimate prompt tokens
            prompt_tokens = sum(len(msg.content) for msg in request.messages) // 4
            
            usage = Usage(
                prompt_tokens=prompt_tokens,
                completion_tokens=completion_tokens,
                total_tokens=prompt_tokens + completion_tokens
            )
            
            # Calculate and consume budget
            cost_calc = await token_manager.calculate_cost(
                model_info.id, model_info.provider, usage
            )
            
            await token_manager.consume_budget(
                request.user_id, request.tenant_id, cost_calc.cost_usd,
                request_id, usage, model_info.id, model_info.provider
            )
            
            # Send usage data
            usage_data = {
                "id": request_id,
                "object": "chat.completion.chunk",
                "created": int(time.time()),
                "model": model_info.id,
                "usage": usage.dict(),
                "cost_usd": float(cost_calc.cost_usd),
                "latency_ms": int((time.time() - start_time) * 1000)
            }
            
            yield f"data: {json.dumps(usage_data)}\n\n"
        
        yield "data: [DONE]\n\n"
        
    except Exception as e:
        # Release budget on streaming error
        await token_manager.release_budget(request.user_id, request.tenant_id, request_id)
        
        error_data = {
            "error": {
                "message": str(e),
                "type": "stream_error"
            }
        }
        yield f"data: {json.dumps(error_data)}\n\n"

@router.get("/models")
async def list_models(
    model_router: ModelRouter = Depends(get_model_router),
    current_user: dict = Depends(get_current_user)
):
    """List available models"""
    
    try:
        all_models = []
        
        for provider_name, provider in model_router.providers.items():
            models = await provider.list_models()
            all_models.extend(models)
        
        return {
            "object": "list",
            "data": [
                {
                    "id": model.id,
                    "object": "model",
                    "created": int(time.time()),
                    "owned_by": model.provider,
                    "permission": [],
                    "root": model.id,
                    "parent": None,
                    "max_tokens": model.max_tokens,
                    "context_length": model.context_length,
                    "cost_per_token": model.cost_per_token,
                    "supports_streaming": model.supports_streaming,
                    "supports_tools": model.supports_tools,
                    "supports_vision": model.supports_vision,
                    "capability_score": model.capability_score
                }
                for model in all_models
            ]
        }
        
    except Exception as e:
        logger.error(f"List models error: {e}")
        raise HTTPException(status_code=500, detail=str(e))

@router.get("/models/{model_id}")
async def get_model(
    model_id: str,
    model_router: ModelRouter = Depends(get_model_router),
    current_user: dict = Depends(get_current_user)
):
    """Get specific model information"""
    
    try:
        # Find model across all providers
        for provider_name, provider in model_router.providers.items():
            models = await provider.list_models()
            for model in models:
                if model.id == model_id:
                    return {
                        "id": model.id,
                        "object": "model",
                        "created": int(time.time()),
                        "owned_by": model.provider,
                        "permission": [],
                        "root": model.id,
                        "parent": None,
                        "max_tokens": model.max_tokens,
                        "context_length": model.context_length,
                        "cost_per_token": model.cost_per_token,
                        "supports_streaming": model.supports_streaming,
                        "supports_tools": model.supports_tools,
                        "supports_vision": model.supports_vision,
                        "capability_score": model.capability_score
                    }
        
        raise HTTPException(status_code=404, detail=f"Model {model_id} not found")
        
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Get model error: {e}")
        raise HTTPException(status_code=500, detail=str(e))

async def _validate_completion_request(request: CompletionRequest, current_user: dict):
    """Validate completion request"""
    
    # Basic validation
    if not request.messages:
        raise HTTPException(status_code=400, detail="Messages cannot be empty")
    
    if not request.user_id or not request.tenant_id:
        raise HTTPException(status_code=400, detail="User ID and Tenant ID are required")
    
    # Security validation
    security_validator = SecurityValidator()
    
    for message in request.messages:
        validation_result = await security_validator.validate_content(message.content)
        
        if not validation_result.is_safe:
            raise HTTPException(
                status_code=400,
                detail=f"Content validation failed: {', '.join(validation_result.violations)}"
            )
    
    # Token limits
    total_content = " ".join(msg.content for msg in request.messages)
    estimated_tokens = len(total_content) // 4
    
    if estimated_tokens > 100000:  # Reasonable limit
        raise HTTPException(
            status_code=400,
            detail="Request too large. Please reduce the content length."
        )
    
    # Rate limiting (basic check - more sophisticated rate limiting in middleware)
    # This would be implemented based on user tier, etc.