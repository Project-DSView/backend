from typing import List, Dict, Any
from datetime import datetime
from app.schemas.playground import ExecutionStepSchema
from app.services.simulators.common.base_simulator import BaseSimulator, FunctionDefinitionTracker
from app.services.simulators.stack.enhanced_stack_operation_parser import EnhancedStackOperationParser
from app.services.simulators.operations.ast_parser import ASTParser
from app.services.simulators.operations.error_handler import ErrorHandler
from app.services.simulators.direct_code_executor import DirectCodeExecutor
from app.utils.messages_th import get_class_defined_message, get_stack_message


class StackSimulator(BaseSimulator):
    """Enhanced Stack simulator with detailed step-by-step execution tracking"""
    
    def __init__(self):
        super().__init__("stack")
        self.ast_parser = ASTParser()
        self.direct_executor = DirectCodeExecutor(use_docker=False, timeout=30)  # Use direct execution
        
    def execute_code(self, code: str, exec_id: str, created_at: datetime) -> List[ExecutionStepSchema]:
        """Execute stack code with detailed step-by-step tracking"""
        steps = []
        step_number = 1  # Initialize before try block to handle early exceptions
        self.reset_context()
        

        try:
            # Check if code should be executed directly (e.g. input/print or custom class)
            # If code contains 'class ArrayStack' it might be a custom implementation or simulation
            # If code DOES NOT contain 'class ArrayStack' but uses ArrayStack operations, it's a simulation
            # If code contains input() or arbitrary classes, use direct executor
            has_input = 'input(' in code
            has_fstring = 'f"' in code or "f'" in code
            # Relaxed check: Accept ArrayStack or any class with Stack in name as simulation
            is_simulation = 'class ' in code and ('Stack' in code or 'stack' in code) and not has_input
            
            if has_input or has_fstring or not is_simulation:
                # Use direct executor for arbitrary code
                return self.direct_executor.execute(code, data_structure_type="stack", input_values=getattr(self, '_input_values', []))

            # [NEW] Calculate real execution steps for state tracking (Mirroring QueueSimulator)
            execution_steps = self.direct_executor.execute(
                code, 
                data_structure_type="stack", 
                input_values=getattr(self, '_input_values', [])
            )


            # Parse the entire code using AST to catch syntax errors early
            self.ast_parser.parse_code(code)
            # Check if code contains class definition
            has_class_definition = 'class ArrayStack:' in code
            
            # Initialize ArrayStack class only if needed
            if has_class_definition:
                self._initialize_array_stack_class(steps)
                step_number = 2
                
                # [NEW] Analyze method behaviors
                # Must be called AFTER initialization to update the methods structure
                from app.services.simulators.operations.behavior_analyzer import BehaviorAnalyzer
                analyzer = BehaviorAnalyzer(self.context)
                analyzer.parse_class_methods(code)
            else:
                # Just initialize the class in context without creating a step
                self.context["classes"] = {
                    "ArrayStack": {
                        "type": "class",
                        "methods": ["__init__", "size", "is_empty", "push", "pop", "stackTop", "printStack"],
                        "defined": True,
                        "line_number": 1
                    }
                }
                step_number = 1
            
            # Process each line of executable code
            lines = code.split('\n')
            
            operation_parser = EnhancedStackOperationParser(self.context)
            
            # Track function definition state
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
                        step_detail = exec_step.state.get("step_detail", {})
                        # Check for "print" operation OR if there is actual output (for trace-based steps)
                        if exec_step.line == line_number and (step_detail.get("operation") == "print" or step_detail.get("output")):
                            execution_step_for_line = exec_step
                            break
                    # If not found by line number, try to find by order (first print statement = first output)
                    if not execution_step_for_line:
                        print_steps = [s for s in execution_steps if s.state.get("step_detail", {}).get("operation") == "print" or s.state.get("step_detail", {}).get("output")]
                        # Count how many print statements we've seen so far
                        print_count_before = sum(1 for s in steps if s.state.get("step_detail", {}).get("operation") == "print" or s.state.get("step_detail", {}).get("output"))
                        if print_count_before < len(print_steps):
                            execution_step_for_line = print_steps[print_count_before]
                
                # If we found an execution step with actual output, use it instead of processing the line
                if execution_step_for_line:
                    # Use the execution step which has actual executed output
                    # Make sure we use the actual executed output, not the code content
                    actual_output = execution_step_for_line.state.get("step_detail", {}).get("output")
                    if actual_output:
                        accumulated_stdout.append(actual_output)
                        self.context["stdout"] = list(accumulated_stdout)
                        
                        # Replace the step with actual output
                        execution_step_for_line.state["message"] = f"Print: {actual_output}"
                        execution_step_for_line.state["stdout"] = list(accumulated_stdout)
                        steps.append(execution_step_for_line)
                        step_number += 1
                    else:
                        # Fallback to operation parser if no output found
                        try:
                            if operation_parser.parse_and_execute(stripped_line, line_number, step_number, 
                                                                steps, self._create_execution_step):
                                step_number += 1
                        except Exception as e:
                            steps.append(self._create_execution_step(
                                step_number, line_number, stripped_line, error=str(e)
                            ))
                            raise e
                else:
                    # For non-print statements, use operation parser for visualization state tracking
                    try:
                        if operation_parser.parse_and_execute(stripped_line, line_number, step_number, 
                                                            steps, self._create_execution_step):
                            step_number += 1
                        
                    except Exception as e:
                        steps.append(self._create_execution_step(
                            step_number, line_number, stripped_line, error=str(e)
                        ))
                        raise e
                        
                # Update last processed line
                last_processed_line = line_number
            
            # If we only have execution steps (no detailed steps from operation parser), return execution steps
            # This happens when code only has print statements
            # If we only have execution steps (no detailed steps from operation parser), return execution steps
            # This happens when code only has print statements OR when all operations were inside functions (skipped)
            # If we have execution steps, check if they provide more detail than the parser steps
            # This is critical for capturing steps inside functions which the AST parser skips
            if execution_steps:
                # If execution steps are significantly more detailed (e.g. inside loop/function)
                # OR if we have very few parser steps (parser missed logic)
                # OR if it's a simulation (custom class) where we want line-by-line debugging
                if len(execution_steps) > len(steps) or len(steps) <= 1 or is_simulation:
                    return execution_steps
        
        except ValueError as e:
            # Handle syntax errors and other ValueError from AST parser
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
            offset = None
            if hasattr(e, 'offset'):
                offset = e.offset
            elif hasattr(e, 'col_offset'):
                offset = e.col_offset
            
            if hasattr(e, 'error_type'):
                error_info = {
                    "error_type": e.error_type,
                    "error_message": str(e),
                    "thai_message": error_message,
                    "line_number": error_line,
                    "code_line": code_line
                }
            else:
                # Use ErrorHandler to format the error
                error_info = ErrorHandler.format_error(e, error_line, code_line, offset)
                error_message = error_info.get("python_style_message", error_info["thai_message"])
            
            # Create error step with detailed error information
            error_step = self._create_execution_step(
                step_number, error_line, code_line,
                error=error_message,
                state={
                    "error": error_message,
                    "error_type": error_info.get("error_type", "ValueError"),
                    "error_message": error_info.get("error_message", str(e)),
                    "code_line": code_line,
                    "instances": {},
                    "active": None,
                    "stdout": []
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
                error=error_info["thai_message"],
                state={
                    "error": error_info.get("python_style_message", error_info["thai_message"]),
                    "error_type": error_info["error_type"],
                    "error_message": error_info["error_message"],
                    "code_line": code_line,
                    "instances": {},
                    "active": None,
                    "stdout": []
                }
            )
            steps.append(error_step)
            return steps
        except Exception as e:
            # Handle other exceptions
            error_info = ErrorHandler.format_error(e, 0, "")
            
            if not steps:
                error_step = self._create_execution_step(
                    step_number, 0, code.split('\n')[0] if code else "",
                    error=error_info["thai_message"],
                    state={
                        "error": error_info.get("python_style_message", error_info["thai_message"]),
                        "error_type": error_info["error_type"],
                        "error_message": error_info["error_message"],
                        "instances": {},
                        "active": None,
                        "stdout": []
                    }
                )
                steps.append(error_step)
            return steps
        
        return steps
    
    def _initialize_array_stack_class(self, steps: List[ExecutionStepSchema]):
        """Initialize ArrayStack class in context"""
        self.context["classes"] = {
            "ArrayStack": {
                "type": "class",
                "methods": ["__init__", "size", "is_empty", "push", "pop", "stackTop", "printStack"],
                "defined": True,
                "line_number": 1
            }
        }
        
        # Add class definition step
        steps.append(ExecutionStepSchema(
            stepNumber=1,
            line=1,
            code="class ArrayStack:",
            state={
                **self._create_initial_state(),
                "message": get_class_defined_message("ArrayStack")
            }
        ))
    
    def _create_initial_state(self) -> Dict[str, Any]:
        """Create initial state"""
        return {
            "variables": {},
            "stdout": [],
            "active": None,
            "classes_defined": ["ArrayStack"],
            "step_detail": {
                "operation": "class_definition",
                "class_name": "ArrayStack"
            }
        }
    
    def _create_execution_step(self, step_number: int, line_number: int, code: str, 
                             message: str = None, state: Dict[str, Any] = None, error: str = None) -> ExecutionStepSchema:
        """Create a standardized execution step with enhanced details"""
        if state is None:
            state = {
                "instances": {k: self._get_instance_display(v) for k, v in self.context["instances"].items()},
                "variables": self.context["variables"].copy(),
                "stdout": self.context["stdout"].copy(),
                "active": self.context["active_instance"]
            }
        
        if message:
            state["message"] = message
        
        # Extract error from step_detail if not provided directly
        if not error and state.get("step_detail") and isinstance(state["step_detail"], dict):
            error = state["step_detail"].get("error")
        
        if error:
            state["error"] = error
            
        return ExecutionStepSchema(
            stepNumber=step_number,
            line=line_number,
            code=code,
            state=state
        )
    
    def _get_instance_display(self, instance):
        """Get display representation of stack instances"""
        if instance.get("class_type") == "ArrayStack":
            return {
                "type": "ArrayStack",
                "data": instance.get("data", []),
                "size": len(instance.get("data", [])),
                "isEmpty": len(instance.get("data", [])) == 0,
                "top": instance.get("data", [])[-1] if instance.get("data", []) else None
            }
        return super()._get_instance_display(instance)


class FunctionDefinitionTracker:
    """Helper class to track function definition state"""
    
    def __init__(self):
        self.in_function_def = False
        self.function_indent_level = 0
    
    def update_state(self, original_line: str, stripped_line: str) -> bool:
        """Update function definition state and return True if line should be skipped"""
        # Track function definitions
        if stripped_line.startswith('def '):
            self.in_function_def = True
            self.function_indent_level = len(original_line) - len(original_line.lstrip())
            return True
        elif self.in_function_def:
            current_indent = len(original_line) - len(original_line.lstrip())
            if current_indent <= self.function_indent_level and stripped_line:
                self.in_function_def = False
                return False
            else:
                return True  # Skip function body lines
        
        return False