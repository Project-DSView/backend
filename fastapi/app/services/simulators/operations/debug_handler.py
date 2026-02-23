"""
Debug Handler - Parses and executes debug commands
Supports variable modification and inspection commands
"""
import re
from typing import Dict, Any, Optional, Tuple


class DebugHandler:
    """Handles debug commands for runtime variable modification and inspection"""
    
    def __init__(self, execution_context: Dict[str, Any]):
        """
        Initialize debug handler with execution context
        
        Args:
            execution_context: The execution context containing variables, instances, etc.
        """
        self.context = execution_context
    
    def parse_debug_command(self, command: str) -> Dict[str, Any]:
        """
        Parse a debug command string
        
        Args:
            command: Debug command string (e.g., "variable = value", "print(var)", "inspect(instance)")
            
        Returns:
            Dictionary with parsed command information
        """
        command = command.strip()
        
        # Assignment: variable = value
        assignment_match = re.match(r'^(\w+)\s*=\s*(.+)$', command)
        if assignment_match:
            var_name = assignment_match.group(1)
            value_expr = assignment_match.group(2).strip()
            return {
                "type": "assignment",
                "variable": var_name,
                "value": self._parse_value(value_expr),
                "original_command": command
            }
        
        # Print: print(variable)
        print_match = re.match(r'^print\s*\(\s*(\w+)\s*\)$', command)
        if print_match:
            var_name = print_match.group(1)
            return {
                "type": "print",
                "variable": var_name,
                "original_command": command
            }
        
        # Inspect: inspect(instance)
        inspect_match = re.match(r'^inspect\s*\(\s*(\w+)\s*\)$', command)
        if inspect_match:
            instance_name = inspect_match.group(1)
            return {
                "type": "inspect",
                "instance": instance_name,
                "original_command": command
            }
        
        # Method call: instance.method()
        method_call_match = re.match(r'^(\w+)\.(\w+)\s*\(\)$', command)
        if method_call_match:
            instance_name = method_call_match.group(1)
            method_name = method_call_match.group(2)
            return {
                "type": "method_call",
                "instance": instance_name,
                "method": method_name,
                "original_command": command
            }
        
        # Unknown command
        return {
            "type": "unknown",
            "original_command": command,
            "error": f"Unknown debug command: {command}"
        }
    
    def _parse_value(self, value_expr: str) -> Any:
        """
        Parse a value expression
        
        Args:
            value_expr: Value expression string
            
        Returns:
            Parsed value
        """
        value_expr = value_expr.strip()
        
        # String literal
        if value_expr.startswith('"') and value_expr.endswith('"'):
            return value_expr[1:-1]
        if value_expr.startswith("'") and value_expr.endswith("'"):
            return value_expr[1:-1]
        
        # Number
        try:
            if '.' in value_expr:
                return float(value_expr)
            else:
                return int(value_expr)
        except ValueError:
            pass
        
        # Boolean
        if value_expr.lower() == 'true':
            return True
        if value_expr.lower() == 'false':
            return False
        
        # None
        if value_expr.lower() == 'none':
            return None
        
        # Variable reference
        if value_expr in self.context.get("variables", {}):
            return self.context["variables"][value_expr]
        
        # Return as string if can't parse
        return value_expr
    
    def execute_debug_command(self, command: str) -> Dict[str, Any]:
        """
        Execute a debug command
        
        Args:
            command: Debug command string
            
        Returns:
            Dictionary with execution result
        """
        parsed = self.parse_debug_command(command)
        
        if parsed["type"] == "assignment":
            return self._execute_assignment(parsed)
        elif parsed["type"] == "print":
            return self._execute_print(parsed)
        elif parsed["type"] == "inspect":
            return self._execute_inspect(parsed)
        elif parsed["type"] == "method_call":
            return self._execute_method_call(parsed)
        else:
            return {
                "success": False,
                "error": parsed.get("error", "Unknown command type"),
                "command": parsed["original_command"]
            }
    
    def _execute_assignment(self, parsed: Dict[str, Any]) -> Dict[str, Any]:
        """Execute variable assignment"""
        var_name = parsed["variable"]
        value = parsed["value"]
        
        # Ensure variables dict exists
        if "variables" not in self.context:
            self.context["variables"] = {}
        
        # Store the value
        self.context["variables"][var_name] = value
        
        return {
            "success": True,
            "message": f"Variable '{var_name}' set to {value}",
            "variable": var_name,
            "value": value,
            "command": parsed["original_command"]
        }
    
    def _execute_print(self, parsed: Dict[str, Any]) -> Dict[str, Any]:
        """Execute print command"""
        var_name = parsed["variable"]
        variables = self.context.get("variables", {})
        
        if var_name not in variables:
            return {
                "success": False,
                "error": f"Variable '{var_name}' not found",
                "command": parsed["original_command"]
            }
        
        value = variables[var_name]
        return {
            "success": True,
            "message": f"{var_name} = {value}",
            "variable": var_name,
            "value": value,
            "command": parsed["original_command"]
        }
    
    def _execute_inspect(self, parsed: Dict[str, Any]) -> Dict[str, Any]:
        """Execute inspect command"""
        instance_name = parsed["instance"]
        instances = self.context.get("instances", {})
        
        if instance_name not in instances:
            return {
                "success": False,
                "error": f"Instance '{instance_name}' not found",
                "command": parsed["original_command"]
            }
        
        instance = instances[instance_name]
        return {
            "success": True,
            "message": f"Inspecting '{instance_name}'",
            "instance": instance_name,
            "data": instance,
            "command": parsed["original_command"]
        }
    
    def _execute_method_call(self, parsed: Dict[str, Any]) -> Dict[str, Any]:
        """Execute method call command"""
        instance_name = parsed["instance"]
        method_name = parsed["method"]
        instances = self.context.get("instances", {})
        
        if instance_name not in instances:
            return {
                "success": False,
                "error": f"Instance '{instance_name}' not found",
                "command": parsed["original_command"]
            }
        
        # For now, just return info about the method call
        # In a full implementation, this would actually call the method
        return {
            "success": True,
            "message": f"Method '{method_name}' called on '{instance_name}'",
            "instance": instance_name,
            "method": method_name,
            "command": parsed["original_command"]
        }
