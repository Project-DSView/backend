from pydantic import BaseModel
from typing import Optional, Dict, Any, List

class ComplexityAnalysisRequest(BaseModel):
    code: str
    model: str = "qwen2.5-coder:1.5b"  # Default model

class ComplexityAnalysisResponse(BaseModel):
    complexity: str
    explanation: str
    
class FunctionComplexitySchema(BaseModel):
    function_name: str
    time_complexity: str
    space_complexity: str
    line_start: int
    line_end: int
    time_complexity_rank: int

class PerformanceAnalysisResponse(BaseModel):
    time_complexity: str
    space_complexity: str
    time_explanation: str
    space_explanation: str
    analysis_details: Dict[str, Any]
    function_complexities: List[FunctionComplexitySchema]
