from typing import Dict, Any


class QueueFunctionHandler:
    """Handles special queue-related functions and utility functions"""
    
    def __init__(self, context: Dict[str, Any]):
        self.context = context
    
    def simulate_function_call(self, func_name: str, args: str) -> Any:
        """Simulate function calls"""
        function_map = {
            "reverse_queue": self._handle_reverse_queue,
        }
        
        if func_name in function_map:
            return function_map[func_name](args)
        
        return None
    
    def _handle_reverse_queue(self, args: str) -> Any:
        """Handle reverse queue function"""
        arg_parts = [arg.strip() for arg in args.split(',')]
        if len(arg_parts) == 1:
            queue_name = arg_parts[0]
            if queue_name in self.context["instances"]:
                queue = self.context["instances"][queue_name]
                original_data = queue["data"].copy()
                queue["data"] = list(reversed(queue["data"]))
                return {
                    "operation": "reverse_queue",
                    "queue_name": queue_name,
                    "before": original_data,
                    "after": queue["data"].copy()
                }
        return None

