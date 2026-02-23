from fastapi import APIRouter, Depends, HTTPException
from app.api.controllers.complexity_controller import ComplexityController
from app.schemas.complexity import (
    ComplexityAnalysisRequest, 
    ComplexityAnalysisResponse,
    PerformanceAnalysisResponse
)

router = APIRouter()
controller = ComplexityController()

@router.post("/llm", response_model=ComplexityAnalysisResponse)
async def analyze_complexity_llm(request: ComplexityAnalysisRequest):
    """
    Analyze code complexity using LLM (Ollama).
    This provides a detailed explanation of Big O.
    """
    try:
        return await controller.analyze_with_llm(request)
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@router.post("/performance", response_model=PerformanceAnalysisResponse)
async def analyze_performance(request: ComplexityAnalysisRequest):
    """
    Analyze performance using AST analyzer (Legacy method).
    This provides fast metrics on loops, recursion, and estimated complexity.
    """
    try:
        return await controller.analyze_performance(request)
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))
