import subprocess
import json
import sys
import io
import re
from typing import Optional, Dict, Any, List, Union
from dataclasses import dataclass, field
from app.services.simulators.real_python_executor import RealPythonExecutor, ExecutionResult

@dataclass
class InteractiveExecutionResult(ExecutionResult):
    """Result of interactive Python code execution"""
    trace: List[Dict[str, Any]] = field(default_factory=list)
    input_history: List[str] = field(default_factory=list)

class InteractivePythonExecutor(RealPythonExecutor):
    """
    Executes Python code with interactive capabilities and tracing
    """
    
    def execute_interactive(
        self, 
        code: str, 
        stdin_data: str = "",
        input_callback: Optional[Any] = None,
        input_values: Optional[List[str]] = None
    ) -> InteractiveExecutionResult:
        """
        Execute code with interactive support and tracing
        
        Args:
            code: Python code to execute
            stdin_data: Initial stdin data (JSON string)
            input_callback: Callback for handling input requests (not used in direct execution mode)
            input_values: Pre-collected input values to feed into input() calls
            
        Returns:
            InteractiveExecutionResult with trace data
        """
        try:
            # 1. Wrap code to support tracing and input interception
            wrapped_code = self._wrap_code_with_trace(code, input_values)
            
            # 2. Execute wrapped code using the underlying executor mechanism
            # We pass the original stdin_data which might contain initial variables
            # 2. Execute wrapped code using the underlying executor mechanism
            # We pass the original stdin_data which might contain initial variables
            if self.use_docker:
                base_result = self.execute_with_docker(wrapped_code, stdin_data, skip_wrapping=True)
            else:
                base_result = self.execute_with_subprocess(wrapped_code, stdin_data, skip_wrapping=True)
            
            # 3. Parse the output to extract trace and result
            return self._parse_interactive_output(base_result)
            
        except Exception as e:
            return InteractiveExecutionResult(
                stdout="",
                stderr=f"Interactive execution error: {str(e)}",
                exit_code=-1,
                timed_out=False
            )
    
    def _wrap_code_with_trace(self, user_code: str, input_values: Optional[List[str]] = None) -> str:
        """
        Wrap user code with tracing logic and input handling
        """
        # Prepare input values as a Python list literal
        input_values_repr = json.dumps(input_values if input_values else [])
        
        wrapper = '''
import sys
import io
import json
import traceback
import inspect
import time

# Initialize trace data
trace_data = []
input_values = {input_values_repr}
input_index = 0
captured_stdout = io.StringIO()
last_stdout_len = 0
last_trace_entry = None
last_trace_output_start = 0  # Track where this trace's output starts
last_line_time = time.perf_counter()  # Track execution time per line

# Set to track already-seen object ids for circular reference detection
_seen_ids = set()

def safe_serialize(obj, depth=0, max_depth=10, return_size=False):
    """Safely serialize an object, handling circular references and depth limits"""
    global _seen_ids
    
    if depth > max_depth:
        return 0 if return_size else "<max depth exceeded>"
    
    # Handle None
    if obj is None:
        return 0 if return_size else None
    
    # Handle primitives (fast path)
    if isinstance(obj, (bool, int, float, str)):
        return sys.getsizeof(obj) if return_size else obj
    
    # Handle circular references for mutable objects
    obj_id = id(obj)
    if obj_id in _seen_ids:
        if return_size:
            return 0
        # Show more informative circular reference with type and key data
        type_name = type(obj).__name__
        # Try to get a meaningful identifier (data, val, value, id, name)
        if hasattr(obj, '__dict__'):
            obj_dict = obj.__dict__
            for key in ['data', 'val', 'value', 'id', 'name', 'key']:
                if key in obj_dict:
                    try:
                        val = obj_dict[key]
                        if isinstance(val, (int, float, str, bool)) and not isinstance(val, bool) or isinstance(val, bool):
                            return f"{{type_name}}({{val}})"
                    except:
                        pass
        return f"{{type_name}}"
    
    try:
        _seen_ids.add(obj_id)
        
        current_size = sys.getsizeof(obj) if return_size else 0
        
        if isinstance(obj, (list, tuple)):
            if return_size:
                for item in obj:
                    current_size += safe_serialize(item, depth + 1, max_depth, return_size=True)
                return current_size
            
            # Optimization: slice large lists
            if len(obj) > 50:
                return [safe_serialize(item, depth + 1, max_depth) for item in obj[:50]] + ["<truncated>"]
            return [safe_serialize(item, depth + 1, max_depth) for item in obj]
        elif isinstance(obj, dict):
            if return_size:
                for k, v in obj.items():
                    current_size += safe_serialize(k, depth + 1, max_depth, return_size=True)
                    current_size += safe_serialize(v, depth + 1, max_depth, return_size=True)
                return current_size

            # Optimization: limit dict size
            if len(obj) > 50:
                serialized = {{}}
                count = 0
                for k, v in obj.items():
                    if isinstance(k, (str, int, float, bool)):
                        serialized[k] = safe_serialize(v, depth + 1, max_depth)
                        count += 1
                        if count >= 50:
                            serialized["<truncated>"] = "..."
                            break
                return serialized
            return {{k: safe_serialize(v, depth + 1, max_depth) for k, v in obj.items() if isinstance(k, (str, int, float, bool))}}
        elif isinstance(obj, set):
            if return_size:
                for item in obj:
                    current_size += safe_serialize(item, depth + 1, max_depth, return_size=True)
                return current_size

            # Convert set to list for serialization
            items = list(obj)
            if len(items) > 50:
                return [safe_serialize(item, depth + 1, max_depth) for item in items[:50]] + ["<truncated>"]
            return [safe_serialize(item, depth + 1, max_depth) for item in items]
        if hasattr(obj, '__dict__'):
            if return_size:
                current_size += safe_serialize(obj.__dict__, depth + 1, max_depth, return_size=True)
                return current_size

            # Handle class instances
            # Optimization: Check for data structure types first by class name if possible
            # or rely on specific field existence.
            
            # Fast check for simple data structures
            is_data_struct = False
            
            # Check for common attributes found in Linked Lists / Trees / Graphs
            # We use a more targeted check to avoid iterating 100+ items for every object
            obj_dict = obj.__dict__
            
            # List of keys that strongly suggest a data structure
            # List of keys that strongly suggest a data structure
            ds_keys = {{'head', 'next', 'prev', 'data', 'val', 'value', 'count', 'root', 'nodes', 'edges', 'left', 'right', 'adjacency_list', 'graph', 'adj_list'}}
            
            # Check intersection - faster than iterating
            if any(k in obj_dict for k in ds_keys):
                 is_relevant_obj = True
            else:
                 is_relevant_obj = False
            
            if is_relevant_obj:
                 # Serialize the object's dictionary, filtering out internals
                 result = {{"type": type(obj).__name__}}
                 # We only serialize public attributes
                 for k, v in obj_dict.items():
                     if not k.startswith('_'):
                         result[k] = safe_serialize(v, depth + 1, max_depth)
                 return result
            else:
                return str(obj)
        else:
            return sys.getsizeof(obj) if return_size else str(obj)
    finally:
        pass


# Custom input function to handle interactive input
def custom_input(prompt=''):
    global input_index
    
    # Print the prompt to stdout so it's captured
    print(prompt, end='')
    
    # If we have pre-supplied input values, use them
    if input_index < len(input_values):
        value = input_values[input_index]
        input_index += 1
        print(value) # Echo input like a real terminal
        return value
    
    # If no input values remain, return a clear execution error as requested
    raise Exception("No more input provided")

# Trace function to capture line execution and variables
def trace_lines(frame, event, arg):
    global last_stdout_len, last_trace_entry, _seen_ids, input_index, last_line_time
    
    if event != 'line':
        return trace_lines
    
    co = frame.f_code
    func_name = co.co_name
    line_no = frame.f_lineno
    filename = co.co_filename
    
    # Only trace the actual string code (filename will be <string>)
    if filename != "<string>":
        return trace_lines

    # Skip tracing our internal helper functions to prevent infinite recursion
    if func_name in ['trace_lines', 'safe_serialize', 'custom_input']:
        return None

    # 1. Check for output from the PREVIOUS line
    current_stdout_val = captured_stdout.getvalue()
    if last_trace_entry is not None:
        # The output produced SINCE the last entry - don't strip to preserve end=" "
        diff = current_stdout_val[last_stdout_len:]
        if diff:
            last_trace_entry["output"] = diff  # Keep whitespace for end=" " support
    
    # Update length for next check
    last_stdout_len = len(current_stdout_val)

    # 2. Capture local variables NOW (which represents state AFTER previous line executed)
    # This is the key fix: we capture the current frame's variables which reflect
    # the state after the previous line ran (before the current line runs)
    # We filter out internal variables and make sure they are serializable
    # Reset seen ids for this round of serialization
    _seen_ids = set()
    local_vars = {{}}
    for key, value in frame.f_locals.items():
        if key.startswith('__') or key.startswith('_'): 
            continue
        # Skip internal functions and modules
        if callable(value) and not hasattr(value, 'data'):
            continue
        if isinstance(value, type(sys)):  # Skip modules
            continue
            
        # Use safe serialization
        local_vars[key] = safe_serialize(value)
    
    # 3. Update PREVIOUS trace entry with current variable state
    # This ensures the previous line's visualization shows state AFTER it executed
    if last_trace_entry is not None:
        last_trace_entry["variables"] = local_vars
            
    # Calculate execution time for previous entry
    current_time = time.perf_counter()
    if last_trace_entry is not None:
        last_trace_entry["execution_time"] = current_time - last_line_time
    last_line_time = current_time
    
    # Capture memory usage
    try:
        current_mem, peak_mem = tracemalloc.get_traced_memory()
    except:
        current_mem, peak_mem = 0, 0

    # [NEW] Trace back to find the user command (line in <module>) that triggered this
    caller_line = None
    try:
        current_frame = frame
        # Limit depth to prevent infinite loops (though f_back usually terminates)
        depth = 0
        while current_frame and depth < 20:
            if current_frame.f_code.co_name == '<module>':
                caller_line = current_frame.f_lineno
                break
            current_frame = current_frame.f_back
            depth += 1
    except:
        pass

    # 4. Create new trace entry for CURRENT line (variables will be updated on next line entry)
    trace_entry = {{
        "line": line_no,
        "caller_line": caller_line,
        "variables": {{}},  # Will be populated when next line is entered
        "event": event,
        "func": func_name,
        "input_index": input_index,
        "execution_time": 0,  # Will be calculated on next line
        "memory_usage": current_mem 
    }}
    trace_data.append(trace_entry)
    last_trace_entry = trace_entry
    
    return trace_lines


import builtins
builtins.input = custom_input


# Redirect stdout to capture output
original_stdout = sys.stdout
sys.stdout = captured_stdout

try:
    # Start memory tracking
    import tracemalloc
    tracemalloc.start()

    # Set trace
    sys.settrace(trace_lines)
    
    # Execute User Code
    # We use exec/compile to ensure filenames match what tracer expects
    compiled_code = compile("""{user_code}""", "<string>", "exec")
    exec(compiled_code, globals())
    
    # Check for output from the LAST line - don't strip
    current_stdout_val = captured_stdout.getvalue()
    if last_trace_entry is not None:
        diff = current_stdout_val[last_stdout_len:]
        if diff:
            last_trace_entry["output"] = diff  # Keep whitespace
        
        # IMPORTANT: Capture final variable state for the LAST traced line
        # Since there's no "next line" to trigger the look-back capture,
        # we need to manually capture the final state here
        _seen_ids = set()
        final_vars = {{}}
        # Create a copy of globals items to avoid "dictionary changed size during iteration" error
        exec_globals_items = list(globals().items())
        for key, value in exec_globals_items:
            if key.startswith('__') or key.startswith('_'): 
                continue
            # Skip internal trace variables
            if key in ['trace_data', 'input_values', 'input_index', 'captured_stdout', 
                       'last_stdout_len', 'last_trace_entry', 'last_trace_output_start',
                       'last_line_time', '_seen_ids', 'safe_serialize', 'custom_input',
                       'trace_lines', 'original_stdout', 'compiled_code', 'current_stdout_val',
                       'diff', 'final_vars', 'exec_globals_items', 'inspect',
                       'sys', 'io', 'json', 'traceback', 'time', 'builtins', 'tracemalloc']:
                continue
            # Skip internal functions and modules
            if callable(value) and not hasattr(value, '__dict__'):
                continue
            if isinstance(value, type(sys)):  # Skip modules
                continue
            try:
                final_vars[key] = safe_serialize(value)
            except:
                pass
        
        # Update the last trace entry with final state
        if final_vars:
            last_trace_entry["variables"] = final_vars
    
    # Stop memory tracking
    tracemalloc.stop()
            
except Exception as e:
    # Capture exception in trace
    trace_data.append({{
        "event": "exception",
        "error": str(e),
        "traceback": traceback.format_exc()
    }})

finally:
    sys.settrace(None)
    sys.stdout = original_stdout
    
    # Reset seen ids for final serialization
    _seen_ids = set()
    
    # Safe serialize trace_data to avoid circular references
    safe_trace = []
    for entry in trace_data:
        safe_entry = {{
            "line": entry.get("line"),
            "caller_line": entry.get("caller_line"),
            "event": entry.get("event"),
            "func": entry.get("func"),
            "output": entry.get("output"),
            "variables": {{}},
            "memory_usage": entry.get("memory_usage", 0),
            "execution_time": entry.get("execution_time", 0)
        }}
        if "variables" in entry:
            for k, v in entry["variables"].items():
                safe_entry["variables"][k] = safe_serialize(v)
        if "error" in entry:
            safe_entry["error"] = entry.get("error")
        if "traceback" in entry:
            safe_entry["traceback"] = entry.get("traceback")
        safe_trace.append(safe_entry)
    
    # Output the result wrapped in a special format for parsing
    result = {{
        "trace": safe_trace,
        "stdout": captured_stdout.getvalue(),
        "input_history": input_values[:input_index]
    }}
    
    # We print a special delimiter to separate logic from any other output
    print("---INTERACTIVE_EXECUTION_RESULT---")
    print(json.dumps(result, default=str))
'''
        # Indent user code not needed for compile/exec in the wrapper exactly like this
        # but we need to escape backslashes and quotes for the triple-quoted string
        escaped_code = user_code.replace('\\', '\\\\').replace('"', '\\"')
        
        final_code = wrapper.format(input_values_repr=input_values_repr, user_code=escaped_code)
        return final_code

    def _parse_interactive_output(self, base_result: ExecutionResult) -> InteractiveExecutionResult:
        """
        Parse the output from the wrapped execution to extract trace and result
        """
        stdout = base_result.stdout
        stderr = base_result.stderr
        
        trace = []
        input_history = []
        actual_stdout = stdout
        parsed_output = {}
        
        # Look for our special delimiter
        delimiter = "---INTERACTIVE_EXECUTION_RESULT---"
        if delimiter in stdout:
            parts = stdout.split(delimiter)
            actual_stdout = parts[0] # Everything before is actual stdout (if any leaked)
            json_part = parts[1].strip()
            
            try:
                result_data = json.loads(json_part)
                trace = result_data.get("trace", [])
                input_history = result_data.get("input_history", [])
                
                # The captured stdout inside the wrapper is usually more accurate
                if "stdout" in result_data:
                    actual_stdout = result_data["stdout"]
                    
                parsed_output = result_data
                
            except json.JSONDecodeError:
                stderr += "\nFailed to parse interactive result JSON"
        
        # Calculate memory delta from previous step (memory_usage is now captured by tracemalloc)
        previous_memory = 0
        for entry in trace:
            current_memory = entry.get("memory_usage", 0)
            entry["memory_delta"] = current_memory - previous_memory
            previous_memory = current_memory
            
            # Ensure memory_usage is at least 0
            if current_memory < 0:
                entry["memory_usage"] = 0
                
        return InteractiveExecutionResult(
            stdout=actual_stdout,
            stderr=stderr,
            exit_code=base_result.exit_code,
            timed_out=base_result.timed_out,
            output=parsed_output, # Store full parsed result
            trace=trace,
            input_history=input_history
        )
