from typing import List, Dict, Any
from app.schemas.playground import ExecutionStepSchema
from app.services.simulators.operations.v0_5.operation_parser import OperationParser
from app.services.simulators.graph.graph_statement_parser import GraphStatementParser
from app.services.simulators.graph.graph_node_manager import GraphNodeManager


class EnhancedGraphOperationParser(OperationParser):
    """Enhanced Graph-specific operation parser with better step tracking"""
    
    def __init__(self, context: Dict[str, Any]):
        super().__init__(context)
        self.node_manager = GraphNodeManager(context)
        self.statement_parser = GraphStatementParser(context, self.print_handler)
    
    def parse_and_execute(self, line: str, line_number: int, step_number: int, 
                         steps: List[ExecutionStepSchema], 
                         create_step_func) -> bool:
        """Parse and execute graph operations with enhanced step tracking"""
        
        # Handle multiple statements separated by semicolons
        if ';' in line:
            return self._handle_multiple_statements(line, line_number, step_number, steps, create_step_func)
        
        return self.statement_parser.execute_single_statement(
            line, line_number, step_number, steps, create_step_func
        )
    
    def _handle_multiple_statements(self, line: str, line_number: int, step_number: int, 
                                   steps: List[ExecutionStepSchema], create_step_func) -> bool:
        """Handle multiple statements separated by semicolons"""
        statements = [stmt.strip() for stmt in line.split(';') if stmt.strip()]
        handled_any = False
        current_step = step_number
        
        for stmt in statements:
            if self.statement_parser.execute_single_statement(
                stmt, line_number, current_step, steps, create_step_func
            ):
                handled_any = True
                current_step += 1
        
        return handled_any