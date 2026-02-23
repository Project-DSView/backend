from typing import List
from datetime import datetime
from app.schemas.playground import ExecutionStepSchema
from app.services.simulators.common.base_simulator import BaseSimulator
from app.services.simulators.operations.ast_parser import ASTParser
from app.services.simulators.operations.error_handler import ErrorHandler
from app.services.simulators.direct_code_executor import DirectCodeExecutor


class UndirectedGraphSimulator(BaseSimulator):
    """Enhanced Undirected Graph simulator with detailed step-by-step execution tracking"""
    
    def __init__(self):
        super().__init__("undirectedgraph")
        self.ast_parser = ASTParser()
        self.direct_executor = DirectCodeExecutor(use_docker=False, timeout=90)  # Use direct execution, increased timeout
    
    def _get_operation_parser(self):
        # Needed for base class but not used in direct execution
        return None
        
    def execute_code(self, code: str, exec_id: str, created_at: datetime) -> List[ExecutionStepSchema]:
        """Execute undirected graph code with detailed step-by-step tracking"""
        steps = []
        self.reset_context()
        
        try:
            # Get input_values if available (set by controller)
            input_values = getattr(self, '_input_values', None)
            
            # Use direct executor for all execution
            return self.direct_executor.execute(
                code, 
                data_structure_type="undirectedgraph", 
                input_values=input_values
            )
            
        except Exception as e:
            # Handle exceptions
            error_info = ErrorHandler.format_error(e, 0, "")
            
            error_step = self._create_execution_step(
                1, 0, "",
                error=error_info.get("python_style_message", error_info["thai_message"]),
                additional_state={
                    "error_type": error_info["error_type"],
                    "error_message": error_info["error_message"]
                }
            )
            steps.append(error_step)
            return steps

    def _get_graph_display(self, instance):
        """Get graph display representation for undirected graph"""
        if not isinstance(instance, dict):
            return []
        
        graph_data = instance.get("graph", {})
        if not isinstance(graph_data, dict):
            return []
        
        # Convert graph to display format used by BaseSimulator
        display = []
        for vertex, neighbors in graph_data.items():
            display.append(f"{vertex}: {list(neighbors.keys())}")
        
        return display