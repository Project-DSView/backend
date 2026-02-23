"""
Request ID middleware for tracking requests across the application.
"""

import uuid
from contextvars import ContextVar
from fastapi import Request

# Request ID tracking
request_id: ContextVar[str] = ContextVar('request_id')


async def add_request_id_middleware(request: Request, call_next):
    """Add request ID to all requests for tracking"""
    req_id = str(uuid.uuid4())
    request_id.set(req_id)
    response = await call_next(request)
    response.headers["X-Request-ID"] = req_id
    return response
