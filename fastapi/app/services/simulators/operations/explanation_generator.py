"""
Explanation Generator - Generates Thai explanations for code execution steps
ช่วยอธิบายการทำงานของโค้ดทีละบรรทัดให้ผู้เรียนเข้าใจ
"""
import re
from typing import Dict, Any, Optional, List


class ExplanationGenerator:
    """
    Generate Thai explanations for code execution steps.
    ช่วยให้ผู้เรียนเข้าใจว่าแต่ละบรรทัดทำอะไร และทำไม visualization ถึงเปลี่ยน
    """
    
    # Pattern matching for common code patterns
    PATTERNS = {
        # Variable assignment
        r'^(\w+)\s*=\s*(.+)$': 'assignment',
        # Method call with assignment
        r'^(\w+)\s*=\s*(\w+)\.(\w+)\((.*)\)$': 'method_call_assign',
        # Method call without assignment
        r'^(\w+)\.(\w+)\((.*)\)$': 'method_call',
        # Function call with assignment
        r'^(\w+)\s*=\s*(\w+)\((.*)\)$': 'function_call_assign',
        # Function call
        r'^(\w+)\((.*)\)$': 'function_call',
        # Class instantiation
        r'^(\w+)\s*=\s*(\w+)\(\)$': 'instantiation_empty',
        r'^(\w+)\s*=\s*(\w+)\((.+)\)$': 'instantiation_with_args',
        # Attribute assignment
        r'^(\w+)\.(\w+)\s*=\s*(.+)$': 'attribute_assign',
        # Print statement
        r'^print\((.*)\)$': 'print',
        # Return statement
        r'^return\s+(.+)$': 'return',
        # If statement
        r'^if\s+(.+):$': 'if_statement',
        # While loop
        r'^while\s+(.+):$': 'while_loop',
        # For loop
        r'^for\s+(\w+)\s+in\s+(.+):$': 'for_loop',
        # Class definition
        r'^class\s+(\w+).*:$': 'class_def',
        # Function definition
        r'^def\s+(\w+)\((.*)\):$': 'function_def',
    }
    
    # Thai explanations for data structure operations
    DS_OPERATIONS = {
        'stack': {
            'push': 'เพิ่มข้อมูลเข้าไปที่ด้านบนของ Stack (LIFO - Last In First Out)',
            'pop': 'นำข้อมูลออกจากด้านบนของ Stack ตามหลัก LIFO',
            'peek': 'ดูข้อมูลที่อยู่บนสุดโดยไม่นำออก',
            'is_empty': 'ตรวจสอบว่า Stack ว่างหรือไม่',
            'size': 'นับจำนวนสมาชิกใน Stack',
        },
        'queue': {
            'enqueue': 'เพิ่มข้อมูลเข้าไปที่ท้ายแถว (FIFO - First In First Out)',
            'dequeue': 'นำข้อมูลออกจากหัวแถวตามหลัก FIFO',
            'front': 'ดูข้อมูลที่หัวแถวโดยไม่นำออก',
            'rear': 'ดูข้อมูลที่ท้ายแถว',
            'is_empty': 'ตรวจสอบว่า Queue ว่างหรือไม่',
        },
        'linkedlist': {
            'insert': 'แทรก node ใหม่เข้าไปใน Linked List',
            'insertfront': 'แทรก node ใหม่ที่ตำแหน่งหัว (head)',
            'insertend': 'แทรก node ใหม่ที่ตำแหน่งท้าย',
            'delete': 'ลบ node ออกจาก Linked List',
            'traverse': 'วนดูข้อมูลทุก node ตั้งแต่ head ไปจนถึง tail',
            'search': 'ค้นหา node ที่มีค่าตรงกับที่ต้องการ',
        },
        'binarysearchtree': {
            'insert': 'เพิ่ม node ใหม่ โดยค่าน้อยไปซ้าย ค่ามากไปขวา',
            'search': 'ค้นหาค่าโดยเปรียบเทียบและเลือกไปซ้ายหรือขวา',
            'delete': 'ลบ node และจัดโครงสร้างใหม่',
            'inorder': 'ท่องต้นไม้แบบ In-order (Left, Root, Right)',
            'preorder': 'ท่องต้นไม้แบบ Pre-order (Root, Left, Right)',
            'postorder': 'ท่องต้นไม้แบบ Post-order (Left, Right, Root)',
        },
        'graph': {
            'addvertex': 'เพิ่ม vertex (จุด) ใหม่ใน Graph',
            'addedge': 'เพิ่ม edge (เส้นเชื่อม) ระหว่าง 2 vertices',
            'bfs': 'ท่อง Graph แบบกว้างก่อน (Breadth-First Search)',
            'dfs': 'ท่อง Graph แบบลึกก่อน (Depth-First Search)',
            'removeedge': 'ลบ edge ออกจาก Graph',
            'removevertex': 'ลบ vertex และ edge ที่เกี่ยวข้องทั้งหมด',
        },
    }
    
    # Visual change explanations
    VISUAL_CHANGES = {
        'node_added': 'มี node ใหม่ปรากฏใน visualization เพราะกำลังสร้าง node ใหม่',
        'node_removed': 'node หายไปจาก visualization เพราะถูกลบออก',
        'pointer_changed': 'ลูกศร (pointer) เปลี่ยน เพราะกำลังเปลี่ยนการเชื่อมต่อ',
        'highlight_current': 'node ถูก highlight เพราะกำลังถูก access หรือ traverse',
        'stack_grow': 'Stack สูงขึ้น เพราะมี element ใหม่ถูก push เข้าไป',
        'stack_shrink': 'Stack เตี้ยลง เพราะ element ถูก pop ออกไป',
        'queue_grow': 'Queue ยาวขึ้น เพราะมี element ใหม่เข้าแถว',
        'queue_shrink': 'Queue สั้นลง เพราะ element ออกจากหัวแถว',
    }

    def __init__(self, data_structure_type: Optional[str] = None):
        """
        Initialize the explanation generator.
        
        Args:
            data_structure_type: Type of data structure (stack, queue, linkedlist, etc.)
        """
        self.data_structure_type = data_structure_type or 'general'
    
    def generate_explanation(
        self,
        code_line: str,
        operation: Optional[str] = None,
        variables: Optional[Dict[str, Any]] = None,
        prev_state: Optional[Dict[str, Any]] = None,
        curr_state: Optional[Dict[str, Any]] = None,
    ) -> Dict[str, str]:
        """
        Generate Thai explanation for a code line.
        
        Args:
            code_line: The line of code being executed
            operation: The type of operation (insert, delete, etc.)
            variables: Current variable values
            prev_state: Previous execution state
            curr_state: Current execution state
            
        Returns:
            Dictionary with:
                - what: อธิบายว่าบรรทัดนี้ทำอะไร
                - why_visual: อธิบายว่าทำไม visualization ถึงเปลี่ยน
                - concept: อธิบาย concept ที่เกี่ยวข้อง
        """
        code_line = code_line.strip()
        
        # Detect code pattern
        pattern_type, match_groups = self._detect_pattern(code_line)
        
        # Generate "what" explanation
        what = self._generate_what_explanation(code_line, pattern_type, match_groups, variables)
        
        # Generate "why visual changed" explanation
        why_visual = self._generate_visual_explanation(
            operation, pattern_type, match_groups, prev_state, curr_state
        )
        
        # Generate "concept" explanation
        concept = self._generate_concept_explanation(operation, pattern_type, match_groups)
        
        return {
            'what': what,
            'why_visual': why_visual,
            'concept': concept,
        }
    
    def _detect_pattern(self, code_line: str) -> tuple:
        """Detect the pattern type of the code line."""
        for pattern, pattern_type in self.PATTERNS.items():
            match = re.match(pattern, code_line)
            if match:
                return pattern_type, match.groups()
        return 'unknown', ()
    
    def _generate_what_explanation(
        self,
        code_line: str,
        pattern_type: str,
        match_groups: tuple,
        variables: Optional[Dict[str, Any]] = None,
    ) -> str:
        """Generate Thai explanation for what the code line does."""
        
        if pattern_type == 'assignment':
            var_name, value = match_groups
            return f"กำหนดค่า {value} ให้กับตัวแปร '{var_name}'"
        
        elif pattern_type == 'instantiation_empty':
            var_name, class_name = match_groups
            return f"สร้าง object ใหม่จาก class '{class_name}' และเก็บไว้ในตัวแปร '{var_name}'"
        
        elif pattern_type == 'instantiation_with_args':
            var_name, class_name, args = match_groups
            return f"สร้าง object ใหม่จาก class '{class_name}' ด้วยค่า {args} และเก็บไว้ในตัวแปร '{var_name}'"
        
        elif pattern_type == 'method_call':
            obj_name, method_name, args = match_groups
            return self._explain_method_call(obj_name, method_name, args)
        
        elif pattern_type == 'method_call_assign':
            var_name, obj_name, method_name, args = match_groups
            method_explain = self._explain_method_call(obj_name, method_name, args)
            return f"{method_explain} และเก็บผลลัพธ์ไว้ที่ '{var_name}'"
        
        elif pattern_type == 'attribute_assign':
            obj_name, attr_name, value = match_groups
            return f"กำหนดค่า {value} ให้กับ attribute '{attr_name}' ของ '{obj_name}'"
        
        elif pattern_type == 'print':
            content = match_groups[0] if match_groups else ''
            return f"แสดงผลลัพธ์: {content}"
        
        elif pattern_type == 'return':
            value = match_groups[0] if match_groups else ''
            return f"ส่งค่า {value} กลับออกจากฟังก์ชัน"
        
        elif pattern_type == 'if_statement':
            condition = match_groups[0] if match_groups else ''
            return f"ตรวจสอบเงื่อนไข: {condition}"
        
        elif pattern_type == 'while_loop':
            condition = match_groups[0] if match_groups else ''
            return f"วนซ้ำตราบเท่าที่: {condition}"
        
        elif pattern_type == 'for_loop':
            var_name, iterable = match_groups
            return f"วนซ้ำโดยให้ '{var_name}' รับค่าจาก {iterable} ทีละตัว"
        
        elif pattern_type == 'class_def':
            class_name = match_groups[0] if match_groups else ''
            return f"ประกาศ class ชื่อ '{class_name}'"
        
        elif pattern_type == 'function_def':
            func_name, params = match_groups
            if params:
                return f"ประกาศฟังก์ชัน '{func_name}' ที่รับ parameter: {params}"
            return f"ประกาศฟังก์ชัน '{func_name}'"
        
        else:
            return f"รันบรรทัด: {code_line}"
    
    def _explain_method_call(self, obj_name: str, method_name: str, args: str) -> str:
        """Generate explanation for method calls based on data structure type."""
        method_lower = method_name.lower()
        
        # Check if we have a specific explanation for this method
        ds_type = self._normalize_ds_type(self.data_structure_type)
        
        if ds_type in self.DS_OPERATIONS:
            for method_key, explanation in self.DS_OPERATIONS[ds_type].items():
                if method_key in method_lower:
                    if args:
                        return f"เรียก {method_name}({args}): {explanation}"
                    return f"เรียก {method_name}(): {explanation}"
        
        # Default explanation
        if args:
            return f"เรียก method '{method_name}' ของ '{obj_name}' ด้วยค่า {args}"
        return f"เรียก method '{method_name}' ของ '{obj_name}'"
    
    def _normalize_ds_type(self, ds_type: str) -> str:
        """Normalize data structure type name."""
        if not ds_type:
            return 'general'
        
        ds_type = ds_type.lower()
        
        if 'stack' in ds_type:
            return 'stack'
        elif 'queue' in ds_type:
            return 'queue'
        elif 'linked' in ds_type or 'list' in ds_type:
            return 'linkedlist'
        elif 'bst' in ds_type or 'tree' in ds_type or 'binary' in ds_type:
            return 'binarysearchtree'
        elif 'graph' in ds_type:
            return 'graph'
        
        return 'general'
    
    def _generate_visual_explanation(
        self,
        operation: Optional[str],
        pattern_type: str,
        match_groups: tuple,
        prev_state: Optional[Dict[str, Any]] = None,
        curr_state: Optional[Dict[str, Any]] = None,
    ) -> str:
        """Generate explanation for why visualization changed."""
        
        ds_type = self._normalize_ds_type(self.data_structure_type)
        
        # Detect visual changes based on states
        if prev_state and curr_state:
            visual_change = self._detect_visual_change(prev_state, curr_state)
            if visual_change:
                return visual_change
        
        # Generate based on operation type
        if operation:
            op_lower = operation.lower()
            
            if 'insert' in op_lower or 'push' in op_lower or 'add' in op_lower or 'enqueue' in op_lower:
                if ds_type == 'stack':
                    return self.VISUAL_CHANGES['stack_grow']
                elif ds_type == 'queue':
                    return self.VISUAL_CHANGES['queue_grow']
                else:
                    return self.VISUAL_CHANGES['node_added']
            
            elif 'delete' in op_lower or 'pop' in op_lower or 'remove' in op_lower or 'dequeue' in op_lower:
                if ds_type == 'stack':
                    return self.VISUAL_CHANGES['stack_shrink']
                elif ds_type == 'queue':
                    return self.VISUAL_CHANGES['queue_shrink']
                else:
                    return self.VISUAL_CHANGES['node_removed']
            
            elif 'traverse' in op_lower or 'search' in op_lower:
                return self.VISUAL_CHANGES['highlight_current']
        
        # Check for pointer assignments
        if pattern_type == 'attribute_assign' and len(match_groups) >= 2:
            attr_name = match_groups[1].lower()
            if attr_name in ['next', 'prev', 'left', 'right', 'head', 'tail', 'root']:
                return self.VISUAL_CHANGES['pointer_changed']
        
        return ""
    
    def _detect_visual_change(
        self,
        prev_state: Dict[str, Any],
        curr_state: Dict[str, Any],
    ) -> str:
        """Detect what changed between two states."""
        
        prev_instances = prev_state.get('instances', {})
        curr_instances = curr_state.get('instances', {})
        
        # Check for new instances
        for key in curr_instances:
            if key not in prev_instances:
                return self.VISUAL_CHANGES['node_added']
        
        # Check for removed instances
        for key in prev_instances:
            if key not in curr_instances:
                return self.VISUAL_CHANGES['node_removed']
        
        # Check for size changes in data structures
        for key in curr_instances:
            if key in prev_instances:
                curr_size = curr_instances[key].get('size', 0)
                prev_size = prev_instances[key].get('size', 0)
                
                if curr_size > prev_size:
                    ds_type = self._normalize_ds_type(self.data_structure_type)
                    if ds_type == 'stack':
                        return self.VISUAL_CHANGES['stack_grow']
                    elif ds_type == 'queue':
                        return self.VISUAL_CHANGES['queue_grow']
                    return self.VISUAL_CHANGES['node_added']
                
                elif curr_size < prev_size:
                    ds_type = self._normalize_ds_type(self.data_structure_type)
                    if ds_type == 'stack':
                        return self.VISUAL_CHANGES['stack_shrink']
                    elif ds_type == 'queue':
                        return self.VISUAL_CHANGES['queue_shrink']
                    return self.VISUAL_CHANGES['node_removed']
        
        return ""
    
    def _generate_concept_explanation(
        self,
        operation: Optional[str],
        pattern_type: str,
        match_groups: tuple,
    ) -> str:
        """Generate concept explanation related to the operation."""
        
        ds_type = self._normalize_ds_type(self.data_structure_type)
        
        if ds_type == 'stack':
            return "Stack ทำงานแบบ LIFO (Last In, First Out) - ข้อมูลที่ใส่เข้าไปทีหลังจะถูกนำออกก่อน"
        
        elif ds_type == 'queue':
            return "Queue ทำงานแบบ FIFO (First In, First Out) - ข้อมูลที่ใส่เข้าไปก่อนจะถูกนำออกก่อน"
        
        elif ds_type == 'linkedlist':
            if pattern_type == 'attribute_assign' and len(match_groups) >= 2:
                attr = match_groups[1].lower()
                if attr == 'next':
                    return "Pointer 'next' ใช้เชื่อมต่อ node ปัจจุบันไปยัง node ถัดไป"
                elif attr == 'prev':
                    return "Pointer 'prev' ใช้เชื่อมกลับไปยัง node ก่อนหน้า (Doubly Linked List)"
            return "Linked List เก็บข้อมูลเป็น node ที่เชื่อมต่อกันด้วย pointer"
        
        elif ds_type == 'binarysearchtree':
            return "BST: node ที่มีค่าน้อยกว่าอยู่ทางซ้าย, node ที่มีค่ามากกว่าอยู่ทางขวา"
        
        elif ds_type == 'graph':
            return "Graph ประกอบด้วย vertices (จุด) และ edges (เส้นเชื่อม) ที่เชื่อมต่อกัน"
        
        return ""


def generate_step_explanation(
    code_line: str,
    data_structure_type: str,
    operation: Optional[str] = None,
    variables: Optional[Dict[str, Any]] = None,
    prev_state: Optional[Dict[str, Any]] = None,
    curr_state: Optional[Dict[str, Any]] = None,
) -> Dict[str, str]:
    """
    Convenience function to generate explanation for a single step.
    
    Returns:
        Dictionary with 'what', 'why_visual', 'concept' keys
    """
    generator = ExplanationGenerator(data_structure_type)
    return generator.generate_explanation(
        code_line=code_line,
        operation=operation,
        variables=variables,
        prev_state=prev_state,
        curr_state=curr_state,
    )
