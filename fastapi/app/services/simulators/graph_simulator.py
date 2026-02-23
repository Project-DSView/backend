from typing import List, Dict, Any
from datetime import datetime
from app.schemas.playground import ExecutionStepSchema
from app.services.simulators.common.base_simulator import BaseSimulator, FunctionDefinitionTracker
from app.services.simulators.graph.enhanced_graph_operation_parser import EnhancedGraphOperationParser
from app.services.simulators.operations.ast_parser import ASTParser
from app.services.simulators.operations.error_handler import ErrorHandler
from app.services.simulators.direct_code_executor import DirectCodeExecutor


class GraphSimulator(BaseSimulator):
    """Enhanced Graph simulator with detailed step-by-step execution tracking"""
    
    def __init__(self):
        super().__init__("graph")
        self.ast_parser = ASTParser()
        self.direct_executor = DirectCodeExecutor(use_docker=False, timeout=30)
        
    def execute_code(self, code: str, exec_id: str, created_at: datetime) -> List[ExecutionStepSchema]:
        """Execute graph code with detailed step-by-step tracking"""
        steps = []
        step_number = 1  # Initialize before try block to handle early exceptions
        self.reset_context()
        
        try:
            # Check if code should be executed directly (e.g. input/print or custom class)
            has_input = 'input(' in code
            has_fstring = 'f"' in code or "f'" in code
            is_simulation = 'class ' in code and ('Graph' in code or 'graph' in code)
            
            if has_input or has_fstring or not is_simulation:
                # Use direct executor for arbitrary code
                return self.direct_executor.execute(code, data_structure_type="graph", input_values=getattr(self, '_input_values', []))
            
            # Calculate real execution steps for state tracking
            execution_steps = self.direct_executor.execute(
                code, 
                data_structure_type="graph", 
                input_values=getattr(self, '_input_values', [])
            )
            
            # Parse the entire code using AST to catch syntax errors early
            self.ast_parser.parse_code(code)
            # Initialize Graph class
            self._initialize_graph_class(steps)
            
            # Process each line of executable code
            lines = code.split('\n')
            step_number = 2
            
            operation_parser = EnhancedGraphOperationParser(self.context)
            
            # Track function definition state
            function_tracker = FunctionDefinitionTracker()
            
            for line_number, line in enumerate(lines, 1):
                original_line = line
                stripped_line = line.strip()
                
                # Skip empty lines and comments
                if not stripped_line or stripped_line.startswith('#'):
                    continue
                
                # Handle function definitions
                if function_tracker.update_state(original_line, stripped_line):
                    continue
                
                # Skip class definitions 
                if stripped_line.startswith('class '):
                    continue
                
                # Skip function body indented lines
                if original_line.startswith('    ') and not stripped_line:
                    continue
                
                try:
                    # Execute the line
                    if operation_parser.parse_and_execute(stripped_line, line_number, step_number, 
                                                        steps, self._create_execution_step):
                        step_number += 1
                    
                except Exception as e:
                        steps.append(self._create_execution_step(
                            step_number, line_number, stripped_line, error=str(e)
                        ))
                        raise e
            
            # If execution_steps are more detailed, return them instead
            # This ensures line-by-line trace data is used for debugging
            if execution_steps:
                if len(execution_steps) > len(steps) or len(steps) <= 1 or is_simulation:
                    return execution_steps
        
        except ValueError as e:
            # Handle syntax errors and other ValueError from AST parser
            error_message = str(e)
            
            # Try to extract line number from error message or code
            error_line = 0
            code_line = ""
            if hasattr(e, 'line_number'):
                error_line = e.line_number
            elif hasattr(e, 'lineno'):
                error_line = e.lineno
            
            # Try to get the problematic line from code
            if error_line > 0:
                lines = code.split('\n')
                if error_line <= len(lines):
                    code_line = lines[error_line - 1]
            
            # Format error with ErrorHandler if we have more details
            if hasattr(e, 'error_type'):
                error_info = {
                    "error_type": e.error_type,
                    "error_message": str(e),
                    "thai_message": error_message,
                    "line_number": error_line,
                    "code_line": code_line
                }
            else:
                # Use ErrorHandler to format the error
                error_info = ErrorHandler.format_error(e, error_line, code_line, offset)
                error_message = error_info.get("python_style_message", error_info["thai_message"])
            
            # Create error step with detailed error information
            error_step = self._create_execution_step(
                step_number, error_line, code_line,
                error=error_message,
                state={
                    "error": error_message,
                    "error_type": error_info.get("error_type", "ValueError"),
                    "error_message": error_info.get("error_message", str(e)),
                    "code_line": code_line,
                    "instances": {},
                    "active": None,
                    "print_output": []
                }
            )
            steps.append(error_step)
            return steps
        except SyntaxError as e:
            # Handle syntax errors directly
            error_line = e.lineno if hasattr(e, 'lineno') and e.lineno else 0
            code_line = e.text if hasattr(e, 'text') and e.text else ""
            
            error_info = ErrorHandler.format_error(e, error_line, code_line, offset)
            
            error_step = self._create_execution_step(
                step_number, error_line, code_line,
                error=error_info["thai_message"],
                state={
                    "error": error_info["thai_message"],
                    "error_type": error_info["error_type"],
                    "error_message": error_info["error_message"],
                    "code_line": code_line,
                    "instances": {},
                    "active": None,
                    "print_output": []
                }
            )
            steps.append(error_step)
            return steps
        except Exception as e:
            # Handle other exceptions
            error_info = ErrorHandler.format_error(e, 0, "")
            
            if not steps:
                error_step = self._create_execution_step(
                    step_number, 0, code.split('\n')[0] if code else "",
                    error=error_info["thai_message"],
                    state={
                        "error": error_info.get("python_style_message", error_info["thai_message"]),
                        "error_type": error_info["error_type"],
                        "error_message": error_info["error_message"],
                        "instances": {},
                        "active": None,
                        "print_output": []
                    }
                )
                steps.append(error_step)
            return steps
        
        return steps
    
    def _initialize_graph_class(self, steps: List[ExecutionStepSchema]):
        """Initialize Graph class in context"""
        self.context["classes"] = {
            "Graph": {
                "type": "class",
                "methods": ["__init__", "add_vertex", "add_edge", "remove_edge", "remove_vertex", 
                           "display", "bfs", "dfs"],
                "defined": True,
                "line_number": 1
            }
        }
        
        # Add class definition step
        steps.append(ExecutionStepSchema(
            stepNumber=1,
            line=1,
            code="class Graph:",
            message="Defined Graph class with methods: __init__, add_vertex, add_edge, remove_edge, remove_vertex, display, bfs, dfs",
            state=self._create_initial_state()
        ))
    
    def _create_initial_state(self) -> Dict[str, Any]:
        """Create initial state"""
        return {
            "instances": {},
            "variables": {},
            "print_output": [],
            "active": None,
            "classes_defined": ["Graph"],
            "step_detail": {
                "operation": "class_definition",
                "class_name": "Graph"
            }
        }
    
    def _create_execution_step(self, step_number: int, line_number: int, code: str, 
                             message: str = None, state: Dict[str, Any] = None, error: str = None) -> ExecutionStepSchema:
        """Create a standardized execution step with enhanced details"""
        if state is None:
            state = {
                "instances": {k: self._get_instance_display(v) for k, v in self.context["instances"].items()},
                "variables": self.context["variables"].copy(),
                "print_output": self.context["print_output"].copy(),
                "active": self.context["active_instance"]
            }
        
        if message:
            state["message"] = message
        if error:
            state["error"] = error
            
        return ExecutionStepSchema(
            stepNumber=step_number,
            line=line_number,
            code=code,
            state=state
        )
    
    def _get_instance_display(self, instance):
        """Get display representation of graph instances"""
        if instance.get("class_type") == "Graph":
            adjacency_list = instance.get("adjacency_list", {})
            return {
                "type": "Graph",
                "adjacency_list": adjacency_list,
                "vertices": list(adjacency_list.keys()),
                "vertex_count": len(adjacency_list),
                "edge_count": self._count_edges(adjacency_list),
                "is_empty": len(adjacency_list) == 0
            }
        return super()._get_instance_display(instance)
    
    def _count_edges(self, adjacency_list: Dict[str, List[str]]) -> int:
        """Count total edges in the graph (undirected)"""
        if not adjacency_list:
            return 0
        
        total_connections = sum(len(neighbors) for neighbors in adjacency_list.values())
        # For undirected graph, divide by 2 since each edge is counted twice
        return total_connections // 2


class FunctionDefinitionTracker:
    """Helper class to track function definition state"""
    
    def __init__(self):
        self.in_function_def = False
        self.function_indent_level = 0
        self.in_class_def = False
        self.class_indent_level = 0
    
    def update_state(self, original_line: str, stripped_line: str) -> bool:
        """Update function definition state and return True if line should be skipped"""
        current_indent = len(original_line) - len(original_line.lstrip())
        
        # Track class definitions
        if stripped_line.startswith('class '):
            self.in_class_def = True
            self.class_indent_level = current_indent
            return True
        elif self.in_class_def:
            if current_indent <= self.class_indent_level and stripped_line and not stripped_line.startswith('class '):
                self.in_class_def = False
            elif current_indent > self.class_indent_level:
                if stripped_line.startswith('def '):
                    self.in_function_def = True
                    self.function_indent_level = current_indent
                    return True
                elif self.in_function_def:
                    if current_indent <= self.function_indent_level and stripped_line:
                        self.in_function_def = False
                        return False
                    else:
                        return True
                else:
                    return True
        
        # Track standalone function definitions
        if stripped_line.startswith('def '):
            self.in_function_def = True
            self.function_indent_level = current_indent
            return True
        elif self.in_function_def:
            if current_indent <= self.function_indent_level and stripped_line:
                self.in_function_def = False
                return False
            else:
                return True
        
        return False