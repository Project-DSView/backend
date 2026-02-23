from typing import Dict, Any
from app.services.simulators.binarysearchtree.bst_node_manager import BSTNodeManager


class BSTMethodExecutor:
    """Handles execution of BST methods with enhanced tracking"""
    
    def __init__(self, context: Dict[str, Any]):
        self.context = context
        self.node_manager = BSTNodeManager(context)
    
    def execute_method(self, instance: Dict[str, Any], instance_name: str, 
                      method_name: str, params: str) -> Dict[str, Any]:
        """Execute BST methods with enhanced tracking"""
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
                         "insert": self._handle_insert,
                         "delete": self._handle_delete,
                         "is_empty": self._handle_is_empty,
                         "findMin": self._handle_find_min,
                         "findMax": self._handle_find_max,
                         "traverse": self._handle_traverse,
                         "preorder": self._handle_preorder,
                         "inorder": self._handle_inorder,
                         "postorder": self._handle_postorder
                     }
                     
                     if behavior_type in behavior_map:
                         return behavior_map[behavior_type](instance, instance_name, params)

        if instance.get("class_type") == "BST":
            method_map = {
                "insert": self._handle_insert,
                "delete": self._handle_delete,
                "is_empty": self._handle_is_empty,
                "findMin": self._handle_find_min,
                "findMax": self._handle_find_max,
                "traverse": self._handle_traverse,
                "preorder": self._handle_preorder,
                "inorder": self._handle_inorder,
                "postorder": self._handle_postorder
            }
        elif instance.get("class_type") == "BSTNode":
            return {
                "message": f"BSTNode instances don't have callable methods", 
                "operation": method_name,
                "error": "invalid_method_call"
            }
        else:
            return {
                "message": f"Unknown class type for method {method_name}", 
                "operation": method_name,
                "error": "unknown_class"
            }
        
        if method_name in method_map:
            return method_map[method_name](instance, instance_name, params)
        else:
            return {
                "message": f"Unknown method {method_name}", 
                "operation": method_name, 
                "error": "unknown_method"
            }
    
    def _handle_insert(self, instance: Dict[str, Any], instance_name: str, params: str) -> Dict[str, Any]:
        """Handle insert operation"""
        if params:
            value = self._parse_parameter(params)
            return self.node_manager.bst_insert(instance, value, instance_name)
        return {
            "message": "insert requires a parameter", 
            "operation": "insert", 
            "error": "missing_parameter"
        }
    
    def _handle_delete(self, instance: Dict[str, Any], instance_name: str, params: str) -> Dict[str, Any]:
        """Handle delete operation"""
        if params:
            value = self._parse_parameter(params)
            return self.node_manager.bst_delete(instance, value, instance_name)
        return {
            "message": "delete requires a parameter", 
            "operation": "delete", 
            "error": "missing_parameter"
        }
    
    def _handle_is_empty(self, instance: Dict[str, Any], instance_name: str, params: str) -> Dict[str, Any]:
        """Handle is_empty operation"""
        return self.node_manager.bst_is_empty(instance, instance_name)
    
    def _handle_find_min(self, instance: Dict[str, Any], instance_name: str, params: str) -> Dict[str, Any]:
        """Handle findMin operation"""
        return self.node_manager.bst_find_min(instance, instance_name)
    
    def _handle_find_max(self, instance: Dict[str, Any], instance_name: str, params: str) -> Dict[str, Any]:
        """Handle findMax operation"""
        return self.node_manager.bst_find_max(instance, instance_name)
    
    def _handle_traverse(self, instance: Dict[str, Any], instance_name: str, params: str) -> Dict[str, Any]:
        """Handle traverse operation"""
        return self.node_manager.bst_traverse(instance, instance_name)
    
    def _handle_preorder(self, instance: Dict[str, Any], instance_name: str, params: str) -> Dict[str, Any]:
        """Handle preorder traversal"""
        # Parse root parameter if provided
        root_node = instance.get("root")
        if params and params != instance_name + ".root":
            # Handle case where specific node is passed
            root_node = instance.get("root")  # For now, use tree root
        
        return self.node_manager.bst_preorder(instance, root_node, instance_name)
    
    def _handle_inorder(self, instance: Dict[str, Any], instance_name: str, params: str) -> Dict[str, Any]:
        """Handle inorder traversal"""
        # Parse root parameter if provided
        root_node = instance.get("root")
        if params and params != instance_name + ".root":
            # Handle case where specific node is passed
            root_node = instance.get("root")  # For now, use tree root
        
        return self.node_manager.bst_inorder(instance, root_node, instance_name)
    
    def _handle_postorder(self, instance: Dict[str, Any], instance_name: str, params: str) -> Dict[str, Any]:
        """Handle postorder traversal"""
        # Parse root parameter if provided
        root_node = instance.get("root")
        if params and params != instance_name + ".root":
            # Handle case where specific node is passed
            root_node = instance.get("root")  # For now, use tree root
        
        return self.node_manager.bst_postorder(instance, root_node, instance_name)
    
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