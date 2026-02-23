example = '''import collections

class UndirectedGraph:
    def __init__(self):
        self.adjacency_list = {}

    def add_vertex(self, vertex):
        if vertex not in self.adjacency_list:
            self.adjacency_list[vertex] = []
            print(f"Vertex '{vertex}' added.")
        else:
            print(f"Vertex '{vertex}' already exists.")

    def add_edge(self, vertex1, vertex2):
        if vertex1 in self.adjacency_list and vertex2 in self.adjacency_list:
            if vertex2 not in self.adjacency_list[vertex1]:
                self.adjacency_list[vertex1].append(vertex2)
            if vertex1 not in self.adjacency_list[vertex2]:
                self.adjacency_list[vertex2].append(vertex1)
            print(f"Edge added between '{vertex1}' and '{vertex2}'.")
        else:
            print("Error: One or both vertices not found.")

    def remove_edge(self, vertex1, vertex2):
        if vertex1 in self.adjacency_list and vertex2 in self.adjacency_list:
            try:
                self.adjacency_list[vertex1].remove(vertex2)
                self.adjacency_list[vertex2].remove(vertex1)
                print(f"Edge removed between '{vertex1}' and '{vertex2}'.")
            except ValueError:
                print(f"Error: No edge exists between '{vertex1}' and '{vertex2}'.")

    def remove_vertex(self, vertex_to_remove):
        if vertex_to_remove in self.adjacency_list:
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
        print("\n--- Graph Adjacency List ---")
        for vertex, neighbors in self.adjacency_list.items():
            print(f"'{vertex}': {', '.join(map(str, neighbors)) if neighbors else 'No neighbors'}")
        print("----------------------------\n")
        
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
        def dfs_cycle_check(vertex, parent):
            visited.add(vertex)
            for neighbor in self.adjacency_list[vertex]:
                if neighbor not in visited:
                    if dfs_cycle_check(neighbor, vertex):
                        return True
                elif neighbor != parent:
                    return True
            return False
        for vertex in self.adjacency_list:
            if vertex not in visited:
                if dfs_cycle_check(vertex, None):
                    return True
        return False

    def is_connected(self):
        if not self.adjacency_list:
            return True
        visited = set()
        start_vertex = next(iter(self.adjacency_list))
        def dfs_visit(vertex):
            visited.add(vertex)
            for neighbor in self.adjacency_list[vertex]:
                if neighbor not in visited:
                    dfs_visit(neighbor)
        dfs_visit(start_vertex)
        return len(visited) == len(self.adjacency_list)

# Example usage
print("=== Undirected Graph Example ===")
myGraph = UndirectedGraph()

print("--- Adding Vertices and Edges ---")
myGraph.add_vertex("A")
myGraph.add_vertex("B")
myGraph.add_vertex("C")
myGraph.add_vertex("D")
myGraph.add_edge("A", "B")
myGraph.add_edge("A", "C")
myGraph.add_edge("B", "D")
myGraph.add_edge("C", "D")
myGraph.display()

print("--- Graph Traversal ---")
myGraph.bfs("A")
myGraph.dfs("A")

print("--- Graph Properties ---")
print(f"Is connected: {myGraph.is_connected()}")
print(f"Has cycle: {myGraph.has_cycle()}")

print("--- Removing Edge ---")
myGraph.remove_edge("C", "D")
myGraph.display()

print("--- Final Traversal ---")
myGraph.bfs("A")
myGraph.dfs("A")
'''
