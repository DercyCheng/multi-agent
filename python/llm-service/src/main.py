"""
Multi-Agent Platform LLM Service
Main application entry point
"""

import asyncio
import logging
import signal
import sys
from contextlib import asynccontextmanager
from typing import AsyncGenerator

import uvicorn
from fastapi import FastAPI, Request
from fastapi.middleware.cors import CORSMiddleware
from fastapi.middleware.gzip import GZipMiddleware
from fastapi.responses import JSONResponse
from opentelemetry import trace
from opentelemetry.exporter.jaeger.thrift import JaegerExporter
from opentelemetry.instrumentation.fastapi import FastAPIInstrumentor
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor

from src.config import Settings
from src.api.routes import completion, health, metrics, models
from src.api.routes import flags as flags_router
from src.feature_flags.manager import FeatureFlagManager
from src.core.model_router import ModelRouter
from src.core.context_engine import ContextEngine
from src.core.token_manager import TokenManager
from src.integrations.mcp_client import MCPClient
from src.middleware.auth import AuthMiddleware
from src.middleware.rate_limit import RateLimitMiddleware
from src.middleware.metrics import MetricsMiddleware
from src.utils.logger import setup_logging
from datetime import datetime

# Global instances
model_router: ModelRouter = None
context_engine: ContextEngine = None
token_manager: TokenManager = None
mcp_client: MCPClient = None

@asynccontextmanager
async def lifespan(app: FastAPI) -> AsyncGenerator[None, None]:
    """Application lifespan manager"""
    global model_router, context_engine, token_manager, mcp_client
    
    # Startup
    logging.info("Starting Multi-Agent LLM Service")
    try:
        # Initialize core components
        settings = Settings()

        # Setup distributed tracing
        if settings.tracing.enabled:
            setup_tracing(settings)

        # Initialize model router
        model_router = ModelRouter(settings.models)
        await model_router.initialize()

        # Initialize context engine
        context_engine = ContextEngine(settings.context)
        await context_engine.initialize()

        # Initialize token manager
        token_manager = TokenManager(settings.tokens)
        await token_manager.initialize()

        # Initialize MCP client
        mcp_client = MCPClient(settings.mcp)
        await mcp_client.initialize()

        # Store in app state
        app.state.model_router = model_router
        app.state.context_engine = context_engine
        app.state.token_manager = token_manager
        app.state.mcp_client = mcp_client
        app.state.settings = settings

        logging.info("LLM Service initialized successfully")

        feature_flags = FeatureFlagManager()
        # Example: seed flags from env/settings if desired
        await feature_flags.set_flag("ollama.enabled", settings.models.ollama.enabled)
        # Seed cron flag (default disabled)
        await feature_flags.set_flag("cron.enabled", False)

        # Store feature flags in app state
        app.state.feature_flags = feature_flags

        # Start background cron worker
        app.state._cron_task = asyncio.create_task(_cron_worker(app))

        yield

    except Exception as e:
        logging.error(f"Failed to initialize LLM Service: {e}")
        raise
    finally:
        # Shutdown
        logging.info("Shutting down Multi-Agent LLM Service")

        try:
            # Cancel background task if running
            cron_task = getattr(app.state, "_cron_task", None)
            if cron_task:
                cron_task.cancel()

            if mcp_client:
                await mcp_client.shutdown()
            if token_manager:
                await token_manager.shutdown()
            if context_engine:
                await context_engine.shutdown()
            if model_router:
                await model_router.shutdown()

            logging.info("LLM Service shutdown completed")

        except Exception as e:
            logging.error(f"Error during shutdown: {e}")

def create_app() -> FastAPI:
    """Create and configure FastAPI application"""
    
    app = FastAPI(
        title="Multi-Agent LLM Service",
        description="Enterprise-grade LLM service with multi-provider support and MCP integration",
        version="1.0.0",
        docs_url="/docs",
        redoc_url="/redoc",
        lifespan=lifespan
    )
    
    # Add middleware
    app.add_middleware(
        CORSMiddleware,
        allow_origins=["*"],  # Configure appropriately for production
        allow_credentials=True,
        allow_methods=["*"],
        allow_headers=["*"],
    )
    
    app.add_middleware(GZipMiddleware, minimum_size=1000)
    app.add_middleware(MetricsMiddleware)
    app.add_middleware(RateLimitMiddleware)
    app.add_middleware(AuthMiddleware)
    
    # Add routes
    app.include_router(health.router, prefix="/health", tags=["health"])
    app.include_router(metrics.router, prefix="/metrics", tags=["metrics"])
    app.include_router(models.router, prefix="/models", tags=["models"])
    app.include_router(completion.router, prefix="/v1", tags=["completion"])
    app.include_router(flags_router.router, prefix="/flags", tags=["flags"])
    
    # Global exception handler
    @app.exception_handler(Exception)
    async def global_exception_handler(request: Request, exc: Exception):
        logging.error(f"Global exception: {exc}", exc_info=True)
        return JSONResponse(
            status_code=500,
            content={"error": "Internal server error", "detail": str(exc)}
        )
    
    return app


async def _cron_worker(app: FastAPI):
    """Simple background worker that runs when feature flag 'cron.enabled' is true."""
    logger = logging.getLogger("cron")
    while True:
        try:
            feature_flags: FeatureFlagManager = getattr(app.state, "feature_flags", None)
            if feature_flags:
                enabled = await feature_flags.get_flag("cron.enabled")
                if enabled:
                    # Run a sample cron job — in real usage this would trigger scheduled tasks
                    logger.info(f"[cron] Running scheduled task at {datetime.utcnow().isoformat()} UTC")
                    # Example: call a health endpoint or cleanup routine; here we just sleep briefly
                    await asyncio.sleep(0.1)
            await asyncio.sleep(5)
        except Exception:
            logger.exception("Error in cron worker")


def setup_tracing(settings: Settings):
    """Setup distributed tracing"""
    if not settings.tracing.enabled:
        return
    
    # Configure tracer provider
    trace.set_tracer_provider(TracerProvider())
    tracer = trace.get_tracer(__name__)
    
    # Configure Jaeger exporter
    jaeger_exporter = JaegerExporter(
        agent_host_name=settings.tracing.jaeger_host,
        agent_port=settings.tracing.jaeger_port,
    )
    
    # Add span processor
    span_processor = BatchSpanProcessor(jaeger_exporter)
    trace.get_tracer_provider().add_span_processor(span_processor)
    
    logging.info("Distributed tracing configured")

def setup_signal_handlers():
    """Setup signal handlers for graceful shutdown"""
    def signal_handler(signum, frame):
        logging.info(f"Received signal {signum}, shutting down...")
        sys.exit(0)
    
    signal.signal(signal.SIGINT, signal_handler)
    signal.signal(signal.SIGTERM, signal_handler)

async def main():
    """Main application entry point"""
    # Setup logging
    setup_logging()
    
    # Setup signal handlers
    setup_signal_handlers()
    
    # Load settings
    settings = Settings()
    
    # Create app
    app = create_app()
    
    # Instrument with OpenTelemetry
    if settings.tracing.enabled:
        FastAPIInstrumentor.instrument_app(app)
    
    # Configure uvicorn
    config = uvicorn.Config(
        app,
        host=settings.server.host,
        port=settings.server.port,
        log_level=settings.logging.level.lower(),
        access_log=settings.logging.access_log,
        reload=settings.server.reload,
        workers=settings.server.workers if not settings.server.reload else 1,
    )
    
    # Start server
    server = uvicorn.Server(config)
    
    logging.info(f"Starting LLM Service on {settings.server.host}:{settings.server.port}")
    
    try:
        await server.serve()
    except KeyboardInterrupt:
        logging.info("Received keyboard interrupt, shutting down...")
    except Exception as e:
        logging.error(f"Server error: {e}")
        raise

if __name__ == "__main__":
    asyncio.run(main())