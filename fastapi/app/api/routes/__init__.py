"""
API routes module for DSView Backend API.

This module provides all API route definitions and router configuration.
"""

from fastapi import APIRouter
from .playground import router as playground_router
from .complexity import router as complexity_router

# Create main API router
api_router = APIRouter()

# Guest routes (no auth required)
api_router.include_router(
    playground_router,
    prefix="/playground", 
    tags=["Playground (Guest)"]
)

# Complexity routes
api_router.include_router(
    complexity_router,
    prefix="/complexity",
    tags=["Complexity"]
)

# Export for easy import
__all__ = ["api_router", "playground_router", "complexity_router"]