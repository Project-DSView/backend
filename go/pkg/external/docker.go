package external

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type DockerConfig struct {
	Image   string
	Timeout time.Duration
	Memory  string // e.g. "256m"
	CPUs    string // e.g. "0.5"
}

type ExecResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
	TimedOut bool
}

type DockerExecutor struct {
	cfg DockerConfig
}

func NewDockerExecutor(cfg DockerConfig) *DockerExecutor {
	if cfg.Image == "" {
		cfg.Image = "python:3.11"
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 5 * time.Second
	}
	if cfg.Memory == "" {
		cfg.Memory = "256m"
	}
	if cfg.CPUs == "" {
		cfg.CPUs = "0.5"
	}
	return &DockerExecutor{cfg: cfg}
}

// wrapUserCode wraps user's Python code with JSON output wrapper
func (e *DockerExecutor) wrapUserCode(userCode string) string {
	wrapper := `import json
import sys
import io
import re
from contextlib import redirect_stdout

# Create StringIO to capture stdout
output_buffer = io.StringIO()

def parse_and_wrap_output(raw_output, use_last_only=False):
    """Parse output and wrap in appropriate structure"""
    raw_output = raw_output.strip()
    
    if not raw_output:
        return ""
    
    # Split into lines
    lines = [line.strip() for line in raw_output.split('\n') if line.strip()]
    
    if not lines:
        return ""
    
    # Process each line
    parsed_lines = []
    for line in lines:
        # Try to evaluate as Python literal first
        try:
            import ast
            evaluated = ast.literal_eval(line)
            parsed_lines.append(evaluated)
        except:
            # If not a literal, keep as string
            parsed_lines.append(line)
    
    # If use_last_only is True and we have multiple lines, use only the last one
    if use_last_only and len(parsed_lines) > 1:
        parsed_lines = [parsed_lines[-1]]
    
    # Determine result
    if len(parsed_lines) == 1:
        result = parsed_lines[0]
    else:
        result = parsed_lines
    
    # If result is a simple string, number, or boolean - wrap it
    if isinstance(result, (str, int, float, bool)):
        # Check if it looks like it should be wrapped in "output" key
        return {"output": result}
    elif isinstance(result, list) and all(isinstance(item, (str, int, float, bool)) for item in result):
        # For simple lists, wrap in output too
        if len(result) == 1:
            return {"output": result[0]}
        else:
            return {"output": result}
    else:
        # For complex objects, return as-is
        return result

def execute_operations(operations, instance_name="mylist"):
    """Execute operations array on the specified instance"""
    # Always create a fresh instance for operations execution
    # This ensures each test case starts with a clean state
    if 'SinglyLinkedList' not in globals():
        return None
    
    # Create a new instance for this test case
    test_instance = globals()['SinglyLinkedList']()
    instance = test_instance
    last_output = None
    
    for op in operations:
        try:
            # Parse operation like "insertFront(\"Tony\")" or "traverse()"
            # Extract method name and parameters
            match = re.match(r'(\w+)\((.*)\)', op)
            if match:
                method_name = match.group(1)
                params_str = match.group(2).strip()
                
                # Get the method
                if hasattr(instance, method_name):
                    method = getattr(instance, method_name)
                    
                    # Parse parameters
                    params = []
                    if params_str:
                        # Handle string parameters with quotes
                        # Extract quoted strings
                        param_matches = re.findall(r'"([^"]*)"', params_str)
                        if param_matches:
                            params = param_matches
                        else:
                            # Try to evaluate as Python literal
                            try:
                                import ast
                                # Handle multiple parameters separated by comma
                                if ',' in params_str:
                                    # Split and evaluate each
                                    parts = [p.strip() for p in params_str.split(',')]
                                    params = [ast.literal_eval(p) for p in parts]
                                else:
                                    params = [ast.literal_eval(params_str)]
                            except:
                                params = [params_str]
                    
                    # If this is traverse(), capture the output
                    if method_name == "traverse":
                        # Create a temporary buffer to capture traverse() output
                        temp_buffer = io.StringIO()
                        with redirect_stdout(temp_buffer):
                            if params:
                                method(*params)
                            else:
                                method()
                        last_output = temp_buffer.getvalue().strip()
                    else:
                        # For other operations, execute normally (no output capture needed)
                        if params:
                            method(*params)
                        else:
                            method()
        except Exception as e:
            # Continue with next operation if one fails
            pass
    
    return last_output

try:
    # Read JSON input from stdin (if any)
    input_data = {}
    try:
        stdin_input = sys.stdin.read().strip()
        if stdin_input:
            input_data = json.loads(stdin_input)
            # Make input variables available globally
            for key, value in input_data.items():
                globals()[key] = value
    except:
        pass
    
    # First, execute user code to create class definitions
    # Use a separate buffer to avoid capturing example code output
    class_def_buffer = io.StringIO()
    with redirect_stdout(class_def_buffer):
        # Execute user code here
USER_CODE_PLACEHOLDER
    
    # Check if we have operations to execute
    if "operations" in input_data and isinstance(input_data["operations"], list):
        # Execute operations and capture only the last traverse() output
        last_output = execute_operations(input_data["operations"])
        
        if last_output:
            # Use the last output from traverse()
            result = {"output": last_output}
        else:
            # If no traverse() was called, return empty output
            result = {"output": ""}
    else:
        # No operations, use output from user code as normal
        raw_output = class_def_buffer.getvalue()
        result = parse_and_wrap_output(raw_output)
    
    # Output as JSON
    print(json.dumps(result, ensure_ascii=False, separators=(',', ':')))

except Exception as e:
    # Handle errors gracefully
    error_info = {"error": str(e)}
    print(json.dumps(error_info, ensure_ascii=False))
`

	// Indent user code properly
	indentedCode := "        " + strings.ReplaceAll(userCode, "\n", "\n        ")
	finalCode := strings.ReplaceAll(wrapper, "USER_CODE_PLACEHOLDER", indentedCode)

	return finalCode
}

// RunPython runs user code (Python) with JSON input on STDIN.
func (e *DockerExecutor) RunPython(code string, stdinJSON string) (*ExecResult, error) {
	// Wrap user code with JSON output wrapper
	wrappedCode := e.wrapUserCode(code)

	// For Docker-in-Docker, we'll pass the code directly via stdin instead of mounting files
	// This avoids volume mounting issues in DinD environments
	args := []string{
		"run", "--rm", "-i",
		// Disable network access inside the container
		"--network", "none",
		// Enforce read-only root filesystem
		"--read-only",
		// Run as root to avoid permission issues
		"--user", "0:0",
		// Drop all Linux capabilities and prevent gaining new privileges
		"--cap-drop", "ALL",
		"--security-opt", "no-new-privileges",
		// Limit processes and file descriptors
		"--pids-limit", "128",
		"--ulimit", "nofile=256:256",
		// Provide a writable tmpfs for temporary files
		"--tmpfs", "/tmp:rw,nosuid,nodev,noexec,size=64m",
		// Resource limits
		"-m", e.cfg.Memory,
		"--cpus", e.cfg.CPUs,
		// Set working directory
		"-w", "/tmp",
		e.cfg.Image,
		"python", "-c", wrappedCode,
	}

	ctx, cancel := context.WithTimeout(context.Background(), e.cfg.Timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "docker", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Stdin = bytes.NewBufferString(stdinJSON)

	err := cmd.Run()

	result := &ExecResult{
		Stdout: strings.TrimSpace(stdout.String()), // ลบ whitespace ส่วนเกิน
		Stderr: stderr.String(),
	}
	if ctx.Err() == context.DeadlineExceeded {
		result.TimedOut = true
		return result, nil
	}

	if ee, ok := err.(*exec.ExitError); ok {
		result.ExitCode = ee.ExitCode()
		return result, nil
	}
	if err != nil {
		return nil, fmt.Errorf("docker run error: %w", err)
	}
	return result, nil
}
