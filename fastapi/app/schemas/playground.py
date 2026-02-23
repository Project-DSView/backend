from datetime import datetime
from typing import List, Literal, Optional, Dict, Any
from pydantic import BaseModel, Field, field_validator
from app.core.config import settings

# Enum
DataType = Literal["stack", "singlylinkedlist", "doublylinkedlist", "binarysearchtree", "undirectedgraph", "directedgraph", "queue", "auto"]
StepStatus = Literal["success", "error", "timeout", "waiting"]

class ExecutionStepSchema(BaseModel):
    stepNumber: int
    line: int
    code: str
    state: dict

class ExecutionCreateRequest(BaseModel):    
    code: str = Field(..., min_length=1, max_length=settings.MAX_CODE_LENGTH, description="The code to execute")
    dataType: Optional[DataType] = Field(default=None, description="Type of data structure (optional, will auto-detect if not provided or 'auto')")
    inputValues: Optional[List[str]] = Field(default=[], description="Pre-collected input values for input() calls")
    autoDetect: Optional[bool] = Field(default=False, description="Whether to auto-detect data structure type")
    
    @field_validator('code')
    @classmethod
    def validate_code(cls, v):
        """Validate and sanitize code input"""
        if not v or not v.strip():
            raise ValueError("Code cannot be empty")
        
        # Check for dangerous patterns
        # Note: input() is allowed in stepthrough mode (playground) for educational purposes
        dangerous_patterns = [
            'import os', 'import sys', 'import subprocess',
            'exec(', 'eval(', '__import__', 'open(',
            'file(', 'raw_input(',  # input() is allowed in stepthrough mode
            'compile(', 'reload(', 'vars(', 'globals(',
            'locals(', 'dir(', 'getattr(', 'setattr(',
            'delattr(', 'hasattr(', 'callable('
        ]
        
        code_lower = v.lower()
        for pattern in dangerous_patterns:
            if pattern in code_lower:
                raise ValueError(f"Dangerous code pattern detected: {pattern}")
        
        # Check for excessive loops that might cause DoS
        # Count 'for' and 'while' at the beginning of lines (actual loop statements)
        lines = v.split('\n')
        for_count = sum(1 for line in lines if line.strip().startswith('for ') and ':' in line)
        while_count = sum(1 for line in lines if line.strip().startswith('while ') and ':' in line)
        
        if while_count > settings.MAX_LOOPS or for_count > settings.MAX_FOR_LOOPS:
            raise ValueError(f"Too many loops detected - potential DoS risk (found {for_count} for loops, {while_count} while loops, max allowed: {settings.MAX_FOR_LOOPS} for, {settings.MAX_LOOPS} while)")
        
        # Check for excessive recursion
        if code_lower.count('def ') > settings.MAX_FUNCTIONS:
            raise ValueError("Too many function definitions - potential DoS risk")
        
        return v.strip()

class FunctionComplexity(BaseModel):
    """Complexity analysis for a single function"""
    functionName: str = Field(description="Name of the function")
    timeComplexity: str = Field(description="Time complexity e.g. O(n)")
    spaceComplexity: str = Field(description="Space complexity e.g. O(1)")
    lineStart: int = Field(description="Starting line number of the function")
    lineEnd: int = Field(description="Ending line number of the function")
    timeComplexityRank: int = Field(description="Numeric rank for comparison (higher = slower)")


class ComplexityAnalysis(BaseModel):
    """Big O complexity analysis result"""
    timeComplexity: str = Field(description="Time complexity e.g. O(n), O(n²)")
    spaceComplexity: str = Field(description="Space complexity e.g. O(1), O(n)")
    timeExplanation: str = Field(description="Thai explanation for time complexity")
    spaceExplanation: str = Field(description="Thai explanation for space complexity")
    llmExplanation: Optional[str] = Field(default=None, description="Detailed LLM explanation in Thai")
    analysisDetails: Optional[Dict[str, Any]] = Field(default=None, description="Detailed analysis breakdown")
    functionComplexities: Optional[List[FunctionComplexity]] = Field(default=None, description="Per-function complexity breakdown")


class ExecutionResponse(BaseModel):
    executionId: str
    code: str
    dataType: DataType
    steps: List[ExecutionStepSchema]
    totalSteps: int
    status: StepStatus
    errorMessage: Optional[str] = None
    executedAt: datetime
    createdAt: datetime
    output: Optional[str] = None
    complexity: Optional[ComplexityAnalysis] = Field(default=None, description="Big O complexity analysis")


# AST Preview Schemas
class ASTNodeMetadata(BaseModel):
    """Metadata for a single AST node"""
    type: str
    typeDisplay: Optional[str] = None
    line: Optional[int] = None
    colOffset: Optional[int] = None
    category: Optional[str] = None
    functionName: Optional[str] = None
    methodName: Optional[str] = None
    objectName: Optional[str] = None
    variableName: Optional[str] = None
    className: Optional[str] = None
    operator: Optional[str] = None
    isBuiltin: Optional[bool] = None
    importance: Optional[str] = None  # "high", "medium", "low"


class ASTPreviewRequest(BaseModel):
    """Request to preview AST structure without execution"""
    code: str = Field(..., min_length=1, max_length=settings.MAX_CODE_LENGTH)


class ASTPreviewResponse(BaseModel):
    """Response with AST structure information"""
    astNodes: List[ASTNodeMetadata]
    nodeCount: int
    hasInput: bool
    executableLines: List[int]
    classes: List[str]
    inputCalls: List[Dict[str, Any]]  # List of input() call locations with prompts