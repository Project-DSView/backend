import ast
from typing import List, Set, Dict, Callable, Any
from app.schemas.playground import ExecutionStepSchema
from app.utils.messages_th import get_class_defined_message, get_message
from app.services.simulators.operations.error_handler import ErrorHandler


class ASTParser:
    """Handles AST parsing and code analysis"""
    
    @staticmethod
    def parse_code(code: str) -> ast.AST:
        """Parse code string into AST"""
        try:
            return ast.parse(code)
        except SyntaxError as e:
            # Extract detailed syntax error information
            line_number = e.lineno if hasattr(e, 'lineno') and e.lineno else 0
            offset = e.offset if hasattr(e, 'offset') and e.offset else 0
            text = e.text if hasattr(e, 'text') and e.text else ""
            filename = getattr(e, 'filename', None) or "<string>"
            
            # Format error message with line number and code
            error_info = ErrorHandler.format_error(e, line_number, text, offset, filename)
            
            # Create a custom exception that preserves syntax error details
            syntax_error = ValueError(error_info["python_style_message"])
            # Attach error details to the exception for later use
            syntax_error.error_type = error_info["error_type"]
            syntax_error.error_message = error_info["error_message"]
            syntax_error.thai_message = error_info["thai_message"]
            syntax_error.python_style_message = error_info["python_style_message"]
            syntax_error.line_number = line_number
            syntax_error.offset = offset
            syntax_error.code_line = text
            syntax_error.lineno = line_number  # For compatibility
            
            raise syntax_error
    
    @staticmethod
    def extract_class_attributes(class_node: ast.ClassDef) -> Dict[str, Dict]:
        """
        Extract class attributes from __init__ method.
        Returns a dict of attribute names with their inferred roles.
        
        Roles:
        - data_value: Stores actual data (name, data, value, val)
        - next_pointer: Points to next node (next, pointer)
        - prev_pointer: Points to previous node (prev, previous) 
        - head_pointer: Points to head of list (head, first)
        - tail_pointer: Points to tail of list (tail, last)
        - size_counter: Counts elements (count, size, length)
        - left_child: Left child in tree (left)
        - right_child: Right child in tree (right)
        - root_pointer: Root of tree (root)
        - generic: Unknown/generic attribute
        """
        attributes = {}
        
        # Find __init__ method
        init_method = None
        for item in class_node.body:
            if isinstance(item, ast.FunctionDef) and item.name == "__init__":
                init_method = item
                break
        
        if not init_method:
            return attributes
        
        # Extract self.xxx = yyy assignments
        for stmt in ast.walk(init_method):
            if isinstance(stmt, ast.Assign):
                for target in stmt.targets:
                    if isinstance(target, ast.Attribute):
                        if isinstance(target.value, ast.Name) and target.value.id == "self":
                            attr_name = target.attr
                            
                            # Infer role from attribute name
                            role = "generic"
                            attr_lower = attr_name.lower()
                            
                            # Data value patterns
                            if attr_lower in ["data", "value", "val", "name", "key", "info"]:
                                role = "data_value"
                            # Next pointer patterns
                            elif attr_lower in ["next", "pointer", "link", "next_node"]:
                                role = "next_pointer"
                            # Previous pointer patterns
                            elif attr_lower in ["prev", "previous", "prev_node", "back"]:
                                role = "prev_pointer"
                            # Head pointer patterns
                            elif attr_lower in ["head", "first", "front", "start"]:
                                role = "head_pointer"
                            # Tail pointer patterns
                            elif attr_lower in ["tail", "last", "rear", "end"]:
                                role = "tail_pointer"
                            # Size counter patterns
                            elif attr_lower in ["count", "size", "length", "num_elements", "_size"]:
                                role = "size_counter"
                            # Tree patterns
                            elif attr_lower in ["left", "left_child"]:
                                role = "left_child"
                            elif attr_lower in ["right", "right_child"]:
                                role = "right_child"
                            elif attr_lower in ["root", "root_node"]:
                                role = "root_pointer"
                            
                            attributes[attr_name] = {
                                "name": attr_name,
                                "role": role
                            }
        
        return attributes
    
    @staticmethod
    def extract_classes(tree: ast.AST) -> Dict[str, Dict]:
        """Extract class definitions from AST with their attributes"""
        classes = {}
        for node in ast.walk(tree):
            if isinstance(node, ast.ClassDef):
                # Extract attributes from __init__
                attributes = ASTParser.extract_class_attributes(node)
                
                classes[node.name] = {
                    "type": "class",
                    "methods": [method.name for method in node.body if isinstance(method, ast.FunctionDef)],
                    "attributes": attributes,
                    "defined": True,
                    "line_number": node.lineno
                }
        return classes
    
    @staticmethod
    def get_executable_lines(tree: ast.AST) -> Set[int]:
        """Get line numbers that should be executed (not inside class definitions)"""
        executable_lines = set()
        
        # Add all top-level statements
        for node in tree.body:
            if not isinstance(node, (ast.ClassDef, ast.FunctionDef)):
                if hasattr(node, 'lineno'):
                    executable_lines.add(node.lineno)
        
        return executable_lines
    
    @staticmethod
    def create_class_definition_steps(classes: Dict[str, Dict], context: Dict, 
                                    get_instance_display_func: Callable = None) -> List[ExecutionStepSchema]:
        """Create execution steps for class definitions"""
        steps = []
        step_number = 1
        
        # Default instance display function if none provided
        if get_instance_display_func is None:
            get_instance_display_func = lambda x: x.get("data", []) if isinstance(x, dict) else []
        
        # Ensure context has proper structure with safe initialization
        if not isinstance(context, dict):
            context = {}
        
        # Initialize all required context keys as dictionaries/lists
        if "classes" not in context or not isinstance(context["classes"], dict):
            context["classes"] = {}
        if "instances" not in context or not isinstance(context["instances"], dict):
            context["instances"] = {}
        if "stdout" not in context or not isinstance(context["stdout"], list):
            context["stdout"] = []
        if "variables" not in context or not isinstance(context["variables"], dict):
            context["variables"] = {}
        if "nodes" not in context or not isinstance(context["nodes"], dict):
            context["nodes"] = {}
        
        for class_name, class_info in classes.items():
            # Safely update classes context
            context["classes"][class_name] = class_info
            
            # Create base state with safe access
            state = {
                "message": get_class_defined_message(class_name),
                "instances": {},
                "active": context.get("active_instance"),
                "stdout": context.get("stdout", []).copy()
            }
            
            # Safely add instances with proper type checking
            instances = context.get("instances", {})
            if isinstance(instances, dict):
                for k, v in instances.items():
                    try:
                        if isinstance(v, dict):
                            state["instances"][k] = get_instance_display_func(v)
                        else:
                            state["instances"][k] = []
                    except Exception:
                        # If get_instance_display_func fails, provide empty list
                        state["instances"][k] = []
            
            # Add data structure specific fields only if they exist and have proper structure
            nodes = context.get("nodes", {})
            if isinstance(nodes, dict) and nodes:
                state["nodes"] = {}
                for k, v in nodes.items():
                    try:
                        if isinstance(v, dict):
                            state["nodes"][k] = v.get("name", "")
                        else:
                            state["nodes"][k] = str(v) if v is not None else ""
                    except Exception:
                        state["nodes"][k] = ""
            
            # Add variables if they exist and are properly structured
            variables = context.get("variables", {})
            if isinstance(variables, dict) and variables:
                # Only include safe variable representations
                safe_variables = {}
                for k, v in variables.items():
                    try:
                        # Convert to safe string representation
                        if isinstance(v, (str, int, float, bool)):
                            safe_variables[k] = v
                        elif v is None:
                            safe_variables[k] = None
                        else:
                            safe_variables[k] = str(v)
                    except Exception:
                        safe_variables[k] = "undefined"
                
                if safe_variables:
                    state["variables"] = safe_variables
            
            # Only add linkedlist field for linked list simulators
            # Check if context indicates this is a linkedlist simulator
            if context.get("include_linkedlist", False) or hasattr(context, 'data_structure_type'):
                linkedlist = context.get("linkedlist", [])
                if isinstance(linkedlist, list):
                    state["linkedlist"] = linkedlist.copy()
                else:
                    state["linkedlist"] = []
            
            # Create execution step with safe line number
            line_number = class_info.get("line_number", 1)
            if not isinstance(line_number, int):
                line_number = 1
            
            steps.append(ExecutionStepSchema(
                stepNumber=step_number,
                line=line_number,
                code=f"class {class_name}:",
                state=state
            ))
            step_number += 1
        
        return steps
    
    @staticmethod
    def classify_operations(tree: ast.AST) -> Dict[str, Any]:
        """
        Classify operations in AST for visualization approach
        
        Args:
            tree: AST tree to analyze
            
        Returns:
            Dictionary with operation classification
        """
        classification = {
            "has_class_definitions": False,
            "has_method_calls": False,
            "has_assignments": False,
            "has_loops": False,
            "has_conditionals": False,
            "has_input": False,
            "has_print": False,
            "method_calls": [],
            "class_names": [],
            "function_names": [],
            "variable_assignments": [],
            "visualization_type": "basic"
        }
        
        for node in ast.walk(tree):
            # Check for class definitions
            if isinstance(node, ast.ClassDef):
                classification["has_class_definitions"] = True
                classification["class_names"].append(node.name)
            
            # Check for method calls
            elif isinstance(node, ast.Call):
                classification["has_method_calls"] = True
                if isinstance(node.func, ast.Name):
                    func_name = node.func.id
                    classification["function_names"].append(func_name)
                    if func_name == "input":
                        classification["has_input"] = True
                    elif func_name == "print":
                        classification["has_print"] = True
                elif isinstance(node.func, ast.Attribute):
                    # Method call on object
                    method_name = node.func.attr
                    classification["method_calls"].append(method_name)
                    if isinstance(node.func.value, ast.Name):
                        obj_name = node.func.value.id
                        classification["method_calls"].append(f"{obj_name}.{method_name}")
            
            # Check for assignments
            elif isinstance(node, ast.Assign):
                classification["has_assignments"] = True
                for target in node.targets:
                    if isinstance(target, ast.Name):
                        classification["variable_assignments"].append(target.id)
            
            # Check for loops
            elif isinstance(node, (ast.For, ast.While)):
                classification["has_loops"] = True
            
            # Check for conditionals
            elif isinstance(node, ast.If):
                classification["has_conditionals"] = True
        
        # Determine visualization type based on operations
        if classification["has_class_definitions"]:
            if any("LinkedList" in name for name in classification["class_names"]):
                classification["visualization_type"] = "linked_list"
            elif any("Stack" in name for name in classification["class_names"]):
                classification["visualization_type"] = "stack"
            elif any("Queue" in name for name in classification["class_names"]):
                classification["visualization_type"] = "queue"
            elif any("Tree" in name or "BST" in name for name in classification["class_names"]):
                classification["visualization_type"] = "tree"
            elif any("Graph" in name for name in classification["class_names"]):
                classification["visualization_type"] = "graph"
        
        return classification
    
    @staticmethod
    def extract_ast_node_metadata(tree: ast.AST) -> List[Dict[str, Any]]:
        """
        Extract detailed AST node metadata for visualization and learning
        
        Args:
            tree: AST tree to analyze
            
        Returns:
            List of AST node metadata with educational information
        """
        nodes = []
        
        for node in ast.walk(tree):
            node_metadata = {
                "type": type(node).__name__,
                "type_display": type(node).__name__.replace("ast.", ""),
                "line": getattr(node, 'lineno', None),
                "col_offset": getattr(node, 'col_offset', None),
            }
            
            # Add type-specific information
            if isinstance(node, ast.Call):
                node_metadata["category"] = "function_call"
                if isinstance(node.func, ast.Name):
                    node_metadata["function_name"] = node.func.id
                    node_metadata["is_builtin"] = node.func.id in ['print', 'input', 'len', 'range', 'str', 'int']
                elif isinstance(node.func, ast.Attribute):
                    node_metadata["method_name"] = node.func.attr
                    if isinstance(node.func.value, ast.Name):
                        node_metadata["object_name"] = node.func.value.id
                        node_metadata["category"] = "method_call"
            
            elif isinstance(node, ast.Assign):
                node_metadata["category"] = "assignment"
                if node.targets:
                    target = node.targets[0]
                    if isinstance(target, ast.Name):
                        node_metadata["variable_name"] = target.id
                    elif isinstance(target, ast.Attribute):
                        node_metadata["category"] = "attribute_assignment"
            
            elif isinstance(node, ast.ClassDef):
                node_metadata["category"] = "class_definition"
                node_metadata["class_name"] = node.name
            
            elif isinstance(node, ast.FunctionDef):
                node_metadata["category"] = "function_definition"
                node_metadata["function_name"] = node.name
            
            elif isinstance(node, ast.If):
                node_metadata["category"] = "conditional"
            
            elif isinstance(node, (ast.For, ast.While)):
                node_metadata["category"] = "loop"
            
            elif isinstance(node, ast.BinOp):
                node_metadata["category"] = "binary_operation"
                node_metadata["operator"] = type(node.op).__name__
            
            elif isinstance(node, ast.Compare):
                node_metadata["category"] = "comparison"
            
            nodes.append(node_metadata)
        
        return nodes