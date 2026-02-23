"""Error handler for formatting and translating errors to Thai messages"""

import re
from typing import Dict, Any
from app.utils.messages_th import get_message


class ErrorHandler:
    """Handles error formatting and translation"""
    
    @staticmethod
    def format_error(error: Exception, line_number: int, code_line: str, offset: int = None, filename: str = None) -> Dict[str, Any]:
        """
        Format error with line number, type, and Thai message
        
        Args:
            error: The exception that occurred
            line_number: Line number where error occurred
            code_line: The code line that caused the error
            offset: Character offset where error occurred (for pointer)
            filename: Optional filename for error display
            
        Returns:
            Dictionary with error information including:
            - error_type: Type of error (e.g., 'SyntaxError', 'NameError')
            - error_message: Original error message
            - thai_message: Thai translated error message
            - python_style_message: Python-style error message with pointer
            - line_number: Line number where error occurred
            - code_line: Code line that caused the error
            - offset: Character offset where error occurred
        """
        error_type = type(error).__name__
        error_message = str(error)
        
        # Extract offset from error if available
        if offset is None and hasattr(error, 'offset'):
            offset = error.offset
        
        # Get Thai error message
        thai_message = ErrorHandler._get_thai_error_message(
            error_type, error_message, line_number, code_line
        )
        
        # Create Python-style error message
        python_style_message = ErrorHandler._create_python_style_error(
            error_type, error_message, line_number, code_line, offset, filename
        )
        
        return {
            "error_type": error_type,
            "error_message": error_message,
            "thai_message": thai_message,
            "python_style_message": python_style_message,
            "line_number": line_number,
            "code_line": code_line.strip() if code_line else "",
            "offset": offset
        }
    
    @staticmethod
    def _create_python_style_error(error_type: str, error_message: str, line_number: int, 
                                   code_line: str, offset: int = None, filename: str = None) -> str:
        """
        Create Python-style error message with file path, line number, code line, and pointer
        
        Args:
            error_type: Type of error (e.g., 'SyntaxError')
            error_message: Original error message
            line_number: Line number where error occurred
            code_line: The code line that caused the error
            offset: Character offset where error occurred
            filename: Optional filename for error display
            
        Returns:
            Python-style error message string
        """
        lines = []
        
        # File path line
        if filename:
            lines.append(f'  File "{filename}", line {line_number}')
        else:
            lines.append(f'  File "<string>", line {line_number}')
        
        # Code line
        if code_line:
            lines.append(f"    {code_line.rstrip()}")
        else:
            lines.append("    ")
        
        # Pointer line (^) pointing to error location
        if offset is not None and offset > 0:
            # Calculate spaces before pointer
            # offset is 1-based in Python, so we need to adjust
            spaces = " " * (offset - 1 + 4)  # +4 for "    " prefix
            lines.append(f"{spaces}^")
        elif code_line:
            # If no offset, point to end of line
            spaces = " " * (len(code_line.rstrip()) + 4)
            lines.append(f"{spaces}^")
        
        # Error type and message
        lines.append(f"{error_type}: {error_message}")
        
        return "\n".join(lines)
    
    @staticmethod
    def _get_thai_error_message(error_type: str, error_message: str, line: int, code: str) -> str:
        """
        Get Thai error message based on error type
        
        Args:
            error_type: Type of error (e.g., 'SyntaxError', 'NameError')
            error_message: Original error message
            line: Line number
            code: Code line that caused error
            
        Returns:
            Thai error message
        """
        # Extract variable/attribute names from error message
        name_match = re.search(r"name '(\w+)'", error_message)
        attr_match = re.search(r"'(\w+)'", error_message)
        
        if error_type == "SyntaxError":
            # Extract syntax error details
            message = error_message
            if "invalid syntax" in error_message.lower():
                message = "ไวยากรณ์ไม่ถูกต้อง"
            elif "unexpected" in error_message.lower():
                message = "พบตัวอักษรที่ไม่คาดคิด"
            elif "expected" in error_message.lower():
                message = "ขาดตัวอักษรที่คาดหวัง"
            
            return get_message("syntax_error", line=line, message=message)
        
        elif error_type == "NameError":
            name = name_match.group(1) if name_match else "ไม่ทราบ"
            return get_message("name_error", name=name, line=line)
        
        elif error_type == "AttributeError":
            attr = attr_match.group(1) if attr_match else "ไม่ทราบ"
            # Try to extract object name
            obj_match = re.search(r"'(\w+)' object", error_message)
            obj = obj_match.group(1) if obj_match else "object"
            return get_message("attribute_error", attr=attr, obj=obj, line=line)
        
        elif error_type == "TypeError":
            message = error_message
            if "unsupported operand" in error_message.lower():
                message = "ไม่สามารถดำเนินการกับประเภทข้อมูลนี้ได้"
            elif "takes" in error_message.lower() and "positional" in error_message.lower():
                message = "จำนวนพารามิเตอร์ไม่ถูกต้อง"
            
            return get_message("type_error", line=line, message=message)
        
        elif error_type == "ValueError":
            message = error_message
            if "invalid literal" in error_message.lower():
                message = "ค่าที่ระบุไม่ถูกต้อง"
            
            return get_message("value_error", line=line, message=message)
        
        elif error_type == "IndexError":
            return get_message("index_error", line=line)
        
        else:
            # Generic runtime error
            return get_message("runtime_error", line=line, message=error_message)



