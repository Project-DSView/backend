import re
from typing import List, Dict, Any
from app.schemas.playground import ExecutionStepSchema
from app.services.simulators.stack.stack_method_executor import StackMethodExecutor
from app.services.simulators.stack.stack_function_handler import StackFunctionHandler
from app.services.simulators.stack.stack_node_manager import StackNodeManager


class StackStatementParser:
    """Handles parsing and execution of individual statements"""
    
    def __init__(self, context: Dict[str, Any], print_handler):
        self.context = context
        self.print_handler = print_handler
        self.method_executor = StackMethodExecutor(context)
        self.function_handler = StackFunctionHandler(context)
        self.node_manager = StackNodeManager(context)
    
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
            
            if class_name == "ArrayStack":
                message = self.node_manager.create_instance(var_name, class_name)
                state = self._create_current_state()
                state["step_detail"] = {
                    "operation": "create_instance",
                    "instance_name": var_name,
                    "class_type": class_name,
                    "data": []
                }
                # Add structured data to state
                state["explanation"] = f"Create new {class_name} instance '{var_name}'"
                state["operation"] = "create_instance"
                state["stack"] = []
                state["value"] = None
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
                
                # Special handling for printStack method to show print output
                if method_name == "printStack" and "stdout" in result:
                    # Create a print step similar to PrintHandler
                    print_state = self._create_current_state()
                    print_state["step_detail"] = {
                        "operation": "print",
                        "content": f"{instance_name}.{method_name}()",
                        "output": result["stdout"],
                        "method_call": True
                    }
                    # Promote fields
                    if "explanation" in result:
                        print_state["explanation"] = result["explanation"]
                    if "operation" in result:
                        print_state["operation"] = result["operation"]
                    if "value" in result:
                        print_state["value"] = result["value"]
                    if "after_data" in result:
                        print_state["stack"] = result["after_data"]
                    elif instance.get("class_type") == "ArrayStack":
                        print_state["stack"] = instance.get("data", [])
                        
                    steps.append(create_step_func(step_number, line_number, line, f"Print: {result['stdout']}", print_state))
                else:
                    # Promote fields
                    if "explanation" in result:
                        state["explanation"] = result["explanation"]
                    if "operation" in result:
                        state["operation"] = result["operation"]
                    if "value" in result:
                        state["value"] = result["value"]
                    if "after_data" in result:
                        state["stack"] = result["after_data"]
                    elif instance.get("class_type") == "ArrayStack":
                        state["stack"] = instance.get("data", [])
                        
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
                
                if method_name == "pop" and instance.get("class_type") == "ArrayStack":
                    result = self.node_manager.stack_pop(instance, instance_name)
                    if result.get("value") is not None:
                        self.context["variables"][var_name] = result["value"]
                        result["message"] = f"Assigned {var_name} = {instance_name}.pop() → {result['value']}"
                        result["assignment"] = {"variable": var_name, "value": result["value"]}
                    
                    state = self._create_current_state()
                    state["step_detail"] = result
                    
                    # Promote fields
                    if "explanation" in result:
                        state["explanation"] = result["explanation"]
                    if "operation" in result:
                        state["operation"] = result["operation"]
                    if "value" in result:
                        state["value"] = result["value"]
                    if "after_data" in result:
                        state["stack"] = result["after_data"]
                    elif instance.get("class_type") == "ArrayStack":
                        state["stack"] = instance.get("data", [])
                        
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
            
            if func_name == "copyStack":
                arg_parts = [arg.strip() for arg in args.split(',')]
                if len(arg_parts) == 2:
                    s1_name, s2_name = arg_parts
                    if s1_name in self.context["instances"] and s2_name in self.context["instances"]:
                        copy_result = self.function_handler.simulate_copy_stack(s1_name, s2_name)
                        message = f"Called copyStack({s1_name}, {s2_name}) → copied {s1_name} to {s2_name}"
                        state = self._create_current_state()
                        state["step_detail"] = copy_result
                        steps.append(create_step_func(step_number, line_number, line, message, state))
                        return True
            
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
            if instance.get("class_type") == "ArrayStack":
                state["instances"][name] = {
                    "type": "ArrayStack",
                    "data": instance["data"].copy(),
                    "size": len(instance["data"]),
                    "isEmpty": len(instance["data"]) == 0,
                    "top": instance["data"][-1] if instance["data"] else None
                }
        
        return state