
import sys
import os
import json
from pprint import pprint

sys.path.append(os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__)))))

from app.services.simulators.direct_code_executor import DirectCodeExecutor

def test_directed_graph_stepthrough():
    print("\n=== Testing Directed Graph Stepthrough ===")
    code = """
class DirectedGraph:
    def __init__(self):
        self.graph = {}

    def add_vertex(self, vertex):
        if vertex not in self.graph:
            self.graph[vertex] = {}

    def add_edge(self, u, v, weight=1):
        self.add_vertex(u)
        self.add_vertex(v)
        self.graph[u][v] = weight

    def traverse_bfs(self, start_node):
        visited = set()
        queue = [start_node]
        visited.add(start_node)

        while queue:
            current = queue.pop(0)
            print(f"Visiting {current}")
            
            for neighbor in self.graph.get(current, {}):
                if neighbor not in visited:
                    visited.add(neighbor)
                    queue.append(neighbor)

g = DirectedGraph()
g.add_edge('A', 'B')
g.add_edge('B', 'C')
g.add_edge('A', 'C')
g.traverse_bfs('A')
"""
    executor = DirectCodeExecutor()
    steps = executor.execute(code, data_structure_type="directedgraph")

    print(f"Total Steps: {len(steps)}")
    
    # Check for instance detection
    graph_instances = [s for s in steps if s.state.get("instances")]
    print(f"Steps with instances: {len(graph_instances)}")
    
    if graph_instances:
        last_instance = graph_instances[-1].state["instances"]
        print("Last Instance State:")
        # pprint(last_instance)
        # Check type and content
        found = False
        for name, inst in last_instance.items():
            if inst.get("type") == "DirectedGraph":
                print(f"Found DirectedGraph '{name}': {inst.get('graph')}")
                found = True
        if not found:
            print("FAILED: DirectedGraph instance not detected correctly.")
    else:
        print("FAILED: No instances detected.")

    # Check for highlighting (current_node)
    highlight_steps = [
        s for s in steps 
        if s.state.get("step_detail", {}).get("current_node") 
    ]
    print(f"Steps with current_node highlight: {len(highlight_steps)}")
    if highlight_steps:
        print("Examples of highlights:")
        for i, s in enumerate(highlight_steps[:5]):
            print(f"  Step {s.stepNumber} (Line {s.line}): current_node = {s.state['step_detail']['current_node']}")
    else:
        print("WARNING: No current_node highlights found.")


def test_undirected_graph_stepthrough():
    print("\n=== Testing Undirected Graph Stepthrough ===")
    code = """
class UndirectedGraph:
    def __init__(self):
        self.graph = {}

    def add_vertex(self, vertex):
        if vertex not in self.graph:
            self.graph[vertex] = {}

    def add_edge(self, u, v, weight=1):
        self.add_vertex(u)
        self.add_vertex(v)
        self.graph[u][v] = weight
        self.graph[v][u] = weight

    def traverse_dfs(self, start_node):
        visited = set()
        stack = [start_node]

        while stack:
            vertex = stack.pop()
            if vertex not in visited:
                print(f"Visiting {vertex}")
                visited.add(vertex)
                for neighbor in reversed(list(self.graph.get(vertex, {}).keys())):
                    if neighbor not in visited:
                        stack.append(neighbor)

g = UndirectedGraph()
g.add_edge('A', 'B')
g.add_edge('B', 'C')
g.add_edge('A', 'D')
g.traverse_dfs('A')
"""
    executor = DirectCodeExecutor()
    steps = executor.execute(code, data_structure_type="undirectedgraph")

    print(f"Total Steps: {len(steps)}")
    
    # Check for instance detection
    graph_instances = [s for s in steps if s.state.get("instances")]
    print(f"Steps with instances: {len(graph_instances)}")
    
    if graph_instances:
        last_instance = graph_instances[-1].state["instances"]
        # Check type
        found = False
        for name, inst in last_instance.items():
            if inst.get("type") == "UndirectedGraph":
                 print(f"Found UndirectedGraph '{name}': {inst.get('graph')}")
                 found = True
        if not found:
             print("FAILED: UndirectedGraph instance not detected correctly.")

    # Check for highlighting
    highlight_steps = [
        s for s in steps 
        if s.state.get("step_detail", {}).get("current_node")
    ]
    print(f"Steps with current_node highlight: {len(highlight_steps)}")
    if highlight_steps:
        print("Examples of highlights:")
        for s in highlight_steps[:5]:
             print(f"  Step {s.stepNumber}: current_node = {s.state['step_detail']['current_node']}")

if __name__ == "__main__":
    test_directed_graph_stepthrough()
    test_undirected_graph_stepthrough()
