from typing import Dict, Any
from app.services.simulators.queue.queue_node_manager import QueueNodeManager


class QueueMethodExecutor:
    """Handles execution of queue methods with enhanced tracking"""
    
    def __init__(self, context: Dict[str, Any]):
        self.context = context
        self.node_manager = QueueNodeManager(context)
    
    def execute_method(self, instance: Dict[str, Any], instance_name: str, 
                      method_name: str, params: str) -> Dict[str, Any]:
        """Execute ArrayQueue methods with enhanced tracking"""
        if instance.get("class_type") != "ArrayQueue":
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
                         "enqueue": self._handle_enqueue,
                         "dequeue": self._handle_dequeue,
                         "size": self._handle_size,
                         "is_empty": self._handle_is_empty,
                         "front": self._handle_front,
                         "back": self._handle_back,
                         "printQueue": self._handle_print_queue
                     }
                     
                     if behavior_type in behavior_map:
                         return behavior_map[behavior_type](instance, instance_name, params)

        method_map = {
            "enqueue": self._handle_enqueue,
            "add": self._handle_enqueue,
            "push": self._handle_enqueue,
            "insert": self._handle_enqueue,
            
            "dequeue": self._handle_dequeue,
            "pop": self._handle_dequeue,
            "remove": self._handle_dequeue,
            "delete": self._handle_dequeue,
            
            "size": self._handle_size,
            "get_size": self._handle_size,
            "count": self._handle_size,
            "length": self._handle_size,
            "__len__": self._handle_size,
            
            "is_empty": self._handle_is_empty,
            "empty": self._handle_is_empty,
            
            "front": self._handle_front,
            "peek": self._handle_front,
            "get_front": self._handle_front,
            "head": self._handle_front,
            
            "back": self._handle_back,
            "rear": self._handle_back,
            "tail": self._handle_back,
            "get_back": self._handle_back,
            
            "printQueue": self._handle_print_queue,
            "print_queue": self._handle_print_queue,
            "display": self._handle_print_queue, 
            "show": self._handle_print_queue
        }
        
        if method_name in method_map:
            return method_map[method_name](instance, instance_name, params)
        else:
            return {
                "message": f"Unknown method {method_name}", 
                "operation": method_name, 
                "error": "unknown_method"
            }
    
    def _handle_enqueue(self, instance: Dict[str, Any], instance_name: str, params: str) -> Dict[str, Any]:
        """Handle enqueue operation"""
        if params:
            value = self._parse_parameter(params)
            return self.node_manager.queue_enqueue(instance, value, instance_name)
        return {
            "message": "enqueue requires a parameter", 
            "operation": "enqueue", 
            "error": "missing_parameter"
        }
    
    def _handle_dequeue(self, instance: Dict[str, Any], instance_name: str, params: str) -> Dict[str, Any]:
        """Handle dequeue operation"""
        return self.node_manager.queue_dequeue(instance, instance_name)
    
    def _handle_size(self, instance: Dict[str, Any], instance_name: str, params: str) -> Dict[str, Any]:
        """Handle size operation"""
        return self.node_manager.queue_size(instance, instance_name)
    
    def _handle_is_empty(self, instance: Dict[str, Any], instance_name: str, params: str) -> Dict[str, Any]:
        """Handle is_empty operation"""
        return self.node_manager.queue_is_empty(instance, instance_name)
    
    def _handle_front(self, instance: Dict[str, Any], instance_name: str, params: str) -> Dict[str, Any]:
        """Handle front operation"""
        return self.node_manager.queue_front(instance, instance_name)
    
    def _handle_back(self, instance: Dict[str, Any], instance_name: str, params: str) -> Dict[str, Any]:
        """Handle back operation"""
        return self.node_manager.queue_back(instance, instance_name)
    
    def _handle_print_queue(self, instance: Dict[str, Any], instance_name: str, params: str) -> Dict[str, Any]:
        """Handle printQueue operation"""
        return self.node_manager.queue_print(instance, instance_name)
    
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

