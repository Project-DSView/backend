"""
DSView Backend API - Data Structure Visualization Backend

This is the main application module for the DSView Backend API.
It provides a stateless architecture for code execution and visualization
of various data structures including stacks, linked lists, binary search trees, and graphs.

The application is organized into the following modules:
- api: API routes and controllers
- core: Core configuration and security
- errors: Error handling and exceptions
- schemas: Pydantic models and data validation
- services: Business logic and simulators
- utils: Utility functions and helpers
"""

# Version information
__version__ = "0.0.5-alpha"
__author__ = "DSView Team"
__description__ = "Data Structure Visualization Backend API"

# Main application components
from .main import app

# Import utilities (lazy loading to avoid circular imports)
def get_utils():
    """Get utils module with lazy loading."""
    from . import utils
    return utils

# Import services (lazy loading to avoid circular imports)
def get_services():
    """Get services module with lazy loading."""
    from . import services
    return services

# Import errors (lazy loading to avoid circular imports)
def get_errors():
    """Get errors module with lazy loading."""
    from . import errors
    return errors

__all__ = [
    # Application
    "app",
    "__version__",
    "__author__",
    "__description__",
    
    # Lazy loading functions
    "get_utils",
    "get_services", 
    "get_errors",
]
