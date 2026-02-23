"""
Startup and shutdown event handlers for the FastAPI application.
"""

from contextlib import asynccontextmanager
from fastapi import FastAPI
import logging

logger = logging.getLogger(__name__)


@asynccontextmanager
async def lifespan(app: FastAPI):
    """
    Lifespan event handler for the FastAPI application.
    Handles startup and shutdown events.
    
    Args:
        app: FastAPI application instance
    """
    # Startup
    logger.info("🚀 DSView Backend API is starting up...")
    logger.info("✅ Application startup completed")
    
    yield
    
    # Shutdown
    logger.info("🛑 DSView Backend API is shutting down...")
    logger.info("✅ Application shutdown completed")


def setup_startup_events(app: FastAPI) -> None:
    """
    Setup startup and shutdown events for the FastAPI application.
    This function is kept for backward compatibility but lifespan is used instead.
    
    Args:
        app: FastAPI application instance
    """
    # Note: lifespan is now handled in main.py
    pass
