"""
Big O Complexity Analyzer with Per-Function Analysis

Analyzes Python code to determine:
- Overall time complexity (O(1), O(n), O(n²), etc.)
- Overall space complexity
- Per-function complexity breakdown
"""

import ast
from dataclasses import dataclass, field
from typing import List, Dict, Any, Optional


@dataclass
class FunctionComplexity:
    """Complexity analysis for a single function"""
    function_name: str
    time_complexity: str
    space_complexity: str
    line_start: int
    line_end: int
    time_complexity_rank: int  # 1=O(1), 2=O(log n), 3=O(n), 4=O(n log n), 5=O(n²), 6=O(2^n)
    loop_count: int = 0
    max_nesting: int = 0
    has_recursion: bool = False


@dataclass
class ComplexityResult:
    """Result of complexity analysis"""
    time_complexity: str
    space_complexity: str
    time_explanation: str
    space_explanation: str
    analysis_details: Dict[str, Any] = field(default_factory=dict)
    function_complexities: List[FunctionComplexity] = field(default_factory=list)


class ComplexityAnalyzer:
    """Analyzes code complexity including per-function breakdown"""
    
    # Complexity ranking for comparison
    COMPLEXITY_RANK = {
        "O(1)": 1,
        "O(log n)": 2,
        "O(n)": 3,
        "O(n log n)": 4,
        "O(n²)": 5,
        "O(n³)": 6,
        "O(2^n)": 7,
        "O(n!)": 8,
    }
    
    def __init__(self):
        self.loop_count = 0
        self.max_nesting = 0
        self.current_nesting = 0
        self.has_recursion = False
        self.recursion_type = None
        self.function_names = set()
        self.function_calls = set()
        self.space_allocations = []
        self.has_growing_structures = False
    
    def analyze(self, code: str) -> ComplexityResult:
        """Analyze code and return complexity result with per-function breakdown"""
        self._reset()
        
        try:
            tree = ast.parse(code)
        except SyntaxError:
            return ComplexityResult(
                time_complexity="Unknown",
                space_complexity="Unknown",
                time_explanation="ไม่สามารถวิเคราะห์โค้ดได้ (syntax error)",
                space_explanation="ไม่สามารถวิเคราะห์โค้ดได้ (syntax error)",
                analysis_details={},
                function_complexities=[]
            )
        
        # Collect function names first (for recursion detection)
        for node in ast.walk(tree):
            if isinstance(node, (ast.FunctionDef, ast.AsyncFunctionDef)):
                self.function_names.add(node.name)
        
        # Analyze per-function complexity
        function_complexities = []
        for node in ast.walk(tree):
            if isinstance(node, (ast.FunctionDef, ast.AsyncFunctionDef)):
                if node.name == '__init__':
                    continue
                func_complexity = self._analyze_function(node)
                function_complexities.append(func_complexity)
        
        # Analyze overall complexity
        self._analyze_node(tree)
        
        # Determine overall time complexity
        time_complexity = self._determine_time_complexity()
        space_complexity = self._determine_space_complexity()
        
        return ComplexityResult(
            time_complexity=time_complexity,
            space_complexity=space_complexity,
            time_explanation=self._get_time_explanation(time_complexity),
            space_explanation=self._get_space_explanation(space_complexity),
            analysis_details={
                "loop_count": self.loop_count,
                "max_nesting": self.max_nesting,
                "has_recursion": self.has_recursion,
                "recursion_type": self.recursion_type,
                "space_allocations": self.space_allocations[:5],  # Limit to first 5
                "has_growing_structures": self.has_growing_structures,
            },
            function_complexities=function_complexities
        )
    
    def _reset(self):
        """Reset analyzer state"""
        self.loop_count = 0
        self.max_nesting = 0
        self.current_nesting = 0
        self.has_recursion = False
        self.recursion_type = None
        self.function_names = set()
        self.function_calls = set()
        self.space_allocations = []
        self.has_growing_structures = False
    
    def _analyze_function(self, node: ast.FunctionDef) -> FunctionComplexity:
        """Analyze a single function's complexity"""
        func_name = node.name
        line_start = node.lineno
        line_end = node.end_lineno or node.lineno
        
        # Reset counters for this function
        loop_count = 0
        max_nesting = 0
        current_nesting = 0
        has_recursion = False
        has_growing = False
        
        def count_loops(n, depth=0):
            nonlocal loop_count, max_nesting, current_nesting, has_recursion, has_growing
            
            if isinstance(n, (ast.For, ast.While)):
                loop_count += 1
                current_nesting = depth + 1
                max_nesting = max(max_nesting, current_nesting)
                
                # Recurse into loop body (not using ast.walk to avoid infinite recursion)
                for child in ast.iter_child_nodes(n):
                    count_loops(child, depth + 1)
                return
            
            if isinstance(n, ast.Call):
                # Check for recursion
                if isinstance(n.func, ast.Name) and n.func.id == func_name:
                    has_recursion = True
                # Check for growing structures
                if isinstance(n.func, ast.Attribute):
                    if n.func.attr in ['append', 'extend', 'insert', 'add']:
                        has_growing = True
            
            # Recurse into children
            for child in ast.iter_child_nodes(n):
                count_loops(child, depth)
        
        count_loops(node)
        
        # Determine time complexity for this function
        if has_recursion:
            # Recursion with loops
            if max_nesting >= 2:
                time_complexity = "O(2^n)"
            elif max_nesting == 1:
                time_complexity = "O(n²)"  # Recursion + 1 loop = O(n) * O(n) = O(n²)
            else:
                time_complexity = "O(n)"  # Simple recursion
        elif max_nesting >= 3:
            time_complexity = "O(n³)"
        elif max_nesting == 2:
            time_complexity = "O(n²)"
        elif max_nesting == 1 or loop_count > 0:
            time_complexity = "O(n)"
        else:
            time_complexity = "O(1)"
        
        # Space complexity
        if has_recursion:
            space_complexity = "O(n)"
        elif has_growing:
            space_complexity = "O(n)"
        else:
            space_complexity = "O(1)"
        
        rank = self.COMPLEXITY_RANK.get(time_complexity, 3)
        
        return FunctionComplexity(
            function_name=func_name,
            time_complexity=time_complexity,
            space_complexity=space_complexity,
            line_start=line_start,
            line_end=line_end,
            time_complexity_rank=rank,
            loop_count=loop_count,
            max_nesting=max_nesting,
            has_recursion=has_recursion
        )
    
    def _is_constant_loop(self, node: ast.AST) -> bool:
        """Check if a loop has a constant range (not n-dependent)"""
        if isinstance(node, ast.For):
            if isinstance(node.iter, ast.Call):
                if isinstance(node.iter.func, ast.Name) and node.iter.func.id == 'range':
                    # Check all range arguments - if all are constants, it's constant
                    for arg in node.iter.args:
                        if not isinstance(arg, ast.Constant):
                            return False  # Has a variable, so n-dependent
                    return True  # All args are constants
        return False  # While loops or other patterns are treated as n-dependent
    
    def _analyze_node(self, node: ast.AST, nesting: int = 0):
        """Recursively analyze AST nodes for overall complexity"""
        if isinstance(node, (ast.FunctionDef, ast.AsyncFunctionDef)) and node.name == '__init__':
            return

        if isinstance(node, (ast.For, ast.While)):
            self.loop_count += 1
            
            # Only count n-dependent loops for nesting depth
            is_constant = self._is_constant_loop(node)
            new_nesting = nesting if is_constant else nesting + 1
            
            if not is_constant:
                self.current_nesting = new_nesting
                self.max_nesting = max(self.max_nesting, self.current_nesting)
            
            for child in ast.iter_child_nodes(node):
                self._analyze_node(child, new_nesting)
            return
        
        if isinstance(node, ast.Call):
            # Check for recursion
            if isinstance(node.func, ast.Name):
                func_name = node.func.id
                self.function_calls.add(func_name)
                if func_name in self.function_names:
                    self.has_recursion = True
                    self.recursion_type = "simple"
            
            # Check for growing structures (append, extend, etc.)
            if isinstance(node.func, ast.Attribute):
                if node.func.attr in ['append', 'extend', 'insert', 'add', 'update']:
                    self.has_growing_structures = True
                    self.space_allocations.append(node.func.attr)
        
        if isinstance(node, ast.ListComp):
            self.loop_count += 1
            self.max_nesting = max(self.max_nesting, 1)
            self.has_growing_structures = True
        
        if isinstance(node, (ast.List, ast.Dict, ast.Set)):
            self.space_allocations.append(type(node).__name__)
        
        # Recurse into children
        for child in ast.iter_child_nodes(node):
            self._analyze_node(child, nesting)
    
    def _determine_time_complexity(self) -> str:
        """Determine overall time complexity based on analysis
        
        For overall complexity, we take the worst-case from nested loops.
        We don't combine recursion with nesting since they may be in different functions.
        """
        # Simply use max nesting depth for overall complexity
        # This is more accurate than trying to combine recursion with nesting
        if self.max_nesting >= 3:
            return "O(n³)"
        elif self.max_nesting == 2:
            return "O(n²)"
        elif self.max_nesting == 1 or self.loop_count > 0 or self.has_recursion:
            return "O(n)"
        else:
            return "O(1)"
    
    def _determine_space_complexity(self) -> str:
        """Determine overall space complexity"""
        if self.has_recursion:
            return "O(n)"
        
        if self.has_growing_structures:
            return "O(n)"
        
        if len(self.space_allocations) > 10:
            return "O(n)"
        
        return "O(1)"
    
    def _get_time_explanation(self, complexity: str) -> str:
        """Get Thai explanation for time complexity"""
        explanations = {
            "O(1)": "ประสิทธิภาพดีมาก - ทำงานเร็วคงที่ไม่ว่าข้อมูลจะมากแค่ไหน",
            "O(log n)": "ประสิทธิภาพดี - เวลาเพิ่มช้ามากเมื่อข้อมูลเพิ่ม (เช่น binary search)",
            "O(n)": "ประสิทธิภาพปานกลาง - เวลาเพิ่มตามจำนวนข้อมูล",
            "O(n log n)": "ประสิทธิภาพปานกลาง - เช่น sorting algorithms ที่ดี",
            "O(n²)": "ประสิทธิภาพช้า - มี nested loop ทำให้ช้าลงมากเมื่อข้อมูลเพิ่ม",
            "O(n³)": "ประสิทธิภาพช้ามาก - มี 3 loops ซ้อนกัน",
            "O(2^n)": "ประสิทธิภาพแย่มาก - เวลาเพิ่มเป็นทวีคูณ (exponential)",
            "O(n!)": "ประสิทธิภาพแย่ที่สุด - ใช้ได้กับข้อมูลขนาดเล็กมากเท่านั้น",
        }
        return explanations.get(complexity, "ไม่สามารถระบุได้แน่ชัด")
    
    def _get_space_explanation(self, complexity: str) -> str:
        """Get Thai explanation for space complexity"""
        explanations = {
            "O(1)": "ใช้หน่วยความจำคงที่ - ไม่สร้าง structure เพิ่มตามขนาดข้อมูล",
            "O(log n)": "ใช้หน่วยความจำน้อย - เช่น recursive call stack ของ binary search",
            "O(n)": "ใช้หน่วยความจำตามขนาดข้อมูล - สร้าง list/array ใหม่",
            "O(n²)": "ใช้หน่วยความจำมาก - เช่น สร้าง 2D array",
        }
        return explanations.get(complexity, "ใช้หน่วยความจำตามการทำงานของโค้ด")
