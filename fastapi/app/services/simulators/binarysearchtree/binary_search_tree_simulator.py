from typing import List
from datetime import datetime

from app.schemas.playground import ExecutionStepSchema
from app.services.simulators.common.base_simulator import BaseSimulator
from app.services.simulators.operations.ast_parser import ASTParser
from app.services.simulators.operations.enhanced_operation_parser import EnhancedOperationParser
from app.services.simulators.operations.error_handler import ErrorHandler
from app.services.simulators.direct_code_executor import DirectCodeExecutor


class BinarySearchTreeSimulator(BaseSimulator):
    """
    Main BinarySearchTree simulator class that orchestrates the simulation process
    Now supports direct execution for step-through debugging
    """
    
    def __init__(self):
        super().__init__("binarysearchtree")
        self.ast_parser = ASTParser()
        self.operation_parser = None  # Initialize in execute_code to pass context
        self.direct_executor = DirectCodeExecutor(use_docker=False, timeout=60)  # Use direct execution
    
    def _get_operation_parser(self):
        """Get the operation parser for BST"""
        if self.operation_parser is None:
            self.operation_parser = EnhancedOperationParser(self.context)
        return self.operation_parser
    
    def execute_code(self, code: str, exec_id: str, created_at: datetime) -> List[ExecutionStepSchema]:
        """
        Execute BST code using AST-first + real Python execution
        Combines AST parsing with real execution results
        """
        steps = []
        step_number = 1
        self.reset_context()  # Ensure clean context
        self.operation_parser = EnhancedOperationParser(self.context)
        
        try:
            # Get input_values if available (set by controller)
            input_values = getattr(self, '_input_values', None)
            
            # Use direct executor for all execution (BST logic is now handled in DirectCodeExecutor)
            # This provides better step-through experience than the old simulation approach
            
            # Execution steps used for merging output later if in simulation mode
            execution_steps = self.direct_executor.execute(
                code,
                stdin_data="",
                data_structure_type="binarysearchtree",
                input_values=input_values
            )
            
            # If we simply want to rely on DirectCodeExecutor (which now has BST support), we can just return its steps
            # This is cleaner and less error-prone than trying to mix old AST simulation with new execution
            return execution_steps
            
        except ValueError as e:
            # Handle syntax errors and other ValueError from AST parser
            error_message = str(e)
            
            # Format error with ErrorHandler
            error_info = ErrorHandler.format_error(e, 0, "")
            
            # Create error step with detailed error information
            error_step = self._create_execution_step(
                step_number, 0, "",
                error=error_info.get("python_style_message", error_info["thai_message"]),
                additional_state={
                    "error_type": error_info.get("error_type", "ValueError"),
                    "error_message": error_info.get("error_message", str(e))
                }
            )
            steps.append(error_step)
            return steps
        except Exception as e:
            # Handle other exceptions
            error_info = ErrorHandler.format_error(e, 0, "")
            
            error_step = self._create_execution_step(
                step_number, 0, "",
                error=error_info.get("python_style_message", error_info["thai_message"]),
                additional_state={
                    "error_type": error_info["error_type"],
                    "error_message": error_info["error_message"]
                }
            )
            steps.append(error_step)
            return steps

    def _get_instance_display(self, instance):
        """Get display representation of a BST instance"""
        if not isinstance(instance, dict):
            return []
            
        # BST display logic if needed (DirectCodeExecutor already handles structure)
        # But BaseSimulator might call this if we used the mixed approach
        return []
