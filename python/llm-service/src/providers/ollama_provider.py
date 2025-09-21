"""""""""""""""

Ollama Provider for Multi-Agent LLM Service

"""Enhanced Ollama Provider for Multi-Agent LLM Service



import asyncio"""Enhanced Ollama Provider for Multi-Agent LLM Service

import json

import logging

import time

import uuidimport asyncio"""Enhanced Ollama Provider for Multi-Agent LLM ServiceEnhanced Ollama Provider for Multi-Agent LLM Service

from typing import List, Optional, AsyncGenerator, Dict, Any

from datetime import datetimeimport json

from dataclasses import dataclass, field

import logging

import httpx

from httpx import AsyncClientimport time



from src.providers.base_provider import BaseProviderimport uuidimport asyncio

from src.core.models import (

    CompletionRequest, CompletionResponse, StreamResponse, ModelInfo,from typing import List, Optional, AsyncGenerator, Dict, Any

    Usage, Choice, StreamChoice, Message, MessageRole

)from datetime import datetimeimport json



logger = logging.getLogger(__name__)from dataclasses import dataclass, field



import loggingAdvanced implementation with multi-model support, health checking,Advanced implementation with multi-model support, health checking,

@dataclass

class ModelHealth:import httpx

    """Model health tracking"""

    model_id: strfrom httpx import AsyncClientimport time

    is_healthy: bool = True

    last_check: datetime = field(default_factory=datetime.utcnow)

    consecutive_failures: int = 0

    response_times: List[float] = field(default_factory=list)from src.providers.base_provider import BaseProviderimport uuidload balancing, and dynamic model management.load balancing, and dynamic model management.

    error_count: int = 0

    success_count: int = 0from src.core.models import (



    @property    CompletionRequest, CompletionResponse, StreamResponse, ModelInfo,from typing import List, Optional, AsyncGenerator, Dict, Any

    def average_response_time(self) -> float:

        if not self.response_times:    Usage, Choice, StreamChoice, Message, MessageRole

            return 0.0

        return sum(self.response_times[-10:]) / len(self.response_times[-10:]))from datetime import datetime""""""



    @property

    def success_rate(self) -> float:

        total = self.success_count + self.error_countlogger = logging.getLogger(__name__)from dataclasses import dataclass, field

        if total == 0:

            return 1.0

        return self.success_count / total





class OllamaProvider(BaseProvider):@dataclass

    """Ollama provider implementation with multi-model support"""

class ModelHealth:import httpx

    def __init__(self, config):

        super().__init__(config)    """Model health tracking"""

        

        # Initialize model tracking    model_id: strfrom httpx import AsyncClientimport asyncioimport asyncio

        self.model_health: Dict[str, ModelHealth] = {}

        self.available_models: Dict[str, ModelInfo] = {}    is_healthy: bool = True

        self.model_aliases: Dict[str, str] = {

            'default': 'mistral',    last_check: datetime = field(default_factory=datetime.utcnow)

            'chat': 'mistral',

            'code': 'codellama'    consecutive_failures: int = 0

        }

            response_times: List[float] = field(default_factory=list)from src.providers.base_provider import BaseProviderimport jsonimport json

        # Performance tracking

        self.request_counts: Dict[str, int] = {}    error_count: int = 0

        self.response_times: Dict[str, List[float]] = {}

            success_count: int = 0from src.core.models import (

        # Configuration

        self.health_check_interval = getattr(config, 'health_check_interval', 60)

        self.max_retries = getattr(config, 'max_retries', 3)

            @property    CompletionRequest, CompletionResponse, StreamResponse, ModelInfo,import loggingimport logging

        # Initialize available models

        self._initialize_default_models()    def average_response_time(self) -> float:

        

        # Start async initialization        if not self.response_times:    Usage, Choice, StreamChoice, Message, MessageRole

        asyncio.create_task(self._initialize_async())

            return 0.0

    def _initialize_default_models(self):

        """Initialize default models"""        return sum(self.response_times[-10:]) / len(self.response_times[-10:]))import timeimport time

        default_models = ['mistral', 'codellama', 'llama2', 'mixtral']

        

        for model_name in default_models:

            self.available_models[model_name] = ModelInfo(    @property

                id=model_name,

                provider="ollama",    def success_rate(self) -> float:

                name=f"{model_name} (ollama)",

                description=f"Local {model_name} model via Ollama",        total = self.success_count + self.error_countlogger = logging.getLogger(__name__)import uuidimport uuid

                max_tokens=4096,

                context_length=4096,        if total == 0:

                cost_per_token=0.0,

                supports_streaming=True,            return 1.0

                supports_tools=False,

                supports_vision=False,        return self.success_count / total

                capability_score=0.7

            )from typing import List, Optional, AsyncGenerator, Dict, Anyfrom typing import List, Optional, AsyncGenerator, Dict, Any

            self.model_health[model_name] = ModelHealth(model_id=model_name)



    async def _initialize_async(self):

        """Initialize async components"""class OllamaProvider(BaseProvider):@dataclass

        try:

            await self._discover_models()    """Enhanced Ollama provider with multi-model support"""

            logger.info(f"Ollama provider initialized with {len(self.available_models)} models")

        except Exception as e:class ModelHealth:from datetime import datetimefrom datetime import datetime, timedelta

            logger.error(f"Failed to initialize Ollama provider: {e}")

    def __init__(self, config):

    async def _discover_models(self):

        """Discover available models from Ollama server"""        super().__init__(config)    """Model health tracking"""

        try:

            async with AsyncClient(timeout=10.0) as client:        

                response = await client.get(f"{self.base_url}/api/tags")

                response.raise_for_status()        # Enhanced configuration    model_id: strfrom dataclasses import dataclass, fieldfrom dataclasses import dataclass, field

                data = response.json()

                        self.model_health: Dict[str, ModelHealth] = {}

                models = data.get('models', [])

                for model_data in models:        self.available_models: Dict[str, ModelInfo] = {}    is_healthy: bool = True

                    model_name = model_data.get('name', '').split(':')[0]

                    if model_name and model_name not in self.available_models:        self.model_aliases: Dict[str, str] = {

                        # Add discovered model

                        self.available_models[model_name] = ModelInfo(            'default': 'mistral',    last_check: datetime = field(default_factory=datetime.utcnow)

                            id=model_name,

                            provider="ollama",            'chat': 'mistral',

                            name=model_data.get('name', model_name),

                            description=f"Ollama model: {model_name}",            'code': 'codellama'    consecutive_failures: int = 0

                            max_tokens=self._get_model_max_tokens(model_name),

                            context_length=self._get_model_max_tokens(model_name),        }

                            cost_per_token=0.0,

                            supports_streaming=True,            response_times: List[float] = field(default_factory=list)import httpximport httpx

                            supports_tools=False,

                            supports_vision=self._model_supports_vision(model_name),        # Performance monitoring

                            capability_score=self._calculate_capability_score(model_name)

                        )        self.request_counts: Dict[str, int] = {}    error_count: int = 0

                        

                        if model_name not in self.model_health:        self.response_times: Dict[str, List[float]] = {}

                            self.model_health[model_name] = ModelHealth(model_id=model_name)

                            success_count: int = 0from httpx import AsyncClient, ConnectTimeout, ReadTimeoutfrom httpx import AsyncClient, ConnectTimeout, ReadTimeout

                logger.info(f"Discovered {len(models)} models from Ollama")

                        # Health checking

        except Exception as e:

            logger.error(f"Failed to discover models: {e}")        self.health_check_interval = getattr(config, 'health_check_interval', 60)



    def _get_model_max_tokens(self, model_name: str) -> int:        self.max_retries = getattr(config, 'max_retries', 3)

        """Get model's maximum token limit"""

        token_limits = {            @property

            'tinyllama': 2048,

            'phi': 2048,        # Initialize async

            'mistral': 8192,

            'llama2': 4096,        asyncio.create_task(self._initialize_async())    def average_response_time(self) -> float:

            'codellama': 16384,

            'mixtral': 32768,

            'neural-chat': 4096

        }    async def _initialize_async(self):        if not self.response_times:from src.providers.base_provider import BaseProviderfrom src.providers.base_provider import BaseProvider

        

        for key, limit in token_limits.items():        """Initialize async components"""

            if key in model_name.lower():

                return limit        try:            return 0.0

        

        return 4096            await self._discover_models()



    def _model_supports_vision(self, model_name: str) -> bool:            logger.info(f"Ollama provider initialized with {len(self.available_models)} models")        return sum(self.response_times[-10:]) / len(self.response_times[-10:])from src.core.models import (from src.core.models import (

        """Check if model supports vision"""

        vision_models = ['llava', 'bakllava']        except Exception as e:

        return any(model in model_name.lower() for model in vision_models)

            logger.error(f"Failed to initialize Ollama provider: {e}")

    def _calculate_capability_score(self, model_name: str) -> float:

        """Calculate capability score based on model"""

        scores = {

            'mixtral': 0.95,    async def _discover_models(self):    @property    CompletionRequest, CompletionResponse, StreamResponse, ModelInfo,    CompletionRequest, CompletionResponse, StreamResponse, ModelInfo,

            'llama2': 0.85,

            'codellama': 0.80,        """Discover available models"""

            'mistral': 0.80,

            'neural-chat': 0.75,        try:    def success_rate(self) -> float:

            'phi': 0.60,

            'tinyllama': 0.50            async with AsyncClient(timeout=10.0) as client:

        }

                        response = await client.get(f"{self.base_url}/api/tags")        total = self.success_count + self.error_count    Usage, Choice, StreamChoice, Message, MessageRole    Usage, Choice, StreamChoice, Message, MessageRole

        for key, score in scores.items():

            if key in model_name.lower():                response.raise_for_status()

                return score

                        data = response.json()        if total == 0:

        return 0.70

                

    def _resolve_model_name(self, model: str) -> str:

        """Resolve model name from alias"""                models = data.get('models', [])            return 1.0))

        if model in self.model_aliases:

            return self.model_aliases[model]                for model_data in models:

        return model

                    model_name = model_data.get('name', '').split(':')[0]        return self.success_count / total

    async def complete(self, request: CompletionRequest) -> CompletionResponse:

        """Complete a chat completion request"""                    if model_name:

        model_name = self._resolve_model_name(request.model or 'default')

                                model_info = ModelInfo(

        # Fallback to available model if requested model not found

        if model_name not in self.available_models:                            id=model_name,

            if self.available_models:

                model_name = list(self.available_models.keys())[0]                            provider="ollama",

            else:

                raise ValueError("No models available")                            name=model_data.get('name', model_name),

        

        for attempt in range(self.max_retries):                            description=f"Ollama model: {model_name}",class OllamaProvider(BaseProvider):logger = logging.getLogger(__name__)logger = logging.getLogger(__name__)

            try:

                start_time = time.time()                            max_tokens=4096,

                

                async with AsyncClient(timeout=self.timeout) as client:                            context_length=4096,    """Enhanced Ollama provider with multi-model support"""

                    payload = {

                        "model": model_name,                            cost_per_token=0.0,

                        "messages": self._format_messages(request.messages),

                        "stream": False,                            supports_streaming=True,

                        "options": {}

                    }                            supports_tools=False,

                    

                    if request.temperature is not None:                            supports_vision=False,    def __init__(self, config):

                        payload["options"]["temperature"] = request.temperature

                    if request.max_tokens is not None:                            capability_score=0.7

                        payload["options"]["num_predict"] = request.max_tokens

                                            )        super().__init__(config)

                    response = await client.post(f"{self.base_url}/api/chat", json=payload)

                    response.raise_for_status()                        

                    data = response.json()

                                            self.available_models[model_name] = model_info        

                    response_time = time.time() - start_time

                    self._update_metrics(model_name, response_time, True)                        

                    

                    message_content = data.get('message', {}).get('content', '')                        if model_name not in self.model_health:        # Enhanced configuration@dataclass@dataclass

                    

                    return CompletionResponse(                            self.model_health[model_name] = ModelHealth(model_id=model_name)

                        id=str(uuid.uuid4()),

                        model=model_name,                        self.model_health: Dict[str, ModelHealth] = {}

                        choices=[Choice(

                            index=0,                logger.info(f"Discovered {len(models)} models from Ollama")

                            message=Message(

                                role=MessageRole.ASSISTANT,                        self.available_models: Dict[str, ModelInfo] = {}class ModelHealth:class ModelHealth:

                                content=message_content

                            ),        except Exception as e:

                            finish_reason="stop"

                        )],            logger.error(f"Failed to discover models: {e}")        self.model_aliases: Dict[str, str] = {

                        usage=Usage(

                            prompt_tokens=data.get('prompt_eval_count', 0),            # Fallback

                            completion_tokens=data.get('eval_count', 0),

                            total_tokens=data.get('prompt_eval_count', 0) + data.get('eval_count', 0)            configured_models = getattr(self.config, "models", ["mistral"]) or ["mistral"]            'default': 'mistral',    """Model health status tracking"""    """Model health status tracking"""

                        )

                    )            for model in configured_models:

                    

            except Exception as e:                self.available_models[model] = ModelInfo(            'chat': 'mistral',

                self._update_metrics(model_name, 0, False)

                if attempt == self.max_retries - 1:                    id=model,

                    raise

                logger.warning(f"Attempt {attempt + 1} failed: {e}")                    provider="ollama",            'code': 'codellama'    model_id: str    model_id: str



    async def stream_complete(self, request: CompletionRequest) -> AsyncGenerator[StreamResponse, None]:                    name=f"{model} (ollama)",

        """Stream a chat completion request"""

        model_name = self._resolve_model_name(request.model or 'default')                    description=f"Local {model} model via Ollama",        }

        

        if model_name not in self.available_models:                    max_tokens=4096,

            if self.available_models:

                model_name = list(self.available_models.keys())[0]                    context_length=4096,            is_healthy: bool = True    is_healthy: bool = True

            else:

                raise ValueError("No models available")                    cost_per_token=0.0,

        

        try:                    supports_streaming=True,        # Performance monitoring

            start_time = time.time()

                                supports_tools=False,

            async with AsyncClient(timeout=self.timeout) as client:

                payload = {                    supports_vision=False,        self.request_counts: Dict[str, int] = {}    last_check: datetime = field(default_factory=datetime.utcnow)    last_check: datetime = field(default_factory=datetime.utcnow)

                    "model": model_name,

                    "messages": self._format_messages(request.messages),                    capability_score=0.7

                    "stream": True,

                    "options": {}                )        self.response_times: Dict[str, List[float]] = {}

                }

                                if model not in self.model_health:

                if request.temperature is not None:

                    payload["options"]["temperature"] = request.temperature                    self.model_health[model] = ModelHealth(model_id=model)            consecutive_failures: int = 0    consecutive_failures: int = 0

                if request.max_tokens is not None:

                    payload["options"]["num_predict"] = request.max_tokens

                

                async with client.stream('POST', f"{self.base_url}/api/chat", json=payload) as response:    def _resolve_model_name(self, model: str) -> str:        # Health checking

                    response.raise_for_status()

                            """Resolve model name from alias"""

                    async for line in response.aiter_lines():

                        if line.strip():        if model in self.model_aliases:        self.health_check_interval = getattr(config, 'health_check_interval', 60)    response_times: List[float] = field(default_factory=list)    response_times: List[float] = field(default_factory=list)

                            try:

                                data = json.loads(line)            return self.model_aliases[model]

                                

                                if 'message' in data:        return model        self.max_retries = getattr(config, 'max_retries', 3)

                                    content = data['message'].get('content', '')

                                    if content:

                                        yield StreamResponse(

                                            id=str(uuid.uuid4()),    async def complete(self, request: CompletionRequest) -> CompletionResponse:            error_count: int = 0    error_count: int = 0

                                            model=model_name,

                                            choices=[StreamChoice(        """Enhanced completion with health tracking"""

                                                index=0,

                                                delta=Message(        model_name = self._resolve_model_name(request.model or 'default')        # Background tasks

                                                    role=MessageRole.ASSISTANT,

                                                    content=content        

                                                ),

                                                finish_reason=None        if model_name not in self.available_models:        self._health_check_task = None    success_count: int = 0    success_count: int = 0

                                            )]

                                        )            if self.available_models:

                                

                                if data.get('done', False):                model_name = list(self.available_models.keys())[0]        self._model_discovery_task = None

                                    response_time = time.time() - start_time

                                    self._update_metrics(model_name, response_time, True)            else:

                                    

                                    yield StreamResponse(                raise ValueError("No models available")        

                                        id=str(uuid.uuid4()),

                                        model=model_name,        

                                        choices=[StreamChoice(

                                            index=0,        for attempt in range(self.max_retries):        # Initialize

                                            delta=Message(role=MessageRole.ASSISTANT, content=""),

                                            finish_reason="stop"            try:

                                        )]

                                    )                start_time = time.time()        asyncio.create_task(self._initialize_async())    @property    @property

                                    break

                                                    

                            except json.JSONDecodeError:

                                continue                async with AsyncClient(timeout=self.timeout) as client:

                                

        except Exception as e:                    payload = {

            self._update_metrics(model_name, 0, False)

            raise                        "model": model_name,    async def _initialize_async(self):    def average_response_time(self) -> float:    def average_response_time(self) -> float:



    def _update_metrics(self, model_name: str, response_time: float, success: bool):                        "messages": self._format_messages(request.messages),

        """Update performance metrics"""

        if model_name not in self.request_counts:                        "stream": False,        """Initialize async components"""

            self.request_counts[model_name] = 0

            self.response_times[model_name] = []                        "options": {}

        

        self.request_counts[model_name] += 1                    }        try:        """Calculate average response time"""        """Calculate average response time"""

        

        if success and response_time > 0:                    

            self.response_times[model_name].append(response_time)

            # Keep only last 1000 response times                    if request.temperature is not None:            await self._discover_models()

            if len(self.response_times[model_name]) > 1000:

                self.response_times[model_name] = self.response_times[model_name][-1000:]                        payload["options"]["temperature"] = request.temperature

        

        # Update health tracking                    if request.max_tokens is not None:            self._start_background_tasks()        if not self.response_times:        if not self.response_times:

        if model_name in self.model_health:

            health = self.model_health[model_name]                        payload["options"]["num_predict"] = request.max_tokens

            if success:

                health.success_count += 1                                logger.info(f"Ollama provider initialized with {len(self.available_models)} models")

                health.consecutive_failures = 0

                if response_time > 0:                    response = await client.post(f"{self.base_url}/api/chat", json=payload)

                    health.response_times.append(response_time)

                    # Keep only last 100 response times for health tracking                    response.raise_for_status()        except Exception as e:            return 0.0            return 0.0

                    if len(health.response_times) > 100:

                        health.response_times = health.response_times[-100:]                    data = response.json()

            else:

                health.error_count += 1                                logger.error(f"Failed to initialize Ollama provider: {e}")

                health.consecutive_failures += 1

                    response_time = time.time() - start_time

    async def list_models(self) -> List[ModelInfo]:

        """List available models with health information"""                    self._update_metrics(model_name, response_time, True)        return sum(self.response_times[-10:]) / len(self.response_times[-10:])        return sum(self.response_times[-10:]) / len(self.response_times[-10:])  # Last 10 responses

        models = []

        for model_name, model_info in self.available_models.items():                    

            health = self.model_health.get(model_name)

            if health:                    message_content = data.get('message', {}).get('content', '')    def _start_background_tasks(self):

                # Enhance description with health info

                health_status = "healthy" if health.is_healthy else "unhealthy"                    

                avg_time = health.average_response_time

                success_rate = health.success_rate                    return CompletionResponse(        """Start background tasks"""

                

                enhanced_desc = f"{model_info.description} [Status: {health_status}, Avg: {avg_time:.2f}s, Success: {success_rate:.1%}]"                        id=str(uuid.uuid4()),

                

                enhanced_model = ModelInfo(                        model=model_name,        self._health_check_task = asyncio.create_task(self._health_check_loop())

                    id=model_info.id,

                    provider=model_info.provider,                        choices=[Choice(

                    name=model_info.name,

                    description=enhanced_desc,                            index=0,        self._model_discovery_task = asyncio.create_task(self._model_discovery_loop())    @property    @property

                    max_tokens=model_info.max_tokens,

                    context_length=model_info.context_length,                            message=Message(

                    cost_per_token=model_info.cost_per_token,

                    supports_streaming=model_info.supports_streaming,                                role=MessageRole.ASSISTANT,

                    supports_tools=model_info.supports_tools,

                    supports_vision=model_info.supports_vision,                                content=message_content

                    capability_score=model_info.capability_score * success_rate

                )                            ),    async def _discover_models(self):    def success_rate(self) -> float:    def success_rate(self) -> float:

                models.append(enhanced_model)

            else:                            finish_reason="stop"

                models.append(model_info)

                                )],        """Discover available models"""

        return models

                        usage=Usage(

    async def validate_api_key(self) -> bool:

        """Validate API connection to Ollama server"""                            prompt_tokens=data.get('prompt_eval_count', 0),        try:        """Calculate success rate"""        """Calculate success rate"""

        try:

            async with AsyncClient(timeout=5.0) as client:                            completion_tokens=data.get('eval_count', 0),

                response = await client.get(f"{self.base_url}/api/version")

                response.raise_for_status()                            total_tokens=data.get('prompt_eval_count', 0) + data.get('eval_count', 0)            async with AsyncClient(timeout=10.0) as client:

                return True

        except Exception:                        )

            return False

                    )                response = await client.get(f"{self.base_url}/api/tags")        total = self.success_count + self.error_count        total = self.success_count + self.error_count

    def get_model_stats(self) -> Dict[str, Any]:

        """Get comprehensive model statistics"""                    

        return {

            "total_models": len(self.available_models),            except Exception as e:                response.raise_for_status()

            "healthy_models": sum(1 for h in self.model_health.values() if h.is_healthy),

            "total_requests": sum(self.request_counts.values()),                self._update_metrics(model_name, 0, False)

            "model_aliases": self.model_aliases,

            "models": {                if attempt == self.max_retries - 1:                data = response.json()        if total == 0:        if total == 0:

                name: {

                    "healthy": health.is_healthy,                    raise

                    "requests": self.request_counts.get(name, 0),

                    "avg_response_time": health.average_response_time,                logger.warning(f"Attempt {attempt + 1} failed: {e}")                

                    "success_rate": health.success_rate,

                    "consecutive_failures": health.consecutive_failures

                }

                for name, health in self.model_health.items()    async def stream_complete(self, request: CompletionRequest) -> AsyncGenerator[StreamResponse, None]:                models = data.get('models', [])            return 1.0            return 1.0

            }

        }        """Enhanced streaming with health tracking"""



    def _format_messages(self, messages: List[Message]) -> List[Dict[str, str]]:        model_name = self._resolve_model_name(request.model or 'default')                for model_data in models:

        """Format messages for Ollama API"""

        return [        

            {

                "role": msg.role.value,        if model_name not in self.available_models:                    model_name = model_data.get('name', '').split(':')[0]        return self.success_count / total        return self.success_count / total

                "content": msg.content

            }            if self.available_models:

            for msg in messages

        ]                model_name = list(self.available_models.keys())[0]                    if model_name:

            else:

                raise ValueError("No models available")                        # Create enhanced model info

        

        try:                        model_info = ModelInfo(

            start_time = time.time()

                                        id=model_name,

            async with AsyncClient(timeout=self.timeout) as client:

                payload = {                            provider="ollama",

                    "model": model_name,

                    "messages": self._format_messages(request.messages),                            name=model_data.get('name', model_name),@dataclass@dataclass

                    "stream": True,

                    "options": {}                            description=self._get_model_description(model_name),

                }

                                            max_tokens=self._get_model_max_tokens(model_name),class OllamaServer:class OllamaServer:

                if request.temperature is not None:

                    payload["options"]["temperature"] = request.temperature                            context_length=self._get_model_context_length(model_name),

                if request.max_tokens is not None:

                    payload["options"]["num_predict"] = request.max_tokens                            cost_per_token=0.0,    """Ollama server instance configuration"""    """Ollama server instance configuration"""

                

                async with client.stream('POST', f"{self.base_url}/api/chat", json=payload) as response:                            supports_streaming=True,

                    response.raise_for_status()

                                                supports_tools=self._model_supports_tools(model_name),    base_url: str    base_url: str

                    async for line in response.aiter_lines():

                        if line.strip():                            supports_vision=self._model_supports_vision(model_name),

                            try:

                                data = json.loads(line)                            capability_score=self._calculate_capability_score(model_name)    weight: float = 1.0    weight: float = 1.0

                                

                                if 'message' in data:                        )

                                    content = data['message'].get('content', '')

                                    if content:                            max_concurrent: int = 10    max_concurrent: int = 10

                                        yield StreamResponse(

                                            id=str(uuid.uuid4()),                        self.available_models[model_name] = model_info

                                            model=model_name,

                                            choices=[StreamChoice(                            timeout: int = 30    timeout: int = 30

                                                index=0,

                                                delta=Message(                        # Initialize health tracking

                                                    role=MessageRole.ASSISTANT,

                                                    content=content                        if model_name not in self.model_health:    is_healthy: bool = True    is_healthy: bool = True

                                                ),

                                                finish_reason=None                            self.model_health[model_name] = ModelHealth(model_id=model_name)

                                            )]

                                        )                    last_health_check: datetime = field(default_factory=datetime.utcnow)    last_health_check: datetime = field(default_factory=datetime.utcnow)

                                

                                if data.get('done', False):                logger.info(f"Discovered {len(models)} models from Ollama")

                                    response_time = time.time() - start_time

                                    self._update_metrics(model_name, response_time, True)                    current_load: int = 0    current_load: int = 0

                                    

                                    yield StreamResponse(        except Exception as e:

                                        id=str(uuid.uuid4()),

                                        model=model_name,            logger.error(f"Failed to discover models: {e}")

                                        choices=[StreamChoice(

                                            index=0,            # Fallback to configured models

                                            delta=Message(role=MessageRole.ASSISTANT, content=""),

                                            finish_reason="stop"            configured_models = getattr(self.config, "models", ["mistral"]) or ["mistral"]    @property    @property

                                        )]

                                    )            for model in configured_models:

                                    break

                                                    self.available_models[model] = ModelInfo(    def load_factor(self) -> float:    def load_factor(self) -> float:

                            except json.JSONDecodeError:

                                continue                    id=model,

                                

        except Exception as e:                    provider="ollama",        """Calculate current load factor"""        """Calculate current load factor"""

            self._update_metrics(model_name, 0, False)

            raise                    name=f"{model} (ollama)",



    def _update_metrics(self, model_name: str, response_time: float, success: bool):                    description=f"Local {model} model via Ollama",        if self.max_concurrent == 0:        if self.max_concurrent == 0:

        """Update performance metrics"""

        if model_name not in self.request_counts:                    max_tokens=8192,

            self.request_counts[model_name] = 0

            self.response_times[model_name] = []                    context_length=8192,            return 1.0            return 1.0

        

        self.request_counts[model_name] += 1                    cost_per_token=0.0,

        

        if success and response_time > 0:                    supports_streaming=True,        return self.current_load / self.max_concurrent        return self.current_load / self.max_concurrent

            self.response_times[model_name].append(response_time)

            if len(self.response_times[model_name]) > 1000:                    supports_tools=False,

                self.response_times[model_name] = self.response_times[model_name][-1000:]

                            supports_vision=False,

        if model_name in self.model_health:

            health = self.model_health[model_name]                    capability_score=0.7

            if success:

                health.success_count += 1                )

                health.consecutive_failures = 0

                if response_time > 0:                if model not in self.model_health:

                    health.response_times.append(response_time)

                    if len(health.response_times) > 100:                    self.model_health[model] = ModelHealth(model_id=model)class OllamaProvider(BaseProvider):class EnhancedOllamaProvider(BaseProvider):

                        health.response_times = health.response_times[-100:]

            else:

                health.error_count += 1

                health.consecutive_failures += 1    def _get_model_description(self, model_name: str) -> str:    """Enhanced Ollama provider with multi-server and multi-model support"""    """Enhanced Ollama provider with multi-server and multi-model support"""



    async def list_models(self) -> List[ModelInfo]:        """Get enhanced model description"""

        """List available models with health info"""

        models = []        descriptions = {

        for model_name, model_info in self.available_models.items():

            health = self.model_health.get(model_name)            'llama2': 'Llama 2 - Open foundation and fine-tuned chat models',

            if health:

                health_status = "healthy" if health.is_healthy else "unhealthy"            'codellama': 'Code Llama - Code generation model based on Llama 2',    def __init__(self, config):    def __init__(self, config):

                avg_time = health.average_response_time

                success_rate = health.success_rate            'mistral': 'Mistral 7B - High-quality language model',

                

                enhanced_desc = f"{model_info.description} [Status: {health_status}, Avg: {avg_time:.2f}s, Success: {success_rate:.1%}]"            'mixtral': 'Mixtral 8x7B - Mixture of experts model',        super().__init__(config)        super().__init__(config)

                

                enhanced_model = ModelInfo(            'neural-chat': 'Neural Chat - Fine-tuned for conversational use cases',

                    id=model_info.id,

                    provider=model_info.provider,            'phi': 'Phi - Small language model by Microsoft',                

                    name=model_info.name,

                    description=enhanced_desc,            'tinyllama': 'TinyLlama - Compact Llama model'

                    max_tokens=model_info.max_tokens,

                    context_length=model_info.context_length,        }        # Multi-server configuration        # Multi-server configuration

                    cost_per_token=model_info.cost_per_token,

                    supports_streaming=model_info.supports_streaming,        

                    supports_tools=model_info.supports_tools,

                    supports_vision=model_info.supports_vision,        for key, desc in descriptions.items():        self.servers = self._parse_servers(config)        self.servers = self._parse_servers(config)

                    capability_score=model_info.capability_score * success_rate

                )            if key in model_name.lower():

                models.append(enhanced_model)

            else:                return desc                self.current_server_index = 0

                models.append(model_info)

                

        return models

        return f"Ollama model: {model_name}"        # Model management        

    async def validate_api_key(self) -> bool:

        """Validate API connection"""

        try:

            async with AsyncClient(timeout=5.0) as client:    def _get_model_max_tokens(self, model_name: str) -> int:        self.model_health: Dict[str, ModelHealth] = {}        # Model management

                response = await client.get(f"{self.base_url}/api/version")

                response.raise_for_status()        """Get model's maximum token limit"""

                return True

        except Exception:        token_limits = {        self.available_models: Dict[str, ModelInfo] = {}        self.model_health: Dict[str, ModelHealth] = {}

            return False

            'tinyllama': 2048,

    def get_model_stats(self) -> Dict[str, Any]:

        """Get comprehensive model statistics"""            'phi': 2048,        self.model_aliases: Dict[str, str] = {}        self.available_models: Dict[str, ModelInfo] = {}

        return {

            "total_models": len(self.available_models),            'mistral': 8192,

            "healthy_models": sum(1 for h in self.model_health.values() if h.is_healthy),

            "total_requests": sum(self.request_counts.values()),            'llama2': 4096,                self.model_aliases: Dict[str, str] = {}

            "models": {

                name: {            'codellama': 16384,

                    "healthy": health.is_healthy,

                    "requests": self.request_counts.get(name, 0),            'mixtral': 32768,        # Configuration        

                    "avg_response_time": health.average_response_time,

                    "success_rate": health.success_rate,            'neural-chat': 4096

                    "consecutive_failures": health.consecutive_failures

                }        }        self.health_check_interval = getattr(config, 'health_check_interval', 60)        # Load balancing and health checking

                for name, health in self.model_health.items()

            },        

            "aliases": self.model_aliases

        }        for key, limit in token_limits.items():        self.max_retries = getattr(config, 'max_retries', 3)        self.health_check_interval = getattr(config, 'health_check_interval', 60)



    def _format_messages(self, messages: List[Message]) -> List[Dict[str, str]]:            if key in model_name.lower():

        """Format messages for Ollama API"""

        return [                return limit        self.circuit_breaker_threshold = getattr(config, 'circuit_breaker_threshold', 5)        self.max_retries = getattr(config, 'max_retries', 3)

            {

                "role": msg.role.value,        

                "content": msg.content

            }        return 4096                self.circuit_breaker_threshold = getattr(config, 'circuit_breaker_threshold', 5)

            for msg in messages

        ]

    def _get_model_context_length(self, model_name: str) -> int:        # Performance monitoring        

        """Get model's context length"""

        return self._get_model_max_tokens(model_name)        self.request_counts: Dict[str, int] = {}        # Performance monitoring



    def _model_supports_tools(self, model_name: str) -> bool:        self.response_times: Dict[str, List[float]] = {}        self.request_counts: Dict[str, int] = {}

        """Check if model supports function calling"""

        tool_capable = ['mixtral', 'llama2', 'mistral']                self.response_times: Dict[str, List[float]] = {}

        return any(model in model_name.lower() for model in tool_capable)

        # Background tasks        

    def _model_supports_vision(self, model_name: str) -> bool:

        """Check if model supports vision"""        self._health_check_task = None        # Background tasks

        vision_models = ['llava', 'bakllava']

        return any(model in model_name.lower() for model in vision_models)        self._model_discovery_task = None        self._health_check_task = None



    def _calculate_capability_score(self, model_name: str) -> float:                self._model_discovery_task = None

        """Calculate capability score"""

        scores = {        # Initialize        

            'mixtral': 0.95,

            'llama2': 0.85,        asyncio.create_task(self._initialize_async())        # Initialize async components

            'codellama': 0.80,

            'mistral': 0.80,        asyncio.create_task(self._initialize_async())

            'neural-chat': 0.75,

            'phi': 0.60,    def _parse_servers(self, config) -> List[OllamaServer]:

            'tinyllama': 0.50

        }        """Parse server configuration"""    def _format_messages(self, messages: List[Message]) -> List[Dict[str, str]]:

        

        for key, score in scores.items():        servers = []        """Format messages for Ollama API"""

            if key in model_name.lower():

                return score                return [

        

        return 0.70        # Single server configuration (backward compatibility)            {



    async def _health_check_loop(self):        base_url = getattr(config, 'base_url', 'http://localhost:11434')                "role": msg.role.value,

        """Background health checking"""

        while True:        servers.append(OllamaServer(                "content": msg.content

            try:

                await asyncio.sleep(self.health_check_interval)            base_url=base_url,            }

                await self._check_server_health()

            except Exception as e:            weight=1.0,            for msg in messages

                logger.error(f"Health check error: {e}")

            max_concurrent=getattr(config, 'max_concurrent', 10),        ]

    async def _check_server_health(self):

        """Check server health"""            timeout=getattr(config, 'timeout', 30)

        try:

            async with AsyncClient(timeout=5.0) as client:        ))

                response = await client.get(f"{self.base_url}/api/version")

                response.raise_for_status()        # Backward compatibility alias

            logger.debug(f"Ollama server health check passed")

        except Exception as e:        return serversOllamaProvider = EnhancedOllamaProvider

            logger.warning(f"Ollama server health check failed: {e}")



    async def _model_discovery_loop(self):

        """Background model discovery"""    async def _initialize_async(self):    def _parse_servers(self, config) -> List[OllamaServer]:

        while True:

            try:        """Initialize async components"""        """Parse server configuration"""

                await asyncio.sleep(300)  # Every 5 minutes

                await self._discover_models()        try:        servers = []

            except Exception as e:

                logger.error(f"Model discovery error: {e}")            await self._discover_models()        



    def _resolve_model_name(self, model: str) -> str:            self._health_check_task = asyncio.create_task(self._health_check_loop())        # Single server configuration (backward compatibility)

        """Resolve model name from alias"""

        if model in self.model_aliases:            self._model_discovery_task = asyncio.create_task(self._model_discovery_loop())        if hasattr(config, 'base_url'):

            return self.model_aliases[model]

        return model            logger.info(f"Ollama provider initialized with {len(self.available_models)} models")            servers.append(OllamaServer(



    async def complete(self, request: CompletionRequest) -> CompletionResponse:        except Exception as e:                base_url=config.base_url,

        """Enhanced completion with health tracking"""

        model_name = self._resolve_model_name(request.model or 'default')            logger.error(f"Failed to initialize Ollama provider: {e}")                weight=1.0,

        

        # Fallback to available model if requested model not found                max_concurrent=getattr(config, 'max_concurrent', 10),

        if model_name not in self.available_models:

            if self.available_models:    async def _discover_models(self):                timeout=getattr(config, 'timeout', 30)

                model_name = list(self.available_models.keys())[0]

            else:        """Discover available models from servers"""            ))

                raise ValueError("No models available")

                all_models = {}        

        for attempt in range(self.max_retries):

            try:                # Multi-server configuration

                start_time = time.time()

                        for server in self.servers:        if hasattr(config, 'servers'):

                async with AsyncClient(timeout=self.timeout) as client:

                    payload = {            try:            for server_config in config.servers:

                        "model": model_name,

                        "messages": self._format_messages(request.messages),                async with AsyncClient(timeout=10.0) as client:                servers.append(OllamaServer(

                        "stream": False,

                        "options": {}                    response = await client.get(f"{server.base_url}/api/tags")                    base_url=server_config.get('base_url'),

                    }

                                        response.raise_for_status()                    weight=server_config.get('weight', 1.0),

                    if request.temperature is not None:

                        payload["options"]["temperature"] = request.temperature                    data = response.json()                    max_concurrent=server_config.get('max_concurrent', 10),

                    if request.max_tokens is not None:

                        payload["options"]["num_predict"] = request.max_tokens                                        timeout=server_config.get('timeout', 30)

                    

                    response = await client.post(f"{self.base_url}/api/chat", json=payload)                    models = data.get('models', [])                ))

                    response.raise_for_status()

                    data = response.json()                    for model_data in models:        

                    

                    response_time = time.time() - start_time                        model_name = model_data.get('name', '').split(':')[0]        if not servers:

                    self._update_metrics(model_name, response_time, True)

                                            if model_name and model_name not in all_models:            # Default configuration

                    message_content = data.get('message', {}).get('content', '')

                                                model_info = ModelInfo(            servers.append(OllamaServer(

                    return CompletionResponse(

                        id=str(uuid.uuid4()),                                id=model_name,                base_url="http://localhost:11434",

                        model=model_name,

                        choices=[Choice(                                provider="ollama",                weight=1.0,

                            index=0,

                            message=Message(                                name=model_data.get('name', model_name),                max_concurrent=10,

                                role=MessageRole.ASSISTANT,

                                content=message_content                                description=f"Ollama model: {model_name}",                timeout=30

                            ),

                            finish_reason="stop"                                max_tokens=8192,            ))

                        )],

                        usage=Usage(                                context_length=8192,        

                            prompt_tokens=data.get('prompt_eval_count', 0),

                            completion_tokens=data.get('eval_count', 0),                                cost_per_token=0.0,        return servers

                            total_tokens=data.get('prompt_eval_count', 0) + data.get('eval_count', 0)

                        )                                supports_streaming=True,

                    )

                                                    supports_tools=False,    async def _initialize_async(self):

            except Exception as e:

                self._update_metrics(model_name, 0, False)                                supports_vision=False,        """Initialize async components"""

                if attempt == self.max_retries - 1:

                    raise                                capability_score=0.7        try:

                logger.warning(f"Attempt {attempt + 1} failed: {e}")

                            )            # Discover available models from all servers

    async def stream(self, request: CompletionRequest) -> AsyncGenerator[StreamResponse, None]:

        """Enhanced streaming with health tracking"""                                        await self._discover_models()

        model_name = self._resolve_model_name(request.model or 'default')

                                    all_models[model_name] = model_info            

        if model_name not in self.available_models:

            if self.available_models:                                        # Start background tasks

                model_name = list(self.available_models.keys())[0]

            else:                            if model_name not in self.model_health:            self._health_check_task = asyncio.create_task(self._health_check_loop())

                raise ValueError("No models available")

                                        self.model_health[model_name] = ModelHealth(model_id=model_name)            self._model_discovery_task = asyncio.create_task(self._model_discovery_loop())

        try:

            start_time = time.time()                                

            

            async with AsyncClient(timeout=self.timeout) as client:                    server.is_healthy = True            logger.info(f"Enhanced Ollama provider initialized with {len(self.servers)} servers and {len(self.available_models)} models")

                payload = {

                    "model": model_name,                    logger.info(f"Discovered {len(models)} models from {server.base_url}")            

                    "messages": self._format_messages(request.messages),

                    "stream": True,                            except Exception as e:

                    "options": {}

                }            except Exception as e:            logger.error(f"Failed to initialize Ollama provider: {e}")

                

                if request.temperature is not None:                logger.error(f"Failed to discover models from {server.base_url}: {e}")

                    payload["options"]["temperature"] = request.temperature

                if request.max_tokens is not None:                server.is_healthy = False    async def _discover_models(self):

                    payload["options"]["num_predict"] = request.max_tokens

                                """Discover available models from all servers"""

                async with client.stream('POST', f"{self.base_url}/api/chat", json=payload) as response:

                    response.raise_for_status()        self.available_models = all_models        all_models = {}

                    

                    async for line in response.aiter_lines():        self._setup_model_aliases()        

                        if line.strip():

                            try:        for server in self.servers:

                                data = json.loads(line)

                                    def _setup_model_aliases(self):            try:

                                if 'message' in data:

                                    content = data['message'].get('content', '')        """Setup model aliases"""                async with AsyncClient(timeout=10.0) as client:

                                    if content:

                                        yield StreamResponse(        self.model_aliases = {                    response = await client.get(f"{server.base_url}/api/tags")

                                            id=str(uuid.uuid4()),

                                            model=model_name,            'default': 'mistral',                    response.raise_for_status()

                                            choices=[StreamChoice(

                                                index=0,            'chat': 'mistral',                    data = response.json()

                                                delta=Message(

                                                    role=MessageRole.ASSISTANT,            'code': 'codellama'                    

                                                    content=content

                                                ),        }                    models = data.get('models', [])

                                                finish_reason=None

                                            )]                            for model_data in models:

                                        )

                                        # Filter to available models                        model_name = model_data.get('name', '').split(':')[0]  # Remove tag

                                if data.get('done', False):

                                    response_time = time.time() - start_time        available_aliases = {}                        if model_name and model_name not in all_models:

                                    self._update_metrics(model_name, response_time, True)

                                            for alias, model in self.model_aliases.items():                            # Create model info

                                    yield StreamResponse(

                                        id=str(uuid.uuid4()),            if model in self.available_models:                            model_info = ModelInfo(

                                        model=model_name,

                                        choices=[StreamChoice(                available_aliases[alias] = model                                id=model_name,

                                            index=0,

                                            delta=Message(role=MessageRole.ASSISTANT, content=""),                                        provider="ollama",

                                            finish_reason="stop"

                                        )]        self.model_aliases = available_aliases                                name=model_data.get('name', model_name),

                                    )

                                    break                                description=self._get_model_description(model_name),

                                    

                            except json.JSONDecodeError:    async def _health_check_loop(self):                                max_tokens=self._get_model_max_tokens(model_name),

                                continue

                                        """Background health checking"""                                context_length=self._get_model_context_length(model_name),

        except Exception as e:

            self._update_metrics(model_name, 0, False)        while True:                                cost_per_token=0.0,  # Local models are free

            raise

            try:                                supports_streaming=True,

    def _update_metrics(self, model_name: str, response_time: float, success: bool):

        """Update performance metrics"""                await asyncio.sleep(self.health_check_interval)                                supports_tools=self._model_supports_tools(model_name),

        if model_name not in self.request_counts:

            self.request_counts[model_name] = 0                await self._check_all_health()                                supports_vision=self._model_supports_vision(model_name),

            self.response_times[model_name] = []

                    except Exception as e:                                capability_score=self._calculate_capability_score(model_name)

        self.request_counts[model_name] += 1

                        logger.error(f"Health check loop error: {e}")                            )

        if success and response_time > 0:

            self.response_times[model_name].append(response_time)                            

            if len(self.response_times[model_name]) > 1000:

                self.response_times[model_name] = self.response_times[model_name][-1000:]    async def _check_all_health(self):                            all_models[model_name] = model_info

        

        if model_name in self.model_health:        """Check health of all servers and models"""                            

            health = self.model_health[model_name]

            if success:        for server in self.servers:                            # Initialize health tracking

                health.success_count += 1

                health.consecutive_failures = 0            await self._check_server_health(server)                            if model_name not in self.model_health:

                if response_time > 0:

                    health.response_times.append(response_time)                                self.model_health[model_name] = ModelHealth(model_id=model_name)

                    if len(health.response_times) > 100:

                        health.response_times = health.response_times[-100:]    async def _check_server_health(self, server: OllamaServer):                    

            else:

                health.error_count += 1        """Check health of a server"""                    server.is_healthy = True

                health.consecutive_failures += 1

        try:                    logger.info(f"Discovered {len(models)} models from {server.base_url}")

    async def list_models(self) -> List[ModelInfo]:

        """List available models with health info"""            async with AsyncClient(timeout=5.0) as client:                    

        models = []

        for model_name, model_info in self.available_models.items():                response = await client.get(f"{server.base_url}/api/version")            except Exception as e:

            health = self.model_health.get(model_name)

            if health:                response.raise_for_status()                logger.error(f"Failed to discover models from {server.base_url}: {e}")

                # Enhance description with health info

                health_status = "healthy" if health.is_healthy else "unhealthy"                            server.is_healthy = False

                avg_time = health.average_response_time

                success_rate = health.success_rate            server.is_healthy = True        

                

                enhanced_desc = f"{model_info.description} [Status: {health_status}, Avg: {avg_time:.2f}s, Success: {success_rate:.1%}]"            server.last_health_check = datetime.utcnow()        self.available_models = all_models

                

                enhanced_model = ModelInfo(                    self._setup_model_aliases()

                    id=model_info.id,

                    provider=model_info.provider,        except Exception as e:

                    name=model_info.name,

                    description=enhanced_desc,            server.is_healthy = False    def _get_model_description(self, model_name: str) -> str:

                    max_tokens=model_info.max_tokens,

                    context_length=model_info.context_length,            logger.warning(f"Server {server.base_url} health check failed: {e}")        """Get model description based on name"""

                    cost_per_token=model_info.cost_per_token,

                    supports_streaming=model_info.supports_streaming,        descriptions = {

                    supports_tools=model_info.supports_tools,

                    supports_vision=model_info.supports_vision,    async def _model_discovery_loop(self):            'llama2': 'Llama 2 - Open foundation and fine-tuned chat models',

                    capability_score=model_info.capability_score * success_rate

                )        """Background model discovery"""            'codellama': 'Code Llama - Code generation model based on Llama 2',

                models.append(enhanced_model)

            else:        while True:            'mistral': 'Mistral 7B - High-quality language model',

                models.append(model_info)

                    try:            'mixtral': 'Mixtral 8x7B - Mixture of experts model',

        return models

                await asyncio.sleep(300)  # Every 5 minutes            'neural-chat': 'Neural Chat - Fine-tuned for conversational use cases',

    def get_model_stats(self) -> Dict[str, Any]:

        """Get comprehensive model statistics"""                await self._discover_models()            'starcode': 'StarCoder - Code generation model',

        return {

            "total_models": len(self.available_models),            except Exception as e:            'vicuna': 'Vicuna - Open-source chatbot trained by fine-tuning LLaMA',

            "healthy_models": sum(1 for h in self.model_health.values() if h.is_healthy),

            "total_requests": sum(self.request_counts.values()),                logger.error(f"Model discovery loop error: {e}")            'orca-mini': 'Orca Mini - Compact version of Orca model',

            "models": {

                name: {            'phi': 'Phi - Small language model by Microsoft',

                    "healthy": health.is_healthy,

                    "requests": self.request_counts.get(name, 0),    def _select_best_server(self) -> Optional[OllamaServer]:            'tinyllama': 'TinyLlama - Compact Llama model for resource-constrained environments'

                    "avg_response_time": health.average_response_time,

                    "success_rate": health.success_rate,        """Select the best available server"""        }

                    "consecutive_failures": health.consecutive_failures

                }        healthy_servers = [s for s in self.servers if s.is_healthy]        

                for name, health in self.model_health.items()

            },                for key, desc in descriptions.items():

            "aliases": self.model_aliases

        }        if not healthy_servers:            if key in model_name.lower():



    async def shutdown(self):            logger.error("No healthy Ollama servers available")                return desc

        """Graceful shutdown"""

        logger.info("Shutting down Ollama Provider...")            return None        

        

        # Cancel background tasks                return f"Ollama model: {model_name}"

        tasks = [self._health_check_task, self._model_discovery_task]

        for task in tasks:        # Simple round-robin for now

            if task:

                task.cancel()        return healthy_servers[0]    def _get_model_max_tokens(self, model_name: str) -> int:

        

        # Wait for cancellation        """Get model's maximum token limit"""

        for task in tasks:

            if task:    def _resolve_model_name(self, model: str) -> str:        token_limits = {

                try:

                    await task        """Resolve model name from alias"""            'tinyllama': 2048,

                except asyncio.CancelledError:

                    pass        if model in self.model_aliases:            'phi': 2048,



    def _format_messages(self, messages: List[Message]) -> List[Dict[str, str]]:            return self.model_aliases[model]            'orca-mini': 2048,

        """Format messages for Ollama API"""

        return [        return model            'mistral': 8192,

            {

                "role": msg.role.value,            'llama2': 4096,

                "content": msg.content

            }    async def complete(self, request: CompletionRequest) -> CompletionResponse:            'codellama': 16384,

            for msg in messages

        ]        """Complete a request"""            'mixtral': 32768,

        model_name = self._resolve_model_name(request.model or 'default')            'neural-chat': 4096,

                    'starcode': 8192,

        if model_name not in self.available_models:            'vicuna': 4096

            # Fallback to first available model        }

            if self.available_models:        

                model_name = list(self.available_models.keys())[0]        for key, limit in token_limits.items():

            else:            if key in model_name.lower():

                raise ValueError("No models available")                return limit

                

        server = self._select_best_server()        return 4096  # Default

        if not server:

            raise ConnectionError("No healthy Ollama servers available")    def _get_model_context_length(self, model_name: str) -> int:

                """Get model's context length"""

        try:        # Context length is typically the same as max tokens for most models

            server.current_load += 1        return self._get_model_max_tokens(model_name)

            start_time = time.time()

                def _model_supports_tools(self, model_name: str) -> bool:

            async with AsyncClient(timeout=server.timeout) as client:        """Check if model supports function calling/tools"""

                payload = {        tool_capable_models = ['mixtral', 'llama2', 'mistral']

                    "model": model_name,        return any(model in model_name.lower() for model in tool_capable_models)

                    "messages": self._format_messages(request.messages),

                    "stream": False,    def _model_supports_vision(self, model_name: str) -> bool:

                    "options": {}        """Check if model supports vision/image understanding"""

                }        vision_models = ['llava', 'bakllava']

                        return any(model in model_name.lower() for model in vision_models)

                if request.temperature is not None:

                    payload["options"]["temperature"] = request.temperature    def _calculate_capability_score(self, model_name: str) -> float:

                if request.max_tokens is not None:        """Calculate capability score based on model characteristics"""

                    payload["options"]["num_predict"] = request.max_tokens        base_scores = {

                            'mixtral': 0.95,

                response = await client.post(f"{server.base_url}/api/chat", json=payload)            'llama2': 0.85,

                response.raise_for_status()            'codellama': 0.80,

                data = response.json()            'mistral': 0.80,

                            'neural-chat': 0.75,

                response_time = time.time() - start_time            'vicuna': 0.75,

                self._update_metrics(model_name, response_time, True)            'starcode': 0.70,

                            'orca-mini': 0.65,

                message_content = data.get('message', {}).get('content', '')            'phi': 0.60,

                            'tinyllama': 0.50

                return CompletionResponse(        }

                    id=str(uuid.uuid4()),        

                    model=model_name,        for key, score in base_scores.items():

                    choices=[Choice(            if key in model_name.lower():

                        index=0,                return score

                        message=Message(        

                            role=MessageRole.ASSISTANT,        return 0.70  # Default score

                            content=message_content

                        ),    def _setup_model_aliases(self):

                        finish_reason="stop"        """Setup model aliases for easier access"""

                    )],        self.model_aliases = {

                    usage=Usage(            'chat': 'mistral',

                        prompt_tokens=data.get('prompt_eval_count', 0),            'code': 'codellama',

                        completion_tokens=data.get('eval_count', 0),            'small': 'tinyllama',

                        total_tokens=data.get('prompt_eval_count', 0) + data.get('eval_count', 0)            'large': 'mixtral',

                    )            'default': 'mistral'

                )        }

                        

        except Exception as e:        # Filter aliases to only include available models

            self._update_metrics(model_name, 0, False)        available_aliases = {}

            raise        for alias, model in self.model_aliases.items():

        finally:            if model in self.available_models:

            server.current_load = max(0, server.current_load - 1)                available_aliases[alias] = model

        

    async def stream(self, request: CompletionRequest) -> AsyncGenerator[StreamResponse, None]:        self.model_aliases = available_aliases

        """Stream completion"""

        model_name = self._resolve_model_name(request.model or 'default')    async def _health_check_loop(self):

                """Background health checking for models and servers"""

        if model_name not in self.available_models:        while True:

            if self.available_models:            try:

                model_name = list(self.available_models.keys())[0]                await asyncio.sleep(self.health_check_interval)

            else:                await self._check_all_health()

                raise ValueError("No models available")            except Exception as e:

                        logger.error(f"Health check loop error: {e}")

        server = self._select_best_server()

        if not server:    async def _check_all_health(self):

            raise ConnectionError("No healthy Ollama servers available")        """Check health of all servers and models"""

                for server in self.servers:

        try:            await self._check_server_health(server)

            server.current_load += 1        

            start_time = time.time()        for model_name in list(self.available_models.keys()):

                        await self._check_model_health(model_name)

            async with AsyncClient(timeout=server.timeout) as client:

                payload = {    async def _check_server_health(self, server: OllamaServer):

                    "model": model_name,        """Check health of a specific server"""

                    "messages": self._format_messages(request.messages),        try:

                    "stream": True,            start_time = time.time()

                    "options": {}            async with AsyncClient(timeout=5.0) as client:

                }                response = await client.get(f"{server.base_url}/api/version")

                                response.raise_for_status()

                if request.temperature is not None:            

                    payload["options"]["temperature"] = request.temperature            response_time = time.time() - start_time

                if request.max_tokens is not None:            server.is_healthy = True

                    payload["options"]["num_predict"] = request.max_tokens            server.last_health_check = datetime.utcnow()

                            

                async with client.stream('POST', f"{server.base_url}/api/chat", json=payload) as response:            logger.debug(f"Server {server.base_url} is healthy (response time: {response_time:.2f}s)")

                    response.raise_for_status()            

                            except Exception as e:

                    async for line in response.aiter_lines():            server.is_healthy = False

                        if line.strip():            logger.warning(f"Server {server.base_url} health check failed: {e}")

                            try:

                                data = json.loads(line)    async def _check_model_health(self, model_name: str):

                                        """Check health of a specific model"""

                                if 'message' in data:        if model_name not in self.model_health:

                                    content = data['message'].get('content', '')            return

                                    if content:        

                                        yield StreamResponse(        health = self.model_health[model_name]

                                            id=str(uuid.uuid4()),        

                                            model=model_name,        try:

                                            choices=[StreamChoice(            # Simple health check - generate a short response

                                                index=0,            start_time = time.time()

                                                delta=Message(            server = self._select_best_server()

                                                    role=MessageRole.ASSISTANT,            

                                                    content=content            if not server or not server.is_healthy:

                                                ),                health.is_healthy = False

                                                finish_reason=None                health.consecutive_failures += 1

                                            )]                return

                                        )            

                                            async with AsyncClient(timeout=10.0) as client:

                                if data.get('done', False):                payload = {

                                    response_time = time.time() - start_time                    "model": model_name,

                                    self._update_metrics(model_name, response_time, True)                    "prompt": "Test",

                                                        "stream": False,

                                    yield StreamResponse(                    "options": {"num_predict": 1}

                                        id=str(uuid.uuid4()),                }

                                        model=model_name,                

                                        choices=[StreamChoice(                response = await client.post(f"{server.base_url}/api/generate", json=payload)

                                            index=0,                response.raise_for_status()

                                            delta=Message(role=MessageRole.ASSISTANT, content=""),            

                                            finish_reason="stop"            response_time = time.time() - start_time

                                        )]            

                                    )            # Update health metrics

                                    break            health.is_healthy = True

                                                health.consecutive_failures = 0

                            except json.JSONDecodeError:            health.last_check = datetime.utcnow()

                                continue            health.response_times.append(response_time)

                                            health.success_count += 1

        except Exception as e:            

            self._update_metrics(model_name, 0, False)            # Keep only last 100 response times

            raise            if len(health.response_times) > 100:

        finally:                health.response_times = health.response_times[-100:]

            server.current_load = max(0, server.current_load - 1)            

            logger.debug(f"Model {model_name} health check passed (response time: {response_time:.2f}s)")

    def _update_metrics(self, model_name: str, response_time: float, success: bool):            

        """Update performance metrics"""        except Exception as e:

        if model_name not in self.request_counts:            health.is_healthy = False

            self.request_counts[model_name] = 0            health.consecutive_failures += 1

            self.response_times[model_name] = []            health.error_count += 1

                    

        self.request_counts[model_name] += 1            logger.warning(f"Model {model_name} health check failed: {e}")

                    

        if success and response_time > 0:            # Circuit breaker logic

            self.response_times[model_name].append(response_time)            if health.consecutive_failures >= self.circuit_breaker_threshold:

            if len(self.response_times[model_name]) > 1000:                logger.error(f"Model {model_name} circuit breaker opened due to {health.consecutive_failures} consecutive failures")

                self.response_times[model_name] = self.response_times[model_name][-1000:]

            async def _model_discovery_loop(self):

        if model_name in self.model_health:        """Background model discovery"""

            health = self.model_health[model_name]        while True:

            if success:            try:

                health.success_count += 1                await asyncio.sleep(300)  # Check every 5 minutes

                health.consecutive_failures = 0                await self._discover_models()

                if response_time > 0:            except Exception as e:

                    health.response_times.append(response_time)                logger.error(f"Model discovery loop error: {e}")

            else:

                health.error_count += 1    def _select_best_server(self) -> Optional[OllamaServer]:

                health.consecutive_failures += 1        """Select the best available server using load balancing"""

        healthy_servers = [s for s in self.servers if s.is_healthy]

    def get_available_models(self) -> List[ModelInfo]:        

        """Get list of available models"""        if not healthy_servers:

        return list(self.available_models.values())            logger.error("No healthy Ollama servers available")

            return None

    def get_model_stats(self) -> Dict[str, Any]:        

        """Get model statistics"""        # Weighted round-robin with load consideration

        stats = {        best_server = None

            "servers": [        best_score = float('inf')

                {        

                    "url": server.base_url,        for server in healthy_servers:

                    "healthy": server.is_healthy,            # Score based on load factor and weight

                    "load": server.current_load,            score = server.load_factor / server.weight

                    "max_concurrent": server.max_concurrent,            if score < best_score:

                    "weight": server.weight                best_score = score

                }                best_server = server

                for server in self.servers        

            ],        return best_server

            "models": {},

            "aliases": self.model_aliases,    def _resolve_model_name(self, model: str) -> str:

            "total_requests": sum(self.request_counts.values()),        """Resolve model name from alias"""

            "healthy_models": sum(1 for h in self.model_health.values() if h.is_healthy),        if model in self.model_aliases:

            "total_models": len(self.available_models)            return self.model_aliases[model]

        }        return model

        

        for model_name, health in self.model_health.items():    async def complete(self, request: CompletionRequest) -> CompletionResponse:

            stats["models"][model_name] = {        """Complete a request with enhanced error handling and load balancing"""

                "healthy": health.is_healthy,        model_name = self._resolve_model_name(request.model or 'default')

                "last_check": health.last_check.isoformat(),        

                "consecutive_failures": health.consecutive_failures,        # Check if model is available and healthy

                "average_response_time": health.average_response_time,        if model_name not in self.available_models:

                "success_rate": health.success_rate,            raise ValueError(f"Model {model_name} not available")

                "request_count": self.request_counts.get(model_name, 0),        

                "error_count": health.error_count,        model_health = self.model_health.get(model_name)

                "success_count": health.success_count        if model_health and not model_health.is_healthy:

            }            raise ValueError(f"Model {model_name} is currently unhealthy")

                

        return stats        for attempt in range(self.max_retries):

            server = self._select_best_server()

    async def shutdown(self):            if not server:

        """Graceful shutdown"""                raise ConnectionError("No healthy Ollama servers available")

        logger.info("Shutting down Ollama Provider...")            

                    try:

        if self._health_check_task:                server.current_load += 1

            self._health_check_task.cancel()                start_time = time.time()

        if self._model_discovery_task:                

            self._model_discovery_task.cancel()                async with AsyncClient(timeout=server.timeout) as client:

                            payload = {

        try:                        "model": model_name,

            if self._health_check_task:                        "messages": self._format_messages(request.messages),

                await self._health_check_task                        "stream": False,

        except asyncio.CancelledError:                        "options": {

            pass                            "temperature": request.temperature,

                                    "num_predict": request.max_tokens,

        try:                        }

            if self._model_discovery_task:                    }

                await self._model_discovery_task                    

        except asyncio.CancelledError:                    # Remove None values

            pass                    payload["options"] = {k: v for k, v in payload["options"].items() if v is not None}

                    

    def _format_messages(self, messages: List[Message]) -> List[Dict[str, str]]:                    response = await client.post(f"{server.base_url}/api/chat", json=payload)

        """Format messages for Ollama API"""                    response.raise_for_status()

        return [                    data = response.json()

            {                    

                "role": msg.role.value,                    response_time = time.time() - start_time

                "content": msg.content                    

            }                    # Update metrics

            for msg in messages                    self._update_metrics(model_name, response_time, True)

        ]                    
                    # Extract response
                    message_content = data.get('message', {}).get('content', '')
                    
                    return CompletionResponse(
                        id=str(uuid.uuid4()),
                        model=model_name,
                        choices=[Choice(
                            index=0,
                            message=Message(
                                role=MessageRole.ASSISTANT,
                                content=message_content
                            ),
                            finish_reason="stop"
                        )],
                        usage=Usage(
                            prompt_tokens=data.get('prompt_eval_count', 0),
                            completion_tokens=data.get('eval_count', 0),
                            total_tokens=data.get('prompt_eval_count', 0) + data.get('eval_count', 0)
                        )
                    )
                    
            except (ConnectTimeout, ReadTimeout) as e:
                logger.warning(f"Timeout on server {server.base_url}, attempt {attempt + 1}/{self.max_retries}: {e}")
                self._update_metrics(model_name, 0, False)
                
                if attempt == self.max_retries - 1:
                    raise ConnectionError(f"All retry attempts failed for model {model_name}")
                
            except Exception as e:
                logger.error(f"Error on server {server.base_url}, attempt {attempt + 1}/{self.max_retries}: {e}")
                self._update_metrics(model_name, 0, False)
                
                if attempt == self.max_retries - 1:
                    raise
                    
            finally:
                server.current_load = max(0, server.current_load - 1)

    async def stream(self, request: CompletionRequest) -> AsyncGenerator[StreamResponse, None]:
        """Stream completion with enhanced features"""
        model_name = self._resolve_model_name(request.model or 'default')
        
        # Check model availability
        if model_name not in self.available_models:
            raise ValueError(f"Model {model_name} not available")
        
        server = self._select_best_server()
        if not server:
            raise ConnectionError("No healthy Ollama servers available")
        
        try:
            server.current_load += 1
            start_time = time.time()
            
            async with AsyncClient(timeout=server.timeout) as client:
                payload = {
                    "model": model_name,
                    "messages": self._format_messages(request.messages),
                    "stream": True,
                    "options": {
                        "temperature": request.temperature,
                        "num_predict": request.max_tokens,
                    }
                }
                
                payload["options"] = {k: v for k, v in payload["options"].items() if v is not None}
                
                async with client.stream('POST', f"{server.base_url}/api/chat", json=payload) as response:
                    response.raise_for_status()
                    
                    async for line in response.aiter_lines():
                        if line.strip():
                            try:
                                data = json.loads(line)
                                
                                if 'message' in data:
                                    content = data['message'].get('content', '')
                                    if content:
                                        yield StreamResponse(
                                            id=str(uuid.uuid4()),
                                            model=model_name,
                                            choices=[StreamChoice(
                                                index=0,
                                                delta=Message(
                                                    role=MessageRole.ASSISTANT,
                                                    content=content
                                                ),
                                                finish_reason=None
                                            )]
                                        )
                                
                                if data.get('done', False):
                                    response_time = time.time() - start_time
                                    self._update_metrics(model_name, response_time, True)
                                    
                                    yield StreamResponse(
                                        id=str(uuid.uuid4()),
                                        model=model_name,
                                        choices=[StreamChoice(
                                            index=0,
                                            delta=Message(role=MessageRole.ASSISTANT, content=""),
                                            finish_reason="stop"
                                        )]
                                    )
                                    break
                                    
                            except json.JSONDecodeError:
                                continue
                                
        except Exception as e:
            self._update_metrics(model_name, 0, False)
            raise
        finally:
            server.current_load = max(0, server.current_load - 1)

    def _update_metrics(self, model_name: str, response_time: float, success: bool):
        """Update performance metrics"""
        if model_name not in self.request_counts:
            self.request_counts[model_name] = 0
            self.response_times[model_name] = []
        
        self.request_counts[model_name] += 1
        
        if success and response_time > 0:
            self.response_times[model_name].append(response_time)
            # Keep only last 1000 response times
            if len(self.response_times[model_name]) > 1000:
                self.response_times[model_name] = self.response_times[model_name][-1000:]
        
        # Update model health
        if model_name in self.model_health:
            health = self.model_health[model_name]
            if success:
                health.success_count += 1
                health.consecutive_failures = 0
                if response_time > 0:
                    health.response_times.append(response_time)
            else:
                health.error_count += 1
                health.consecutive_failures += 1

    def get_available_models(self) -> List[ModelInfo]:
        """Get list of available models with health status"""
        models = []
        for model_name, model_info in self.available_models.items():
            # Update model info with current health status
            health = self.model_health.get(model_name)
            if health:
                # Add health information to model description
                health_status = "healthy" if health.is_healthy else "unhealthy"
                avg_response_time = health.average_response_time
                success_rate = health.success_rate
                
                enhanced_description = f"{model_info.description} (Status: {health_status}, Avg Response: {avg_response_time:.2f}s, Success Rate: {success_rate:.2%})"
                
                enhanced_model = ModelInfo(
                    id=model_info.id,
                    provider=model_info.provider,
                    name=model_info.name,
                    description=enhanced_description,
                    max_tokens=model_info.max_tokens,
                    context_length=model_info.context_length,
                    cost_per_token=model_info.cost_per_token,
                    supports_streaming=model_info.supports_streaming,
                    supports_tools=model_info.supports_tools,
                    supports_vision=model_info.supports_vision,
                    capability_score=model_info.capability_score * (success_rate if health.is_healthy else 0.1)
                )
                models.append(enhanced_model)
            else:
                models.append(model_info)
        
        return models

    def get_model_stats(self) -> Dict[str, Any]:
        """Get comprehensive model statistics"""
        stats = {
            "servers": [
                {
                    "url": server.base_url,
                    "healthy": server.is_healthy,
                    "load": server.current_load,
                    "max_concurrent": server.max_concurrent,
                    "load_factor": server.load_factor,
                    "weight": server.weight
                }
                for server in self.servers
            ],
            "models": {},
            "aliases": self.model_aliases,
            "total_requests": sum(self.request_counts.values()),
            "healthy_models": sum(1 for h in self.model_health.values() if h.is_healthy),
            "total_models": len(self.available_models)
        }
        
        for model_name, health in self.model_health.items():
            stats["models"][model_name] = {
                "healthy": health.is_healthy,
                "last_check": health.last_check.isoformat(),
                "consecutive_failures": health.consecutive_failures,
                "average_response_time": health.average_response_time,
                "success_rate": health.success_rate,
                "request_count": self.request_counts.get(model_name, 0),
                "error_count": health.error_count,
                "success_count": health.success_count
            }
        
        return stats

    async def shutdown(self):
        """Graceful shutdown"""
        logger.info("Shutting down Enhanced Ollama Provider...")
        
        # Cancel background tasks
        if self._health_check_task:
            self._health_check_task.cancel()
        if self._model_discovery_task:
            self._model_discovery_task.cancel()
        
        # Wait for tasks to complete
        try:
            if self._health_check_task:
                await self._health_check_task
        except asyncio.CancelledError:
            pass
        
        try:
            if self._model_discovery_task:
                await self._model_discovery_task
        except asyncio.CancelledError:
            pass
        
        logger.info("Enhanced Ollama Provider shutdown complete")

    def _format_messages(self, messages: List[Message]) -> List[Dict[str, str]]:
        """Format messages for Ollama API"""
        return [
            {
                "role": msg.role.value,
                "content": msg.content
            }
            for msg in messages
        ]


# Backward compatibility alias
OllamaProvider = EnhancedOllamaProvider

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
            "model": request.model or (getattr(self.config, "models", ["mistral"])[0] if getattr(self.config, "models", None) else "mistral"),
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
