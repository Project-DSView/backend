from typing import Dict, Any
from app.services.simulators.stack.stack_node_manager import StackNodeManager


class StackMethodExecutor:
    """Handles execution of stack methods with enhanced tracking"""
    
    def __init__(self, context: Dict[str, Any]):
        self.context = context
        self.node_manager = StackNodeManager(context)
    
    def execute_method(self, instance: Dict[str, Any], instance_name: str, 
                      method_name: str, params: str) -> Dict[str, Any]:
        """Execute ArrayStack methods with enhanced tracking"""
        if instance.get("class_type") != "ArrayStack":
            return {"message": f"Unknown method {method_name}", "operation": method_name}
        
        # Check available behaviors from context first
        if "classes" in self.context and isinstance(self.context["classes"], dict):
            # Find the class and its method behaviors
            class_type = instance.get("class_type")
            if class_type and class_type in self.context["classes"]:
                 methods = self.context["classes"][class_type].get("methods", {})
                 if isinstance(methods, dict) and method_name in methods:
                     method_info = methods[method_name]
                     behavior_type = method_info.get("behavior_type")
                     
                     # Map behavior types to handlers
                     behavior_map = {
                         "push": self._handle_push,
                         "pop": self._handle_pop,
                         "size": self._handle_size,
                         "handler_size": self._handle_size, # behavior analyzer might return this
                         "is_empty": self._handle_is_empty,
                         "stackTop": self._handle_stack_top,
                         "peek": self._handle_stack_top,
                         "printStack": self._handle_print_stack,
                         "print": self._handle_print_stack
                     }
                     
                     if behavior_type in behavior_map:
                         return behavior_map[behavior_type](instance, instance_name, params)

        method_map = {
            # Push variants
            "push": self._handle_push,
            "add": self._handle_push,
            "insert": self._handle_push,
            
            # Pop variants
            "pop": self._handle_pop,
            "remove": self._handle_pop,
            "delete": self._handle_pop,
            
            # Size variants
            "size": self._handle_size,
            "get_size": self._handle_size,
            "count": self._handle_size,
            "length": self._handle_size,
            "__len__": self._handle_size,
            
            # Empty check variants
            "is_empty": self._handle_is_empty,
            "empty": self._handle_is_empty,
            "isempty": self._handle_is_empty,
            
            # Top/Peek variants
            "stackTop": self._handle_stack_top,
            "get_stack_top": self._handle_stack_top,
            "peek": self._handle_stack_top,
            "top": self._handle_stack_top,
            "get_top": self._handle_stack_top,
            
            # Print variants
            "printStack": self._handle_print_stack,
            "print_stack": self._handle_print_stack,
            "display": self._handle_print_stack,
            "show": self._handle_print_stack,
            "__str__": self._handle_print_stack,
            "__repr__": self._handle_print_stack
        }
        
        if method_name in method_map:
            return method_map[method_name](instance, instance_name, params)
        else:
            return {
                "message": f"Unknown method {method_name}", 
                "operation": method_name, 
                "error": "unknown_method"
            }
    
    def _handle_push(self, instance: Dict[str, Any], instance_name: str, params: str) -> Dict[str, Any]:
        """Handle push operation"""
        if params:
            value = self._parse_parameter(params)
            return self.node_manager.stack_push(instance, value, instance_name)
        return {
            "message": "push requires a parameter", 
            "operation": "push", 
            "error": "missing_parameter"
        }
    
    def _handle_pop(self, instance: Dict[str, Any], instance_name: str, params: str) -> Dict[str, Any]:
        """Handle pop operation"""
        return self.node_manager.stack_pop(instance, instance_name)
    
    def _handle_size(self, instance: Dict[str, Any], instance_name: str, params: str) -> Dict[str, Any]:
        """Handle size operation"""
        return self.node_manager.stack_size(instance, instance_name)
    
    def _handle_is_empty(self, instance: Dict[str, Any], instance_name: str, params: str) -> Dict[str, Any]:
        """Handle is_empty operation"""
        return self.node_manager.stack_is_empty(instance, instance_name)
    
    def _handle_stack_top(self, instance: Dict[str, Any], instance_name: str, params: str) -> Dict[str, Any]:
        """Handle stackTop operation"""
        return self.node_manager.stack_top(instance, instance_name)
    
    def _handle_print_stack(self, instance: Dict[str, Any], instance_name: str, params: str) -> Dict[str, Any]:
        """Handle printStack operation"""
        return self.node_manager.stack_print(instance, instance_name)
    
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