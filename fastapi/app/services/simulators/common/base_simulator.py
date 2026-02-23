"""
Common base simulator functionality for DSView Backend API.

This module provides shared functionality for all data structure simulators.
"""

from typing import List
from app.schemas.playground import ExecutionStepSchema
from app.services.simulators.operations.error_handler import ErrorHandler


class FunctionDefinitionTracker:
    """Tracks function definition state during code execution"""
    
    def __init__(self):
        self.in_function = False
        self.function_indent = 0
    
    def update_state(self, original_line: str, stripped_line: str) -> bool:
        """Update function definition state and return True if line should be skipped"""
        # Check if we're starting a function definition
        if stripped_line.startswith('def '):
            self.in_function = True
            self.function_indent = len(original_line) - len(original_line.lstrip())
            return True
        
        # Check if we're in a function body
        if self.in_function:
            if not stripped_line:  # Empty line
                return True
            
            current_indent = len(original_line) - len(original_line.lstrip())
            
            # If we're back to the same or less indentation, we're out of the function
            if current_indent <= self.function_indent and stripped_line:
                self.in_function = False
                self.function_indent = 0
                return False  # Process this line
            
            return True  # Skip function body lines
        
        return False


class BaseSimulator:
    """Base class for all data structure simulators with common functionality"""
    
    def __init__(self, data_structure_type: str):
        self.data_structure_type = data_structure_type
        self.reset_context()
    
    def reset_context(self):
        """Reset the simulator context with proper initialization"""
        self.context = {
            "classes": {}, 
            "instances": {},
            "variables": {},
            "active_instance": None,
            "nodes": {},
            "stdout": []
        }
        
        # Add data structure specific context
        if self.data_structure_type in ["singlylinkedlist", "doublylinkedlist"]:
            self.context["linkedlist"] = []
            self.context["include_linkedlist"] = True
        elif self.data_structure_type in ["undirectedgraph", "directedgraph"]:
            self.context["graph"] = {}
            self.context["edges"] = []
            self.context["vertices"] = []
    
    def _create_execution_step(self, step_number: int, line_number: int, code: str, 
                             message: str = None, error: str = None, additional_state: dict = None) -> ExecutionStepSchema:
        """Create a standardized execution step"""
        # Ensure context is properly structured
        self._ensure_context_structure()
        
        # Build base state safely
        state = {
            "instances": {},
            "active": self.context.get("active_instance"),
            "stdout": self.context.get("stdout", []).copy()
        }
        
        # Safely add instances with detailed information
        instances = self.context.get("instances", {})
        if isinstance(instances, dict):
            for k, v in instances.items():
                try:
                    instance_data = self._get_instance_display(v)
                    state["instances"][k] = instance_data
                    
                    # Add detailed instance state information
                    if isinstance(v, dict):
                        state[f"{k}_count"] = v.get("count", 0)
                        state[f"{k}_head"] = v.get("head")
                        state[f"{k}_class_type"] = v.get("class_type", "Unknown")
                except Exception:
                    state["instances"][k] = []
        
        # Add data structure specific state
        self._add_data_structure_state(state)
        
        # Add variables safely
        variables = self.context.get("variables", {})
        if isinstance(variables, dict) and variables:
            safe_vars = {}
            for k, v in variables.items():
                try:
                    if isinstance(v, (str, int, float, bool)) or v is None:
                        safe_vars[k] = v
                    else:
                        safe_vars[k] = str(v)
                except Exception:
                    safe_vars[k] = "undefined"
            if safe_vars:
                state["variables"] = safe_vars
        
        # Add any additional state (but exclude error field from step_detail)
        if additional_state and isinstance(additional_state, dict):
            # Filter out error field from step_detail if it exists
            filtered_state = {}
            for key, value in additional_state.items():
                if key == "step_detail" and isinstance(value, dict):
                    # Remove error field from step_detail
                    filtered_step_detail = {k: v for k, v in value.items() if k != "error"}
                    filtered_state[key] = filtered_step_detail
                else:
                    filtered_state[key] = value
            state.update(filtered_state)
            
        if message:
            state["message"] = message
        if error and isinstance(error, str) and error.strip():
            state["error"] = error
            
        return ExecutionStepSchema(
            stepNumber=step_number,
            line=line_number,
            code=code,
            state=state
        )
    
    def _add_data_structure_state(self, state: dict):
        """Add data structure specific state - to be overridden by subclasses"""
        if self.data_structure_type in ["singlylinkedlist", "doublylinkedlist"]:
            linkedlist = self.context.get("linkedlist", [])
            if isinstance(linkedlist, list):
                state["linkedlist"] = linkedlist.copy()
            else:
                state["linkedlist"] = []
            
            # Add nodes for linkedlist simulator
            nodes = self.context.get("nodes", {})
            if isinstance(nodes, dict):
                state["nodes"] = {}
                for k, v in nodes.items():
                    try:
                        if isinstance(v, dict):
                            state["nodes"][k] = {
                                "name": v.get("name", ""),
                                "next": v.get("next"),
                                "prev": v.get("prev"),  # For doubly linked list
                                "id": k
                            }
                        else:
                            state["nodes"][k] = {"name": str(v) if v is not None else "", "next": None, "prev": None, "id": k}
                    except Exception:
                        state["nodes"][k] = {"name": "", "next": None, "prev": None, "id": k}
        
        elif self.data_structure_type in ["undirectedgraph", "directedgraph"]:
            graph = self.context.get("graph", {})
            edges = self.context.get("edges", [])
            vertices = self.context.get("vertices", [])
            
            if isinstance(graph, dict):
                state["graph"] = graph.copy()
            else:
                state["graph"] = {}
            
            if isinstance(edges, list):
                state["edges"] = edges.copy()
            else:
                state["edges"] = []
            
            if isinstance(vertices, list):
                state["vertices"] = vertices.copy()
            else:
                state["vertices"] = []
    
    def _ensure_context_structure(self):
        """Ensure context has proper structure"""
        if not isinstance(self.context, dict):
            self.context = {}
        
        required_keys = {
            "classes": {},
            "instances": {},
            "variables": {},
            "nodes": {},
            "stdout": []
        }
        
        for key, default_value in required_keys.items():
            if key not in self.context:
                self.context[key] = default_value
            elif not isinstance(self.context[key], type(default_value)):
                self.context[key] = default_value
        
        # Add data structure specific keys
        if self.data_structure_type in ["singlylinkedlist", "doublylinkedlist"]:
            list_key = self.data_structure_type
            if list_key not in self.context or not isinstance(self.context[list_key], list):
                self.context[list_key] = []
            self.context["include_linkedlist"] = True
        elif self.data_structure_type in ["undirectedgraph", "directedgraph"]:
            if "graph" not in self.context or not isinstance(self.context["graph"], dict):
                self.context["graph"] = {}
            if "edges" not in self.context or not isinstance(self.context["edges"], list):
                self.context["edges"] = []
            if "vertices" not in self.context or not isinstance(self.context["vertices"], list):
                self.context["vertices"] = []
    
    def _get_instance_display(self, instance):
        """Get display representation of an instance - to be overridden by subclasses"""
        if not isinstance(instance, dict):
            return []
            
        class_type = instance.get("class_type")
        if class_type in ["SinglyLinkedList", "DoublyLinkedList"]:
            return self._traverse_linked_list(instance)
        elif class_type in ["ArrayStack", "Queue"]:
            data = instance.get("data", [])
            return data if isinstance(data, list) else []
        elif class_type in ["DirectedGraph", "UndirectedGraph"]:
            return self._get_graph_display(instance)
        
        # Fallback
        data = instance.get("data", [])
        return data if isinstance(data, list) else []
    
    def _traverse_linked_list(self, instance):
        """Traverse a linked list instance and return display data"""
        if not isinstance(instance, dict):
            return []
            
        result = []
        head_id = instance.get("head")
        if head_id is None:
            return []
        
        visited = set()  # Prevent infinite loops
        nodes = self.context.get("nodes", {})
        
        if not isinstance(nodes, dict):
            return []
        
        current_node_id = head_id
        while current_node_id is not None and current_node_id not in visited:
            visited.add(current_node_id)
            
            if current_node_id not in nodes:
                break
                
            node = nodes[current_node_id]
            if not isinstance(node, dict):
                break
                
            node_name = node.get("name", "")
            result.append(str(node_name))
            
            current_node_id = node.get("next")
                
        return result
    
    def _get_graph_display(self, instance):
        """Get graph display representation - to be overridden by subclasses"""
        return []
    
    def _initialize_class(self, class_name: str, steps: List[ExecutionStepSchema]):
        """Initialize a data structure class - to be overridden by subclasses"""
        pass
    
    def _process_line(self, line: str, line_number: int, step_number: int, 
                     operation_parser, function_tracker: FunctionDefinitionTracker) -> List[ExecutionStepSchema]:
        """Process a single line of code with common logic"""
        steps = []
        
        try:
            # Use the operation parser to handle the line
            # The EnhancedOperationParser uses parse_and_execute method
            if hasattr(operation_parser, 'parse_and_execute'):
                # Use the new method signature
                handled = operation_parser.parse_and_execute(
                    line, line_number, step_number, steps, self._create_execution_step
                )
                if not handled:
                    # No specific result, create a basic step
                    step = self._create_execution_step(step_number, line_number, line)
                    steps.append(step)
            else:
                # Fallback to old method if it exists
                if hasattr(operation_parser, 'parse_and_execute_line'):
                    result = operation_parser.parse_and_execute_line(line, line_number)
                    
                    if result:
                        if isinstance(result, list):
                            # Multiple steps returned
                            for i, step_data in enumerate(result):
                                step = self._create_execution_step(
                                    step_number + i, line_number, line,
                                    message=step_data.get('message'),
                                    error=step_data.get('error'),
                                    additional_state=step_data.get('state')
                                )
                                steps.append(step)
                        else:
                            # Single step returned
                            step = self._create_execution_step(
                                step_number, line_number, line,
                                message=result.get('message'),
                                error=result.get('error'),
                                additional_state=result.get('state')
                            )
                            steps.append(step)
                    else:
                        # No specific result, create a basic step
                        step = self._create_execution_step(step_number, line_number, line)
                        steps.append(step)
                else:
                    # No specific result, create a basic step
                    step = self._create_execution_step(step_number, line_number, line)
                    steps.append(step)
                
        except Exception as e:
            # Format error with ErrorHandler
            error_info = ErrorHandler.format_error(e, line_number, line)
            
            # Create error step with detailed error information
            error_step = self._create_execution_step(
                step_number, line_number, line,
                error=error_info.get("python_style_message", error_info["thai_message"]),
                additional_state={
                    "error_type": error_info["error_type"],
                    "error_message": error_info["error_message"],
                    "code_line": error_info["code_line"]
                }
            )
            steps.append(error_step)
        
        return steps
    
    def execute_code(self, code: str, exec_id: str, created_at) -> List[ExecutionStepSchema]:
        """Common code execution logic - can be overridden by subclasses"""
        steps = []
        self.reset_context()
        
        try:
            # Initialize data structure class
            self._initialize_class(steps)
            
            # Process each line of executable code
            lines = code.split('\n')
            step_number = len(steps) + 1
            
            # Get operation parser (to be implemented by subclasses)
            operation_parser = self._get_operation_parser()
            
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
                
                # Process the line
                line_steps = self._process_line(
                    original_line, line_number, step_number, 
                    operation_parser, function_tracker
                )
                steps.extend(line_steps)
                step_number += len(line_steps)
            
            return steps
            
        except Exception as e:
            # Format error with ErrorHandler
            error_info = ErrorHandler.format_error(e, 0, "")
            
            # Create error step with detailed error information
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
    
    def _get_operation_parser(self):
        """Get the appropriate operation parser - to be implemented by subclasses"""
        raise NotImplementedError("Subclasses must implement _get_operation_parser")
    
    def _initialize_class(self, steps: List[ExecutionStepSchema]):
        """Initialize the data structure class - to be implemented by subclasses"""
        pass
