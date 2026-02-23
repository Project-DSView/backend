import re
from typing import List, Dict, Any
from app.schemas.playground import ExecutionStepSchema
from app.services.simulators.binarysearchtree.bst_method_executor import BSTMethodExecutor
from app.services.simulators.binarysearchtree.bst_node_manager import BSTNodeManager


class BSTStatementParser:
    """Handles parsing and execution of individual BST statements"""
    
    def __init__(self, context: Dict[str, Any], print_handler):
        self.context = context
        self.print_handler = print_handler
        self.method_executor = BSTMethodExecutor(context)
        self.node_manager = BSTNodeManager(context)
    
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
        instantiation_match = re.match(r"(\w+)\s*=\s*(\w+)\s*\(\s*([^)]*)\s*\)", line)
        if instantiation_match:
            var_name = instantiation_match.group(1)
            class_name = instantiation_match.group(2)
            params = instantiation_match.group(3).strip()
            
            if class_name == "BST":
                message = self.node_manager.create_instance(var_name, class_name)
                state = self._create_current_state()
                state["step_detail"] = {
                    "operation": "create_instance",
                    "instance_name": var_name,
                    "class_type": class_name,
                    "root": None
                }
                steps.append(create_step_func(step_number, line_number, line, message, state))
                return True
            elif class_name == "BSTNode" and params:
                # Parse the parameter (data value)
                data_value = self._parse_parameter(params)
                message = self.node_manager.create_node_instance(var_name, data_value)
                state = self._create_current_state()
                state["step_detail"] = {
                    "operation": "create_node",
                    "instance_name": var_name,
                    "class_type": class_name,
                    "data": data_value
                }
                steps.append(create_step_func(step_number, line_number, line, message, state))
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
                state["step_detail"] = result
                steps.append(create_step_func(step_number, line_number, line, result["message"], state))
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
                if method_name in ["delete", "findMin", "findMax", "is_empty"]:
                    result = self.method_executor.execute_method(instance, instance_name, method_name, params)
                    if result.get("value") is not None:
                        self.context["variables"][var_name] = result["value"]
                        result["message"] = f"Assigned {var_name} = {instance_name}.{method_name}({params}) → {result['value']}"
                        result["assignment"] = {"variable": var_name, "value": result["value"]}
                    
                    state = self._create_current_state()
                    state["step_detail"] = result
                    steps.append(create_step_func(step_number, line_number, line, result["message"], state))
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
                steps.append(create_step_func(step_number, line_number, line, message, state))
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
                steps.append(create_step_func(step_number, line_number, line, message, state))
                return True
            except ValueError:
                pass
        return False
    
    def _parse_parameter(self, params: str) -> Any:
        """Parse method parameters"""
        params = params.strip()
        
        # String literal
        if (params.startswith('"') and params.endswith('"')) or \
           (params.startswith("'") and params.endswith("'")):
            return params[1:-1]
        
        # Number
        try:
            if '.' in params:
                return float(params)
            else:
                return int(params)
        except ValueError:
            pass
        
        # Variable reference
        if params in self.context["variables"]:
            return self.context["variables"][params]
        
        return params
    
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
            if instance.get("class_type") == "BST":
                state["instances"][name] = {
                    "type": "BST",
                    "root": self._serialize_tree_node(instance.get("root")),
                    "isEmpty": instance.get("root") is None,
                    "size": self._count_nodes(instance.get("root")),
                    "height": self._calculate_height(instance.get("root"))
                }
            elif instance.get("class_type") == "BSTNode":
                state["instances"][name] = {
                    "type": "BSTNode",
                    "data": instance.get("data"),
                    "left": self._serialize_tree_node(instance.get("left")),
                    "right": self._serialize_tree_node(instance.get("right"))
                }
        
        return state
    
    def _serialize_tree_node(self, node):
        """Serialize a tree node for display"""
        if node is None:
            return None
        return {
            "data": node.get("data"),
            "left": self._serialize_tree_node(node.get("left")),
            "right": self._serialize_tree_node(node.get("right"))
        }
    
    def _count_nodes(self, node):
        """Count total nodes in the tree"""
        if node is None:
            return 0
        return 1 + self._count_nodes(node.get("left")) + self._count_nodes(node.get("right"))
    
    def _calculate_height(self, node):
        """Calculate height of the tree"""
        if node is None:
            return 0
        left_height = self._calculate_height(node.get("left"))
        right_height = self._calculate_height(node.get("right"))
        return 1 + max(left_height, right_height)