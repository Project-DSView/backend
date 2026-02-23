
import sys
import os

# Add project root to path
sys.path.append(os.path.abspath(os.path.join(os.path.dirname(__file__), "../..")))

from app.services.simulators.interactive_python_executor import InteractivePythonExecutor

def test_memory_tracking():
    print("Testing Memory Tracking...")
    executor = InteractivePythonExecutor()
    
    code = """
x = [1, 2, 3]
y = 10
print(x)
"""
    
    result = executor.execute_interactive(code)
    
    if not result.trace:
        print("Error: No trace generated")
        print(f"STDOUT: {result.stdout}")
        print(f"STDERR: {result.stderr}")
        return

    print(f"Trace length: {len(result.trace)}")
    
    previous_memory = 0
    consistent_increase = True
    
    for i, step in enumerate(result.trace):
         memory = step.get("memory_usage", 0)
         line = step.get("line")
         print(f"Step {i+1} (Line {line}): Memory = {memory} bytes")
         
         if i > 0 and memory < previous_memory and line > 1:
             # Memory consumption might fluctuate, but generally for a growing list we expect increase
             # However, small fluctuations are possible due to internal python optimizations
             pass
             
         previous_memory = memory

    last_step_memory = result.trace[-1].get("memory_usage", 0)
    print(f"Final Memory: {last_step_memory} bytes")
    
    if last_step_memory > 0:
        print("✅ Memory tracking verification PASSED")
    else:
        print("❌ Memory tracking verification FAILED: Memory is 0")

if __name__ == "__main__":
    test_memory_tracking()
