from typing import List
from datetime import datetime
from app.schemas.playground import ExecutionStepSchema
from app.services.simulators.common.base_simulator import BaseSimulator, FunctionDefinitionTracker


class DoublyLinkedListSimulator(BaseSimulator):
    """Main DoublyLinkedList simulator class that orchestrates the simulation process
    
    Simplified approach: Uses DirectCodeExecutor with detailed tracing to handle
    all execution and step generation.
    """
    
    def __init__(self):
        super().__init__("doublylinkedlist")
        from app.services.simulators.direct_code_executor import DirectCodeExecutor
        self.direct_executor = DirectCodeExecutor(use_docker=False, timeout=30)
    
    def execute_code(self, code: str, exec_id: str, created_at: datetime) -> List[ExecutionStepSchema]:
        """Execute doubly linked list code with proper step-by-step tracking
        
        Uses DirectCodeExecutor which now has proper tracing for doublylinkedlist
        data structure type, providing detailed execution steps.
        """
        self.reset_context()  # Ensure clean context
        
        # Get input_values if available (set by controller)
        input_values = getattr(self, '_input_values', None)
        
        # Use DirectCodeExecutor for all cases - it now supports detailed tracing
        # for doublylinkedlist data structure type
        return self.direct_executor.execute(
            code,
            stdin_data="",
            data_structure_type="doublylinkedlist",
            input_values=input_values
        )
