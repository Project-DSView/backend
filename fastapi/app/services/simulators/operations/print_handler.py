import re
from typing import List, Dict, Any
from app.schemas.playground import ExecutionStepSchema


class PrintHandler:
    """Enhanced print statement handling with step details for stack visualization"""
    
    def __init__(self, context: Dict[str, Any]):
        self.context = context
    
    def handle_print_statement(self, line: str, line_number: int, step_number: int, 
                             steps: List[ExecutionStepSchema], create_step_func) -> bool:
        """Handle print statements with enhanced step details"""
        print_match = re.match(r"print\((.*)\)", line)
        if not print_match:
            return False
        
        print_content = print_match.group(1).strip()
        
        # Handle complex print statements with multiple arguments
        if ',' in print_content and not self._is_single_string_with_commas(print_content):
            output_parts = []
            args = self._parse_print_arguments(print_content)
            
            for arg in args:
                output_parts.append(self._evaluate_print_content(arg.strip()))
            
            output_value = ' '.join(str(part) for part in output_parts)
        else:
            output_value = self._evaluate_print_content(print_content)
        
        # Add to print output list
        self.context["stdout"].append(output_value)
        
        # Create step with detailed information
        state = self._create_current_state()
        state["step_detail"] = {
            "operation": "print",
            "content": print_content,
            "output": output_value,
            "arguments": self._parse_print_arguments(print_content) if ',' in print_content else [print_content]
        }
        
        steps.append(create_step_func(step_number, line_number, line, f"Print: {output_value}", state))
        return True
    
    def _is_single_string_with_commas(self, content: str) -> bool:
        """Check if the content is a single string that contains commas"""
        content = content.strip()
        return ((content.startswith('"') and content.endswith('"')) or 
                (content.startswith("'") and content.endswith("'")))
    
    def _parse_print_arguments(self, print_content: str) -> List[str]:
        """Parse print arguments, handling quoted strings properly"""
        args = []
        current_arg = ""
        in_quotes = False
        quote_char = None
        
        for char in print_content:
            if char in ['"', "'"] and not in_quotes:
                in_quotes = True
                quote_char = char
                current_arg += char
            elif char == quote_char and in_quotes:
                in_quotes = False
                quote_char = None
                current_arg += char
            elif char == ',' and not in_quotes:
                args.append(current_arg.strip())
                current_arg = ""
            else:
                current_arg += char
        
        if current_arg.strip():
            args.append(current_arg.strip())
        
        return args
    
    def _evaluate_print_content(self, print_content: str) -> str:
        """Evaluate the content of a print statement"""
        try:
            # Handle different types of print content
            if not print_content:
                return ""
            elif self._is_string_literal(print_content):
                return print_content[1:-1]
            elif print_content in self.context["variables"]:
                return self._format_variable_value(print_content)
            elif "." in print_content and "(" in print_content and ")" in print_content:
                # Handle method calls like newStack.is_empty()
                return self._evaluate_method_call(print_content)
            elif "." in print_content:
                return self._evaluate_attribute_access(print_content)
            else:
                # Try to evaluate as expression or return as-is
                try:
                    return str(eval(print_content, {"__builtins__": {}}, self.context["variables"]))
                except:
                    return print_content
        except Exception:
            return print_content
    
    def _is_string_literal(self, content: str) -> bool:
        """Check if content is a string literal"""
        return ((content.startswith('"') and content.endswith('"')) or 
                (content.startswith("'") and content.endswith("'")))
    
    def _format_variable_value(self, var_name: str) -> str:
        """Format a variable value for printing"""
        if var_name in self.context["variables"]:
            value = self.context["variables"][var_name]
            if isinstance(value, str):
                return value
            elif value is None:
                return "None"
            elif isinstance(value, bool):
                return "True" if value else "False"
            else:
                return str(value)
        return var_name
    
    def _evaluate_attribute_access(self, expression: str) -> str:
        """Evaluate attribute access expressions like mylist.head, pNew.name"""
        try:
            parts = expression.split('.')
            if len(parts) == 2:
                obj_name, attr_name = parts
                
                # Check if it's an instance attribute
                if obj_name in self.context["instances"]:
                    return self._evaluate_instance_attribute(obj_name, attr_name)
                
                # Check if it's a variable (node) attribute
                elif obj_name in self.context["variables"]:
                    return self._evaluate_variable_attribute(obj_name, attr_name)
            
            return expression
        except Exception:
            return expression
    
    def _evaluate_instance_attribute(self, obj_name: str, attr_name: str) -> str:
        """Evaluate instance attribute access"""
        instance = self.context["instances"][obj_name]
        
        if attr_name == "data":
            return str(instance.get("data", []))
        elif attr_name == "size":
            return str(len(instance.get("data", [])))
        elif attr_name == "isEmpty":
            return str(len(instance.get("data", [])) == 0)
        elif attr_name == "is_empty":
            return str(len(instance.get("data", [])) == 0)
        elif attr_name == "top":
            data = instance.get("data", [])
            return str(data[-1] if data else None)
        elif attr_name == "stackTop":
            data = instance.get("data", [])
            return str(data[-1] if data else None)
        
        return f"{obj_name}.{attr_name}"
    
    def _evaluate_variable_attribute(self, obj_name: str, attr_name: str) -> str:
        """Evaluate variable attribute access"""
        return f"{obj_name}.{attr_name}"
    
    def _evaluate_method_call(self, expression: str) -> str:
        """Evaluate method calls like newStack.is_empty()"""
        try:
            # Parse method call: obj.method()
            match = re.match(r"(\w+)\.(\w+)\(\)", expression)
            if match:
                obj_name = match.group(1)
                method_name = match.group(2)
                
                if obj_name in self.context["instances"]:
                    instance = self.context["instances"][obj_name]
                    
                    # [NEW] Check logic-based behavior first
                    class_type = instance.get("class_type")
                    behavior_type = "unknown"
                    
                    if class_type and "classes" in self.context and class_type in self.context["classes"]:
                        methods = self.context["classes"][class_type].get("methods", {})
                        if method_name in methods:
                            behavior_type = methods[method_name].get("behavior_type")
                    
                    # Logic-based evaluation
                    if behavior_type == "size" or method_name in ["size", "get_size", "count", "__len__"]:
                         return str(len(instance.get("data", [])))
                    elif behavior_type == "is_empty" or method_name in ["is_empty", "empty", "isempty"]:
                         return str(len(instance.get("data", [])) == 0)
                    elif behavior_type == "stackTop" or method_name in ["stackTop", "get_stack_top", "peek", "top", "get_top"]:
                         data = instance.get("data", [])
                         return str(data[-1] if data else None)
                    elif behavior_type == "pop" or method_name in ["pop", "remove", "delete"]:
                         if instance.get("data"):
                             val = instance["data"].pop()
                             return str(val)
                         return "None"
                    
                    # Fallback for explicit method names (legacy)
                    if instance.get("class_type") == "ArrayStack":
                        if method_name == "size":
                            return str(len(instance.get("data", [])))
                        elif method_name == "is_empty":
                            return str(len(instance.get("data", [])) == 0)
                        elif method_name == "stackTop":
                            data = instance.get("data", [])
                            return str(data[-1] if data else None)
            
            return expression
        except Exception:
            return expression
    
    def _create_current_state(self) -> Dict[str, Any]:
        """Create current state snapshot"""
        state = {
            "instances": {},
            "variables": self.context["variables"].copy(),
            "variables": self.context["variables"].copy(),
            "stdout": self.context.get("stdout", []).copy(),
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