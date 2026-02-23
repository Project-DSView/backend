from typing import Dict, Any
from app.services.simulators.graph.graph_node_manager import GraphNodeManager


class GraphMethodExecutor:
    """Handles execution of Graph methods with enhanced tracking"""
    
    def __init__(self, context: Dict[str, Any]):
        self.context = context
        self.node_manager = GraphNodeManager(context)
    
    def execute_method(self, instance: Dict[str, Any], instance_name: str, 
                      method_name: str, params: str) -> Dict[str, Any]:
        """Execute Graph methods with enhanced tracking"""
        class_type = instance.get("class_type")
        if class_type not in ["Graph", "UndirectedGraph", "DirectedGraph"]:
            return {
                "message": f"Unknown class type for method {method_name}", 
                "operation": method_name,
                "error": "unknown_class"
            }
        
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
                         "add_vertex": self._handle_add_vertex,
                         "add_edge": self._handle_add_edge,
                         "bfs": self._handle_bfs,
                         "dfs": self._handle_dfs
                         # Other behaviors can be added as analyzer supports them
                     }
                     
                     if behavior_type in behavior_map:
                         return behavior_map[behavior_type](instance, instance_name, params)

        method_map = {
            "add_vertex": self._handle_add_vertex,
            "add_edge": self._handle_add_edge,
            "remove_edge": self._handle_remove_edge,
            "remove_vertex": self._handle_remove_vertex,
            "display": self._handle_display,
            "bfs": self._handle_bfs,
            "dfs": self._handle_dfs,
            "has_cycle": self._handle_has_cycle,
            "is_connected": self._handle_is_connected,
            "topological_sort": self._handle_topological_sort,
            "get_in_degree": self._handle_get_in_degree,
            "get_out_degree": self._handle_get_out_degree
        }
        
        if method_name in method_map:
            return method_map[method_name](instance, instance_name, params)
        else:
            return {
                "message": f"Unknown method {method_name}", 
                "operation": method_name, 
                "error": "unknown_method"
            }
    
    def _handle_add_vertex(self, instance: Dict[str, Any], instance_name: str, params: str) -> Dict[str, Any]:
        """Handle add_vertex operation"""
        if params:
            vertex = self._parse_parameter(params)
            return self.node_manager.add_vertex(instance, vertex, instance_name)
        return {
            "message": "add_vertex requires a parameter", 
            "operation": "add_vertex", 
            "error": "missing_parameter"
        }
    
    def _handle_add_edge(self, instance: Dict[str, Any], instance_name: str, params: str) -> Dict[str, Any]:
        """Handle add_edge operation"""
        if params and ',' in params:
            param_parts = [p.strip().strip('"\'') for p in params.split(',')]
            if len(param_parts) == 2:
                vertex1, vertex2 = param_parts
                return self.node_manager.add_edge(instance, vertex1, vertex2, instance_name)
        return {
            "message": "add_edge requires two parameters", 
            "operation": "add_edge", 
            "error": "missing_parameters"
        }
    
    def _handle_remove_edge(self, instance: Dict[str, Any], instance_name: str, params: str) -> Dict[str, Any]:
        """Handle remove_edge operation"""
        if params and ',' in params:
            param_parts = [p.strip().strip('"\'') for p in params.split(',')]
            if len(param_parts) == 2:
                vertex1, vertex2 = param_parts
                return self.node_manager.remove_edge(instance, vertex1, vertex2, instance_name)
        return {
            "message": "remove_edge requires two parameters", 
            "operation": "remove_edge", 
            "error": "missing_parameters"
        }
    
    def _handle_remove_vertex(self, instance: Dict[str, Any], instance_name: str, params: str) -> Dict[str, Any]:
        """Handle remove_vertex operation"""
        if params:
            vertex = self._parse_parameter(params)
            return self.node_manager.remove_vertex(instance, vertex, instance_name)
        return {
            "message": "remove_vertex requires a parameter", 
            "operation": "remove_vertex", 
            "error": "missing_parameter"
        }
    
    def _handle_display(self, instance: Dict[str, Any], instance_name: str, params: str) -> Dict[str, Any]:
        """Handle display operation"""
        return self.node_manager.display_graph(instance, instance_name)
    
    def _handle_bfs(self, instance: Dict[str, Any], instance_name: str, params: str) -> Dict[str, Any]:
        """Handle BFS traversal"""
        if params:
            start_vertex = self._parse_parameter(params)
            return self.node_manager.bfs_traversal(instance, start_vertex, instance_name)
        return {
            "message": "bfs requires a start vertex parameter", 
            "operation": "bfs", 
            "error": "missing_parameter"
        }
    
    def _handle_dfs(self, instance: Dict[str, Any], instance_name: str, params: str) -> Dict[str, Any]:
        """Handle DFS traversal"""
        if params:
            start_vertex = self._parse_parameter(params)
            return self.node_manager.dfs_traversal(instance, start_vertex, instance_name)
        return {
            "message": "dfs requires a start vertex parameter", 
            "operation": "dfs", 
            "error": "missing_parameter"
        }
    
    def _handle_has_cycle(self, instance: Dict[str, Any], instance_name: str, params: str) -> Dict[str, Any]:
        """Handle has_cycle method"""
        return self.node_manager.has_cycle(instance, instance_name)
    
    def _handle_is_connected(self, instance: Dict[str, Any], instance_name: str, params: str) -> Dict[str, Any]:
        """Handle is_connected method"""
        return self.node_manager.is_connected(instance, instance_name)
    
    def _handle_topological_sort(self, instance: Dict[str, Any], instance_name: str, params: str) -> Dict[str, Any]:
        """Handle topological_sort method"""
        return self.node_manager.topological_sort(instance, instance_name)
    
    def _handle_get_in_degree(self, instance: Dict[str, Any], instance_name: str, params: str) -> Dict[str, Any]:
        """Handle get_in_degree method"""
        if params:
            vertex = self._parse_parameter(params)
            return self.node_manager.get_in_degree(instance, vertex, instance_name)
        return {
            "message": "get_in_degree requires a vertex parameter", 
            "operation": "get_in_degree", 
            "error": "missing_parameter"
        }
    
    def _handle_get_out_degree(self, instance: Dict[str, Any], instance_name: str, params: str) -> Dict[str, Any]:
        """Handle get_out_degree method"""
        if params:
            vertex = self._parse_parameter(params)
            return self.node_manager.get_out_degree(instance, vertex, instance_name)
        return {
            "message": "get_out_degree requires a vertex parameter", 
            "operation": "get_out_degree", 
            "error": "missing_parameter"
        }
    
    def _parse_parameter(self, params: str) -> Any:
        """Parse method parameters"""
        params = params.strip()
        
        # String literal
        if (params.startswith('"') and params.endswith('"')) or \
           (params.startswith("'") and params.endswith("'")):
            return params[1:-1]
        
        # Variable reference
        if params in self.context["variables"]:
            return self.context["variables"][params]
        
        return params