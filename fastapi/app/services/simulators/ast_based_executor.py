"""
AST-Based Executor - Combines AST parsing with real Python execution
Parses AST first to identify operations, then executes code and captures state at AST node boundaries
"""
import ast
import json
import sys
from typing import List, Dict, Any, Optional, Set
from datetime import datetime

from app.schemas.playground import ExecutionStepSchema
from app.services.simulators.operations.ast_parser import ASTParser
from app.services.simulators.real_python_executor import RealPythonExecutor, ExecutionResult
from app.services.simulators.interactive_python_executor import InteractivePythonExecutor, InputRequest


class ASTBasedExecutor:
    """
    Executes Python code using AST-first approach:
    1. Parse AST to identify operations and structure
    2. Execute code in real Python
    3. Capture state at AST node boundaries
    4. Return steps with AST info + execution results
    """
    
    def __init__(self, use_docker: bool = False, timeout: int = 30, memory: str = "256m"):
        """
        Initialize the AST-based executor
        
        Args:
            use_docker: If True, use Docker for execution
            timeout: Execution timeout in seconds
            memory: Memory limit for Docker (if used)
        """
        self.ast_parser = ASTParser()
        self.real_executor = RealPythonExecutor(use_docker, timeout, memory)
        self.interactive_executor = InteractivePythonExecutor(use_docker, timeout, memory)
        self.use_docker = use_docker
        self.timeout = timeout
    
    def parse_ast_structure(self, code: str) -> Dict[str, Any]:
        """
        Parse AST to identify operations and structure
        
        Args:
            code: Python code to parse
            
        Returns:
            Dictionary with AST structure information
        """
        try:
            tree = self.ast_parser.parse_code(code)
            
            # Extract classes
            classes = self.ast_parser.extract_classes(tree)
            
            # Get executable lines
            executable_lines = self.ast_parser.get_executable_lines(tree)
            
            # Extract AST node information
            ast_nodes = self._extract_ast_nodes(tree)
            
            return {
                "tree": tree,
                "classes": classes,
                "executable_lines": executable_lines,
                "ast_nodes": ast_nodes,
                "has_input": self._has_input_call(tree),
                "node_count": len(ast_nodes)
            }
        except Exception as e:
            raise ValueError(f"AST parsing failed: {str(e)}")
    
    def _extract_ast_nodes(self, tree: ast.AST) -> List[Dict[str, Any]]:
        """
        Extract AST node information for visualization
        
        Args:
            tree: AST tree
            
        Returns:
            List of AST node information
        """
        nodes = []
        
        for node in ast.walk(tree):
            node_info = {
                "type": type(node).__name__,
                "line": getattr(node, 'lineno', None),
                "col_offset": getattr(node, 'col_offset', None),
            }
            
            # Add specific information based on node type
            if isinstance(node, ast.Call):
                if isinstance(node.func, ast.Name):
                    node_info["function_name"] = node.func.id
                elif isinstance(node.func, ast.Attribute):
                    node_info["function_name"] = node.func.attr
                    if isinstance(node.func.value, ast.Name):
                        node_info["object_name"] = node.func.value.id
            
            elif isinstance(node, ast.Assign):
                if node.targets:
                    target = node.targets[0]
                    if isinstance(target, ast.Name):
                        node_info["variable_name"] = target.id
                    elif isinstance(target, ast.Attribute):
                        if isinstance(target.value, ast.Name):
                            node_info["object_name"] = target.value.id
                            node_info["attribute_name"] = target.attr
            
            elif isinstance(node, ast.ClassDef):
                node_info["class_name"] = node.name
            
            elif isinstance(node, ast.FunctionDef):
                node_info["function_name"] = node.name
            
            nodes.append(node_info)
        
        return nodes
    
    def _has_input_call(self, tree: ast.AST) -> bool:
        """Check if code contains input() calls"""
        for node in ast.walk(tree):
            if isinstance(node, ast.Call):
                if isinstance(node.func, ast.Name) and node.func.id == 'input':
                    return True
        return False
    
    def instrument_code_with_ast_markers(self, code: str, ast_nodes: List[Dict[str, Any]]) -> str:
        """
        Instrument code with AST node markers to capture state
        
        Args:
            code: Original Python code
            ast_nodes: List of AST node information
            
        Returns:
            Instrumented code with state capture markers
        """
        # For now, we'll use a simpler approach:
        # Execute code and capture state at key points
        # In a more advanced implementation, we could inject markers at each AST node
        
        wrapper = '''import json
import sys
import io
from contextlib import redirect_stdout

# State capture function
_state_captures = []

def _capture_state(node_type, line_number, context=None):
    """Capture execution state at AST node"""
    state = {
        "node_type": node_type,
        "line_number": line_number,
        "timestamp": None  # Could add timestamp if needed
    }
    if context:
        state.update(context)
    _state_captures.append(state)

try:
    # Read JSON input from stdin (if any)
    input_data = {}
    try:
        stdin_input = sys.stdin.read().strip()
        if stdin_input:
            input_data = json.loads(stdin_input)
            for key, value in input_data.items():
                globals()[key] = value
    except:
        pass
    
    # Execute user code
    {user_code}
    
    # Output state captures
    print(json.dumps({{"type": "state_captures", "captures": _state_captures}}), file=sys.stderr, flush=True)
    
    # Output success
    print(json.dumps({{"type": "execution_complete"}}), file=sys.stderr, flush=True)

except Exception as e:
    error_info = {{"type": "error", "error": str(e)}}
    print(json.dumps(error_info), file=sys.stderr, flush=True)
    error_info = {{"error": str(e)}}
    print(json.dumps(error_info, ensure_ascii=False))
'''
        # Indent user code
        indented_code = '\n'.join('    ' + line for line in code.split('\n'))
        final_code = wrapper.format(user_code=indented_code)
        return final_code
    
    def execute_with_ast(
        self,
        code: str,
        stdin_data: str = "",
        data_structure_type: Optional[str] = None,
        input_values: Optional[List[str]] = None
    ) -> List[ExecutionStepSchema]:
        """
        Execute code using AST-first approach
        
        Args:
            code: Python code to execute
            stdin_data: Input data (JSON string)
            data_structure_type: Type of data structure (for context)
            
        Returns:
            List of execution steps with AST info + execution results
        """
        steps = []
        step_number = 1
        
        try:
            # Step 1: Parse AST first
            ast_structure = self.parse_ast_structure(code)
            
            # Step 2: Create class definition steps
            if ast_structure["classes"]:
                context = {
                    "classes": ast_structure["classes"],
                    "instances": {},
                    "variables": {},
                    "nodes": {},
                    "stdout": [],
                    "include_linkedlist": data_structure_type in ["singlylinkedlist", "doublylinkedlist"],
                    "data_structure_type": data_structure_type
                }
                
                class_steps = self.ast_parser.create_class_definition_steps(
                    ast_structure["classes"],
                    context,
                    None  # get_instance_display_func - will be set by simulator
                )
                steps.extend(class_steps)
                step_number = len(steps) + 1
            
            # Step 3: Execute code in real Python
            if ast_structure["has_input"]:
                # Use interactive executor for input() support
                result = self.interactive_executor.execute_interactive(
                    code,
                    stdin_data,
                    input_callback=None,  # Not used when input_values provided
                    input_values=input_values  # Use pre-collected input values
                )
            else:
                # Use regular executor
                result = self.real_executor.execute(code, stdin_data)
            
            # Step 4: Create execution steps from results
            if result.exit_code == 0:
                # Success - create steps from execution
                # Get detailed AST node metadata
                tree = self.ast_parser.parse_code(code)
                # Use static method from ASTParser
                from app.services.simulators.operations.ast_parser import ASTParser
                try:
                    ast_node_metadata = ASTParser.extract_ast_node_metadata(tree)
                except Exception:
                    # Fallback if method doesn't exist yet
                    ast_node_metadata = []
                
                # Parse stdout to extract print outputs
                # The wrapper code captures print output and sends it as JSON
                # We need to extract actual print output from the captured buffer
                print_outputs = []
                actual_stdout = ""
                
                # Try to get actual stdout from execution result
                # The wrapper redirects stdout to a buffer and sends JSON
                # Priority: result.output (parsed JSON) > result.stdout (raw JSON string)
                output_value = None
                
                if result.output:
                    # If result.output is a dict with "output" key, that's the actual print output
                    if isinstance(result.output, dict) and "output" in result.output:
                        output_value = result.output["output"]
                    else:
                        output_value = result.output
                elif result.stdout:
                    # Try to parse JSON from stdout
                    try:
                        parsed_output = json.loads(result.stdout.strip())
                        if isinstance(parsed_output, dict) and "output" in parsed_output:
                            output_value = parsed_output["output"]
                        else:
                            # If not JSON or no "output" key, use stdout as-is
                            output_value = result.stdout
                    except json.JSONDecodeError:
                        # Not JSON, use stdout as-is (direct print output)
                        output_value = result.stdout
                
                # Process output_value to create print_outputs list
                if output_value is not None:
                    if isinstance(output_value, str):
                        # Single string output
                        print_outputs = [output_value]
                        actual_stdout = output_value
                    elif isinstance(output_value, list):
                        # List of outputs
                        print_outputs = [str(v) for v in output_value]
                        actual_stdout = '\\n'.join(print_outputs)
                    else:
                        # Other types
                        print_outputs = [str(output_value)]
                        actual_stdout = str(output_value)
                else:
                    # No output found
                    print_outputs = []
                    actual_stdout = ""
                
                # Find print statements in code and create steps for them
                code_lines = code.split('\n')
                print_statement_count = 0
                for line_num, line in enumerate(code_lines, 1):
                    stripped_line = line.strip()
                    # Check if this line contains a print statement
                    if stripped_line.startswith('print(') or 'print(' in stripped_line:
                        # Find AST node for this line
                        line_nodes = [n for n in ast_node_metadata if n.get("line") == line_num]
                        
                        # Get the output for this print statement
                        output_value = ""
                        if print_outputs:
                            # Match output to print statement by order
                            if print_statement_count < len(print_outputs):
                                output_value = print_outputs[print_statement_count]
                            else:
                                # If we have more print statements than outputs, use the last output
                                output_value = print_outputs[-1] if print_outputs else ""
                        
                        # Create step for print statement
                        print_step = ExecutionStepSchema(
                            stepNumber=step_number,
                            line=line_num,
                            code=stripped_line,
                            state={
                                "message": f"Print: {output_value}" if output_value else "Print statement executed",
                                "ast_info": {
                                    "node_count": ast_structure["node_count"],
                                    "has_input": ast_structure["has_input"],
                                    "classes": list(ast_structure["classes"].keys()),
                                    "ast_nodes": line_nodes if line_nodes else []
                                },
                                "stdout": [output_value] if output_value else [],
                                "step_detail": {
                                    "operation": "print",
                                    "content": stripped_line,
                                    "output": output_value  # This is the actual executed output
                                },
                                "stdout": actual_stdout,
                                "stderr": result.stderr
                            }
                        )
                        steps.append(print_step)
                        step_number += 1
                        print_statement_count += 1
                
                # If no print steps were created, create a general execution step
                if not any(s.state.get("step_detail", {}).get("operation") == "print" for s in steps):
                    execution_step = ExecutionStepSchema(
                        stepNumber=step_number,
                        line=1,  # Will be updated based on AST nodes
                        code=code.split('\n')[0] if code.split('\n') else code,
                        state={
                            "message": "Code executed successfully",
                            "ast_info": {
                                "node_count": ast_structure["node_count"],
                                "has_input": ast_structure["has_input"],
                                "classes": list(ast_structure["classes"].keys()),
                                "ast_nodes": ast_node_metadata  # Include detailed node metadata
                            },
                            "execution_result": result.output if result.output else {},
                            "stdout": result.stdout,
                            "stderr": result.stderr,
                            "stdout": print_outputs if print_outputs else []
                        }
                    )
                    steps.append(execution_step)
            else:
                # Error - create error step with AST info
                # Try to get AST node metadata even for errors
                tree = self.ast_parser.parse_code(code)
                ast_node_metadata = []
                try:
                    from app.services.simulators.operations.ast_parser import ASTParser
                    ast_node_metadata = ASTParser.extract_ast_node_metadata(tree)
                except Exception:
                    pass
                
                error_step = ExecutionStepSchema(
                    stepNumber=step_number,
                    line=1,
                    code=code.split('\n')[0] if code.split('\n') else code,
                    state={
                        "error": result.stderr or "Execution failed",
                        "exit_code": result.exit_code,
                        "timed_out": result.timed_out,
                        "ast_info": {
                            "node_count": ast_structure["node_count"],
                            "has_input": ast_structure["has_input"],
                            "classes": list(ast_structure["classes"].keys()),
                            "ast_nodes": ast_node_metadata  # Include AST nodes even for errors
                        }
                    }
                )
                steps.append(error_step)
            
            return steps
            
        except ValueError as e:
            # AST parsing error or validation error
            # Try to parse AST even if there's an error (might be partial)
            ast_node_metadata = []
            try:
                tree = self.ast_parser.parse_code(code)
                from app.services.simulators.operations.ast_parser import ASTParser
                ast_node_metadata = ASTParser.extract_ast_node_metadata(tree)
            except Exception:
                pass
            
            error_step = ExecutionStepSchema(
                stepNumber=step_number,
                line=1,
                code=code.split('\n')[0] if code.split('\n') else code,
                state={
                    "error": str(e),
                    "error_type": "AST_PARSING_ERROR",
                    "ast_info": {
                        "node_count": len(ast_node_metadata),
                        "has_input": False,
                        "ast_nodes": ast_node_metadata  # Include partial AST if available
                    }
                }
            )
            steps.append(error_step)
            return steps
        
        except Exception as e:
            # Unexpected error
            error_step = ExecutionStepSchema(
                stepNumber=step_number,
                line=1,
                code=code.split('\n')[0] if code.split('\n') else code,
                state={
                    "error": f"Unexpected error: {str(e)}",
                    "error_type": "EXECUTION_ERROR"
                }
            )
            steps.append(error_step)
            return steps
    
    def classify_operations(self, code: str) -> Dict[str, Any]:
        """
        Classify operations in code for visualization approach
        
        Args:
            code: Python code to analyze
            
        Returns:
            Dictionary with operation classification
        """
        ast_structure = self.parse_ast_structure(code)
        
        operations = {
            "has_class_definitions": len(ast_structure["classes"]) > 0,
            "has_method_calls": False,
            "has_assignments": False,
            "has_loops": False,
            "has_conditionals": False,
            "has_input": ast_structure["has_input"],
            "visualization_type": "basic"  # Will be determined by data structure type
        }
        
        # Analyze AST nodes for operation types
        for node_info in ast_structure["ast_nodes"]:
            node_type = node_info.get("type", "")
            if "Call" in node_type:
                operations["has_method_calls"] = True
            elif "Assign" in node_type:
                operations["has_assignments"] = True
            elif "For" in node_type or "While" in node_type:
                operations["has_loops"] = True
            elif "If" in node_type:
                operations["has_conditionals"] = True
        
        return operations
