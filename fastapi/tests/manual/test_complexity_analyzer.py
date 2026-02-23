"""
Manual test for Big O complexity analyzer
"""
import sys
sys.path.insert(0, 'c:/Desktop/Project_KMITL/Website/backend/fastapi')

from app.services.simulators.operations.complexity_analyzer import ComplexityAnalyzer


def test_single_loop():
    """Test O(n) detection for single loop"""
    code = """
for i in range(n):
    print(i)
"""
    analyzer = ComplexityAnalyzer()
    result = analyzer.analyze(code)
    print("Single Loop Test:")
    print(f"  Time: {result.time_complexity}")
    print(f"  Space: {result.space_complexity}")
    print(f"  Explanation: {result.time_explanation}")
    assert result.time_complexity == "O(n)"
    assert result.space_complexity == "O(1)"


def test_nested_loops():
    """Test O(n²) detection for nested loops"""
    code = """
for i in range(n):
    for j in range(n):
        print(i, j)
"""
    analyzer = ComplexityAnalyzer()
    result = analyzer.analyze(code)
    print("\nNested Loop Test:")
    print(f"  Time: {result.time_complexity}")
    print(f"  Space: {result.space_complexity}")
    print(f"  Explanation: {result.time_explanation}")
    assert result.time_complexity == "O(n²)"
    assert result.space_complexity == "O(1)"


def test_constant_loop():
    """Test loop with constant range"""
    code = """
for i in range(n):
    for j in range(100):
        print(i, j)
"""
    analyzer = ComplexityAnalyzer()
    result = analyzer.analyze(code)
    print("\nConstant Loop Test:")
    print(f"  Time: {result.time_complexity}")
    print(f"  Space: {result.space_complexity}")
    print(f"  Explanation: {result.time_explanation}")
    assert result.time_complexity == "O(n)"  # 100 is constant


def test_recursion():
    """Test recursion detection"""
    code = """
def f(n):
    if n == 0:
        return
    f(n-1)
"""
    analyzer = ComplexityAnalyzer()
    result = analyzer.analyze(code)
    print("\nRecursion Test:")
    print(f"  Time: {result.time_complexity}")
    print(f"  Space: {result.space_complexity}")
    print(f"  Has recursion: {result.analysis_details['has_recursion']}")
    print(f"  Recursion type: {result.analysis_details['recursion_type']}")
    assert result.time_complexity == "O(n)"
    assert result.space_complexity == "O(n)"


def test_space_complexity():
    """Test space complexity with growing list"""
    code = """
arr = []
for i in range(n):
    arr.append(i)
"""
    analyzer = ComplexityAnalyzer()
    result = analyzer.analyze(code)
    print("\nSpace Complexity Test:")
    print(f"  Time: {result.time_complexity}")
    print(f"  Space: {result.space_complexity}")
    print(f"  Has growing structures: {result.analysis_details['has_growing_structures']}")
    assert result.time_complexity == "O(n)"
    assert result.space_complexity == "O(n)"


def test_no_loops():
    """Test O(1) for no loops"""
    code = """
x = 1
y = 2
print(x + y)
"""
    analyzer = ComplexityAnalyzer()
    result = analyzer.analyze(code)
    print("\nNo Loops Test:")
    print(f"  Time: {result.time_complexity}")
    print(f"  Space: {result.space_complexity}")
    assert result.time_complexity == "O(1)"
    assert result.space_complexity == "O(1)"

def test_per_function_analysis():
    """Test per-function Big O analysis"""
    code = """
def simple():
    x = 1
    return x

def linear(n):
    for i in range(n):
        print(i)

def quadratic(n):
    for i in range(n):
        for j in range(n):
            print(i, j)

def with_constant_inner(n):
    for i in range(n):
        for j in range(100):  # constant loop
            print(i, j)
"""
    analyzer = ComplexityAnalyzer()
    result = analyzer.analyze(code)
    
    print("\n" + "=" * 50)
    print("Per-Function Big O Analysis Test")
    print("=" * 50)
    
    print(f"\nOverall: Time={result.time_complexity}, Space={result.space_complexity}")
    
    print("\n📊 Per-Function Breakdown:")
    print("-" * 50)
    
    # Sort by rank (highest complexity first)
    sorted_funcs = sorted(result.function_complexities, 
                         key=lambda f: f.time_complexity_rank, 
                         reverse=True)
    
    for func in sorted_funcs:
        print(f"  {func.function_name}()")
        print(f"    Time: {func.time_complexity} (rank: {func.time_complexity_rank})")
        print(f"    Space: {func.space_complexity}")
        print(f"    Lines: {func.line_start}-{func.line_end}")
        print()
    
    # Verify
    func_dict = {f.function_name: f for f in result.function_complexities}
    assert func_dict["simple"].time_complexity == "O(1)", f"simple should be O(1), got {func_dict['simple'].time_complexity}"
    assert func_dict["linear"].time_complexity == "O(n)", f"linear should be O(n), got {func_dict['linear'].time_complexity}"
    assert func_dict["quadratic"].time_complexity == "O(n²)", f"quadratic should be O(n²), got {func_dict['quadratic'].time_complexity}"
    
    print("✓ Per-function analysis working correctly!")


if __name__ == "__main__":
    print("=" * 50)
    print("Big O Complexity Analyzer Tests")
    print("=" * 50)
    
    test_single_loop()
    test_nested_loops()
    test_constant_loop()
    test_recursion()
    test_space_complexity()
    test_no_loops()
    test_per_function_analysis()
    
    print("\n" + "=" * 50)
    print("All tests passed! ✓")
    print("=" * 50)

