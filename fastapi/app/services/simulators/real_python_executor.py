import subprocess
import json
import sys
import io
from typing import Optional, Dict, Any
from dataclasses import dataclass
from contextlib import redirect_stdout


@dataclass
class ExecutionResult:
    """Result of Python code execution"""
    stdout: str
    stderr: str
    exit_code: int
    timed_out: bool = False
    output: Optional[Dict[str, Any]] = None  # Parsed JSON output if available


class RealPythonExecutor:
    """Executes Python code in real Python interpreter (not simulation)"""
    
    def __init__(self, use_docker: bool = False, timeout: int = 30, memory: str = "256m"):
        """
        Initialize the executor
        
        Args:
            use_docker: If True, use Docker for execution (more secure but requires Docker)
            timeout: Execution timeout in seconds
            memory: Memory limit for Docker (if used)
        """
        self.use_docker = use_docker
        self.timeout = timeout
        self.memory = memory
        self.python_image = "python:3.11"
    
    def wrap_user_code(self, user_code: str) -> str:
        """
        Wrap user code with JSON output wrapper
        Similar to Go backend's wrapUserCode function
        """
        wrapper = '''import json
import sys
import io
import re
from contextlib import redirect_stdout

try:
    # Read JSON input from stdin (if any)
    input_data = {{}}
    try:
        stdin_input = sys.stdin.read().strip()
        if stdin_input:
            input_data = json.loads(stdin_input)
            # Make input variables available globally
            for key, value in input_data.items():
                globals()[key] = value
    except:
        pass
    
    # Capture print output using redirect_stdout
    class_def_buffer = io.StringIO()
    with redirect_stdout(class_def_buffer):
        # Execute user code here
{user_code}
    
    # Get captured output
    raw_output = class_def_buffer.getvalue()
    
    # Process output - split by newlines and keep each line as-is
    # This preserves the actual print output without trying to evaluate it
    output_lines = [line.rstrip('\\n\\r') for line in raw_output.split('\\n') if line.strip()]
    
    # Wrap output appropriately
    if len(output_lines) == 1:
        result = {{"output": output_lines[0]}}
    elif len(output_lines) > 1:
        result = {{"output": output_lines}}
    else:
        result = {{"output": ""}}
    
    # Output as JSON
    print(json.dumps(result, ensure_ascii=False, separators=(',', ':')))

except Exception as e:
    # Handle errors gracefully
    error_info = {{"error": str(e)}}
    print(json.dumps(error_info, ensure_ascii=False))
'''
        # Indent user code properly
        indented_code = '\n'.join('        ' + line for line in user_code.split('\n'))
        final_code = wrapper.format(user_code=indented_code)
        return final_code
    
    def execute_with_subprocess(self, code: str, stdin_data: str = "", skip_wrapping: bool = False) -> ExecutionResult:
        """
        Execute Python code using subprocess (direct Python execution)
        
        Args:
            code: Python code to execute
            stdin_data: Input data to send to stdin (JSON string)
            skip_wrapping: If True, do not wrap code with JSON output wrapper
            
        Returns:
            ExecutionResult with stdout, stderr, exit_code
        """
        try:
            if skip_wrapping:
                wrapped_code = code
            else:
                wrapped_code = self.wrap_user_code(code)
            
            # Execute Python code
            process = subprocess.Popen(
                [sys.executable, '-c', wrapped_code],
                stdin=subprocess.PIPE,
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
                text=True,
                bufsize=1
            )
            
            try:
                stdout, stderr = process.communicate(
                    input=stdin_data,
                    timeout=self.timeout
                )
                exit_code = process.returncode
                timed_out = False
            except subprocess.TimeoutExpired:
                process.kill()
                stdout, stderr = process.communicate()
                exit_code = -1
                timed_out = True
                stderr = "Code execution timed out"
            
            # Try to parse JSON output
            output = None
            if stdout.strip():
                try:
                    output = json.loads(stdout.strip())
                except json.JSONDecodeError:
                    # If not JSON, keep as string
                    pass
            
            # Debug: Log execution result for troubleshooting
            # The wrapper code should execute the user code and capture evaluated print output
            # If output contains literal strings like "f\"{result:.2f}\"", there might be an execution issue
            import logging
            logger = logging.getLogger(__name__)
            if output and isinstance(output, dict) and "output" in output:
                output_val = output["output"]
                if isinstance(output_val, str) and output_val.startswith('f"') and '{' in output_val:
                    logger.warning(f"Potential issue: Output appears to be literal f-string: {output_val[:100]}")
                    logger.warning(f"Full stdout: {stdout[:500]}")
                    logger.warning(f"Full stderr: {stderr[:500]}")
            
            return ExecutionResult(
                stdout=stdout.strip(),
                stderr=stderr,
                exit_code=exit_code,
                timed_out=timed_out,
                output=output
            )
            
        except Exception as e:
            return ExecutionResult(
                stdout="",
                stderr=f"Execution error: {str(e)}",
                exit_code=-1,
                timed_out=False
            )
    
    def execute_with_docker(self, code: str, stdin_data: str = "", skip_wrapping: bool = False) -> ExecutionResult:
        """
        Execute Python code using Docker (more secure, isolated)
        
        Args:
            code: Python code to execute
            stdin_data: Input data to send to stdin (JSON string)
            skip_wrapping: If True, do not wrap code with JSON output wrapper
            
        Returns:
            ExecutionResult with stdout, stderr, exit_code
        """
        try:
            if skip_wrapping:
                wrapped_code = code
            else:
                wrapped_code = self.wrap_user_code(code)
            
            # Docker command arguments (similar to Go backend)
            docker_args = [
                "docker", "run", "--rm", "-i",
                "--network", "none",
                "--read-only",
                "--user", "0:0",
                "--cap-drop", "ALL",
                "--security-opt", "no-new-privileges",
                "--pids-limit", "128",
                "--ulimit", "nofile=256:256",
                "--tmpfs", "/tmp:rw,nosuid,nodev,noexec,size=64m",
                "-m", self.memory,
                "--cpus", "0.5",
                "-w", "/tmp",
                self.python_image,
                "python", "-c", wrapped_code
            ]
            
            process = subprocess.Popen(
                docker_args,
                stdin=subprocess.PIPE,
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
                text=True,
                bufsize=1
            )
            
            try:
                stdout, stderr = process.communicate(
                    input=stdin_data,
                    timeout=self.timeout
                )
                exit_code = process.returncode
                timed_out = False
            except subprocess.TimeoutExpired:
                process.kill()
                stdout, stderr = process.communicate()
                exit_code = -1
                timed_out = True
                stderr = "Code execution timed out"
            
            # Try to parse JSON output
            output = None
            if stdout.strip():
                try:
                    output = json.loads(stdout.strip())
                except json.JSONDecodeError:
                    # If not JSON, keep as string
                    pass
            
            # Debug: Log execution result for troubleshooting
            # The wrapper code should execute the user code and capture evaluated print output
            # If output contains literal strings like "f\"{result:.2f}\"", there might be an execution issue
            import logging
            logger = logging.getLogger(__name__)
            if output and isinstance(output, dict) and "output" in output:
                output_val = output["output"]
                if isinstance(output_val, str) and output_val.startswith('f"') and '{' in output_val:
                    logger.warning(f"Potential issue: Output appears to be literal f-string: {output_val[:100]}")
                    logger.warning(f"Full stdout: {stdout[:500]}")
                    logger.warning(f"Full stderr: {stderr[:500]}")
            
            return ExecutionResult(
                stdout=stdout.strip(),
                stderr=stderr,
                exit_code=exit_code,
                timed_out=timed_out,
                output=output
            )
            
        except FileNotFoundError:
            # Docker not available, fall back to subprocess
            return self.execute_with_subprocess(code, stdin_data, skip_wrapping)
        except Exception as e:
            return ExecutionResult(
                stdout="",
                stderr=f"Docker execution error: {str(e)}",
                exit_code=-1,
                timed_out=False
            )
    
    def execute(self, code: str, stdin_data: str = "", skip_wrapping: bool = False) -> ExecutionResult:
        """
        Execute Python code using the configured method
        
        Args:
            code: Python code to execute
            stdin_data: Input data to send to stdin (JSON string)
            skip_wrapping: If True, do not wrap code with JSON output wrapper
            
        Returns:
            ExecutionResult with execution results
        """
        if self.use_docker:
            return self.execute_with_docker(code, stdin_data, skip_wrapping)
        else:
            return self.execute_with_subprocess(code, stdin_data, skip_wrapping)
    
    def execute_simple(self, code: str) -> ExecutionResult:
        """
        Execute Python code without wrapping (for testing/debugging)
        
        Args:
            code: Python code to execute
            
        Returns:
            ExecutionResult with execution results
        """
        try:
            process = subprocess.Popen(
                [sys.executable, '-c', code],
                stdin=subprocess.PIPE,
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
                text=True
            )
            
            try:
                stdout, stderr = process.communicate(timeout=self.timeout)
                exit_code = process.returncode
                timed_out = False
            except subprocess.TimeoutExpired:
                process.kill()
                stdout, stderr = process.communicate()
                exit_code = -1
                timed_out = True
                stderr = "Code execution timed out"
            
            return ExecutionResult(
                stdout=stdout.strip(),
                stderr=stderr,
                exit_code=exit_code,
                timed_out=timed_out
            )
            
        except Exception as e:
            return ExecutionResult(
                stdout="",
                stderr=f"Execution error: {str(e)}",
                exit_code=-1,
                timed_out=False
            )
