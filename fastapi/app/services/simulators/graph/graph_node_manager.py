from typing import Dict, Any
from collections import deque
from app.services.simulators.operations.node_manager import NodeManager


class GraphNodeManager(NodeManager):
    """Enhanced Graph-specific node manager for Graph operations"""
    
    def create_instance_data(self, class_name: str) -> Dict[str, Any]:
        """Create new instance data structure for Graph"""
        if class_name in ["Graph", "UndirectedGraph", "DirectedGraph"]:
            return {
                "adjacency_list": {},
                "class_type": class_name,
                "attributes": {},
                "history": []  # Track operation history
            }
        return super().create_instance_data(class_name)
    
    def create_instance(self, var_name: str, class_name: str) -> str:
        """Create a new instance of a class"""
        if class_name in ["Graph", "UndirectedGraph", "DirectedGraph"]:
            self.context["instances"][var_name] = self.create_instance_data(class_name)
            self.context["active_instance"] = var_name
            # Initialize edges and vertices lists
            if "edges" not in self.context:
                self.context["edges"] = []
            if "vertices" not in self.context:
                self.context["vertices"] = []
            return f"Created {class_name}() → {var_name}.adjacency_list = {{}}"
        else:
            return super().create_instance(var_name, class_name)
    
    def add_vertex(self, instance: Dict[str, Any], vertex: Any, instance_name: str = "") -> Dict[str, Any]:
        """Add vertex to graph with detailed tracking"""
        adjacency_list = instance.get("adjacency_list", {})
        old_adjacency_list = adjacency_list.copy()
        
        if vertex not in adjacency_list:
            adjacency_list[vertex] = []
            instance["adjacency_list"] = adjacency_list
            
            # Update context for individual vertex tracking
            if "vertices" not in self.context:
                self.context["vertices"] = []
            if vertex not in self.context["vertices"]:
                self.context["vertices"].append(vertex)
            
            # Record operation in history
            instance["history"].append({
                "operation": "add_vertex",
                "vertex": vertex,
                "before": old_adjacency_list,
                "after": adjacency_list.copy()
            })
            
            return {
                "message": f"add_vertex('{vertex}') → vertex '{vertex}' added",
                "operation": "add_vertex",
                "vertex": vertex,
                "before_graph": old_adjacency_list,
                "after_graph": adjacency_list.copy(),
                "vertices": self.context["vertices"].copy(),
                "edges": [list(edge) for edge in self.context.get("edges", [])],  # Convert tuples to arrays for frontend
                "instance_name": instance_name
            }
        else:
            return {
                "message": f"add_vertex('{vertex}') → vertex '{vertex}' already exists",
                "operation": "add_vertex",
                "vertex": vertex,
                "warning": "vertex_exists",
                "vertices": self.context["vertices"].copy(),
                "edges": [list(edge) for edge in self.context.get("edges", [])],  # Convert tuples to arrays for frontend
                "instance_name": instance_name
            }
    
    def add_edge(self, instance: Dict[str, Any], vertex1: str, vertex2: str, instance_name: str = "") -> Dict[str, Any]:
        """Add edge between two vertices with detailed tracking"""
        adjacency_list = instance.get("adjacency_list", {})
        old_adjacency_list = {k: v.copy() for k, v in adjacency_list.items()}
        class_type = instance.get("class_type", "Graph")
        
        if vertex1 in adjacency_list and vertex2 in adjacency_list:
            # Add edge from vertex1 to vertex2
            if vertex2 not in adjacency_list[vertex1]:
                adjacency_list[vertex1].append(vertex2)
            
            # For undirected graphs, add edge from vertex2 to vertex1
            if class_type in ["Graph", "UndirectedGraph"]:
                if vertex1 not in adjacency_list[vertex2]:
                    adjacency_list[vertex2].append(vertex1)
            
            # Update context for individual edge tracking
            if "edges" not in self.context:
                self.context["edges"] = []
            
            # Store edge in context
            if class_type == "DirectedGraph":
                # For directed graphs, store only the actual edge direction
                edge = (vertex1, vertex2)
                if edge not in self.context["edges"]:
                    self.context["edges"].append(edge)
            else:
                # For undirected graphs, store as single edge
                edge = (vertex1, vertex2) if vertex1 < vertex2 else (vertex2, vertex1)
                if edge not in self.context["edges"]:
                    self.context["edges"].append(edge)
            
            # Record operation in history
            instance["history"].append({
                "operation": "add_edge",
                "vertex1": vertex1,
                "vertex2": vertex2,
                "before": old_adjacency_list,
                "after": {k: v.copy() for k, v in adjacency_list.items()}
            })
            
            edge_type = "undirected edge" if class_type in ["Graph", "UndirectedGraph"] else "directed edge"
            return {
                "message": f"add_edge('{vertex1}', '{vertex2}') → {edge_type} added between '{vertex1}' and '{vertex2}'",
                "operation": "add_edge",
                "vertex1": vertex1,
                "vertex2": vertex2,
                "before_graph": old_adjacency_list,
                "after_graph": {k: v.copy() for k, v in adjacency_list.items()},
                "edges": [list(edge) for edge in self.context["edges"]],  # Convert tuples to arrays for frontend
                "vertices": self.context.get("vertices", []).copy(),
                "instance_name": instance_name
            }
        else:
            missing_vertices = []
            if vertex1 not in adjacency_list:
                missing_vertices.append(vertex1)
            if vertex2 not in adjacency_list:
                missing_vertices.append(vertex2)
            
            return {
                "message": f"add_edge('{vertex1}', '{vertex2}') failed → vertices not found: {missing_vertices}",
                "operation": "add_edge",
                "vertex1": vertex1,
                "vertex2": vertex2,
                "error": "vertices_not_found",
                "missing_vertices": missing_vertices,
                "edges": [list(edge) for edge in self.context.get("edges", [])],  # Convert tuples to arrays for frontend
                "vertices": self.context.get("vertices", []).copy(),
                "instance_name": instance_name
            }
    
    def remove_edge(self, instance: Dict[str, Any], vertex1: str, vertex2: str, instance_name: str = "") -> Dict[str, Any]:
        """Remove edge between two vertices with detailed tracking"""
        adjacency_list = instance.get("adjacency_list", {})
        old_adjacency_list = {k: v.copy() for k, v in adjacency_list.items()}
        class_type = instance.get("class_type", "Graph")
        
        if vertex1 in adjacency_list and vertex2 in adjacency_list:
            try:
                if vertex2 in adjacency_list[vertex1]:
                    adjacency_list[vertex1].remove(vertex2)
                
                # For undirected graphs, remove edge from vertex2 to vertex1
                if class_type in ["Graph", "UndirectedGraph"]:
                    if vertex1 in adjacency_list[vertex2]:
                        adjacency_list[vertex2].remove(vertex1)
                
                # Update context for individual edge tracking
                if "edges" in self.context:
                    if class_type == "DirectedGraph":
                        # For directed graphs, remove only the actual edge direction
                        edge = (vertex1, vertex2)
                        if edge in self.context["edges"]:
                            self.context["edges"].remove(edge)
                    else:
                        # For undirected graphs, remove single edge
                        edge = (vertex1, vertex2) if vertex1 < vertex2 else (vertex2, vertex1)
                        if edge in self.context["edges"]:
                            self.context["edges"].remove(edge)
                
                # Record operation in history
                instance["history"].append({
                    "operation": "remove_edge",
                    "vertex1": vertex1,
                    "vertex2": vertex2,
                    "before": old_adjacency_list,
                    "after": {k: v.copy() for k, v in adjacency_list.items()}
                })
                
                edge_type = "undirected edge" if class_type in ["Graph", "UndirectedGraph"] else "directed edge"
                return {
                    "message": f"remove_edge('{vertex1}', '{vertex2}') → {edge_type} removed between '{vertex1}' and '{vertex2}'",
                    "operation": "remove_edge",
                    "vertex1": vertex1,
                    "vertex2": vertex2,
                    "before_graph": old_adjacency_list,
                    "after_graph": {k: v.copy() for k, v in adjacency_list.items()},
                    "edges": [list(edge) for edge in self.context.get("edges", [])],  # Convert tuples to arrays for frontend
                    "vertices": self.context.get("vertices", []).copy(),
                    "instance_name": instance_name
                }
            except ValueError:
                return {
                    "message": f"remove_edge('{vertex1}', '{vertex2}') failed → no edge exists between '{vertex1}' and '{vertex2}'",
                    "operation": "remove_edge",
                    "vertex1": vertex1,
                    "vertex2": vertex2,
                    "error": "edge_not_found",
                    "edges": [list(edge) for edge in self.context.get("edges", [])],  # Convert tuples to arrays for frontend
                    "vertices": self.context.get("vertices", []).copy(),
                    "instance_name": instance_name
                }
        else:
            return {
                "message": f"remove_edge('{vertex1}', '{vertex2}') failed → one or both vertices not found",
                "operation": "remove_edge",
                "vertex1": vertex1,
                "vertex2": vertex2,
                "error": "vertices_not_found",
                "edges": [list(edge) for edge in self.context.get("edges", [])],  # Convert tuples to arrays for frontend
                "vertices": self.context.get("vertices", []).copy(),
                "instance_name": instance_name
            }
    
    def remove_vertex(self, instance: Dict[str, Any], vertex_to_remove: str, instance_name: str = "") -> Dict[str, Any]:
        """Remove vertex and all its edges with detailed tracking"""
        adjacency_list = instance.get("adjacency_list", {})
        old_adjacency_list = {k: v.copy() for k, v in adjacency_list.items()}
        
        if vertex_to_remove in adjacency_list:
            # Get neighbors before removal for tracking
            neighbors = adjacency_list[vertex_to_remove].copy()
            
            # Remove all edges pointing to this vertex
            for vertex in adjacency_list:
                if vertex_to_remove in adjacency_list[vertex]:
                    adjacency_list[vertex].remove(vertex_to_remove)
            
            # Remove the vertex itself
            del adjacency_list[vertex_to_remove]
            
            # Update context for individual vertex tracking
            if "vertices" in self.context and vertex_to_remove in self.context["vertices"]:
                self.context["vertices"].remove(vertex_to_remove)
            
            # Update context for individual edge tracking - remove all edges involving this vertex
            if "edges" in self.context:
                self.context["edges"] = [edge for edge in self.context["edges"] 
                                       if vertex_to_remove not in edge]
            
            # Record operation in history
            instance["history"].append({
                "operation": "remove_vertex",
                "vertex": vertex_to_remove,
                "neighbors": neighbors,
                "before": old_adjacency_list,
                "after": {k: v.copy() for k, v in adjacency_list.items()}
            })
            
            return {
                "message": f"remove_vertex('{vertex_to_remove}') → vertex '{vertex_to_remove}' and its edges removed",
                "operation": "remove_vertex",
                "vertex": vertex_to_remove,
                "removed_neighbors": neighbors,
                "before_graph": old_adjacency_list,
                "after_graph": {k: v.copy() for k, v in adjacency_list.items()},
                "edges": [list(edge) for edge in self.context.get("edges", [])],  # Convert tuples to arrays for frontend
                "vertices": self.context.get("vertices", []).copy(),
                "instance_name": instance_name
            }
        else:
            return {
                "message": f"remove_vertex('{vertex_to_remove}') failed → vertex '{vertex_to_remove}' not found",
                "operation": "remove_vertex",
                "vertex": vertex_to_remove,
                "error": "vertex_not_found",
                "edges": [list(edge) for edge in self.context.get("edges", [])],  # Convert tuples to arrays for frontend
                "vertices": self.context.get("vertices", []).copy(),
                "instance_name": instance_name
            }
    
    def display_graph(self, instance: Dict[str, Any], instance_name: str = "") -> Dict[str, Any]:
        """Display graph adjacency list with detailed tracking"""
        adjacency_list = instance.get("adjacency_list", {})
        
        if not adjacency_list:
            self.context["stdout"].append("The graph is empty.")
            return {
                "message": f"{instance_name}.display() → graph is empty",
                "operation": "display",
                "graph_state": "empty",
                "edges": [list(edge) for edge in self.context.get("edges", [])],  # Convert tuples to arrays for frontend
                "vertices": self.context.get("vertices", []).copy(),
                "instance_name": instance_name
            }
        
        # Format display output
        display_lines = ["", "--- Graph Adjacency List ---"]
        for vertex, neighbors in adjacency_list.items():
            if neighbors:
                neighbor_str = ', '.join(map(str, neighbors))
                display_lines.append(f"'{vertex}': {neighbor_str}")
            else:
                display_lines.append(f"'{vertex}': No neighbors")
        display_lines.append("----------------------------")
        display_lines.append("")
        
        # Add to print output
        for line in display_lines:
            self.context["stdout"].append(line)
        
        return {
            "message": f"{instance_name}.display() → displayed adjacency list",
            "operation": "display",
            "adjacency_list": adjacency_list.copy(),
            "vertex_count": len(adjacency_list),
            "edge_count": sum(len(neighbors) for neighbors in adjacency_list.values()) // 2,
            "display_output": display_lines,
            "edges": [list(edge) for edge in self.context.get("edges", [])],  # Convert tuples to arrays for frontend
            "vertices": self.context.get("vertices", []).copy(),
            "instance_name": instance_name
        }
    
    def bfs_traversal(self, instance: Dict[str, Any], start_vertex: str, instance_name: str = "") -> Dict[str, Any]:
        """Perform BFS traversal with detailed tracking"""
        adjacency_list = instance.get("adjacency_list", {})
        
        if start_vertex not in adjacency_list:
            return {
                "message": f"{instance_name}.bfs('{start_vertex}') failed → starting vertex not found",
                "operation": "bfs",
                "start_vertex": start_vertex,
                "error": "vertex_not_found",
                "edges": [list(edge) for edge in self.context.get("edges", [])],  # Convert tuples to arrays for frontend
                "vertices": self.context.get("vertices", []).copy(),
                "instance_name": instance_name
            }
        
        visited = set()
        queue = deque([start_vertex])
        result = []
        traversal_steps = []
        
        visited.add(start_vertex)
        
        bfs_output = f"BFS (from '{start_vertex}'): "
        
        while queue:
            current_vertex = queue.popleft()
            result.append(current_vertex)
            bfs_output += f"-> {current_vertex} "
            
            traversal_steps.append({
                "current": current_vertex,
                "queue_before": list(queue),
                "visited_before": list(visited)
            })
            
            for neighbor in adjacency_list[current_vertex]:
                if neighbor not in visited:
                    visited.add(neighbor)
                    queue.append(neighbor)
        
        # Add to print output
        self.context["stdout"].append(bfs_output)
        
        return {
            "message": f"{instance_name}.bfs('{start_vertex}') → traversal: {' -> '.join(result)}",
            "operation": "bfs",
            "start_vertex": start_vertex,
            "traversal_result": result,
            "traversal_steps": traversal_steps,
            "visited_vertices": list(visited),
            "edges": self.context.get("edges", []).copy(),
            "vertices": self.context.get("vertices", []).copy(),
            "instance_name": instance_name
        }
    
    def dfs_traversal(self, instance: Dict[str, Any], start_vertex: str, instance_name: str = "") -> Dict[str, Any]:
        """Perform DFS traversal with detailed tracking"""
        adjacency_list = instance.get("adjacency_list", {})
        
        if start_vertex not in adjacency_list:
            return {
                "message": f"{instance_name}.dfs('{start_vertex}') failed → starting vertex not found",
                "operation": "dfs",
                "start_vertex": start_vertex,
                "error": "vertex_not_found",
                "edges": [list(edge) for edge in self.context.get("edges", [])],  # Convert tuples to arrays for frontend
                "vertices": self.context.get("vertices", []).copy(),
                "instance_name": instance_name
            }
        
        visited = set()
        result = []
        traversal_steps = []
        
        def _dfs_recursive(vertex):
            visited.add(vertex)
            result.append(vertex)
            traversal_steps.append({
                "current": vertex,
                "visited_before": list(visited),
                "neighbors": adjacency_list[vertex].copy()
            })
            
            for neighbor in adjacency_list[vertex]:
                if neighbor not in visited:
                    _dfs_recursive(neighbor)
        
        dfs_output = f"DFS (from '{start_vertex}'): "
        _dfs_recursive(start_vertex)
        
        for vertex in result:
            dfs_output += f"-> {vertex} "
        
        # Add to print output
        self.context["stdout"].append(dfs_output)
        
        return {
            "message": f"{instance_name}.dfs('{start_vertex}') → traversal: {' -> '.join(result)}",
            "operation": "dfs",
            "start_vertex": start_vertex,
            "traversal_result": result,
            "traversal_steps": traversal_steps,
            "visited_vertices": list(visited),
            "edges": self.context.get("edges", []).copy(),
            "vertices": self.context.get("vertices", []).copy(),
            "instance_name": instance_name
        }
    
    def has_cycle(self, instance: Dict[str, Any], instance_name: str = "") -> Dict[str, Any]:
        """Check if graph has cycle"""
        adjacency_list = instance.get("adjacency_list", {})
        class_type = instance.get("class_type", "Graph")
        
        if class_type in ["Graph", "UndirectedGraph"]:
            # Undirected graph cycle detection
            visited = set()
            
            def dfs_cycle_check(vertex, parent):
                visited.add(vertex)
                for neighbor in adjacency_list[vertex]:
                    if neighbor not in visited:
                        if dfs_cycle_check(neighbor, vertex):
                            return True
                    elif neighbor != parent:
                        return True
                return False
            
            for vertex in adjacency_list:
                if vertex not in visited:
                    if dfs_cycle_check(vertex, None):
                        return {
                            "message": f"{instance_name}.has_cycle() → True (cycle detected)",
                            "operation": "has_cycle",
                            "result": True,
                            "edges": [list(edge) for edge in self.context.get("edges", [])],  # Convert tuples to arrays for frontend
                            "vertices": self.context.get("vertices", []).copy(),
                            "instance_name": instance_name
                        }
        else:
            # Directed graph cycle detection
            visited = set()
            rec_stack = set()
            
            def dfs_cycle_check(vertex):
                visited.add(vertex)
                rec_stack.add(vertex)
                
                for neighbor in adjacency_list[vertex]:
                    if neighbor not in visited:
                        if dfs_cycle_check(neighbor):
                            return True
                    elif neighbor in rec_stack:
                        return True
                
                rec_stack.remove(vertex)
                return False
            
            for vertex in adjacency_list:
                if vertex not in visited:
                    if dfs_cycle_check(vertex):
                        return {
                            "message": f"{instance_name}.has_cycle() → True (cycle detected)",
                            "operation": "has_cycle",
                            "result": True,
                            "edges": [list(edge) for edge in self.context.get("edges", [])],  # Convert tuples to arrays for frontend
                            "vertices": self.context.get("vertices", []).copy(),
                            "instance_name": instance_name
                        }
        
        return {
            "message": f"{instance_name}.has_cycle() → False (no cycle)",
            "operation": "has_cycle",
            "result": False,
            "edges": self.context.get("edges", []).copy(),
            "vertices": self.context.get("vertices", []).copy(),
            "instance_name": instance_name
        }
    
    def is_connected(self, instance: Dict[str, Any], instance_name: str = "") -> Dict[str, Any]:
        """Check if undirected graph is connected"""
        adjacency_list = instance.get("adjacency_list", {})
        
        if not adjacency_list:
            return {
                "message": f"{instance_name}.is_connected() → True (empty graph)",
                "operation": "is_connected",
                "result": True,
                "edges": [list(edge) for edge in self.context.get("edges", [])],  # Convert tuples to arrays for frontend
                "vertices": self.context.get("vertices", []).copy(),
                "instance_name": instance_name
            }
        
        visited = set()
        start_vertex = next(iter(adjacency_list))
        
        def dfs_visit(vertex):
            visited.add(vertex)
            for neighbor in adjacency_list[vertex]:
                if neighbor not in visited:
                    dfs_visit(neighbor)
        
        dfs_visit(start_vertex)
        is_connected = len(visited) == len(adjacency_list)
        
        return {
            "message": f"{instance_name}.is_connected() → {is_connected}",
            "operation": "is_connected",
            "result": is_connected,
            "visited_count": len(visited),
            "total_vertices": len(adjacency_list),
            "edges": self.context.get("edges", []).copy(),
            "vertices": self.context.get("vertices", []).copy(),
            "instance_name": instance_name
        }
    
    def topological_sort(self, instance: Dict[str, Any], instance_name: str = "") -> Dict[str, Any]:
        """Perform topological sort on directed graph"""
        adjacency_list = instance.get("adjacency_list", {})
        
        # Check if graph has cycle first
        cycle_check = self.has_cycle(instance, instance_name)
        if cycle_check.get("result", False):
            return {
                "message": f"{instance_name}.topological_sort() → Cannot sort: Graph contains a cycle",
                "operation": "topological_sort",
                "result": None,
                "error": "cycle_detected",
                "edges": [list(edge) for edge in self.context.get("edges", [])],  # Convert tuples to arrays for frontend
                "vertices": self.context.get("vertices", []).copy(),
                "instance_name": instance_name
            }
        
        visited = set()
        stack = []
        
        def dfs_topological(vertex):
            visited.add(vertex)
            for neighbor in adjacency_list[vertex]:
                if neighbor not in visited:
                    dfs_topological(neighbor)
            stack.append(vertex)
        
        for vertex in adjacency_list:
            if vertex not in visited:
                dfs_topological(vertex)
        
        result = stack[::-1]
        
        return {
            "message": f"{instance_name}.topological_sort() → {result}",
            "operation": "topological_sort",
            "result": result,
            "edges": self.context.get("edges", []).copy(),
            "vertices": self.context.get("vertices", []).copy(),
            "instance_name": instance_name
        }
    
    def get_in_degree(self, instance: Dict[str, Any], vertex: str, instance_name: str = "") -> Dict[str, Any]:
        """Get in-degree of a vertex"""
        adjacency_list = instance.get("adjacency_list", {})
        
        if vertex not in adjacency_list:
            return {
                "message": f"{instance_name}.get_in_degree('{vertex}') → Error: Vertex not found",
                "operation": "get_in_degree",
                "result": -1,
                "error": "vertex_not_found",
                "edges": [list(edge) for edge in self.context.get("edges", [])],  # Convert tuples to arrays for frontend
                "vertices": self.context.get("vertices", []).copy(),
                "instance_name": instance_name
            }
        
        in_degree = 0
        for v in adjacency_list:
            if vertex in adjacency_list[v]:
                in_degree += 1
        
        return {
            "message": f"{instance_name}.get_in_degree('{vertex}') → {in_degree}",
            "operation": "get_in_degree",
            "vertex": vertex,
            "result": in_degree,
            "edges": self.context.get("edges", []).copy(),
            "vertices": self.context.get("vertices", []).copy(),
            "instance_name": instance_name
        }
    
    def get_out_degree(self, instance: Dict[str, Any], vertex: str, instance_name: str = "") -> Dict[str, Any]:
        """Get out-degree of a vertex"""
        adjacency_list = instance.get("adjacency_list", {})
        
        if vertex not in adjacency_list:
            return {
                "message": f"{instance_name}.get_out_degree('{vertex}') → Error: Vertex not found",
                "operation": "get_out_degree",
                "result": -1,
                "error": "vertex_not_found",
                "edges": [list(edge) for edge in self.context.get("edges", [])],  # Convert tuples to arrays for frontend
                "vertices": self.context.get("vertices", []).copy(),
                "instance_name": instance_name
            }
        
        out_degree = len(adjacency_list[vertex])
        
        return {
            "message": f"{instance_name}.get_out_degree('{vertex}') → {out_degree}",
            "operation": "get_out_degree",
            "vertex": vertex,
            "result": out_degree,
            "edges": self.context.get("edges", []).copy(),
            "vertices": self.context.get("vertices", []).copy(),
            "instance_name": instance_name
        }