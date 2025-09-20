"""
Logging utilities for Multi-Agent LLM Service
"""

import logging
import logging.config
import sys
import json
from typing import Dict, Any
from datetime import datetime

class JSONFormatter(logging.Formatter):
    """JSON formatter for structured logging"""
    
    def format(self, record: logging.LogRecord) -> str:
        """Format log record as JSON"""
        
        log_entry = {
            "timestamp": datetime.utcnow().isoformat() + "Z",
            "level": record.levelname,
            "logger": record.name,
            "message": record.getMessage(),
            "module": record.module,
            "function": record.funcName,
            "line": record.lineno
        }
        
        # Add exception info if present
        if record.exc_info:
            log_entry["exception"] = self.formatException(record.exc_info)
        
        # Add extra fields
        if hasattr(record, 'user_id'):
            log_entry["user_id"] = record.user_id
        
        if hasattr(record, 'tenant_id'):
            log_entry["tenant_id"] = record.tenant_id
        
        if hasattr(record, 'request_id'):
            log_entry["request_id"] = record.request_id
        
        if hasattr(record, 'execution_id'):
            log_entry["execution_id"] = record.execution_id
        
        return json.dumps(log_entry)

class ColoredFormatter(logging.Formatter):
    """Colored formatter for console output"""
    
    COLORS = {
        'DEBUG': '\033[36m',    # Cyan
        'INFO': '\033[32m',     # Green
        'WARNING': '\033[33m',  # Yellow
        'ERROR': '\033[31m',    # Red
        'CRITICAL': '\033[35m', # Magenta
        'RESET': '\033[0m'      # Reset
    }
    
    def format(self, record: logging.LogRecord) -> str:
        """Format log record with colors"""
        
        color = self.COLORS.get(record.levelname, self.COLORS['RESET'])
        reset = self.COLORS['RESET']
        
        # Format timestamp
        timestamp = datetime.fromtimestamp(record.created).strftime('%Y-%m-%d %H:%M:%S')
        
        # Format message
        message = f"{color}[{timestamp}] {record.levelname:8} {record.name}: {record.getMessage()}{reset}"
        
        # Add exception info if present
        if record.exc_info:
            message += f"\n{self.formatException(record.exc_info)}"
        
        return message

def setup_logging(
    level: str = "INFO",
    format_type: str = "json",
    log_file: str = None
) -> None:
    """Setup logging configuration"""
    
    # Convert level string to logging constant
    numeric_level = getattr(logging, level.upper(), logging.INFO)
    
    # Create formatters
    if format_type.lower() == "json":
        formatter = JSONFormatter()
    else:
        formatter = ColoredFormatter()
    
    # Setup handlers
    handlers = []
    
    # Console handler
    console_handler = logging.StreamHandler(sys.stdout)
    console_handler.setFormatter(formatter)
    console_handler.setLevel(numeric_level)
    handlers.append(console_handler)
    
    # File handler if specified
    if log_file:
        file_handler = logging.FileHandler(log_file)
        file_handler.setFormatter(JSONFormatter())  # Always use JSON for files
        file_handler.setLevel(numeric_level)
        handlers.append(file_handler)
    
    # Configure root logger
    logging.basicConfig(
        level=numeric_level,
        handlers=handlers,
        force=True
    )
    
    # Set specific logger levels
    logging.getLogger("uvicorn").setLevel(logging.WARNING)
    logging.getLogger("uvicorn.access").setLevel(logging.WARNING)
    logging.getLogger("fastapi").setLevel(logging.WARNING)
    logging.getLogger("httpx").setLevel(logging.WARNING)
    logging.getLogger("openai").setLevel(logging.WARNING)
    logging.getLogger("anthropic").setLevel(logging.WARNING)
    
    # Create application logger
    logger = logging.getLogger("llm_service")
    logger.info("Logging configured", extra={
        "level": level,
        "format": format_type,
        "log_file": log_file
    })

class ContextLogger:
    """Logger with context information"""
    
    def __init__(self, name: str, context: Dict[str, Any] = None):
        self.logger = logging.getLogger(name)
        self.context = context or {}
    
    def _log(self, level: int, message: str, *args, **kwargs):
        """Log with context"""
        extra = kwargs.get('extra', {})
        extra.update(self.context)
        kwargs['extra'] = extra
        
        self.logger.log(level, message, *args, **kwargs)
    
    def debug(self, message: str, *args, **kwargs):
        self._log(logging.DEBUG, message, *args, **kwargs)
    
    def info(self, message: str, *args, **kwargs):
        self._log(logging.INFO, message, *args, **kwargs)
    
    def warning(self, message: str, *args, **kwargs):
        self._log(logging.WARNING, message, *args, **kwargs)
    
    def error(self, message: str, *args, **kwargs):
        self._log(logging.ERROR, message, *args, **kwargs)
    
    def critical(self, message: str, *args, **kwargs):
        self._log(logging.CRITICAL, message, *args, **kwargs)
    
    def exception(self, message: str, *args, **kwargs):
        kwargs['exc_info'] = True
        self._log(logging.ERROR, message, *args, **kwargs)
    
    def with_context(self, **context) -> 'ContextLogger':
        """Create new logger with additional context"""
        new_context = self.context.copy()
        new_context.update(context)
        return ContextLogger(self.logger.name, new_context)

def get_logger(name: str, **context) -> ContextLogger:
    """Get context logger"""
    return ContextLogger(name, context)

# Request logging utilities
def log_request_start(logger: logging.Logger, request_id: str, method: str, path: str, user_id: str = None):
    """Log request start"""
    logger.info(
        f"Request started: {method} {path}",
        extra={
            "request_id": request_id,
            "method": method,
            "path": path,
            "user_id": user_id,
            "event": "request_start"
        }
    )

def log_request_end(
    logger: logging.Logger, 
    request_id: str, 
    status_code: int, 
    duration_ms: int,
    user_id: str = None
):
    """Log request end"""
    logger.info(
        f"Request completed: {status_code} ({duration_ms}ms)",
        extra={
            "request_id": request_id,
            "status_code": status_code,
            "duration_ms": duration_ms,
            "user_id": user_id,
            "event": "request_end"
        }
    )

def log_model_request(
    logger: logging.Logger,
    execution_id: str,
    model: str,
    provider: str,
    tokens: int,
    cost: float,
    user_id: str = None
):
    """Log model request"""
    logger.info(
        f"Model request: {provider}:{model} ({tokens} tokens, ${cost:.4f})",
        extra={
            "execution_id": execution_id,
            "model": model,
            "provider": provider,
            "tokens": tokens,
            "cost_usd": cost,
            "user_id": user_id,
            "event": "model_request"
        }
    )

def log_error(
    logger: logging.Logger,
    error: Exception,
    context: Dict[str, Any] = None,
    user_id: str = None
):
    """Log error with context"""
    extra = {
        "error_type": type(error).__name__,
        "error_message": str(error),
        "user_id": user_id,
        "event": "error"
    }
    
    if context:
        extra.update(context)
    
    logger.error(
        f"Error occurred: {type(error).__name__}: {error}",
        extra=extra,
        exc_info=True
    )