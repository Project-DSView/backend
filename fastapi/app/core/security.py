from fastapi import HTTPException, Header, Depends
from typing import Optional
from .config import settings

async def verify_api_key(x_api_key: Optional[str] = Header(None, alias=settings.API_KEY_NAME, include_in_schema=False)) -> bool:
    """Verify API key from header"""
    if not settings.API_KEY:
        return True
    
    if not x_api_key:
        raise HTTPException(status_code=401, detail="API key is required")
    
    if x_api_key != settings.API_KEY:
        raise HTTPException(status_code=401, detail="Invalid API key")
    
    return True
