"""
Centralized logger configuration for FastAPI application.
Provides JSON formatter and structured logging capabilities.
"""

import logging
import sys
import json
import time
from datetime import datetime, UTC
from typing import Dict, Any, Optional
from pathlib import Path

class JSONFormatter(logging.Formatter):
    """Custom JSON formatter for structured logging."""
    
    def format(self, record: logging.LogRecord) -> str:
        """Format log record as JSON."""
        # Use datetime for microseconds support (Windows compatible)
        # Use UTC timezone-aware datetime to avoid deprecation warning
        timestamp = datetime.fromtimestamp(record.created, UTC).strftime("%Y-%m-%dT%H:%M:%S.%fZ")
        log_entry: Dict[str, Any] = {
            "timestamp": timestamp,
            "level": record.levelname,
            "logger": record.name,
            "message": record.getMessage(),
            "module": record.module,
            "function": record.funcName,
            "line": record.lineno,
        }
        
        # Add extra fields from the record
        for key, value in record.__dict__.items():
            if key not in ['name', 'msg', 'args', 'levelname', 'levelno', 'pathname', 
                          'filename', 'module', 'exc_info', 'exc_text', 'stack_info',
                          'lineno', 'funcName', 'created', 'msecs', 'relativeCreated',
                          'thread', 'threadName', 'processName', 'process', 'getMessage']:
                log_entry[key] = value
        
        # Add exception info if present
        if record.exc_info:
            log_entry["exception"] = self.formatException(record.exc_info)
            
        return json.dumps(log_entry, ensure_ascii=False)

def setup_logger(
    name: str = "fastapi",
    level: str = "INFO",
    log_file: Optional[str] = None,
    console_output: bool = True
) -> logging.Logger:
    """
    Set up a structured logger with JSON formatting.
    
    Args:
        name: Logger name
        level: Logging level (DEBUG, INFO, WARNING, ERROR, CRITICAL)
        log_file: Path to log file (optional)
        console_output: Whether to output to console
        
    Returns:
        Configured logger instance
    """
    logger = logging.getLogger(name)
    logger.setLevel(getattr(logging, level.upper()))
    
    # Clear existing handlers
    logger.handlers.clear()
    
    # Console handler
    if console_output:
        console_handler = logging.StreamHandler(sys.stdout)
        console_handler.setLevel(getattr(logging, level.upper()))
        console_handler.setFormatter(JSONFormatter())
        logger.addHandler(console_handler)
    
    # File handler
    if log_file:
        # Ensure log directory exists
        log_path = Path(log_file)
        log_path.parent.mkdir(parents=True, exist_ok=True)
        
        file_handler = logging.FileHandler(log_file)
        file_handler.setLevel(getattr(logging, level.upper()))
        file_handler.setFormatter(JSONFormatter())
        logger.addHandler(file_handler)
    
    # Prevent duplicate logs
    logger.propagate = False
    
    return logger

def get_logger(name: str = "fastapi") -> logging.Logger:
    """Get a logger instance."""
    return logging.getLogger(name)

# Application logger
app_logger = setup_logger(
    name="fastapi.app",
    level="INFO",
    log_file="/app/logs/fastapi_app.log",
    console_output=True
)

# Error logger
error_logger = setup_logger(
    name="fastapi.error",
    level="ERROR",
    log_file="/app/logs/fastapi_error.log",
    console_output=True
)

# Performance logger
perf_logger = setup_logger(
    name="fastapi.performance",
    level="INFO",
    log_file="/app/logs/fastapi_performance.log",
    console_output=False
)

# Business event logger
business_logger = setup_logger(
    name="fastapi.business",
    level="INFO",
    log_file="/app/logs/fastapi_business.log",
    console_output=False
)
