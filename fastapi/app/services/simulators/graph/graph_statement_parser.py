import re
from typing import List, Dict, Any
from app.schemas.playground import ExecutionStepSchema
from app.services.simulators.graph.graph_method_executor import GraphMethodExecutor
from app.services.simulators.graph.graph_node_manager import GraphNodeManager


class GraphStatementParser:
    """Handles parsing and execution of individual Graph statements"""
    
    def __init__(self, context: Dict[str, Any], print_handler):
        self.context = context
        self.print_handler = print_handler
        self.method_executor = GraphMethodExecutor(context)
        self.node_manager = GraphNodeManager(context)
    
    def execute_single_statement(self, line: str, line_number: int, step_number: int, 
                                 steps: List[ExecutionStepSchema], create_step_func) -> bool:
        """Execute a single statement with detailed tracking"""
        
        # Handle class instantiation
        if self._handle_class_instantiation(line, line_number, step_number, steps, create_step_func):
            return True
        
        # Handle method calls
        if self._handle_method_calls(line, line_number, step_number, steps, create_step_func):
            return True
        
        # Handle variable assignments from method calls
        if self._handle_assignment_from_method(line, line_number, step_number, steps, create_step_func):
            return True
        
        # Handle simple variable assignments
        if self._handle_simple_assignment(line, line_number, step_number, steps, create_step_func):
            return True
        
        # Handle print statements
        if self.print_handler.handle_print_statement(line, line_number, step_number, steps, create_step_func):
            return True
        
        return False
    
    def _handle_class_instantiation(self, line: str, line_number: int, step_number: int, 
                                   steps: List[ExecutionStepSchema], create_step_func) -> bool:
        """Handle class instantiation with better parsing"""
        instantiation_match = re.match(r"(\w+)\s*=\s*(\w+)\s*\(\s*\)", line)
        if instantiation_match:
            var_name = instantiation_match.group(1)
            class_name = instantiation_match.group(2)
            
            if class_name in ["Graph", "UndirectedGraph", "DirectedGraph"]:
                message = self.node_manager.create_instance(var_name, class_name)
                state = self._create_current_state()
                state["step_detail"] = {
                    "operation": "create_instance",
                    "instance_name": var_name,
                    "class_type": class_name,
                    "adjacency_list": {}
                }
                # Don't send state as additional_state to avoid error field pollution
                steps.append(create_step_func(step_number, line_number, line, message))
                return True
        return False
    
    def _handle_method_calls(self, line: str, line_number: int, step_number: int, 
                            steps: List[ExecutionStepSchema], create_step_func) -> bool:
        """Handle method calls with instance tracking"""
        method_match = re.match(r"(\w+)\.(\w+)\s*\((.*?)\)", line)
        if method_match:
            instance_name = method_match.group(1)
            method_name = method_match.group(2)
            params = method_match.group(3).strip()
            
            if instance_name in self.context["instances"]:
                instance = self.context["instances"][instance_name]
                result = self.method_executor.execute_method(instance, instance_name, method_name, params)
                state = self._create_current_state()
                # Add step detail without error field
                step_detail = {k: v for k, v in result.items() if k != "error"}
                state["step_detail"] = step_detail
                # Don't send state as additional_state to avoid error field pollution
                steps.append(create_step_func(step_number, line_number, line, result["message"]))
                return True
        return False
    
    def _handle_assignment_from_method(self, line: str, line_number: int, step_number: int, 
                                      steps: List[ExecutionStepSchema], create_step_func) -> bool:
        """Handle variable assignments from method calls"""
        assignment_match = re.match(r"(\w+)\s*=\s*(\w+)\.(\w+)\s*\((.*?)\)", line)
        if assignment_match:
            var_name = assignment_match.group(1)
            instance_name = assignment_match.group(2)
            method_name = assignment_match.group(3)
            params = assignment_match.group(4).strip()
            
            if instance_name in self.context["instances"]:
                instance = self.context["instances"][instance_name]
                
                # Handle methods that return values
                if method_name in ["bfs", "dfs"]:
                    result = self.method_executor.execute_method(instance, instance_name, method_name, params)
                    if result.get("traversal_result") is not None:
                        self.context["variables"][var_name] = result["traversal_result"]
                        result["message"] = f"Assigned {var_name} = {instance_name}.{method_name}({params}) → {result['traversal_result']}"
                        result["assignment"] = {"variable": var_name, "value": result["traversal_result"]}
                    
                    state = self._create_current_state()
                    # Add step detail without error field
                    step_detail = {k: v for k, v in result.items() if k != "error"}
                    state["step_detail"] = step_detail
                    # Don't send state as additional_state to avoid error field pollution
                    steps.append(create_step_func(step_number, line_number, line, result["message"]))
                    return True
        return False
    
    def _handle_simple_assignment(self, line: str, line_number: int, step_number: int, 
                                 steps: List[ExecutionStepSchema], create_step_func) -> bool:
        """Handle simple variable assignments"""
        simple_assignment_match = re.match(r"(\w+)\s*=\s*(.+)", line)
        if simple_assignment_match and not line.count('.') and not line.count('('):
            var_name = simple_assignment_match.group(1)
            value_str = simple_assignment_match.group(2).strip()
            
            # Handle string literals
            if ((value_str.startswith('"') and value_str.endswith('"')) or 
                (value_str.startswith("'") and value_str.endswith("'"))):
                self.context["variables"][var_name] = value_str[1:-1]
                message = f"Assigned {var_name} = {value_str}"
                state = self._create_current_state()
                state["step_detail"] = {
                    "operation": "assignment",
                    "variable": var_name,
                    "value": value_str[1:-1],
                    "type": "string"
                }
                # Don't send state as additional_state to avoid error field pollution
                steps.append(create_step_func(step_number, line_number, line, message))
                return True
            
            # Handle numeric literals
            try:
                if '.' in value_str:
                    value = float(value_str)
                else:
                    value = int(value_str)
                self.context["variables"][var_name] = value
                message = f"Assigned {var_name} = {value}"
                state = self._create_current_state()
                state["step_detail"] = {
                    "operation": "assignment",
                    "variable": var_name,
                    "value": value,
                    "type": "number"
                }
                # Don't send state as additional_state to avoid error field pollution
                steps.append(create_step_func(step_number, line_number, line, message))
                return True
            except ValueError:
                pass
        return False
    
    def _create_current_state(self) -> Dict[str, Any]:
        """Create current state snapshot for step tracking"""
        state = {
            "instances": {},
            "variables": self.context["variables"].copy(),
            "stdout": self.context["stdout"].copy(),
            "active": self.context.get("active_instance")
        }
        
        # Add detailed instance states
        for name, instance in self.context["instances"].items():
            class_type = instance.get("class_type")
            if class_type in ["Graph", "UndirectedGraph", "DirectedGraph"]:
                adjacency_list = instance.get("adjacency_list", {})
                # For directed graphs, don't divide by 2
                edge_count = sum(len(neighbors) for neighbors in adjacency_list.values())
                if class_type in ["Graph", "UndirectedGraph"]:
                    edge_count = edge_count // 2
                
                state["instances"][name] = {
                    "type": class_type,
                    "adjacency_list": adjacency_list,
                    "vertices": list(adjacency_list.keys()),
                    "vertex_count": len(adjacency_list),
                    "edge_count": edge_count,
                    "is_empty": len(adjacency_list) == 0
                }
        
        return state