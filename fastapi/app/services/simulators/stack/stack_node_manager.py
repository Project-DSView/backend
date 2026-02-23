from typing import Dict, Any
from app.services.simulators.operations.node_manager import NodeManager
from app.utils.messages_th import get_stack_message, get_message


class StackNodeManager(NodeManager):
    """Enhanced Stack-specific node manager for ArrayStack operations"""
    
    def create_instance_data(self, class_name: str) -> Dict[str, Any]:
        """Create new instance data structure for ArrayStack"""
        if class_name == "ArrayStack":
            return {
                "data": [],
                "class_type": class_name,
                "attributes": {},
                "history": []  # Track operation history
            }
        return super().create_instance_data(class_name)
    
    def create_instance(self, var_name: str, class_name: str) -> str:
        """Create a new instance of a class"""
        if class_name == "ArrayStack":
            self.context["instances"][var_name] = self.create_instance_data(class_name)
            self.context["active_instance"] = var_name
            return get_message("instance_created_generic", class_name=class_name, var_name=var_name)
        else:
            return super().create_instance(var_name, class_name)
    
    def stack_push(self, instance: Dict[str, Any], value: Any, instance_name: str = "") -> Dict[str, Any]:
        """Push value to stack with detailed tracking"""
        # Check for overflow if max_size is defined
        max_size = instance.get("attributes", {}).get("max_size")
        if max_size is not None and len(instance["data"]) >= max_size:
            return {
                "message": get_stack_message("push", str(value)) + " - Stack is full (overflow)",
                "operation": "push",
                "value": value,
                "error": "overflow",
                "instance_name": instance_name,
                "explanation": f"Cannot push value {value} onto the stack: Stack is full (overflow)"
            }
        
        old_data = instance["data"].copy()
        instance["data"].append(value)
        new_data = instance["data"].copy()
        
        # Record operation in history
        instance["history"].append({
            "operation": "push",
            "value": value,
            "before": old_data,
            "after": new_data
        })
        
        return {
            "message": f"เพิ่มข้อมูล '{value}' ลงใน stack (LIFO - Last In First Out)",
            "operation": "push",
            "value": value,
            "before_data": old_data,
            "after_data": new_data,
            "instance_name": instance_name,
            "explanation": f"Push value {value} onto the top of the stack"
        }
    
    def stack_pop(self, instance: Dict[str, Any], instance_name: str = "") -> Dict[str, Any]:
        """Pop value from stack with detailed tracking"""
        if not instance["data"]:
            return {
                "message": "ลบข้อมูลออกจาก stack ที่ตำแหน่งบนสุด" + " - " + get_message("traverse_empty"),
                "operation": "pop",
                "value": None,
                "error": "underflow",
                "instance_name": instance_name,
                "explanation": "Cannot pop from an empty stack"
            }
        
        old_data = instance["data"].copy()
        value = instance["data"].pop()
        new_data = instance["data"].copy()
        
        # Record operation in history
        instance["history"].append({
            "operation": "pop",
            "value": value,
            "before": old_data,
            "after": new_data
        })
        
        return {
            "message": f"ลบข้อมูลออกจาก stack ที่ตำแหน่งบนสุด → คืนค่า {value}",
            "operation": "pop",
            "value": value,
            "before_data": old_data,
            "after_data": new_data,
            "instance_name": instance_name,
            "explanation": f"Pop value {value} from the top of the stack"
        }
    
    def stack_top(self, instance: Dict[str, Any], instance_name: str = "") -> Dict[str, Any]:
        """Get top value of stack"""
        if not instance["data"]:
            return {
                "message": "ดูข้อมูลที่ตำแหน่งบนสุดของ stack โดยไม่ลบออก" + " - " + get_message("traverse_empty"),
                "operation": "stackTop",
                "value": None,
                "error": "empty_stack",
                "instance_name": instance_name,
                "explanation": "Cannot peek at an empty stack"
            }
        
        top_value = instance["data"][-1]
        return {
            "message": f"ดูข้อมูลที่ตำแหน่งบนสุดของ stack โดยไม่ลบออก → คืนค่า {top_value}",
            "operation": "stackTop",
            "value": top_value,
            "instance_name": instance_name,
            "explanation": f"Peek at value {top_value} from the top of the stack"
        }
    
    def stack_size(self, instance: Dict[str, Any], instance_name: str = "") -> Dict[str, Any]:
        """Get stack size"""
        size = len(instance["data"])
        return {
            "message": f"นับจำนวนข้อมูลใน stack → คืนค่า {size}",
            "operation": "size",
            "value": size,
            "instance_name": instance_name,
            "explanation": f"Get stack size (Size: {size})"
        }
    
    def stack_is_empty(self, instance: Dict[str, Any], instance_name: str = "") -> Dict[str, Any]:
        """Check if stack is empty"""
        is_empty = len(instance["data"]) == 0
        return {
            "message": f"ตรวจสอบว่า stack ว่างเปล่าหรือไม่ → คืนค่า {is_empty}",
            "operation": "is_empty",
            "value": is_empty,
            "instance_name": instance_name,
            "explanation": f"Check if stack is empty (Result: {is_empty})"
        }
    
    def stack_print(self, instance: Dict[str, Any], instance_name: str = "") -> Dict[str, Any]:
        """Print stack contents"""
        stack_data = instance["data"].copy()
        print_output = str(stack_data)
        self.context["stdout"].append(print_output)
        return {
            "message": f"แสดงข้อมูล {instance_name}: {stack_data}",
            "operation": "printStack",
            "value": stack_data,
            "instance_name": instance_name,
            "stdout": print_output,  # Add print output to result
            "explanation": f"Print stack contents: {stack_data}"
        }