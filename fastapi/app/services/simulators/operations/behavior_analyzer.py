import ast
import re
from typing import List, Dict, Any, Optional
from app.utils.messages_th import (
    get_traverse_message, get_insert_message, get_delete_message, 
    get_error_message, get_message
)


class BehaviorAnalyzer:
    """Analyzes method behavior from AST logic with enhanced method name detection"""
    
    def __init__(self, context: Dict[str, Any]):
        self.context = context
        self.method_definitions = {}  # Store parsed method definitions
        # Import NodeManager for actual operations
        from app.services.simulators.operations.node_manager import NodeManager
        self.node_manager = NodeManager(context)
        
        # Ensure context has proper structure
        self._ensure_context_structure()
    
    def _ensure_context_structure(self):
        """Ensure context has the proper dictionary structure"""
        if not isinstance(self.context.get("classes"), dict):
            self.context["classes"] = {}
        if not isinstance(self.context.get("instances"), dict):
            self.context["instances"] = {}
        if not isinstance(self.context.get("variables"), dict):
            self.context["variables"] = {}
        if not isinstance(self.context.get("nodes"), dict):
            self.context["nodes"] = {}
        if not isinstance(self.context.get("stdout"), list):
            self.context["stdout"] = []
            
    def parse_class_methods(self, code: str):
        """Parse all method definitions from the code and analyze their behavior"""
        try:
            tree = ast.parse(code)
            
            for node in ast.walk(tree):
                if isinstance(node, ast.ClassDef):
                    class_name = node.name
                    # Ensure classes dict exists and is properly initialized
                    self._ensure_context_structure()
                    
                    # Initialize class entry if not exists, but don't overwrite if it does
                    if class_name not in self.context["classes"]:
                        self.context["classes"][class_name] = {}
                    
                    # Ensure methods is a dictionary for detailed analysis
                    # We always reset methods because we are about to re-analyze them
                    self.context["classes"][class_name]["methods"] = {}
                    
                    for item in node.body:
                        if isinstance(item, ast.FunctionDef):
                            method_info = self._analyze_method_behavior_from_ast(item)
                            method_key = f"{class_name}.{item.name}"
                            self.method_definitions[method_key] = method_info
                            self.context["classes"][class_name]["methods"][item.name] = method_info
        
        except Exception as e:
            print(f"Error parsing methods: {e}")
    
    def _analyze_method_behavior_from_ast(self, method_node: ast.FunctionDef) -> Dict[str, Any]:
        """Analyze what a method actually does based on its AST with enhanced classification"""
        behavior = {
            "name": method_node.name,
            "params": [arg.arg for arg in method_node.args.args[1:]],  # Skip 'self'
            "behavior_type": "unknown",
            "operations": [],
            "has_loop": False,
            "has_print": False,
            "creates_new_node": False,
            "modifies_head": False,
            "modifies_count": False,
            "modifies_next_pointers": False,
            "checks_empty_list": False,
            "traverses_list": False,
            "finds_target": False,
            "deletes_node": False,
            "deletes_node": False,
            "method_name_hints": self._get_method_name_hints(method_node.name),
            # Stack specific flags
            "uses_append": False,
            "uses_pop": False,
            "uses_peek": False,
            "returns_size": False,
            "checks_empty": False,
            "uses_peek_back": False,
            "uses_peek_front": False,
            "uses_pop_0": False,
            "uses_popleft": False,
            "is_bst_insert": False,
            "is_bst_search": False,
            "is_bst_delete": False,
            "is_graph_add_vertex": False,
            "is_graph_add_edge": False,
            "is_graph_bfs": False,
            "is_graph_dfs": False
        }
        
        # Deep analysis of method body
        self._analyze_statements(method_node.body, behavior)
        
        # Determine behavior type based on comprehensive analysis
        behavior["behavior_type"] = self._classify_behavior_by_logic(behavior)
        return behavior
    
    def _get_method_name_hints(self, method_name: str) -> Dict[str, bool]:
        """Get hints from method name about what it might do"""
        name_lower = method_name.lower()
        return {
            "is_insert": any(word in name_lower for word in ['insert', 'add', 'push', 'append']),
            "is_delete": any(word in name_lower for word in ['delete', 'remove', 'pop', 'del']),
            "is_traverse": any(word in name_lower for word in ['traverse', 'print', 'display', 'show', 'list']),
            "is_front": any(word in name_lower for word in ['front', 'first', 'head', 'begin']),
            "is_last": any(word in name_lower for word in ['last', 'end', 'tail', 'back']),
            "is_getter": name_lower.startswith('get') or name_lower in ['size', 'length', 'count', 'empty'],
            "is_init": method_name == '__init__'
        }
    
    def _analyze_statements(self, statements: List[ast.stmt], behavior: Dict[str, Any]):
        """Recursively analyze all statements in the method"""
        for stmt in statements:
            if isinstance(stmt, ast.While) or isinstance(stmt, ast.For):
                behavior["has_loop"] = True
                behavior["traverses_list"] = True
                
                # Detect loop pattern that finds tail node
                # Pattern: while start.next != None or while current.next
                if isinstance(stmt, ast.While):
                    if self._is_tail_finding_loop(stmt):
                        behavior["traverses_to_end"] = True
                
                # Analyze loop body
                if hasattr(stmt, 'body'):
                    self._analyze_statements(stmt.body, behavior)
            
            elif isinstance(stmt, ast.If):
                # Check if it's checking for empty list
                if self._is_empty_list_check(stmt):
                    behavior["checks_empty_list"] = True
                # Check if it's finding a target
                if self._is_target_finding(stmt):
                    behavior["finds_target"] = True
                # Analyze if/else bodies
                if hasattr(stmt, 'body'):
                    self._analyze_statements(stmt.body, behavior)
                if hasattr(stmt, 'orelse'):
                    self._analyze_statements(stmt.orelse, behavior)
            
            elif isinstance(stmt, ast.Assign):
                self._analyze_assignment(stmt, behavior)
            
            elif isinstance(stmt, ast.Expr) and isinstance(stmt.value, ast.Call):
                # Check for method calls (append, pop, etc.)
                if isinstance(stmt.value.func, ast.Attribute):
                    method_name = stmt.value.func.attr
                    if method_name == 'append':
                        behavior["uses_append"] = True
                    elif method_name == 'pop':
                        behavior["uses_pop"] = True
                        # Check if it's pop(0)
                        if len(stmt.value.args) > 0 and isinstance(stmt.value.args[0], ast.Constant) and stmt.value.args[0].value == 0:
                            behavior["uses_pop_0"] = True
                    elif method_name == 'popleft':
                        behavior["uses_popleft"] = True
                    elif method_name == 'insert':
                         # Check for BST insert patterns (will add more robust checks later if needed)
                         pass
                    elif method_name == 'add_vertex': # Simple name check for graph for now
                         behavior["is_graph_add_vertex"] = True
                    elif method_name == 'add_edge':
                         behavior["is_graph_add_edge"] = True
                
                if hasattr(stmt.value.func, 'id') and stmt.value.func.id == 'print':
                    behavior["has_print"] = True
            
            elif isinstance(stmt, ast.Return):
                if isinstance(stmt.value, ast.Attribute) and stmt.value.attr == 'size':
                    behavior["returns_size"] = True
                # Check for len()
                if isinstance(stmt.value, ast.Call) and hasattr(stmt.value.func, 'id') and stmt.value.func.id == 'len':
                    behavior["returns_size"] = True
                # Check for [-1] (peek/back) and [0] (front)
                if isinstance(stmt.value, ast.Subscript):
                    # Check for [-1]
                    if (isinstance(stmt.value.slice, ast.UnaryOp) and 
                        isinstance(stmt.value.slice.op, ast.USub) and
                        isinstance(stmt.value.slice.operand, ast.Constant) and
                        stmt.value.slice.operand.value == 1):
                        behavior["uses_peek_back"] = True
                    
                    # Check for [0]
                    if (isinstance(stmt.value.slice, ast.Constant) and 
                        stmt.value.slice.value == 0):
                        behavior["uses_peek_front"] = True

                    # Python < 3.9 style (Index) - handling simplisticly for now
                    if hasattr(stmt.value.slice, 'value'): # ast.Index
                         if isinstance(stmt.value.slice.value, ast.UnaryOp):
                             if (isinstance(stmt.value.slice.value.op, ast.USub) and 
                                 stmt.value.slice.value.operand.value == 1):
                                 behavior["uses_peek_back"] = True
                         elif isinstance(stmt.value.slice.value, ast.Constant) and stmt.value.slice.value.value == 0:
                             behavior["uses_peek_front"] = True

    
    def _analyze_assignment(self, stmt: ast.Assign, behavior: Dict[str, Any]):
        """Analyze assignment statements to understand what's being modified"""
        for target in stmt.targets:
            if isinstance(target, ast.Attribute):
                # Check what attribute is being modified
                if target.attr == "head":
                    behavior["modifies_head"] = True
                elif target.attr == "count":
                    behavior["modifies_count"] = True
                elif target.attr == "next":
                    behavior["modifies_next_pointers"] = True
            
            elif isinstance(target, ast.Name):
                # Check if creating new node
                if isinstance(stmt.value, ast.Call):
                    if (hasattr(stmt.value.func, 'id') and 
                        stmt.value.func.id in ['DataNode', 'Node']):
                        behavior["creates_new_node"] = True
                    elif (hasattr(stmt.value.func, 'attr') and 
                          target.id in ['new_node', 'node']):
                        behavior["creates_new_node"] = True
    
    def _is_empty_list_check(self, if_stmt: ast.If) -> bool:
        """Check if this is testing for empty list (self.head is None)"""
        test = if_stmt.test
        if isinstance(test, ast.Compare):
            if (isinstance(test.left, ast.Attribute) and 
                test.left.attr == "head" and
                len(test.comparators) > 0 and
                isinstance(test.comparators[0], ast.Constant) and
                test.comparators[0].value is None):
                return True
        return False
    
    def _is_target_finding(self, if_stmt: ast.If) -> bool:
        """Check if this is looking for a specific target node"""
        test = if_stmt.test
        if isinstance(test, ast.Compare):
            # Look for patterns like current.name == target_name
            if (isinstance(test.left, ast.Attribute) and 
                test.left.attr in ['name', 'data', 'value']):
                return True
        return False
    
    def _is_tail_finding_loop(self, while_stmt: ast.While) -> bool:
        """Check if this while loop is finding the tail node"""
        test = while_stmt.test
        if isinstance(test, ast.Compare):
            # Pattern: while start.next != None or while current.next != None
            if isinstance(test.left, ast.Attribute):
                # Check if it's accessing .next attribute
                if test.left.attr == "next":
                    # Check if comparing with None
                    if len(test.comparators) > 0:
                        comparator = test.comparators[0]
                        if isinstance(comparator, ast.Constant) and comparator.value is None:
                            # Check operator: != or is not
                            if isinstance(test.ops[0], (ast.NotEq, ast.IsNot)):
                                return True
        return False
    
    def _classify_behavior_by_logic(self, behavior: Dict[str, Any]) -> str:
        """Classify behavior based on comprehensive logic analysis and method name hints"""
        hints = behavior["method_name_hints"]
        
        # Constructor behavior
        if hints["is_init"]:
            return "constructor"
        
        # Print/Display operation: has print and loop, doesn't modify structure
        if ((behavior["has_print"] and behavior["traverses_list"]) or hints["is_traverse"]) and \
           not behavior["modifies_head"] and not behavior["creates_new_node"]:
            return "traverse"
        
        # Insert operations - Logic-First Approach: ตรวจสอบ actual behavior ก่อนเสมอ
        if behavior["creates_new_node"]:
            # Priority 1: Insert at end - มี loop หา tail, ไม่ modify head โดยตรง
            # Pattern: while loop หา node สุดท้าย แล้วแทรกท้าย
            # Logic: มี loop + modify next pointer แต่ไม่ modify head โดยตรง
            if behavior["has_loop"] and not behavior["modifies_head"] and behavior["modifies_next_pointers"]:
                return "insert_last"  # แสดงผลเป็น insert_last แม้ชื่อ method จะเป็น insertFront
            
            # Priority 2: Insert at front - modify head โดยตรง, ไม่มี loop
            # Pattern: สร้าง node ใหม่ แล้วตั้งเป็น head ทันที
            # Logic: modify head โดยตรง + ไม่มี loop
            if behavior["modifies_head"] and not behavior["has_loop"]:
                return "insert_front"
            
            # Priority 3: Insert before/after - has loop, finds target node
            # Pattern: loop หา target node แล้วแทรกก่อน/หลัง
            # Logic: มี loop + หา target + modify next pointer
            if behavior["has_loop"] and behavior["finds_target"] and behavior["modifies_next_pointers"]:
                return "insert_before"
            
            # Fallback: ใช้ method name hints เฉพาะเมื่อ logic ไม่ชัดเจน
            # (กรณีนี้ควรเกิดขึ้นน้อยมาก เพราะ logic analysis ควรครอบคลุมแล้ว)
            if hints["is_insert"]:
                if hints["is_front"]:
                    return "insert_front"
                elif hints["is_last"]:
                    return "insert_last"
                elif len(behavior["params"]) >= 2:
                    return "insert_before"
                else:
                    return "insert_front"  # Default to front if unclear
        
        # Delete operation: finds target, modifies structure but doesn't create nodes
        if (behavior["finds_target"] and not behavior["creates_new_node"] and 
            (behavior["modifies_head"] or behavior["modifies_next_pointers"])) or hints["is_delete"]:
            return "delete"
        
        # Simple getter methods
        if hints["is_getter"] or (not behavior["creates_new_node"] and not behavior["modifies_head"] and 
            not behavior["modifies_count"] and not behavior["has_loop"]):
            return "getter"
        
        # Fallback based on method name if behavior analysis is unclear
        if hints["is_traverse"]:
            return "traverse"
        elif hints["is_insert"]:
            if hints["is_front"]:
                return "insert_front"
            elif hints["is_last"]:
                return "insert_last"
            else:
                return "insert_positional"
        elif hints["is_delete"]:
            return "delete"
        
        # Default fallback
        return "unknown"

    def _classify_behavior_by_logic(self, behavior: Dict[str, Any]) -> str:
        """Classify behavior based on comprehensive logic analysis and method name hints"""
        hints = behavior["method_name_hints"]
        
        # Constructor behavior
        if hints["is_init"]:
            return "constructor"

        # --- Stack Specific Logic ---
        # Prioritize logic detection for Stack operations
        
        # Push: Uses append()
        # Logic: calling .append() on a list is the definition of push for ArrayStack
        if behavior.get("uses_append"):
            return "push"
            
        # Pop: Uses pop()
        # Logic: calling .pop() on a list is the definition of pop for ArrayStack
        if behavior.get("uses_pop"):
            return "pop"
            
        # StackTop/Peek: Accesses list[-1]
        # Logic: accessing the last element via [-1]
        if behavior.get("uses_peek"):
            return "stackTop"
            
        # Size: Check for len() usage or return size
        if hints["is_getter"] and (behavior.get("returns_size") or "size" in behavior["name"].lower()):
             return "size"
             
        # Empty: Checks length == 0
        if hints["is_getter"] and (behavior.get("checks_empty") or "empty" in behavior["name"].lower()):
            return "is_empty"
            
        # Print/Traversal
        if behavior["has_print"] and not behavior["creates_new_node"]:
            return "printStack"

        # --- Queue Specific Logic ---
        
        # Enqueue: Uses append() (same as push but context matters, we return 'enqueue' if asked or generic 'push/enqueue')
        # We will handle the differentiation in the simulator specific mapping or allow both keys
        
        # Dequeue: Uses pop(0)
        # Dequeue: Uses pop(0) or popleft()
        if behavior.get("uses_pop_0") or behavior.get("uses_popleft"):
            return "dequeue"
            
        # Front: Accesses [0]
        if behavior.get("uses_peek_front"):
            return "front"
            
        # Back: Accesses [-1] (same as stackTop but for queue context it's back)
        if behavior.get("uses_peek_back"):
            return "back"

        # --- BST Specific Logic ---
        # (Simplified logic based on names + recursive patterns which are hard to detect solely by simple flags)
        # We rely heavily on hints for complex BST operations unless we do partial control flow graph analysis
        
        if hints.get("is_insert"):
            return "insert"
        if hints.get("is_delete"):
            return "delete"
        if hints.get("is_traverse"):
             # Differentiate traversal types if possible, else default to inorder or based on name
             if "pre" in behavior["name"].lower(): return "preorder"
             if "post" in behavior["name"].lower(): return "postorder"
             return "inorder" # Default
        if "min" in behavior["name"].lower(): return "findMin"
        if "max" in behavior["name"].lower(): return "findMax"

        # --- Graph Specific Logic ---
        if behavior.get("is_graph_add_vertex") or "add_vertex" in behavior["name"].lower():
            return "add_vertex"
        if behavior.get("is_graph_add_edge") or "add_edge" in behavior["name"].lower():
            return "add_edge"
        if "bfs" in behavior["name"].lower():
            return "bfs"
        if "dfs" in behavior["name"].lower():
            return "dfs"

        # --- Linked List Logic (Existing) ---
        
        # Print/Display operation: has print and loop, doesn't modify structure
        if ((behavior["has_print"] and behavior["traverses_list"]) or hints["is_traverse"]) and \
           not behavior["modifies_head"] and not behavior["creates_new_node"]:
            return "traverse"
        
        # Insert operations
        if behavior["creates_new_node"]:
            if behavior["has_loop"] and not behavior["modifies_head"] and behavior["modifies_next_pointers"]:
                return "insert_last"
            
            if behavior["modifies_head"] and not behavior["has_loop"]:
                return "insert_front"
            
            if behavior["has_loop"] and behavior["finds_target"] and behavior["modifies_next_pointers"]:
                return "insert_before"
            
            if hints["is_insert"]:
                if hints["is_front"]: return "insert_front"
                elif hints["is_last"]: return "insert_last"
                else: return "insert_front"
        
        # Delete operation
        if (behavior["finds_target"] and not behavior["creates_new_node"] and 
            (behavior["modifies_head"] or behavior["modifies_next_pointers"])) or hints["is_delete"]:
            return "delete"
        
        # Simple getter methods
        if hints["is_getter"] or (not behavior["creates_new_node"] and not behavior["modifies_head"] and 
            not behavior["modifies_count"] and not behavior["has_loop"]):
            return "getter"
        
        return "unknown"

    
    def execute_method_by_behavior(self, instance_name: str, method_name: str, params: str) -> str:
        """Execute method based on its analyzed behavior and name"""
        
        # Ensure context structure
        self._ensure_context_structure()
        
        # Get the class type of the instance
        instances = self.context.get("instances", {})
        if not isinstance(instances, dict) or instance_name not in instances:
            return f"Instance '{instance_name}' not found"
            
        instance = instances[instance_name]
        if not isinstance(instance, dict):
            return f"Instance '{instance_name}' has invalid structure"
            
        class_type = instance.get("class_type")
        
        # Look up the method behavior
        method_key = f"{class_type}.{method_name}"
        if method_key not in self.method_definitions:
            return f"Method {method_name} not found in {class_type}"
        
        method_info = self.method_definitions[method_key]
        behavior_type = method_info["behavior_type"]
        
        # Execute based on behavior, not name
        if behavior_type == "constructor":
            return get_message("constructor_called", method_name=method_name, instance_name=instance_name)
        
        elif behavior_type == "traverse":
            return self._execute_traverse_behavior(instance, instance_name, method_name)
        
        elif behavior_type == "insert_front":
            if params:
                data = params.strip('"\'')
                return self._execute_insert_front_behavior(instance, instance_name, method_name, data)
            return get_error_message("method_requires_data", method_name=method_name)
        
        elif behavior_type == "insert_last":
            if params:
                data = params.strip('"\'')
                return self._execute_insert_last_behavior(instance, instance_name, method_name, data)
            return get_error_message("method_requires_data", method_name=method_name)
        
        elif behavior_type in ["insert_before", "insert_positional"]:
            param_parts = [p.strip().strip('"\'') for p in params.split(',')]
            if len(param_parts) == 2:
                target, data = param_parts
                return self._execute_insert_before_behavior(instance, instance_name, method_name, target, data)
            elif len(param_parts) == 1 and params:
                data = param_parts[0]
                # Default to insert_last if only one parameter
                return self._execute_insert_last_behavior(instance, instance_name, method_name, data)
            return get_error_message("method_requires_params", method_name=method_name)
        
        elif behavior_type == "delete":
            if params:
                target = params.strip('"\'')
                return self._execute_delete_behavior(instance, instance_name, method_name, target)
            return get_error_message("method_requires_target", method_name=method_name)
        
        elif behavior_type == "getter":
            return get_message("getter_returned", instance_name=instance_name, method_name=method_name)
        
        else:
            # Fallback for Stack/Queue inserts
            if params and method_info["creates_new_node"]:
                data = params.strip('"\'')
                if method_info["has_loop"] or any(word in method_name.lower() for word in ['last', 'end']):
                    return self._execute_insert_last_behavior(instance, instance_name, method_name, data)
                else:
                    return self._execute_insert_front_behavior(instance, instance_name, method_name, data)
            
            # Queue behaviors
            if behavior_type == "enqueue":
                return get_message("using_method", method_name=method_name) # Actual logic handled by NodeManager via Simulator calls
            if behavior_type == "dequeue":
                return get_message("using_method", method_name=method_name)
                
            return get_message("executed_unknown", method_name=method_name, instance_name=instance_name, behavior_type=behavior_type)
    
    def _execute_traverse_behavior(self, instance: Dict[str, Any], instance_name: str, method_name: str) -> str:
        """Execute traversal behavior regardless of method name"""
        display_data = self._get_instance_display(instance)
        if not display_data:
            print_output = self.context.get("stdout", [])
            if isinstance(print_output, list):
                print_output.append(get_message("traverse_empty"))
            return get_traverse_message(instance_name, method_name)
        else:
            traverse_output = "-> " + " -> ".join(display_data)
            print_output = self.context.get("stdout", [])
            if isinstance(print_output, list):
                print_output.append(traverse_output)
            return get_traverse_message(instance_name, method_name, traverse_output)
    
    def _execute_insert_front_behavior(self, instance: Dict[str, Any], instance_name: str, method_name: str, data: str) -> str:
        """Execute insert front behavior using NodeManager"""
        message = self.node_manager.insert_front(instance, data)
        return f"{message} {get_message('using_method', method_name=method_name)}"
    
    def _execute_insert_last_behavior(self, instance: Dict[str, Any], instance_name: str, method_name: str, data: str) -> str:
        """Execute insert last behavior using NodeManager"""
        message = self.node_manager.insert_last(instance, data)
        return f"{message} {get_message('using_method', method_name=method_name)}"
    
    def _execute_insert_before_behavior(self, instance: Dict[str, Any], instance_name: str, method_name: str, target: str, data: str) -> str:
        """Execute insert before behavior using NodeManager"""
        message = self.node_manager.insert_before(instance, target, data)
        return f"{message} {get_message('using_method', method_name=method_name)}"
    
    def _execute_delete_behavior(self, instance: Dict[str, Any], instance_name: str, method_name: str, target: str) -> str:
        """Execute delete behavior using NodeManager"""
        message = self.node_manager.delete_node(instance, target)
        return f"{message} {get_message('using_method', method_name=method_name)}"
    
    def _get_instance_display(self, instance: Dict[str, Any]) -> List[str]:
        """Get display representation of an instance"""
        if not isinstance(instance, dict):
            return []
            
        class_type = instance.get("class_type")
        if class_type in ["SinglyLinkedList", "DoublyLinkedList"]:
            return self._traverse_linked_list(instance)
        return instance.get("data", [])
    
    def _traverse_linked_list(self, instance: Dict[str, Any]) -> List[str]:
        """Traverse a linked list instance and return display data"""
        if not isinstance(instance, dict):
            return []
            
        result = []
        if instance.get("head") is None:
            return []
        
        current_node_id = instance["head"]
        visited = set()  # Prevent infinite loops
        nodes = self.context.get("nodes", {})
        
        if not isinstance(nodes, dict):
            return []
        
        while current_node_id is not None and current_node_id not in visited:
            visited.add(current_node_id)
            if current_node_id in nodes and isinstance(nodes[current_node_id], dict):
                node = nodes[current_node_id]
                result.append(node.get("name", ""))
                current_node_id = node.get("next")
            else:
                break
                
        return result