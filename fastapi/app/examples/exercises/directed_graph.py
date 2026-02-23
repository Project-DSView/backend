example = '''import collections

class DirectedGraph:
    def __init__(self):
        self.adjacency_list = {}

    def add_vertex(self, vertex):
        if vertex not in self.adjacency_list:
            self.adjacency_list[vertex] = []
            print(f"Vertex '{vertex}' added.")
        else:
            print(f"Vertex '{vertex}' already exists.")

    def add_edge(self, from_vertex, to_vertex):
        if from_vertex in self.adjacency_list and to_vertex in self.adjacency_list:
            if to_vertex not in self.adjacency_list[from_vertex]:
                self.adjacency_list[from_vertex].append(to_vertex)
                print(f"Edge added from '{from_vertex}' to '{to_vertex}'.")
            else:
                print(f"Edge from '{from_vertex}' to '{to_vertex}' already exists.")
        else:
            print("Error: One or both vertices not found.")

    def remove_edge(self, from_vertex, to_vertex):
        if from_vertex in self.adjacency_list and to_vertex in self.adjacency_list:
            try:
                self.adjacency_list[from_vertex].remove(to_vertex)
                print(f"Edge removed from '{from_vertex}' to '{to_vertex}'.")
            except ValueError:
                print(f"Error: No edge exists from '{from_vertex}' to '{to_vertex}'.")

    def remove_vertex(self, vertex_to_remove):
        if vertex_to_remove in self.adjacency_list:
            self.adjacency_list[vertex_to_remove] = []
            for vertex in self.adjacency_list:
                if vertex_to_remove in self.adjacency_list[vertex]:
                    self.adjacency_list[vertex].remove(vertex_to_remove)
            del self.adjacency_list[vertex_to_remove]
            print(f"Vertex '{vertex_to_remove}' removed.")
        else:
            print(f"Error: Vertex '{vertex_to_remove}' not found.")

    def display(self):
        if not self.adjacency_list:
            print("The graph is empty.")
            return
        print("\n--- Directed Graph Adjacency List ---")
        for vertex, neighbors in self.adjacency_list.items():
            print(f"'{vertex}': {', '.join(map(str, neighbors)) if neighbors else 'No outgoing edges'}")
        print("------------------------------------\n")
        
    def bfs(self, start_vertex):
        if start_vertex not in self.adjacency_list:
            print("Error: Starting vertex not found.")
            return []
        visited = set()
        queue = collections.deque([start_vertex])
        result = []
        visited.add(start_vertex)
        print(f"BFS (from '{start_vertex}'): ", end="")
        while queue:
            current_vertex = queue.popleft()
            result.append(current_vertex)
            print(f"-> {current_vertex}", end=" ")
            for neighbor in self.adjacency_list[current_vertex]:
                if neighbor not in visited:
                    visited.add(neighbor)
                    queue.append(neighbor)
        print()
        return result

    def dfs(self, start_vertex):
        if start_vertex not in self.adjacency_list:
            print("Error: Starting vertex not found.")
            return []
        visited = set()
        result = []
        def _dfs_recursive(vertex):
            visited.add(vertex)
            result.append(vertex)
            print(f"-> {vertex}", end=" ")
            for neighbor in self.adjacency_list[vertex]:
                if neighbor not in visited:
                    _dfs_recursive(neighbor)
        print(f"DFS (from '{start_vertex}'): ", end="")
        _dfs_recursive(start_vertex)
        print()
        return result

    def has_cycle(self):
        visited = set()
        rec_stack = set()
        def dfs_cycle_check(vertex):
            visited.add(vertex)
            rec_stack.add(vertex)
            for neighbor in self.adjacency_list[vertex]:
                if neighbor not in visited:
                    if dfs_cycle_check(neighbor):
                        return True
                elif neighbor in rec_stack:
                    return True
            rec_stack.remove(vertex)
            return False
        for vertex in self.adjacency_list:
            if vertex not in visited:
                if dfs_cycle_check(vertex):
                    return True
        return False

    def topological_sort(self):
        if self.has_cycle():
            print("Cannot perform topological sort: Graph contains a cycle.")
            return None
        visited = set()
        stack = []
        def dfs_topological(vertex):
            visited.add(vertex)
            for neighbor in self.adjacency_list[vertex]:
                if neighbor not in visited:
                    dfs_topological(neighbor)
            stack.append(vertex)
        for vertex in self.adjacency_list:
            if vertex not in visited:
                dfs_topological(vertex)
        return stack[::-1]

    def get_in_degree(self, vertex):
        if vertex not in self.adjacency_list:
            return -1
        in_degree = 0
        for v in self.adjacency_list:
            if vertex in self.adjacency_list[v]:
                in_degree += 1
        return in_degree

    def get_out_degree(self, vertex):
        if vertex not in self.adjacency_list:
            return -1
        return len(self.adjacency_list[vertex])

# Example usage
print("=== Directed Graph Example ===")
myGraph = DirectedGraph()

print("--- Adding Vertices and Edges ---")
myGraph.add_vertex("A")
myGraph.add_vertex("B")
myGraph.add_vertex("C")
myGraph.add_vertex("D")
myGraph.add_edge("A", "B")
myGraph.add_edge("A", "C")
myGraph.add_edge("B", "D")
myGraph.add_edge("C", "D")
myGraph.add_edge("D", "A")  # This creates a cycle
myGraph.display()

print("--- Graph Traversal ---")
myGraph.bfs("A")
myGraph.dfs("A")

print("--- Graph Properties ---")
print(f"Has cycle: {myGraph.has_cycle()}")
print(f"Vertex A - In-degree: {myGraph.get_in_degree('A')}, Out-degree: {myGraph.get_out_degree('A')}")

print("--- Removing Edge to Break Cycle ---")
myGraph.remove_edge("D", "A")
myGraph.display()

print("--- Topological Sort ---")
topo_order = myGraph.topological_sort()
if topo_order:
    print(f"Topological order: {' -> '.join(topo_order)}")

print("--- Final Traversal ---")
myGraph.bfs("A")
myGraph.dfs("A")
'''
