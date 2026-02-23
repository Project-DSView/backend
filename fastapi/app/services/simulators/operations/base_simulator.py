"""
Base simulator functionality for DSView Backend API.

This module provides base functionality for code simulators.

DEPRECATED: This module is deprecated. Use app.services.simulators.common.base_simulator instead.
"""

from app.schemas.playground import ExecutionStepSchema


class BaseSimulator:
    """Base class for code simulators with common functionality"""
    
    def __init__(self):
        self.data_structure_type = None  # Set this in subclasses
        self.reset_context()
    
    def reset_context(self):
        """Reset the simulator context with proper initialization"""
        self.context = {
            "classes": {}, 
            "instances": {},
            "variables": {},
            "active_instance": None,
            "nodes": {},
            "stdout": []
        }
        
        # Add linkedlist-specific context for linkedlist simulators
        if self.data_structure_type == "linkedlist":
            self.context["linkedlist"] = []
            self.context["include_linkedlist"] = True
    
    def _create_execution_step(self, step_number: int, line_number: int, code: str, 
                             message: str = None, error: str = None, additional_state: dict = None) -> ExecutionStepSchema:
        """Create a standardized execution step"""
        # Ensure context is properly structured
        self._ensure_context_structure()
        
        # Build base state safely
        state = {
            "instances": {},
            "active": self.context.get("active_instance"),
            "stdout": self.context.get("stdout", []).copy()
        }
        
        # Safely add instances with detailed information
        instances = self.context.get("instances", {})
        if isinstance(instances, dict):
            for k, v in instances.items():
                try:
                    instance_data = self._get_instance_display(v)
                    state["instances"][k] = instance_data
                    
                    # Add detailed instance state information
                    if isinstance(v, dict):
                        state[f"{k}_count"] = v.get("count", 0)
                        state[f"{k}_head"] = v.get("head")
                        state[f"{k}_class_type"] = v.get("class_type", "Unknown")
                except Exception:
                    state["instances"][k] = []
        
        # Only add linkedlist if this is a linkedlist simulator
        if self.data_structure_type == "linkedlist":
            linkedlist = self.context.get("linkedlist", [])
            if isinstance(linkedlist, list):
                state["linkedlist"] = linkedlist.copy()
            else:
                state["linkedlist"] = []
        
        # Safely add nodes for linkedlist simulator with enhanced information
        if self.data_structure_type == "linkedlist":
            nodes = self.context.get("nodes", {})
            if isinstance(nodes, dict):
                state["nodes"] = {}
                for k, v in nodes.items():
                    try:
                        if isinstance(v, dict):
                            state["nodes"][k] = {
                                "name": v.get("name", ""),
                                "next": v.get("next"),
                                "id": k
                            }
                        else:
                            state["nodes"][k] = {"name": str(v) if v is not None else "", "next": None, "id": k}
                    except Exception:
                        state["nodes"][k] = {"name": "", "next": None, "id": k}
        
        # Add variables safely
        variables = self.context.get("variables", {})
        if isinstance(variables, dict) and variables:
            safe_vars = {}
            for k, v in variables.items():
                try:
                    if isinstance(v, (str, int, float, bool)) or v is None:
                        safe_vars[k] = v
                    else:
                        safe_vars[k] = str(v)
                except Exception:
                    safe_vars[k] = "undefined"
            if safe_vars:
                state["variables"] = safe_vars
        
        # Add any additional state
        if additional_state and isinstance(additional_state, dict):
            state.update(additional_state)
            
        if message:
            state["message"] = message
        if error:
            state["error"] = error
            
        return ExecutionStepSchema(
            stepNumber=step_number,
            line=line_number,
            code=code,
            state=state
        )
    
    def _ensure_context_structure(self):
        """Ensure context has proper structure"""
        if not isinstance(self.context, dict):
            self.context = {}
        
        required_keys = {
            "classes": {},
            "instances": {},
            "variables": {},
            "nodes": {},
            "stdout": []
        }
        
        for key, default_value in required_keys.items():
            if key not in self.context:
                self.context[key] = default_value
            elif not isinstance(self.context[key], type(default_value)):
                self.context[key] = default_value
        
        # Special handling for linkedlist simulators
        if self.data_structure_type in ["singlylinkedlist", "doublylinkedlist"]:
            list_key = self.data_structure_type
            if list_key not in self.context or not isinstance(self.context[list_key], list):
                self.context[list_key] = []
            self.context["include_linkedlist"] = True
    
    def _get_instance_display(self, instance):
        """Get display representation of an instance - to be overridden by subclasses"""
        if not isinstance(instance, dict):
            return []
            
        class_type = instance.get("class_type")
        if class_type in ["SinglyLinkedList", "DoublyLinkedList"]:
            return self._traverse_linked_list(instance)
        elif class_type in ["ArrayStack", "Queue"]:
            data = instance.get("data", [])
            return data if isinstance(data, list) else []
        
        # Fallback
        data = instance.get("data", [])
        return data if isinstance(data, list) else []
    
    def _traverse_linked_list(self, instance):
        """Traverse a linked list instance and return display data"""
        if not isinstance(instance, dict):
            return []
            
        result = []
        head_id = instance.get("head")
        if head_id is None:
            return []
        
        visited = set()  # Prevent infinite loops
        nodes = self.context.get("nodes", {})
        
        if not isinstance(nodes, dict):
            return []
        
        current_node_id = head_id
        while current_node_id is not None and current_node_id not in visited:
            visited.add(current_node_id)
            
            if current_node_id not in nodes:
                break
                
            node = nodes[current_node_id]
            if not isinstance(node, dict):
                break
                
            node_name = node.get("name", "")
            result.append(str(node_name))
            
            current_node_id = node.get("next")
                
        return result