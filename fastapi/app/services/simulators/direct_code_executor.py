import ast
import json
import re
from typing import List, Dict, Any, Optional
from datetime import datetime

from app.schemas.playground import ExecutionStepSchema
from app.services.simulators.real_python_executor import RealPythonExecutor, ExecutionResult
from app.services.simulators.interactive_python_executor import InteractivePythonExecutor, InteractiveExecutionResult
from app.services.simulators.operations.ast_parser import ASTParser
from app.services.simulators.operations.explanation_generator import ExplanationGenerator


class DirectCodeExecutor:
    """
    Executes Python code directly without AST transformation
    - Execute code through RealPythonExecutor or InteractivePythonExecutor
    - Parse output and state from execution results
    - Create execution steps from results
    - Support Docker execution for security
    """
    
    def __init__(self, use_docker: bool = False, timeout: int = 60, memory: str = "256m"):
        """
        Initialize the direct code executor
        
        Args:
            use_docker: If True, use Docker for execution (more secure, isolated)
            timeout: Execution timeout in seconds
            memory: Memory limit for Docker (if used)
        """
        self.use_docker = use_docker
        self.timeout = timeout
        self.memory = memory
        self.real_executor = RealPythonExecutor(use_docker, timeout, memory)
        self.interactive_executor = InteractivePythonExecutor(use_docker, timeout, memory)
    
    def _has_input_call(self, code: str) -> bool:
        """Check if code contains input() calls"""
        return 'input(' in code or 'input (' in code
    
    def execute(
        self,
        code: str,
        stdin_data: str = "",
        data_structure_type: Optional[str] = None,
        input_values: Optional[List[str]] = None
    ) -> List[ExecutionStepSchema]:
        """
        Execute code directly and return execution steps
        
        Args:
            code: Python code to execute
            stdin_data: Input data (JSON string)
            data_structure_type: Type of data structure (for context)
            input_values: Pre-collected input values for input() calls
            
        Returns:
            List of execution steps with execution results
        """
        steps = []
        step_number = 1
        
        try:
            # Check if code has input() calls
            has_input = self._has_input_call(code)
            
            # Execute code
            # Check if we should force tracing (for simulators that need detailed steps)
            force_tracing = data_structure_type in ["stack", "queue", "linked_list", "singlylinkedlist", "doublylinkedlist", "binary_search_tree", "binarysearchtree", "graph", "tree", "directedgraph", "undirectedgraph"]
            
            # Execute code
            if has_input or force_tracing:
                # Use interactive executor for input() support AND detailed tracing
                # If no input values provided but we want tracing, pass empty list to use the value-based wrapper
                # which preserves line numbers better (via exec)
                effective_input_values = input_values if input_values is not None else []
                
                result = self.interactive_executor.execute_interactive(
                    code,
                    stdin_data,
                    input_callback=None,  # Not used when input_values provided
                    input_values=effective_input_values
                )
            else:
                # Use regular executor
                result = self.real_executor.execute(code, stdin_data)
            
            # Check for waiting signal in stdout
            waiting_signal = None
            if hasattr(result, 'stdout') and result.stdout:
                if "__WAITING_FOR_INPUT__:" in result.stdout:
                    parts = result.stdout.split("__WAITING_FOR_INPUT__:")
                    result.stdout = parts[0] # Remove signal from displayed stdout
                    try:
                        waiting_signal = json.loads(parts[1].strip())
                    except:
                        pass
            
            # Create execution steps from results
            if isinstance(result, InteractiveExecutionResult):
                # Always try to create steps from trace/output first (even if failed)
                steps = self._create_steps_from_interactive_result(
                    code, result, step_number, data_structure_type
                )
                
                # If execution failed, append error step
                if result.exit_code != 0:
                    # Create error step starting after the last trace step
                    next_step_num = step_number + len(steps)
                    error_steps = self._create_error_step(
                        code, result, next_step_num, data_structure_type
                    )
                    steps.extend(error_steps)
            else:
                if result.exit_code == 0:
                    steps = self._create_steps_from_result(
                        code, result, step_number, data_structure_type
                    )
                else:
                    steps = self._create_error_step(
                        code, result, step_number, data_structure_type
                    )
            
            # If waiting signal detected, add a "Waiting for Input" step
            if waiting_signal:
                waiting_step = ExecutionStepSchema(
                    stepNumber=len(steps) + 1,
                    line=waiting_signal.get("line", len(code.split('\n'))),
                    code=code.split('\n')[waiting_signal.get("line", 1) - 1].strip(),
                    state={
                        "message": f"Waiting for input: {waiting_signal.get('prompt', '')}",
                        "waiting_for_input": True,
                        "input_prompt": waiting_signal.get("prompt", ""),
                        "stdout": result.stdout,
                        "stdout": steps[-1].state.get("stdout", []) if steps else []
                    }
                )
                steps.append(waiting_step)
            
            return steps
            
            return steps
            
        except Exception as e:
            # Unexpected error
            import traceback
            error_step = ExecutionStepSchema(
                stepNumber=step_number,
                line=1,
                code=code.split('\n')[0] if code.split('\n') else code,
                state={
                    "error": f"Unexpected error: {str(e)}",
                    "error_type": "EXECUTION_ERROR",
                    "message": f"เกิดข้อผิดพลาดในการรันโค้ด: {str(e)}",
                    "traceback": traceback.format_exc()
                }
            )
            return [error_step]
    
    def _extract_class_metadata(self, code: str) -> Dict[str, Any]:
        """
        Extract class metadata from code for dynamic visualization labels.
        Returns metadata about class attributes organized by their roles.
        """
        try:
            tree = ASTParser.parse_code(code)
            classes = ASTParser.extract_classes(tree)
            
            # Organize by roles for frontend
            metadata = {
                "classes": {},
                "node_class": None,  # Class that represents a node (has next_pointer)
                "list_class": None,  # Class that represents the list (has head_pointer)
            }
            
            for class_name, class_info in classes.items():
                attrs = class_info.get("attributes", {})
                metadata["classes"][class_name] = {
                    "name": class_name,
                    "attributes": attrs,
                    "methods": class_info.get("methods", [])
                }
                
                # Detect node class (has next_pointer and data_value)
                has_next = any(a.get("role") == "next_pointer" for a in attrs.values())
                has_data = any(a.get("role") == "data_value" for a in attrs.values())
                if has_next and has_data:
                    metadata["node_class"] = class_name
                    # Extract specific attribute names
                    for attr_name, attr_info in attrs.items():
                        if attr_info.get("role") == "data_value":
                            metadata["data_attr"] = attr_name
                        elif attr_info.get("role") == "next_pointer":
                            metadata["next_attr"] = attr_name
                        elif attr_info.get("role") == "prev_pointer":
                            metadata["prev_attr"] = attr_name
                
                # Detect list class (has head_pointer)
                has_head = any(a.get("role") == "head_pointer" for a in attrs.values())
                if has_head:
                    metadata["list_class"] = class_name
                    for attr_name, attr_info in attrs.items():
                        if attr_info.get("role") == "head_pointer":
                            metadata["head_attr"] = attr_name
                        elif attr_info.get("role") == "tail_pointer":
                            metadata["tail_attr"] = attr_name
                        elif attr_info.get("role") == "size_counter":
                            metadata["count_attr"] = attr_name
                
                # Detect BST (has root_pointer)
                has_root = any(a.get("role") == "root_pointer" for a in attrs.values())
                if has_root:
                    metadata["tree_class"] = class_name
                    for attr_name, attr_info in attrs.items():
                        if attr_info.get("role") == "root_pointer":
                            metadata["root_attr"] = attr_name
                
                # Detect BST Node (has left_child and right_child)
                has_left = any(a.get("role") == "left_child" for a in attrs.values())
                has_right = any(a.get("role") == "right_child" for a in attrs.values())
                if has_left and has_right:
                    metadata["tree_node_class"] = class_name
                    for attr_name, attr_info in attrs.items():
                        if attr_info.get("role") == "left_child":
                            metadata["left_attr"] = attr_name
                        elif attr_info.get("role") == "right_child":
                            metadata["right_attr"] = attr_name
                        elif attr_info.get("role") == "data_value":
                            metadata["tree_data_attr"] = attr_name
            
            return metadata
        except Exception:
            return {}
    
    def _create_steps_from_result(
        self,
        code: str,
        result: ExecutionResult,
        step_number: int,
        data_structure_type: Optional[str] = None
    ) -> List[ExecutionStepSchema]:
        """Create execution steps from regular execution result"""
        steps = []
        
        # Extract class metadata for dynamic labels
        class_metadata = self._extract_class_metadata(code)
        
        # Track accumulated stdout across all steps - join outputs properly
        # Outputs without newlines should be concatenated, newlines start new entries
        accumulated_stdout = []
        current_line_buffer = ""  # Buffer for outputs without newlines
        
        # [NEW] If trace data is available, use it to generate detailed steps
        if isinstance(result, InteractiveExecutionResult) and hasattr(result, 'trace') and result.trace:
            code_lines = code.split('\n')
            
            # Map trace steps to ExecutionStepSchema
            for trace_step in result.trace:
                # Handle Exception Event
                if trace_step.get("event") == "exception":
                     error_msg = trace_step.get("error", "Unknown Error")
                     steps.append(ExecutionStepSchema(
                        stepNumber=step_number,
                        line=steps[-1].line if steps else 1, # Use last line or 1
                        code=steps[-1].code if steps else "",
                        state={
                            "error": error_msg,
                            "message": f"Execution Error: {error_msg}",
                            "traceback": trace_step.get("traceback", ""),
                            "stdout": list(accumulated_stdout)
                        }
                     ))
                     step_number += 1
                     continue

                line_no = trace_step.get("line", 1)
                
                # Check for valid line number
                current_code = ""
                if 1 <= line_no <= len(code_lines):
                    current_code = code_lines[line_no - 1].strip()
                
                # Skip steps for empty lines or comments (tracer might catch them)
                if not current_code or current_code.startswith('#'):
                     continue
                
                # Format state variables
                variables = trace_step.get("variables", {})
                
                # Determine output (if any)
                step_output = trace_step.get("output")
                
                # Accumulate stdout - join outputs that don't end with newlines
                if step_output:
                    if step_output.endswith('\n'):
                        # Complete line - add buffer + this output
                        full_line = current_line_buffer + step_output.rstrip('\n')
                        if full_line:
                            accumulated_stdout.append(full_line)
                        current_line_buffer = ""
                    else:
                        # Partial line - add to buffer
                        current_line_buffer += step_output
                
                # Create detailed state (matching simulator expectations)
                state = {
                    "variables": variables,
                    "stdout": list(accumulated_stdout),  # Copy of accumulated stdout
                    "message": f"Executed line {line_no}" if not step_output else f"Print: {step_output}",
                    "active": None, # Could enhance this to track active instance
                    "instances": {}, # Populate if variable analysis detects instances
                    "memory": trace_step.get("memory_usage", 0), # Add memory usage from trace
                    "memory_delta": trace_step.get("memory_delta", 0), # Memory change from previous step
                    "execution_time": trace_step.get("execution_time", 0), # Execution time in seconds
                    "step_detail": {
                         "operation": "execution",
                         "content": current_code,
                         "output": step_output
                    },
                    "explanation": None  # Will be populated below
                }
                
                # Enhanced: Update instances if variables contain data structures
                instances = {}
                active_instance = None
                
                for var_name, var_value in variables.items():
                    # Skip internal trace variables
                    if var_name.startswith("trace_") or var_name == "input_values" or var_name == "active_instance":
                        continue

                    # Detect potential Stack instances
                    # We look for objects serialized as dicts containing 'data' list
                    if isinstance(var_value, dict) and "data" in var_value and isinstance(var_value["data"], list):
                        # Use data_structure_type to refine class name if possible
                        class_type = "ArrayStack" # Default for stack
                        if data_structure_type == "queue":
                             class_type = "ArrayQueue"
                        
                        # Format for frontend visualization
                        instances[var_name] = {
                            "type": class_type,
                            "class_type": class_type,
                            "data": var_value["data"],
                            "size": len(var_value["data"]),
                            "isEmpty": len(var_value["data"]) == 0,
                            "top": var_value["data"][-1] if var_value["data"] else None
                        }
                        active_instance = var_name # Set last found as active

                    # [NEW] Detect LinkedList instances (Root Node)
                    # We look for objects that have 'head' and 'count' (Root Node pattern)
                    # OR objects that have 'next' and some data field (Node pattern - for fallback)
                    elif isinstance(var_value, dict) and ("head" in var_value or "next" in var_value):
                         
                         # Case 1: Root Node (LinkedList wrapper)
                         # Explicitly check if 'head' key exists. Value can be None or dict.
                         if "head" in var_value:
                             head_node = var_value.get("head")
                             
                             # Traverse to get nodes
                             nodes_list = []
                             
                             # Only traverse if head_node is a valid object (dict)
                             if isinstance(head_node, dict):
                                 curr = head_node
                                 
                                 # Simple traversal with loop detection
                                 # serialized nodes are dicts
                                 visited_ids = set()
                                 
                                 while isinstance(curr, dict):
                                     # Try to find data field
                                     data_val = curr.get("data") or curr.get("val") or curr.get("value") or curr.get("name")
                                     if data_val is not None:
                                         nodes_list.append(str(data_val))
                                     
                                     # Move to next
                                     # check for circular reference indicator from serializer
                                     if curr.get("next") == "<circular reference>":
                                         break
                                         
                                     curr = curr.get("next")
                                     
                                     # Safety break for very long lists
                                     if len(nodes_list) > 100:
                                         break
                             
                             count = var_value.get("count")
                             if count is None:
                                 count = len(nodes_list)
                             
                             instances[var_name] = {
                                 "type": "LinkedList",
                                 "class_type": "LinkedList",
                                 "nodes": nodes_list,
                                 "count": count,
                                 "head": str(nodes_list[0]) if nodes_list else None,
                                 "tail": str(nodes_list[-1]) if nodes_list else None
                             }
                             active_instance = var_name

                         # Case 2: Raw Node (Head of a list without wrapper, or just a node)
                         # Only treat as active instance if we haven't found a Root Node yet
                         elif "next" in var_value and active_instance is None:
                             # This might be just a node, but let's treat it as a list starting here
                             nodes_list = []
                             curr = var_value
                             
                             while isinstance(curr, dict):
                                 data_val = curr.get("data") or curr.get("val") or curr.get("value") or curr.get("name")
                                 if data_val is not None:
                                     nodes_list.append(str(data_val))
                                 
                                 if curr.get("next") == "<circular reference>":
                                     break
                                 curr = curr.get("next")
                                 if len(nodes_list) > 100: break
                             
                             instances[var_name] = {
                                 "type": "LinkedList",
                                 "class_type": "LinkedList",
                                 "nodes": nodes_list,
                                 "count": len(nodes_list),
                                 "head": str(nodes_list[0]) if nodes_list else None,
                                 "tail": str(nodes_list[-1]) if nodes_list else None
                             }
                             # Don't set active_instance here strongly, prefer Root Node if available later
 
                    # [NEW] Detect Binary Search Tree instances
                    elif isinstance(var_value, dict) and (("root" in var_value) or ("left" in var_value) or ("right" in var_value)):
                         
                         is_bst = False
                         root_data = None
                         is_wrapper = False
                         
                         # Case 1: BST Wrapper (has root)
                         if "root" in var_value:
                             is_bst = True
                             root_data = var_value.get("root")
                             is_wrapper = True
                             
                         # Case 2: Raw Root Node (has left/right)
                         elif "left" in var_value or "right" in var_value:
                             is_bst = True
                             root_data = var_value
                             
                         if is_bst:
                             instances[var_name] = {
                                 "type": "BinarySearchTree",
                                 "class_type": "BinarySearchTree",
                                 "root": root_data, 
                                 "isEmpty": root_data is None
                             }
                             # Prefer Wrapper over Node for active instance
                             if is_wrapper:
                                 active_instance = var_name
                             elif active_instance is None:
                                 active_instance = var_name

                    # [NEW] Detect Graph instances
                    elif isinstance(var_value, dict) and (
                        "graph" in var_value or "adj_list" in var_value or "adjacency_list" in var_value or "nodes" in var_value
                    ):
                         # Determine graph type
                         is_directed = data_structure_type == "directedgraph" or "Directed" in var_value.get("type", "")
                         graph_type = "DirectedGraph" if is_directed else "UndirectedGraph"
                         
                         # Extract Adjacency List
                         graph_data = {}
                         if "graph" in var_value and isinstance(var_value["graph"], dict):
                             graph_data = var_value["graph"]
                         elif "adj_list" in var_value and isinstance(var_value["adj_list"], dict):
                             graph_data = var_value["adj_list"]
                         elif "adjacency_list" in var_value and isinstance(var_value["adjacency_list"], dict):
                             graph_data = var_value["adjacency_list"]
                         
                         instances[var_name] = {
                             "type": graph_type,
                             "class_type": graph_type,
                             "graph": graph_data,
                             "isDirected": is_directed
                         }
                         active_instance = var_name

                # Update state with instances
                state["instances"] = instances
                state["active"] = active_instance
                
                # Add class metadata for dynamic visualization labels
                if class_metadata:
                    state["class_metadata"] = class_metadata

                # [NEW] Enrich step_detail with semantic info based on data structure type
                # This helps frontend visualization (highlighting current node, etc.)
                func_name = trace_step.get("func", "")
                
                # Guess operation from function name if possible
                operation = "execution"
                
                # [ENHANCED] Detect node_creation when inside DataNode.__init__
                # Check if self variable has node-like structure (has name/data AND next pointer)
                self_var = variables.get("self", {})
                is_node_init = False
                node_value = None
                
                
                if func_name == "__init__":
                    if isinstance(self_var, dict):
                        # Check if self looks like a node (has data field AND next pointer)
                        # Skip 'type' key as it's metadata
                        has_data = any(k in self_var for k in ["name", "data", "value", "val"])
                        has_next = "next" in self_var
                        
                        # It's a node creation if self has both data and next pointer
                        if has_data and has_next:
                            is_node_init = True
                            # Get value from self_var, converting to string
                            raw_value = (
                                self_var.get("name") or 
                                self_var.get("data") or 
                                self_var.get("value") or 
                                self_var.get("val")
                            )
                            if raw_value is not None:
                                node_value = str(raw_value)

                    
                    # FALLBACK: Detect based on code patterns when self is not properly serialized
                    # If code is assigning to self.next = None OR self.name = X, and we're in __init__
                    # This catches DataNode.__init__ even when self isn't serialized
                    if not is_node_init:
                        # Check for DataNode-like __init__ patterns
                        if "self.next" in current_code and "None" in current_code:
                            # This is likely self.next = None in DataNode.__init__
                            # Check if we have a 'name' parameter or self.name was set earlier
                            is_node_init = True
                            # Try multiple sources to get the node value (prioritized)
                            # 1. Check 'name' parameter variable (for direct argument)
                            name_param = variables.get("name")
                            if name_param is not None and name_param != "":
                                node_value = str(name_param)
                            # 2. Check self_var dict for already-set values (for input-based values)
                            elif isinstance(self_var, dict):
                                node_value = (
                                    self_var.get("name") or 
                                    self_var.get("data") or 
                                    self_var.get("value") or 
                                    self_var.get("val")
                                )
                                if node_value is not None:
                                    node_value = str(node_value)
                            # 3. Check 'data' or 'val' parameter variables (alternative node constructors)
                            if not node_value:
                                for alt_name in ["data", "val", "value"]:
                                    alt_param = variables.get(alt_name)
                                    if alt_param is not None and alt_param != "":
                                        node_value = str(alt_param)
                                        break

                        elif "self.name" in current_code and "=" in current_code:
                            is_node_init = True
                            # Value might be in the 'name' parameter
                            name_param = variables.get("name")
                            if name_param is not None and name_param != "":
                                node_value = str(name_param)
                            # Also check self_var in case it's already set
                            elif isinstance(self_var, dict):
                                node_value = self_var.get("name")
                                if node_value is not None:
                                    node_value = str(node_value)

                        elif "self.data" in current_code and "=" in current_code:
                            # Also handle DataNode classes that use 'data' instead of 'name'
                            is_node_init = True
                            data_param = variables.get("data")
                            if data_param is not None and data_param != "":
                                node_value = str(data_param)
                            elif isinstance(self_var, dict):
                                node_value = self_var.get("data")
                                if node_value is not None:
                                    node_value = str(node_value)

                
                if is_node_init:
                    operation = "node_creation"
                    # Always set node_value - use fallback if actual value couldn't be determined
                    effective_node_value = str(node_value) if node_value else "new_node"
                    state["step_detail"]["node_value"] = effective_node_value
                    state["step_detail"]["node_variable"] = "self"
                    state["step_detail"]["is_connected"] = False
                
                # [ENHANCED] Detect pointer_assignment when connecting nodes
                # Must check LEFT side of assignment to correctly categorize:
                # - pNew.next = self.head → chained_pointer_assignment (left side has .next)
                # - self.head = pNew → pointer_assignment (left side has .head)
                
                # Split by = first to check left side
                elif "=" in current_code and ("." in current_code):
                    parts = current_code.split("=")
                    if len(parts) >= 2:
                        left_side = parts[0].strip()
                        right_side = parts[1].strip()
                        

                        
                        # Check LEFT side to determine operation type
                        if ".next" in left_side and "None" not in right_side:
                            # pNew.next = self.head (linking new node to existing list)
                            operation = "chained_pointer_assignment"
                            state["step_detail"]["creates_connection"] = True
                            state["step_detail"]["source_var"] = right_side
                            
                            # [NEW] Detect pNew.next = self.head pattern for intermediate step visualization
                            # This helps frontend show arrow from pending node to head before head is reassigned
                            if "self.head" in right_side or ".head" in right_side:
                                state["step_detail"]["next_points_to_head"] = True
                                
                                # Extract the pending node variable name (left side of .next =)
                                if ".next" in left_side:
                                    pending_var = left_side.replace(".next", "").strip()
                                    state["step_detail"]["pending_node_variable"] = pending_var
                                    
                                    # Try to get the pending node's value from variables
                                    if pending_var in variables:
                                        pending_node = variables[pending_var]
                                        if isinstance(pending_node, dict):
                                            pending_val = (
                                                pending_node.get("name") or 
                                                pending_node.get("data") or 
                                                pending_node.get("value") or 
                                                pending_node.get("val")
                                            )
                                            if pending_val is not None:
                                                state["step_detail"]["pending_node_value"] = str(pending_val)
                                        # Fallback: Handle circular reference strings like "DataNode(Ako)"
                                        elif isinstance(pending_node, str):
                                            cr_match = re.match(r'^\w+\((.+)\)$', pending_node)
                                            if cr_match:
                                                state["step_detail"]["pending_node_value"] = cr_match.group(1)
                                
                                # Try to get the current head value from instances
                                for inst_name, inst_data in instances.items():
                                    if isinstance(inst_data, dict) and inst_data.get("type") == "LinkedList":
                                        head_val = inst_data.get("head")
                                        if head_val:
                                            state["step_detail"]["target_head_value"] = str(head_val)
                            
                            # [NEW] Detect pNew.next = start.next pattern (insertBefore intermediate step)
                            # e.g., pNew.next = start.next where start is a traversal pointer
                            elif ".next" in right_side:

                                # Extract potential variables first to validate if this is really an insertion pattern
                                # LHS: pNew.next -> pending_var = pNew
                                # RHS: start.next -> start_var = start
                                # If pending_var == start_var (e.g. start.next = start.next.next), it's a deletion/traversal, NOT insertion.
                                
                                potential_pending_var = None
                                if ".next" in left_side:
                                    potential_pending_var = left_side.replace(".next", "").strip()
                                
                                potential_start_var = None
                                # Handle start.next.next case for RHS
                                if ".next" in right_side:
                                    # Split by .next and take the first part as base variable
                                    potential_start_var = right_side.split(".next")[0].strip()
                                
                                # Only proceed if variables are different (Insertion pattern)
                                if potential_pending_var and potential_start_var and potential_pending_var != potential_start_var:
                                    state["step_detail"]["next_points_to_start_next"] = True

                                    
                                    pending_var = potential_pending_var
                                    state["step_detail"]["pending_node_variable"] = pending_var
                                    
                                    # Try to get the pending node's value from variables
                                    if pending_var in variables:
                                        pending_node = variables[pending_var]

                                        if isinstance(pending_node, dict):
                                            pending_val = (
                                                pending_node.get("name") or 
                                                pending_node.get("data") or 
                                                pending_node.get("value") or 
                                                pending_node.get("val")
                                            )

                                            if pending_val is not None:
                                                state["step_detail"]["pending_node_value"] = str(pending_val)
                                        # Fallback: Handle circular reference strings like "DataNode(Ako)"
                                        elif isinstance(pending_node, str):
                                            cr_match = re.match(r'^\w+\((.+)\)$', pending_node)
                                            if cr_match:
                                                state["step_detail"]["pending_node_value"] = cr_match.group(1)

                                    else:
                                        pass

                                
                                    # Extract the start variable name (right side before .next)
                                    start_var = right_side.replace(".next", "").strip()
                                    state["step_detail"]["start_node_variable"] = start_var
                                    
                                    # Try to get start node's position and next value
                                    if start_var in variables:
                                        start_node = variables[start_var]

                                        
                                        # Extract start_val from dict or circular reference string
                                        start_val = None
                                        if isinstance(start_node, dict):
                                            start_val = (
                                                start_node.get("name") or 
                                                start_node.get("data") or 
                                                start_node.get("value") or 
                                                start_node.get("val")
                                            )
                                        # Fallback: Handle circular reference strings like "DataNode(Mika)"
                                        elif isinstance(start_node, str):
                                            cr_match = re.match(r'^\w+\((.+)\)$', start_node)
                                            if cr_match:
                                                start_val = cr_match.group(1)
                                        
                                        if start_val is not None:
                                            state["step_detail"]["start_node_value"] = str(start_val)
                                            
                                            # Find start node's position in the list
                                            for inst_name, inst_data in instances.items():
                                                if isinstance(inst_data, dict) and inst_data.get("type") == "LinkedList":
                                                    nodes = inst_data.get("nodes", [])

                                                    for idx, node_val in enumerate(nodes):
                                                        if str(node_val) == str(start_val):
                                                            state["step_detail"]["start_node_position"] = idx
                                                            # Target is start.next, so it's at idx+1
                                                            if idx + 1 < len(nodes):
                                                                state["step_detail"]["target_next_value"] = str(nodes[idx + 1])
                                                                state["step_detail"]["target_next_position"] = idx + 1
                                                            break
                                                    break
                                    

                                else:
                                    pass

                        
                        elif ".head" in left_side:
                            # self.head = pNew (reassigning head pointer)
                            operation = "pointer_assignment"
                            state["step_detail"]["is_head_assignment"] = True
                            state["step_detail"]["creates_connection"] = True
                            state["step_detail"]["source_var"] = right_side
                
                # Check for input operation
                elif "input(" in current_code or "input" in func_name.lower():
                    operation = "input"
                elif "insert" in func_name.lower() or "add" in func_name.lower() or "push" in func_name.lower() or "append" in func_name.lower():
                    operation = "insert"
                elif "delete" in func_name.lower() or "remove" in func_name.lower() or "pop" in func_name.lower():
                    operation = "delete" 
                elif "search" in func_name.lower() or "find" in func_name.lower() or "contains" in func_name.lower() or "get" in func_name.lower():
                    operation = "search"
                elif "traverse" in func_name.lower():
                    operation = "traverse"
                
                state["step_detail"]["operation"] = operation
                
                # [NEW] Add user command (caller line code)
                # This helps visualize the high-level user command that triggered this step
                caller_line = trace_step.get("caller_line")
                if caller_line and isinstance(caller_line, int) and 1 <= caller_line <= len(code_lines):
                    user_cmd = code_lines[caller_line - 1].strip()
                    # Only add if it's not the same as the current line (to avoid redundancy)
                    if user_cmd and user_cmd != current_code:
                        state["step_detail"]["user_command"] = user_cmd
                
                # [NEW] Detect common pitfalls/warnings for educational feedback
                warnings = []
                
                # Pitfall 1: Head reassignment without saving reference
                # e.g., "self.head = newNode" without temp = self.head first
                if operation == "pointer_assignment" and state["step_detail"].get("is_head_assignment"):
                    # Check if there's no temp variable holding old head
                    has_temp = any(
                        name.lower() in ["temp", "old_head", "prev", "current", "curr", "tmp", "save", "backup", "pnew", "new_node", "newnode", "node"]
                        for name in variables.keys()
                    )
                    if not has_temp:
                        warnings.append({
                            "type": "losing_reference",
                            "severity": "warning",
                            "message": "ถ้าลืมเก็บ reference ของ head ก่อน จะ lose entire list",
                            "tip": "เก็บ head ไว้ใน temp ก่อน reassign เช่น: temp = self.head"
                        })
                
                # Pitfall 2: Potential null pointer access
                # e.g., current.next when current might be None
                if ".next" in current_code or ".data" in current_code or ".name" in current_code:
                    # Check if accessing .next/.data on a variable that could be None
                    # Look for patterns like: current.next, node.data, etc.
                    access_patterns = [
                        (r"(\w+)\.next", "next"),
                        (r"(\w+)\.data", "data"),
                        (r"(\w+)\.name", "name"),
                        (r"(\w+)\.value", "value"),
                    ]
                    for pattern, attr in access_patterns:
                        match = re.search(pattern, current_code)
                        if match:
                            var_name = match.group(1)
                            # Skip 'self' as it's always valid
                            if var_name != "self" and var_name in variables:
                                var_value = variables.get(var_name)
                                # If variable is None or looks like it could be None
                                if var_value is None:
                                    warnings.append({
                                        "type": "null_pointer",
                                        "severity": "error",
                                        "message": f"Null Pointer! ตัวแปร {var_name} เป็น None แต่พยายามเข้าถึง .{attr}",
                                        "tip": f"ตรวจสอบว่า {var_name} != None ก่อนเข้าถึง attribute"
                                    })
                                    break
                
                # Pitfall 3: Delete without proper pointer update (Memory leak warning)
                if operation == "delete":
                    # Check if there's proper cleanup - look for patterns of proper deletion
                    # This is a general educational warning
                    warnings.append({
                        "type": "memory_leak_reminder",
                        "severity": "info",
                        "message": "การลบ node: อย่าลืมปรับ pointer ก่อน-หลัง node ที่จะลบ",
                        "tip": "ต้อง prev.next = current.next เพื่อข้าม node ที่ลบ ไม่งั้นจะเกิด Memory Leak"
                    })
                
                # Add warnings to step_detail if any
                if warnings:
                    state["step_detail"]["warnings"] = warnings
                
                # [ENHANCED] Mark traverse steps for step-by-step animation
                if operation == "traverse":
                    state["step_detail"]["is_traverse_step"] = True
                
                # Extract 'current_node' and 'inserted_node' for visualization highlights
                # Heuristics: look for variables named 'node', 'curr', 'current', 'temp' that are Nodes
                if data_structure_type in ["binary_search_tree", "binarysearchtree", "singlylinkedlist", "doublylinkedlist", "directedgraph", "undirectedgraph"]:
                    current_node_val = None
                    inserted_node_val = None
                    
                    # 1. Find Current Node (Orange Highlight)
                    # Priority: 'curr', 'current', 'node' (in recursive calls), 'temp'
                    # Graph additions: 'vertex', 'v', 'u', 'neighbor'
                    # LinkedList additions: 'start' (common in traverse)
                    for name in variables:
                         if name.lower() in ["curr", "current", "node", "temp", "ptr", "current_node", "vertex", "v", "u", "neighbor", "start"]:
                              val = variables[name]
                              
                              # Case A: Object/Dict Node (BST, LinkedList)
                              if isinstance(val, dict):
                                   # Extract value from node dict - handle potential wrapper with "type"
                                   # The tracer wraps objects with {"type": "ClassName", "data": ...}
                                   
                                   # Check for direct value - include 'name' for DataNode
                                   node_val = val.get("val") or val.get("value") or val.get("data") or val.get("name")
                                   
                                   # If detected value is actually a Type definition (like BSTNode) string, ignore
                                   if node_val == "BSTNode": 
                                        node_val = None
                                   
                                   # If we have nested structure (e.g. from custom serializer)
                                   # Sometimes data is inside another dict if it was wrapped
                                   if node_val is None:
                                        # Iterate keys to find something that looks like data
                                        for k, v in val.items():
                                            if k in ["data", "val", "value", "name"] and v is not None:
                                                node_val = v
                                                break
                                   
                                   if node_val is not None:
                                        current_node_val = str(node_val)
                                        # If found high priority name, stop
                                        if name.lower() in ["curr", "current", "current_node", "vertex", "start"]:
                                             break
                              
                              # Case B: Primitive Value (Graph Vertex)
                              elif data_structure_type in ["directedgraph", "undirectedgraph"] and isinstance(val, (str, int)):
                                   current_node_val = str(val)
                                   if name.lower() in ["curr", "current", "current_node", "vertex", "u", "v"]:
                                        break
                              
                              # Case C: Circular Reference Format - "Node(value)" or "DataNode(value)"
                              # This happens when the serializer encounters a circular reference
                              # and returns format like "Node(5)" or "DataNode(Tony)" instead of a dict
                              elif isinstance(val, str):
                                   # Match patterns like "Node(5)", "DataNode(Tony)", "ListNode(hello)"
                                   cr_match = re.match(r'^(\w+)\((.+)\)$', val)
                                   if cr_match:
                                        type_name = cr_match.group(1)
                                        node_val = cr_match.group(2)
                                        # Only accept known node type names
                                        if type_name.lower() in ["node", "datanode", "listnode", "sllnode", "dllnode", "bstnode", "treenode"]:
                                             current_node_val = node_val
                                             if name.lower() in ["curr", "current", "current_node", "start"]:
                                                  break
                    
                    # 2. Find Inserted Node/Value (Yellow Highlight)
                    # Usually passed as 'val', 'value', 'data' to insert methods
                    if operation == "insert":
                         for name in variables:
                              if name.lower() in ["val", "value", "data", "key"]:
                                   val = variables[name]
                                   if isinstance(val, (int, str, float)):
                                        inserted_node_val = str(val)
                                        break
                    
                    if current_node_val:
                        state["step_detail"]["current_node"] = current_node_val
                        # [ENHANCED] For traverse operations, also set traverse_node explicitly
                        if operation == "traverse":
                            state["step_detail"]["traverse_node"] = current_node_val
                        # Also add to message for clarity
                        if "message" in state and state["message"].startswith("Executed"):
                             if operation != "execution" and operation != "chained_pointer_assignment":
                                  state["message"] = f"{operation.capitalize()}: Visiting node {current_node_val}"
                             else:
                                  state["message"] = f"Visiting node {current_node_val}"
                        
                        # [NEW] Add pointer_position (index) for visualization
                        # Find the index of current_node in the linked list
                        for var_name, instance_data in instances.items():
                            if isinstance(instance_data, dict) and instance_data.get("type") == "LinkedList":
                                nodes = instance_data.get("nodes", [])
                                for idx, node_val in enumerate(nodes):
                                    if str(node_val) == current_node_val:
                                        state["step_detail"]["pointer_position"] = idx
                                        break
                                break  # Only check first LinkedList instance
                        
                        # [ENHANCED] Dynamic pointer variable detection from code patterns
                        # Parse the actual variable name from the code instead of using fixed names
                        
                        detected_pointer_var = None
                        
                        # Pattern 1: xxx = xxx.next (pointer movement)
                        pointer_move_match = re.match(r'^\s*(\w+)\s*=\s*\1\.next', current_code)
                        if pointer_move_match:
                            detected_pointer_var = pointer_move_match.group(1)
                            state["step_detail"]["is_pointer_movement"] = True
                        
                        # Pattern 2: while xxx != None or while xxx is not None
                        if not detected_pointer_var:
                            while_match = re.match(r'^\s*while\s+(\w+)(?:\.next)?\s*(?:!=|is not)\s*None', current_code)
                            if while_match:
                                detected_pointer_var = while_match.group(1)
                                state["step_detail"]["is_loop_iteration"] = True
                        
                        # Pattern 3: if xxx != None or if xxx is not None
                        if not detected_pointer_var:
                            if_match = re.match(r'^\s*if\s+(\w+)(?:\.next)?\s*(?:!=|is not)\s*None', current_code)
                            if if_match:
                                detected_pointer_var = if_match.group(1)
                        
                        # Pattern 4: xxx.next = yyy.next or xxx.next = yyy (chained pointer assignment)
                        if not detected_pointer_var:
                            chained_match = re.match(r'^\s*(\w+)\.next\s*=\s*(\w+)(?:\.next)?', current_code)
                            if chained_match:
                                # Use the right side variable as the pointer to show
                                detected_pointer_var = chained_match.group(2)
                        
                        # Pattern 5: xxx = yyy.next (assignment from another node's next)
                        if not detected_pointer_var:
                            assign_next_match = re.match(r'^\s*(\w+)\s*=\s*(\w+)\.next', current_code)
                            if assign_next_match:
                                detected_pointer_var = assign_next_match.group(1)
                        
                        # Pattern 6: print(xxx.name) or similar access
                        if not detected_pointer_var:
                            access_match = re.match(r'^\s*(?:print\s*\(\s*)?(\w+)\.(name|val|value|data|next)', current_code)
                            if access_match:
                                detected_pointer_var = access_match.group(1)
                        
                        # Pattern 7: if xxx.next.name == yyy (comparison)
                        if not detected_pointer_var:
                            compare_match = re.match(r'^\s*if\s+(\w+)(?:\.next)?\.name\s*==', current_code)
                            if compare_match:
                                detected_pointer_var = compare_match.group(1)
                        
                        # If detected from code, verify and use
                        if detected_pointer_var and detected_pointer_var in variables:
                            val = variables[detected_pointer_var]
                            if isinstance(val, dict):
                                node_val = val.get("val") or val.get("value") or val.get("data") or val.get("name")
                                if node_val is not None and str(node_val) == current_node_val:
                                    state["step_detail"]["pointer_variable_name"] = detected_pointer_var
                        
                        # Always try fallback to find ANY pointer variable pointing to current_node
                        if "pointer_variable_name" not in state["step_detail"]:
                            for var_name in variables:
                                val = variables[var_name]
                                if isinstance(val, dict):
                                    node_val = val.get("val") or val.get("value") or val.get("data") or val.get("name")
                                    if node_val is not None and str(node_val) == current_node_val:
                                        state["step_detail"]["pointer_variable_name"] = var_name
                                        break

                    if inserted_node_val:
                        state["step_detail"]["inserted_node"] = inserted_node_val


                # Generate Thai explanation for this step
                try:
                    prev_step_state = steps[-1].state if steps else None
                    explanation_gen = ExplanationGenerator(data_structure_type)
                    explanation = explanation_gen.generate_explanation(
                        code_line=current_code,
                        operation=state["step_detail"].get("operation"),
                        variables=variables,
                        prev_state=prev_step_state,
                        curr_state=state,
                    )
                    state["explanation"] = explanation
                except Exception:
                    # If explanation generation fails, continue without it
                    state["explanation"] = None

                steps.append(ExecutionStepSchema(
                    stepNumber=step_number,
                    line=line_no,
                    code=current_code,
                    state=state
                ))
                step_number += 1
            
            # Flush any remaining buffer (for partial lines at the end)
            if current_line_buffer:
                accumulated_stdout.append(current_line_buffer)
                # Update the last step's stdout if we have steps
                if steps:
                    steps[-1].state["stdout"] = list(accumulated_stdout)
                
            if steps:
                return steps
        
        # Fallback to print-based steps if no trace
        # Extract output - the wrapper captures actual print output
        print_outputs = []
        actual_stdout = ""
        
        # Debug: Log the result to understand what we're getting
        # The wrapper code executes the user code and captures print output
        # The output should be the actual evaluated print values, not literal strings
        
        # Priority: result.output (parsed JSON) > result.stdout (raw JSON string)
        import logging
        logger = logging.getLogger(__name__)
        
        # Debug: Log what we received
        logger.info(f"DEBUG: result.output type: {type(result.output)}, value: {result.output}")
        logger.info(f"DEBUG: result.stdout: {result.stdout[:200] if result.stdout else 'None'}")
        logger.info(f"DEBUG: result.stderr: {result.stderr[:200] if result.stderr else 'None'}")
        logger.info(f"DEBUG: result.exit_code: {result.exit_code}")
        
        if result.output:
            # If result.output is a dict with "output" key
            if isinstance(result.output, dict) and "output" in result.output:
                output_value = result.output["output"]
                logger.info(f"DEBUG: output_value from result.output: {output_value}, type: {type(output_value)}")
                
                if isinstance(output_value, str):
                    # This is the actual executed print output (not a literal)
                    print_outputs = [output_value]
                    actual_stdout = output_value
                elif isinstance(output_value, list):
                    # Multiple print outputs
                    print_outputs = [str(v) for v in output_value]
                    actual_stdout = '\n'.join(print_outputs)
                else:
                    print_outputs = [str(output_value)]
                    actual_stdout = str(output_value)
            # [NEW] Special handling for InteractiveExecutionResult
            # The result.output is the full metadata dict (trace, stdout, etc.)
            # We should NOT stringify this dict as output. 
            # If it's the metadata dict, we skip "result.output" processing and fall through to result.stdout
            elif isinstance(result, InteractiveExecutionResult) and isinstance(result.output, dict) and "trace" in result.output:
                # This is the wrapper result dict, not user output. Ignore it.
                pass
            else:
                print_outputs = [str(result.output)]
                actual_stdout = str(result.output)
        elif result.stdout:
            # Try to parse JSON from stdout
            try:
                parsed_output = json.loads(result.stdout.strip())
                if isinstance(parsed_output, dict) and "output" in parsed_output:
                    output_value = parsed_output["output"]
                    if isinstance(output_value, str):
                        # This is the actual executed print output
                        # The wrapper code executed the code and this is the evaluated result
                        print_outputs = [output_value]
                        actual_stdout = output_value
                    elif isinstance(output_value, list):
                        # Multiple print outputs
                        print_outputs = [str(v) for v in output_value]
                        actual_stdout = '\n'.join(print_outputs)
                    else:
                        print_outputs = [str(output_value)]
                        actual_stdout = str(output_value)
                elif isinstance(parsed_output, dict) and "error" in parsed_output:
                    # Error in execution
                    error_msg = parsed_output.get("error", "Execution error")
                    error_step = ExecutionStepSchema(
                        stepNumber=step_number,
                        line=1,
                        code=code.split('\n')[0] if code.split('\n') else code,
                        state={
                            "error": error_msg,
                            "message": f"เกิดข้อผิดพลาด: {error_msg}"
                        }
                    )
                    return [error_step]
                else:
                    # Not JSON or no "output" key, use stdout as-is
                    print_outputs = [result.stdout] if result.stdout.strip() else []
                    actual_stdout = result.stdout
            except json.JSONDecodeError:
                # Not JSON, use stdout as-is (direct print output)
                # This means the wrapper didn't wrap it, so it's direct output
                if result.stdout.strip():
                    print_outputs = [result.stdout.strip()]
                    actual_stdout = result.stdout.strip()
                else:
                    print_outputs = []
                    actual_stdout = ""
        
        # If we still don't have output, check stderr for any clues
        if not print_outputs and result.stderr:
            # Check if stderr contains error info
            try:
                error_info = json.loads(result.stderr.strip())
                if isinstance(error_info, dict) and "error" in error_info:
                    error_step = ExecutionStepSchema(
                        stepNumber=step_number,
                        line=1,
                        code=code.split('\n')[0] if code.split('\n') else code,
                        state={
                            "error": error_info["error"],
                            "message": f"เกิดข้อผิดพลาด: {error_info['error']}"
                        }
                    )
                    return [error_step]
            except (json.JSONDecodeError, KeyError):
                pass
        
        # Find print statements in code and create steps for them
        code_lines = code.split('\n')
        print_statement_count = 0
        
        for line_num, line in enumerate(code_lines, 1):
            stripped_line = line.strip()
            
            # Skip commented out lines
            if stripped_line.startswith('#'):
                continue
                
            # Check if this line contains a print statement
            if stripped_line.startswith('print(') or 'print(' in stripped_line:
                # Get the output for this print statement
                output_value = ""
                if print_outputs:
                    # Match output to print statement by order
                    if print_statement_count < len(print_outputs):
                        output_value = print_outputs[print_statement_count]
                    else:
                        # If we have more print statements than outputs, do NOT reuse the last output
                        # This happens when we detect print statements that weren't actually executed
                        output_value = ""
                
                # Create step for print statement
                print_step = ExecutionStepSchema(
                    stepNumber=step_number,
                    line=line_num,
                    code=stripped_line,
                    state={
                        "message": f"Print: {output_value}" if output_value else "Print statement executed",
                        "stdout": [output_value] if output_value else [],
                        "step_detail": {
                            "operation": "print",
                            "content": stripped_line,
                            "output": output_value
                        },
                        "stdout": actual_stdout,
                        "stderr": result.stderr if result.stderr else ""
                    }
                )
                steps.append(print_step)
                step_number += 1
                print_statement_count += 1
        
        # If no print steps were created, create a general execution step
        if not steps:
            execution_step = ExecutionStepSchema(
                stepNumber=step_number,
                line=1,
                code=code.split('\n')[0] if code.split('\n') else code,
                state={
                    "message": "Code executed successfully",
                    "execution_result": result.output if result.output else {},
                    "stdout": result.stdout,
                    "stderr": result.stderr if result.stderr else "",
                    "stdout": print_outputs if print_outputs else []
                }
            )
            steps.append(execution_step)
        
        return steps
    
    def _create_steps_from_interactive_result(
        self,
        code: str,
        result: InteractiveExecutionResult,
        step_number: int,
        data_structure_type: Optional[str] = None
    ) -> List[ExecutionStepSchema]:
        """Create execution steps from interactive execution result"""
        # Similar to regular result but handle input history
        steps = self._create_steps_from_result(code, result, step_number, data_structure_type)
        
        # Add input history to steps if available
        if result.input_history:
            for step in steps:
                if "step_detail" in step.state:
                    step.state["step_detail"]["input_history"] = result.input_history
        
        return steps
    
    def _create_error_step(
        self,
        code: str,
        result: ExecutionResult,
        step_number: int,
        data_structure_type: Optional[str] = None
    ) -> List[ExecutionStepSchema]:
        """Create error step from execution result"""
        error_step = ExecutionStepSchema(
            stepNumber=step_number,
            line=1,
            code=code.split('\n')[0] if code.split('\n') else code,
            state={
                "error": result.stderr or "Execution failed",
                "exit_code": result.exit_code,
                "timed_out": result.timed_out,
                "message": f"เกิดข้อผิดพลาด: {result.stderr}" if result.stderr else "Execution failed"
            }
        )
        return [error_step]




