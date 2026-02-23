import re
from typing import List, Dict, Any
from app.schemas.playground import ExecutionStepSchema
from app.services.simulators.queue.queue_method_executor import QueueMethodExecutor
from app.services.simulators.queue.queue_function_handler import QueueFunctionHandler
from app.services.simulators.queue.queue_node_manager import QueueNodeManager


class QueueStatementParser:
    """Handles parsing and execution of individual statements"""
    
    def __init__(self, context: Dict[str, Any], print_handler):
        self.context = context
        self.print_handler = print_handler
        self.method_executor = QueueMethodExecutor(context)
        self.function_handler = QueueFunctionHandler(context)
        self.node_manager = QueueNodeManager(context)
    
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
        
        # Handle function calls
        if self._handle_function_calls(line, line_number, step_number, steps, create_step_func):
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
            
            if class_name == "ArrayQueue":
                message = self.node_manager.create_instance(var_name, class_name)
                state = self._create_current_state()
                state["step_detail"] = {
                    "operation": "create_instance",
                    "instance_name": var_name,
                    "class_type": class_name,
                    "data": []
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
                
                # Special handling for printQueue method to show print output
                if method_name == "printQueue" and "stdout" in result:
                    # Create a print step similar to PrintHandler
                    print_state = self._create_current_state()
                    print_state["step_detail"] = {
                        "operation": "print",
                        "content": f"{instance_name}.{method_name}()",
                        "output": result["stdout"],
                        "method_call": True
                    }
                    steps.append(create_step_func(step_number, line_number, line, f"Print: {result['stdout']}", print_state))
                else:
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
                
                if method_name == "dequeue" and instance.get("class_type") == "ArrayQueue":
                    result = self.node_manager.queue_dequeue(instance, instance_name)
                    if result.get("value") is not None:
                        self.context["variables"][var_name] = result["value"]
                        result["message"] = f"Assigned {var_name} = {instance_name}.dequeue() → {result['value']}"
                        result["assignment"] = {"variable": var_name, "value": result["value"]}
                    
                    state = self._create_current_state()
                    state["step_detail"] = result
                    steps.append(create_step_func(step_number, line_number, line, result["message"], state))
                    return True
                
                # Handle front() method assignment
                if method_name == "front" and instance.get("class_type") == "ArrayQueue":
                    result = self.node_manager.queue_front(instance, instance_name)
                    if result.get("value") is not None:
                        self.context["variables"][var_name] = result["value"]
                        result["message"] = f"Assigned {var_name} = {instance_name}.front() → {result['value']}"
                        result["assignment"] = {"variable": var_name, "value": result["value"]}
                    
                    state = self._create_current_state()
                    state["step_detail"] = result
                    steps.append(create_step_func(step_number, line_number, line, result["message"], state))
                    return True
                
                # Handle back() method assignment
                if method_name == "back" and instance.get("class_type") == "ArrayQueue":
                    result = self.node_manager.queue_back(instance, instance_name)
                    if result.get("value") is not None:
                        self.context["variables"][var_name] = result["value"]
                        result["message"] = f"Assigned {var_name} = {instance_name}.back() → {result['value']}"
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
        return False
    
    def _handle_function_calls(self, line: str, line_number: int, step_number: int, 
                              steps: List[ExecutionStepSchema], create_step_func) -> bool:
        """Handle function calls with better tracking"""
        
        # Function call with assignment
        assignment_func_match = re.match(r"(\w+)\s*=\s*(\w+)\((.*?)\)", line)
        if assignment_func_match:
            var_name = assignment_func_match.group(1)
            func_name = assignment_func_match.group(2)
            args = assignment_func_match.group(3)
            
            result = self.function_handler.simulate_function_call(func_name, args)
            if result is not None:
                self.context["variables"][var_name] = result
                message = f"Called {func_name}({args}) → assigned result to {var_name}"
                state = self._create_current_state()
                state["step_detail"] = {
                    "operation": "function_call",
                    "function_name": func_name,
                    "arguments": args,
                    "return_value": result,
                    "assignment": {"variable": var_name, "value": result}
                }
                steps.append(create_step_func(step_number, line_number, line, message, state))
                return True
        
        # Direct function call
        func_call_match = re.match(r"(\w+)\((.*?)\)", line)
        if func_call_match:
            func_name = func_call_match.group(1)
            args = func_call_match.group(2)
            
            # Skip print function calls - let PrintHandler handle them
            if func_name == "print":
                return False
            
            # Other function calls
            result = self.function_handler.simulate_function_call(func_name, args)
            message = f"Called {func_name}({args})"
            if result is not None:
                message += f" → returned {result}"
            
            state = self._create_current_state()
            state["step_detail"] = {
                "operation": "function_call",
                "function_name": func_name,
                "arguments": args,
                "return_value": result
            }
            steps.append(create_step_func(step_number, line_number, line, message, state))
            return True
        
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
            if instance.get("class_type") == "ArrayQueue":
                state["instances"][name] = {
                    "type": "ArrayQueue",
                    "data": instance["data"].copy(),
                    "size": len(instance["data"]),
                    "isEmpty": len(instance["data"]) == 0,
                    "front": instance["data"][0] if instance["data"] else None,
                    "back": instance["data"][-1] if instance["data"] else None
                }
        
        return state

