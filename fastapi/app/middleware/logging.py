"""
Structured JSON logging middleware for FastAPI.
Provides request/response logging with correlation IDs and performance metrics.
"""

import json
import time
import uuid
from contextvars import ContextVar
from typing import Dict, Any
import logging
from datetime import datetime
from pathlib import Path
from fastapi import Request, Response
from starlette.middleware.base import BaseHTTPMiddleware

# Request ID context variable
request_id: ContextVar[str] = ContextVar('request_id')

# Configure structured logger
logger = logging.getLogger("fastapi.structured")
logger.setLevel(logging.INFO)

# Create file handler for structured logs
# Ensure log directory exists before creating FileHandler
log_file_path = "/app/logs/fastapi_structured.log"
log_path = Path(log_file_path)
log_path.parent.mkdir(parents=True, exist_ok=True)

file_handler = logging.FileHandler(log_file_path)
file_handler.setLevel(logging.INFO)

# JSON formatter
class JSONFormatter(logging.Formatter):
    def format(self, record):
        # Use datetime for cross-platform timestamp formatting with microseconds
        timestamp = datetime.utcfromtimestamp(record.created).strftime("%Y-%m-%dT%H:%M:%S.%fZ")
        log_entry = {
            "timestamp": timestamp,
            "level": record.levelname,
            "logger": record.name,
            "message": record.getMessage(),
            "module": record.module,
            "function": record.funcName,
            "line": record.lineno,
        }
        
        # Add request context if available
        if hasattr(record, 'request_id'):
            log_entry["request_id"] = record.request_id
        if hasattr(record, 'method'):
            log_entry["method"] = record.method
        if hasattr(record, 'path'):
            log_entry["path"] = record.path
        if hasattr(record, 'status_code'):
            log_entry["status_code"] = record.status_code
        if hasattr(record, 'duration'):
            log_entry["duration"] = record.duration
        if hasattr(record, 'client_ip'):
            log_entry["client_ip"] = record.client_ip
        if hasattr(record, 'user_agent'):
            log_entry["user_agent"] = record.user_agent
        if hasattr(record, 'user_id'):
            log_entry["user_id"] = record.user_id
            
        # Add exception info if present
        if record.exc_info:
            log_entry["exception"] = self.formatException(record.exc_info)
            
        return json.dumps(log_entry, ensure_ascii=False)

file_handler.setFormatter(JSONFormatter())
logger.addHandler(file_handler)

class StructuredLoggingMiddleware(BaseHTTPMiddleware):
    """Middleware for structured JSON logging with request tracing."""
    
    async def dispatch(self, request: Request, call_next):
        # Generate request ID
        req_id = str(uuid.uuid4())
        request_id.set(req_id)
        
        # Extract request information
        start_time = time.time()
        client_ip = request.client.host if request.client else "unknown"
        user_agent = request.headers.get("user-agent", "")
        
        # Extract user ID from JWT if available (for authenticated requests)
        user_id = None
        auth_header = request.headers.get("authorization", "")
        if auth_header.startswith("Bearer "):
            # In a real implementation, you would decode the JWT here
            # For now, we'll just mark it as authenticated
            user_id = "authenticated"
        
        # Log request start
        logger.info(
            "Request started",
            extra={
                "request_id": req_id,
                "method": request.method,
                "path": str(request.url.path),
                "query_params": str(request.query_params),
                "client_ip": client_ip,
                "user_agent": user_agent,
                "user_id": user_id,
                "headers": dict(request.headers),
            }
        )
        
        # Process request
        try:
            response = await call_next(request)
            
            # Calculate duration
            duration = time.time() - start_time
            
            # Log successful response
            logger.info(
                "Request completed",
                extra={
                    "request_id": req_id,
                    "method": request.method,
                    "path": str(request.url.path),
                    "status_code": response.status_code,
                    "duration": round(duration, 4),
                    "client_ip": client_ip,
                    "user_agent": user_agent,
                    "user_id": user_id,
                    "response_size": response.headers.get("content-length", 0),
                }
            )
            
            # Add request ID to response headers
            response.headers["X-Request-ID"] = req_id
            response.headers["X-Response-Time"] = f"{duration:.4f}s"
            
            return response
            
        except Exception as e:
            # Calculate duration for error case
            duration = time.time() - start_time
            
            # Log error
            logger.error(
                f"Request failed: {str(e)}",
                extra={
                    "request_id": req_id,
                    "method": request.method,
                    "path": str(request.url.path),
                    "status_code": 500,
                    "duration": round(duration, 4),
                    "client_ip": client_ip,
                    "user_agent": user_agent,
                    "user_id": user_id,
                    "error": str(e),
                    "error_type": type(e).__name__,
                },
                exc_info=True
            )
            
            # Re-raise the exception
            raise

def get_request_id() -> str:
    """Get the current request ID from context."""
    return request_id.get("")

def log_structured(level: str, message: str, **kwargs):
    """Log a structured message with current request context."""
    extra = {
        "request_id": request_id.get(""),
        **kwargs
    }
    
    if level.upper() == "DEBUG":
        logger.debug(message, extra=extra)
    elif level.upper() == "INFO":
        logger.info(message, extra=extra)
    elif level.upper() == "WARNING":
        logger.warning(message, extra=extra)
    elif level.upper() == "ERROR":
        logger.error(message, extra=extra)
    elif level.upper() == "CRITICAL":
        logger.critical(message, extra=extra)
    else:
        logger.info(message, extra=extra)

def log_performance(operation: str, duration: float, **kwargs):
    """Log performance metrics."""
    log_structured(
        "INFO",
        f"Performance: {operation}",
        operation=operation,
        duration=duration,
        **kwargs
    )

def log_business_event(event: str, **kwargs):
    """Log business events."""
    log_structured(
        "INFO",
        f"Business Event: {event}",
        event=event,
        **kwargs
    )
