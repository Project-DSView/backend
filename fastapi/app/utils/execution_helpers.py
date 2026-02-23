import uuid
from datetime import datetime
from typing import List

from app.schemas.playground import ExecutionResponse, ExecutionStepSchema


class ExecutionHelper:
    """Helper class for execution-related utilities"""
    
    @staticmethod
    def generate_execution_id() -> str:
        """Generate a unique execution ID"""
        return f"exec_{uuid.uuid4().hex[:12]}"
    
    @staticmethod
    def create_success_response(
        exec_id: str,
        code: str, 
        data_type: str,
        steps: List[ExecutionStepSchema],
        created_at: datetime,
        output: str = None
    ) -> ExecutionResponse:
        """Create a successful execution response"""
        return ExecutionResponse(
            executionId=exec_id,
            code=code,
            dataType=data_type,
            steps=steps,
            totalSteps=len(steps),
            status="success",
            executedAt=created_at,
            createdAt=created_at,
            output=output
        )
    
    @staticmethod
    def create_error_response(
        exec_id: str,
        code: str,
        data_type: str, 
        error_message: str,
        created_at: datetime,
        steps: List[ExecutionStepSchema] = None
    ) -> ExecutionResponse:
        """Create an error execution response"""
        # If steps are provided (e.g., error steps from simulator), use them
        # Otherwise, create an empty steps list
        if steps is None:
            steps = []
        
        return ExecutionResponse(
            executionId=exec_id,
            code=code,
            dataType=data_type,
            steps=steps,
            totalSteps=len(steps),
            status="error",
            errorMessage=error_message,
            executedAt=created_at,
            createdAt=created_at
        )