from typing import List
from datetime import datetime
from app.schemas.playground import ExecutionStepSchema
from app.services.simulators.common.base_simulator import BaseSimulator
from app.services.simulators.operations.ast_parser import ASTParser
from app.services.simulators.operations.enhanced_operation_parser import EnhancedOperationParser
from app.services.simulators.operations.error_handler import ErrorHandler
from app.services.simulators.direct_code_executor import DirectCodeExecutor


class SinglyLinkedListSimulator(BaseSimulator):
    """Main SinglyLinkedList simulator class that orchestrates the simulation process"""
    
    def __init__(self):
        super().__init__("singlylinkedlist")
        self.ast_parser = ASTParser()
        self.operation_parser = None  # Initialize in execute_code to pass context
        self.direct_executor = DirectCodeExecutor(use_docker=False, timeout=30)  # Use direct execution
    
    def _get_operation_parser(self):
        """Get the operation parser for singly linked list"""
        if self.operation_parser is None:
            self.operation_parser = EnhancedOperationParser(self.context)
        return self.operation_parser
    
    def execute_code(self, code: str, exec_id: str, created_at: datetime) -> List[ExecutionStepSchema]:
        """
        Execute singly linked list code using AST-first + real Python execution
        Combines AST parsing with real execution results
        """
        steps = []
        step_number = 1
        self.reset_context()  # Ensure clean context
        self.operation_parser = EnhancedOperationParser(self.context)
        
        try:
            # Get input_values if available (set by controller)
            input_values = getattr(self, '_input_values', None)
            
            # Step 1: Check if code should be executed directly (e.g. input/print or custom class)
            # If code contains input() or arbitrary usage, use direct executor
            has_input = 'input(' in code
            has_fstring = 'f"' in code or "f'" in code
            
            # Simple check for simulation vs direct execution
            is_simulation = 'class LinkedList:' in code or ('LinkedList' in code and not has_input)
            
            if has_input or has_fstring or not is_simulation:
                # Use direct executor for arbitrary code
                return self.direct_executor.execute(code, data_structure_type="singlylinkedlist", input_values=input_values)

            # Execution steps used for merging output later if in simulation mode
            execution_steps = self.direct_executor.execute(
                code,
                stdin_data="",
                data_structure_type="singlylinkedlist",
                input_values=input_values
            )
            
            # Step 2: Parse AST for class definitions and state tracking
            tree = self.ast_parser.parse_code(code)
            self.operation_parser.set_class_code(code)
            
            # Step 3: Handle class definitions with proper state
            classes = self.ast_parser.extract_classes(tree)
            if classes:
                class_steps = self.ast_parser.create_class_definition_steps(
                    classes, self.context, self._get_instance_display
                )
                # Merge with AST executor steps, avoiding duplicates
                if not steps or steps[0].code != class_steps[0].code if class_steps else True:
                    steps.extend(class_steps)
                    step_number = len(steps) + 1
            
            # Step 4: Process executable lines for detailed state tracking
            # This maintains backward compatibility with visualization
            lines = code.split('\n')
            from app.services.simulators.common.base_simulator import FunctionDefinitionTracker
            function_tracker = FunctionDefinitionTracker()
            
            # Track accumulated stdout
            accumulated_stdout = []
            
            # Keep track of last processed line to find skipped outputs
            last_processed_line = 0
            
            for line_number, line in enumerate(lines, 1):
                original_line = line
                stripped_line = line.strip()
                
                # Skip empty lines and comments
                if not stripped_line or stripped_line.startswith('#'):
                    continue
                
                # Handle function definitions
                if function_tracker.update_state(original_line, stripped_line):
                    continue
                
                # Check for skipped outputs between last_processed and current
                if execution_steps:
                    for s in execution_steps:
                        if last_processed_line < s.line < line_number:
                             # This step corresponds to a skipped line
                             out = s.state.get("step_detail", {}).get("output")
                             if out and out not in accumulated_stdout:
                                 accumulated_stdout.append(out)
                                 self.context["stdout"] = list(accumulated_stdout)
                
                # Skip class definitions 
                if stripped_line.startswith('class '):
                    continue
                
                # Skip function body indented lines
                if original_line.startswith('    ') and not stripped_line:
                    continue
                
                # Check if this line is a print statement - if so, use output from direct executor
                is_print_statement = stripped_line.startswith('print(') or 'print(' in stripped_line
                execution_step_for_line = None
                if is_print_statement and execution_steps:
                    # Find the corresponding execution step for this print statement
                    for exec_step in execution_steps:
                        if exec_step.line == line_number and exec_step.state.get("step_detail", {}).get("operation") == "print":
                            execution_step_for_line = exec_step
                            break
                
                # If we found an execution step with actual output, use it instead of processing the line
                if execution_step_for_line and execution_step_for_line.state.get("step_detail", {}).get("output"):
                    # Use the execution step which has actual executed output
                    actual_output = execution_step_for_line.state.get("step_detail", {}).get("output")
                    accumulated_stdout.append(actual_output)
                    self.context["stdout"] = list(accumulated_stdout)
                    
                    execution_step_for_line.state["message"] = f"Print: {actual_output}"
                    execution_step_for_line.state["stdout"] = list(accumulated_stdout)
                    
                    steps.append(execution_step_for_line)
                    step_number += 1
                else:
                    # Process the line to get detailed state
                    line_steps = self._process_line(
                        original_line, line_number, step_number, 
                        self.operation_parser, function_tracker
                    )
                    
                    # Enhance steps with real execution results from direct executor
                    for step in line_steps:
                        # For print statements, merge actual output from execution steps
                        if is_print_statement and execution_steps:
                            for exec_step in execution_steps:
                                if exec_step.line == line_number and exec_step.state.get("step_detail", {}).get("operation") == "print":
                                    # Merge actual output from direct executor
                                    exec_output = exec_step.state.get("step_detail", {}).get("output")
                                    if exec_output:
                                        if exec_output not in accumulated_stdout: # Avoid double counting if iterating same line
                                             accumulated_stdout.append(exec_output)
                                             self.context["stdout"] = list(accumulated_stdout)
                                        
                                        if not step.state.get("step_detail"):
                                            step.state["step_detail"] = {}
                                        step.state["step_detail"]["output"] = exec_output
                                        step.state["step_detail"]["operation"] = "print"
                                        step.state["stdout"] = list(accumulated_stdout)
                                        step.state["message"] = f"Print: {exec_output}"
                                    break
                        
                        # Merge execution info if available
                        if execution_steps and len(execution_steps) > 0:
                            last_exec_step = execution_steps[-1]
                            if last_exec_step.state.get("execution_result"):
                                step.state["execution_result"] = last_exec_step.state.get("execution_result")
                        
                        # Also add AST node metadata for current line
                        try:
                            tree = self.ast_parser.parse_code(code)
                            # Use static method from ASTParser
                            from app.services.simulators.operations.ast_parser import ASTParser
                            ast_node_metadata = ASTParser.extract_ast_node_metadata(tree)
                            # Filter nodes for current line
                            line_nodes = [n for n in ast_node_metadata if n.get("line") == line_number]
                            if line_nodes:
                                if not step.state.get("ast_info"):
                                    step.state["ast_info"] = {}
                                if "ast_nodes" not in step.state["ast_info"]:
                                    step.state["ast_info"]["ast_nodes"] = []
                                step.state["ast_info"]["ast_nodes"].extend(line_nodes)
                        except Exception:
                            # If AST parsing fails, continue without AST info
                            pass
                    
                    steps.extend(line_steps)
                    step_number += len(line_steps)
                
                # Update last processed line
                last_processed_line = line_number
            
            # If we have execution steps but no detailed steps, use execution steps
            if not steps and execution_steps:
                return execution_steps
            
            # If execution_steps are more detailed, return them instead
            # This ensures line-by-line trace data is used for debugging
            if execution_steps:
                if len(execution_steps) > len(steps) or len(steps) <= 1 or is_simulation:
                    return execution_steps
            
            return steps
            
        except ValueError as e:
            # Handle syntax errors and other ValueError from AST parser
            # The error message should already be formatted by ErrorHandler
            error_message = str(e)
            
            # Try to extract line number from error message or code
            error_line = 0
            code_line = ""
            if hasattr(e, 'line_number'):
                error_line = e.line_number
            elif hasattr(e, 'lineno'):
                error_line = e.lineno
            
            # Try to get the problematic line from code
            if error_line > 0:
                lines = code.split('\n')
                if error_line <= len(lines):
                    code_line = lines[error_line - 1]
            
            # Format error with ErrorHandler if we have more details
            if hasattr(e, 'python_style_message'):
                error_info = {
                    "error_type": e.error_type,
                    "error_message": e.error_message,
                    "thai_message": e.thai_message,
                    "python_style_message": e.python_style_message,
                    "line_number": error_line,
                    "code_line": code_line
                }
                error_message = e.python_style_message
            elif hasattr(e, 'error_type'):
                error_info = {
                    "error_type": e.error_type,
                    "error_message": str(e),
                    "thai_message": error_message,
                    "line_number": error_line,
                    "code_line": code_line
                }
            else:
                # Use ErrorHandler to format the error
                error_info = ErrorHandler.format_error(e, error_line, code_line, getattr(e, 'offset', None))
                error_message = error_info.get("python_style_message", error_info.get("thai_message", str(e)))
            
            # Create error step with detailed error information
            error_step = self._create_execution_step(
                step_number, error_line, code_line,
                error=error_message,
                additional_state={
                    "error_type": error_info.get("error_type", "ValueError"),
                    "error_message": error_info.get("error_message", str(e)),
                    "code_line": code_line
                }
            )
            steps.append(error_step)
            return steps
        except SyntaxError as e:
            # Handle syntax errors directly
            error_line = e.lineno if hasattr(e, 'lineno') and e.lineno else 0
            code_line = e.text if hasattr(e, 'text') and e.text else ""
            offset = e.offset if hasattr(e, 'offset') and e.offset else None
            
            error_info = ErrorHandler.format_error(e, error_line, code_line, offset)
            
            error_step = self._create_execution_step(
                step_number, error_line, code_line,
                error=error_info.get("python_style_message", error_info["thai_message"]),
                additional_state={
                    "error_type": error_info["error_type"],
                    "error_message": error_info["error_message"],
                    "code_line": code_line
                }
            )
            steps.append(error_step)
            return steps
        except Exception as e:
            # Handle other exceptions
            error_info = ErrorHandler.format_error(e, 0, "")
            
            error_step = self._create_execution_step(
                step_number, 0, "",
                error=error_info.get("python_style_message", error_info["thai_message"]),
                additional_state={
                    "error_type": error_info["error_type"],
                    "error_message": error_info["error_message"]
                }
            )
            steps.append(error_step)
            return steps
    
    def _process_line(self, line: str, line_number: int, step_number: int, 
                     operation_parser, function_tracker) -> List[ExecutionStepSchema]:
        """
        Process a single line of code and return execution steps
        Enhanced to ensure proper state is always included with linked list data
        """
        steps = []
        
        # Create a step creation function that matches the expected signature
        def create_step(step_num, line_num, code, message=None, error=None, additional_state=None):
            step = self._create_execution_step(step_num, line_num, code, message, error, additional_state)
            
            # Ensure step always has proper linked list state
            # This is critical for visualization to work
            
            # 1. Ensure instances exist and have proper display data
            if "instances" not in step.state:
                step.state["instances"] = {}
            
            # Update instances with current context data using _get_instance_display
            instances = self.context.get("instances", {})
            if isinstance(instances, dict):
                for instance_name, instance_data in instances.items():
                    try:
                        # Use _get_instance_display to get proper node list
                        display_data = self._get_instance_display(instance_data)
                        step.state["instances"][instance_name] = display_data
                    except Exception:
                        # Fallback to empty list if display fails
                        step.state["instances"][instance_name] = []
            
            # 2. Ensure linkedlist array is included
            if "linkedlist" not in step.state:
                linkedlist = self.context.get("linkedlist", [])
                if isinstance(linkedlist, list):
                    step.state["linkedlist"] = linkedlist.copy()
                else:
                    step.state["linkedlist"] = []
            
            # 3. Ensure nodes dictionary is included with names
            nodes = self.context.get("nodes", {})
            if isinstance(nodes, dict):
                if "nodes" not in step.state:
                    step.state["nodes"] = {}
                for k, v in nodes.items():
                    if isinstance(v, dict):
                        step.state["nodes"][k] = v.get("name", "")
                    else:
                        step.state["nodes"][k] = str(v) if v is not None else ""
            
            # 4. Ensure variables are included
            variables = self.context.get("variables", {})
            if isinstance(variables, dict) and variables:
                safe_vars = {}
                for k, v in variables.items():
                    try:
                        if isinstance(v, (str, int, float, bool)) or v is None:
                            safe_vars[k] = v
                        else:
                            safe_vars[k] = str(v)
                    except Exception:
                        safe_vars[k] = "undefined"
                if safe_vars:
                    step.state["variables"] = safe_vars
            
            # 5. Ensure message is set if not provided
            if not step.state.get("message") and message:
                step.state["message"] = message
            
            return step
        
        # Use the parse_and_execute method from EnhancedOperationParser
        try:
            success = operation_parser.parse_and_execute(line, line_number, step_number, steps, create_step)
            if not success:
                # If no operation was matched, create a basic step with proper state
                basic_step = create_step(step_number, line_number, line, f"Executed: {line.strip()}")
                steps.append(basic_step)
        except Exception as e:
            error_step = create_step(step_number, line_number, line, error=str(e))
            steps.append(error_step)
        
        return steps
