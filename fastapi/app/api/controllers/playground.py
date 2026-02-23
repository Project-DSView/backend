import asyncio
import ast
from datetime import datetime, timezone
from asyncio import TimeoutError
from typing import Optional, List

from app.schemas.playground import (
    ExecutionCreateRequest,
    ExecutionResponse,
    ExecutionStepSchema,
    ComplexityAnalysis,
    FunctionComplexity,
)
from app.services.simulators.operations.complexity_analyzer import ComplexityAnalyzer
from app.services.simulators.simulator_factory import SimulatorFactory
from app.services.simulators.operations.ast_parser import ASTParser
from app.services.simulators.data_structure_detector import DataStructureDetector
from app.utils.execution_helpers import ExecutionHelper
from app.core.config import settings


class ExecuteController:
    def __init__(self):
        """Initialize controller for playground mode"""
        self.simulator_factory = SimulatorFactory()

    async def run_code_guest(
        self, 
        request: ExecutionCreateRequest
    ) -> ExecutionResponse:
        """Execute code for guest users without storing data"""
        code = request.code.strip()
        data_type = request.dataType
        input_values = request.inputValues or []

        # Auto-detect data structure type if not provided or set to "auto"
        if not data_type or data_type == "auto" or request.autoDetect:
            detected_type = self.detector.detect_from_code(code)
            if detected_type:
                data_type = detected_type
            else:
                # If detection fails and no type provided, raise error
                if not request.dataType or request.dataType == "auto":
                    raise ValueError(
                        "Could not auto-detect data structure type. "
                        "Please specify dataType explicitly. "
                        f"Supported types: {self.simulator_factory.get_supported_types()}"
                    )

        # Validate data type is supported
        if not self.simulator_factory.is_supported(data_type):
            raise NotImplementedError(
                f"DataType '{data_type}' not supported yet. "
                f"Supported types: {self.simulator_factory.get_supported_types()}"
            )

        exec_id = ExecutionHelper.generate_execution_id()
        created_at = datetime.now(timezone.utc)

        try:
            simulator = self.simulator_factory.create_simulator(data_type)
            # Set input_values on simulator if it has ast_executor
            if input_values and hasattr(simulator, 'ast_executor'):
                simulator._input_values = input_values
            steps = await self._execute_with_timeout(
                simulator, code, exec_id, created_at, timeout=settings.EXECUTION_TIMEOUT, input_values=input_values
            )
            
            # Check if any step has an error
            error_message = None
            for step in steps:
                if step.state and isinstance(step.state, dict):
                    error = step.state.get("error")
                    if error and isinstance(error, str) and error.strip():
                        error_message = error
                        break
            
            # If there's an error, return error response with error steps included
            if error_message:
                return ExecutionHelper.create_error_response(
                    exec_id, code, data_type, error_message, created_at, steps
                )
            
            # Extract output from steps (usually available in the last step's state)
            output = None
            if steps:
                # Iterate steps to find one with stdout
                for step in reversed(steps):
                    if step.state and isinstance(step.state, dict):
                        stdout = step.state.get("stdout")
                        if stdout:
                            if isinstance(stdout, list):
                                output = "\n".join(str(s) for s in stdout)
                            else:
                                output = str(stdout)
                            break
            
            # Check if waiting for input
            is_waiting = False
            if steps and steps[-1].state.get("waiting_for_input"):
                is_waiting = True
            
            # Return response directly without storing
            status = "waiting" if is_waiting else "success"
            
            # Analyze complexity (Basic AST analysis for immediate feedback)
            complexity_result = None
            try:
                analyzer = ComplexityAnalyzer()
                result = analyzer.analyze(code)
                complexity_result = ComplexityAnalysis(
                    timeComplexity=result.time_complexity,
                    spaceComplexity=result.space_complexity,
                    timeExplanation=result.time_explanation,
                    spaceExplanation=result.space_explanation,
                    analysisDetails=result.analysis_details,
                    functionComplexities=[
                        FunctionComplexity(
                            functionName=fc.function_name,
                            timeComplexity=fc.time_complexity,
                            spaceComplexity=fc.space_complexity,
                            lineStart=fc.line_start,
                            lineEnd=fc.line_end,
                            timeComplexityRank=fc.time_complexity_rank
                        )
                        for fc in result.function_complexities
                    ] if result.function_complexities else None
                )
            except Exception:
                # If complexity analysis fails, continue without it
                pass
            
            return ExecutionResponse(
                executionId=exec_id,
                code=code,
                dataType=data_type,
                steps=steps,
                totalSteps=len(steps),
                status=status,
                executedAt=created_at,
                createdAt=created_at,
                output=output,
                complexity=complexity_result
            )
            
        except TimeoutError:
            return ExecutionHelper.create_error_response(
                exec_id, code, data_type, "Execution timeout - code took too long to execute", created_at, []
            )
        except Exception as e:
            return ExecutionHelper.create_error_response(
                exec_id, code, data_type, str(e), created_at, []
            )

    async def _execute_with_timeout(
        self, simulator, code: str, exec_id: str, created_at: datetime, timeout: int = 30, input_values: Optional[List[str]] = None
    ) -> list[ExecutionStepSchema]:
        """Execute code with timeout protection"""
        try:
            return await asyncio.wait_for(
                self._execute_with_simulator(simulator, code, exec_id, created_at, input_values),
                timeout=timeout
            )
        except TimeoutError:
            raise TimeoutError(f"Code execution exceeded {timeout} seconds")

    async def _execute_with_simulator(
        self, simulator, code: str, exec_id: str, created_at: datetime, input_values: Optional[List[str]] = None
    ) -> list[ExecutionStepSchema]:
        """Execute code using the appropriate simulator with async optimization"""
        if hasattr(simulator, "execute_code") and callable(
            getattr(simulator, "execute_code")
        ):
            # Set input_values on simulator if it has direct_executor or ast_executor
            if input_values:
                if hasattr(simulator, 'direct_executor'):
                    simulator._input_values = input_values
                elif hasattr(simulator, 'ast_executor'):
                    simulator._input_values = input_values
            
            # Use asyncio.to_thread for CPU-intensive operations
            result = await asyncio.to_thread(
                simulator.execute_code, code, exec_id, created_at
            )
            return result
        else:
            raise ValueError("Simulator does not have execute_code method")

    async def cleanup(self):
        """Cleanup resources - minimal cleanup for guest mode"""
        # Force garbage collection to free memory
        import gc
        gc.collect()