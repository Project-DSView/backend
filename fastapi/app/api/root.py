"""
Root endpoint and API information.
"""

from fastapi import Request, APIRouter

router = APIRouter()


@router.get("/")
def read_root(request: Request):
    return {
        "message": "DSView Backend API (FastAPI)",
        "status": "healthy",
        "timestamp": datetime.now(timezone.utc).isoformat(),
        "version": "0.0.5-alpha",
        "service": "DSView Backend API",
        "uptime": "running",
    }
