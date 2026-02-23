import ast
from typing import Optional, Dict, Any, Set
from app.services.simulators.operations.ast_parser import ASTParser


class DataStructureDetector:
    """
    Detects data structure type from Python code
    Uses class name patterns (primary) and method signatures (fallback)
    """
    
    # Class name patterns mapping
    CLASS_NAME_PATTERNS = {
        "binarysearchtree": ["BST", "BinarySearchTree", "BinaryTree"],
        "singlylinkedlist": ["SinglyLinkedList", "LinkedList"],
        "doublylinkedlist": ["DoublyLinkedList"],
        "stack": ["Stack", "ArrayStack"],
        "queue": ["Queue"],
        "undirectedgraph": ["Graph", "UndirectedGraph"],
        "directedgraph": ["DirectedGraph"],
    }
    
    # Method signature patterns (fallback)
    METHOD_PATTERNS = {
        "stack": {
            "methods": ["push", "pop", "stackTop", "stack_top", "peek"],
            "min_matches": 2
        },
        "singlylinkedlist": {
            "methods": ["insertFront", "insertEnd", "insert_front", "insert_end", 
                       "traverse", "delete", "search"],
            "min_matches": 2
        },
        "doublylinkedlist": {
            "methods": ["insertFront", "insertEnd", "insert_front", "insert_end",
                       "traverse", "traverseReverse", "traverse_reverse"],
            "min_matches": 2
        },
        "binarysearchtree": {
            "methods": ["insert", "delete", "search", "inorder", "preorder", 
                       "postorder", "findMin", "findMax", "find_min", "find_max"],
            "min_matches": 2
        },
        "undirectedgraph": {
            "methods": ["addEdge", "add_edge", "addVertex", "add_vertex",
                       "dfs", "bfs", "isConnected", "is_connected"],
            "min_matches": 2
        },
        "directedgraph": {
            "methods": ["addEdge", "add_edge", "addVertex", "add_vertex",
                      "topologicalSort", "topological_sort", "hasCycle", "has_cycle"],
            "min_matches": 2
        },
        "queue": {
            "methods": ["enqueue", "dequeue", "front", "rear", "isEmpty", "is_empty"],
            "min_matches": 2
        }
    }
    
    def __init__(self):
        """Initialize the detector"""
        self.ast_parser = ASTParser()
    
    def detect_from_code(self, code: str) -> Optional[str]:
        """
        Detect data structure type from code
        
        Args:
            code: Python code to analyze
            
        Returns:
            Detected data structure type (e.g., "stack", "singlylinkedlist") or None
        """
        try:
            # Parse AST
            tree = self.ast_parser.parse_code(code)
            
            # Try class name detection first (primary method)
            detected_type = self._detect_from_class_names(tree)
            if detected_type:
                return detected_type
            
            # Fallback to method signature analysis
            detected_type = self._detect_from_methods(tree)
            if detected_type:
                return detected_type
            
            return None
            
        except Exception:
            # If parsing fails, return None
            return None
    
    def _detect_from_class_names(self, tree: ast.AST) -> Optional[str]:
        """
        Detect data structure from class names (primary method)
        
        Args:
            tree: AST tree to analyze
            
        Returns:
            Detected data structure type or None
        """
        # Extract class names
        classes = self.ast_parser.extract_classes(tree)
        class_names = list(classes.keys())
        
        # Check each class name against patterns
        for ds_type, patterns in self.CLASS_NAME_PATTERNS.items():
            for class_name in class_names:
                for pattern in patterns:
                    if pattern.lower() in class_name.lower():
                        return ds_type
        
        return None
    
    def _detect_from_methods(self, tree: ast.AST) -> Optional[str]:
        """
        Detect data structure from method signatures (fallback method)
        
        Args:
            tree: AST tree to analyze
            
        Returns:
            Detected data structure type or None
        """
        # Extract method calls and method definitions
        method_calls = set()
        method_definitions = set()
        
        for node in ast.walk(tree):
            # Collect method calls
            if isinstance(node, ast.Call):
                if isinstance(node.func, ast.Attribute):
                    method_name = node.func.attr
                    method_calls.add(method_name)
                elif isinstance(node.func, ast.Name):
                    # Function call (not method)
                    pass
            
            # Collect method definitions
            if isinstance(node, ast.FunctionDef):
                method_definitions.add(node.name)
        
        # Combine method calls and definitions
        all_methods = method_calls.union(method_definitions)
        
        # Check each data structure pattern
        best_match = None
        best_score = 0
        
        for ds_type, pattern_info in self.METHOD_PATTERNS.items():
            pattern_methods = pattern_info["methods"]
            min_matches = pattern_info["min_matches"]
            
            # Count matching methods
            matches = sum(1 for method in all_methods if method.lower() in [m.lower() for m in pattern_methods])
            
            if matches >= min_matches and matches > best_score:
                best_score = matches
                best_match = ds_type
        
        return best_match
    
    def get_detection_confidence(self, code: str) -> Dict[str, Any]:
        """
        Get detection result with confidence score
        
        Args:
            code: Python code to analyze
            
        Returns:
            Dictionary with detected_type, confidence, and method used
        """
        try:
            tree = self.ast_parser.parse_code(code)
            
            # Try class name detection
            class_type = self._detect_from_class_names(tree)
            if class_type:
                return {
                    "detected_type": class_type,
                    "confidence": "high",
                    "method": "class_name",
                    "alternative": None
                }
            
            # Try method signature analysis
            method_type = self._detect_from_methods(tree)
            if method_type:
                return {
                    "detected_type": method_type,
                    "confidence": "medium",
                    "method": "method_signature",
                    "alternative": None
                }
            
            return {
                "detected_type": None,
                "confidence": "none",
                "method": None,
                "alternative": None
            }
            
        except Exception as e:
            return {
                "detected_type": None,
                "confidence": "none",
                "method": None,
                "alternative": str(e)
            }




