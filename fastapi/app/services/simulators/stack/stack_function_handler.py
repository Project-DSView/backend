from typing import Dict, Any


class StackFunctionHandler:
    """Handles special stack-related functions like copyStack and utility functions"""
    
    def __init__(self, context: Dict[str, Any]):
        self.context = context
    
    def simulate_copy_stack(self, s1_name: str, s2_name: str) -> Dict[str, Any]:
        """Simulate the copyStack function with detailed tracking"""
        if s1_name not in self.context["instances"] or s2_name not in self.context["instances"]:
            return {"operation": "copyStack", "error": "instances_not_found"}
        
        s1 = self.context["instances"][s1_name]
        s2 = self.context["instances"][s2_name]
        
        # Record original states
        original_s1 = s1["data"].copy()
        original_s2 = s2["data"].copy()
        
        # Simply copy s1's data to s2 (preserve s1, copy to s2)
        s2["data"] = s1["data"].copy()
        
        # Record in history
        s1["history"].append({
            "operation": "copyStack",
            "before": original_s1,
            "after": s1["data"].copy()
        })
        s2["history"].append({
            "operation": "copyStack", 
            "before": original_s2,
            "after": s2["data"].copy()
        })
        
        return {
            "operation": "copyStack",
            "source": s1_name,
            "destination": s2_name,
            "source_before": original_s1,
            "source_after": s1["data"].copy(),
            "destination_before": original_s2,
            "destination_after": s2["data"].copy()
        }
    
    def simulate_function_call(self, func_name: str, args: str) -> Any:
        """Simulate function calls"""
        function_map = {
            "is_parentheses_matching": self._handle_parentheses_matching,
            "infixToPostfix": self._handle_infix_to_postfix,
            "copyStack": self._handle_copy_stack_call
        }
        
        if func_name in function_map:
            return function_map[func_name](args)
        
        return None
    
    def _handle_parentheses_matching(self, args: str) -> bool:
        """Handle parentheses matching function"""
        arg = args.strip().strip('"\'')
        if arg in self.context["variables"]:
            arg = self.context["variables"][arg]
        
        # Simplified parentheses matching
        stack_count = 0
        for char in arg:
            if char == '(':
                stack_count += 1
            elif char == ')':
                if stack_count > 0:
                    stack_count -= 1
                else:
                    stack_count += 1
        
        result = stack_count == 0
        if not result:
            self.context["stdout"].append(f"Parentheses in {arg} are unmatched")
        return result
    
    def _handle_infix_to_postfix(self, args: str) -> str:
        """Handle infix to postfix conversion"""
        arg = args.strip().strip('"\'')
        if arg in self.context["variables"]:
            arg = self.context["variables"][arg]
        return f"{arg}_postfix"
    
    def _handle_copy_stack_call(self, args: str) -> Any:
        """Handle copyStack function call"""
        arg_parts = [arg.strip() for arg in args.split(',')]
        if len(arg_parts) == 2:
            s1_name, s2_name = arg_parts
            if s1_name in self.context["instances"] and s2_name in self.context["instances"]:
                return self.simulate_copy_stack(s1_name, s2_name)
        return None