"""
Health check endpoints.
"""
from datetime import datetime, timezone
from fastapi import APIRouter, Response, status
import httpx
from ..core.config import settings

router = APIRouter()

@router.get("/health")
async def health_check(response: Response):
    """Health probe: checks if process is running and dependencies are reachable."""
    ollama_status = "ok"
    ollama_models = []
    
    ollama_base_url = getattr(settings, "OLLAMA_BASE_URL", "http://ollama:11434")
    
    # Check Ollama dependency and fetch available models with a fast timeout (2s)
    try:
        async with httpx.AsyncClient() as client:
            res = await client.get(f"{ollama_base_url}/api/tags", timeout=2.0)
            res.raise_for_status()
            
            # Parse models from response
            data = res.json()
            models = data.get("models", [])
            ollama_models = [m.get("name") for m in models if "name" in m]
            if not ollama_models:
                ollama_status = "ok (no models pulled)"
                
    except Exception as e:
        ollama_status = f"down: {str(e)}"
    
    health_status = "ok"
    if ollama_status.startswith("down"):
        health_status = "error"
        response.status_code = status.HTTP_503_SERVICE_UNAVAILABLE

    return {
        "status": health_status,
        "service": "fastapi",
        "timestamp": datetime.now(timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ"),
        "version": "0.0.5-alpha",
        "env": settings.ENVIRONMENT,
        "dependencies": {
            "ollama": {
                "status": "ok" if not ollama_status.startswith("down") else "down",
                "details": ollama_status,
                "models": ollama_models
            }
        }
    }
