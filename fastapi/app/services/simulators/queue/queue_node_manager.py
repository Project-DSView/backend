from typing import Dict, Any
from app.services.simulators.operations.node_manager import NodeManager
from app.utils.messages_th import get_message


class QueueNodeManager(NodeManager):
    """Enhanced Queue-specific node manager for ArrayQueue operations"""
    
    def create_instance_data(self, class_name: str) -> Dict[str, Any]:
        """Create new instance data structure for ArrayQueue"""
        if class_name == "ArrayQueue":
            return {
                "data": [],
                "class_type": class_name,
                "attributes": {},
                "history": []  # Track operation history
            }
        return super().create_instance_data(class_name)
    
    def create_instance(self, var_name: str, class_name: str) -> str:
        """Create a new instance of a class"""
        if class_name == "ArrayQueue":
            self.context["instances"][var_name] = self.create_instance_data(class_name)
            self.context["active_instance"] = var_name
            return get_message("instance_created_generic", class_name=class_name, var_name=var_name)
        else:
            return super().create_instance(var_name, class_name)
    
    def queue_enqueue(self, instance: Dict[str, Any], value: Any, instance_name: str = "") -> Dict[str, Any]:
        """Enqueue value to queue with detailed tracking"""
        # Check for overflow if max_size is defined
        max_size = instance.get("attributes", {}).get("max_size")
        if max_size is not None and len(instance["data"]) >= max_size:
            return {
                "message": f"เพิ่มข้อมูล '{value}' ลงใน queue" + " - Queue is full (overflow)",
                "operation": "enqueue",
                "value": value,
                "error": "overflow",
                "instance_name": instance_name
            }
        
        old_data = instance["data"].copy()
        instance["data"].append(value)
        new_data = instance["data"].copy()
        
        # Record operation in history
        instance["history"].append({
            "operation": "enqueue",
            "value": value,
            "before": old_data,
            "after": new_data
        })
        
        return {
            "message": f"เพิ่มข้อมูล '{value}' ลงใน queue (FIFO - First In First Out)",
            "operation": "enqueue",
            "value": value,
            "before_data": old_data,
            "after_data": new_data,
            "instance_name": instance_name
        }
    
    def queue_dequeue(self, instance: Dict[str, Any], instance_name: str = "") -> Dict[str, Any]:
        """Dequeue value from queue with detailed tracking"""
        if not instance["data"]:
            return {
                "message": "ลบข้อมูลออกจาก queue ที่ตำแหน่งหน้า" + " - " + get_message("traverse_empty"),
                "operation": "dequeue",
                "value": None,
                "error": "underflow",
                "instance_name": instance_name
            }
        
        old_data = instance["data"].copy()
        value = instance["data"].pop(0)  # FIFO - remove from front
        new_data = instance["data"].copy()
        
        # Record operation in history
        instance["history"].append({
            "operation": "dequeue",
            "value": value,
            "before": old_data,
            "after": new_data
        })
        
        return {
            "message": f"ลบข้อมูลออกจาก queue ที่ตำแหน่งหน้า → คืนค่า {value}",
            "operation": "dequeue",
            "value": value,
            "before_data": old_data,
            "after_data": new_data,
            "instance_name": instance_name
        }
    
    def queue_front(self, instance: Dict[str, Any], instance_name: str = "") -> Dict[str, Any]:
        """Get front value of queue"""
        if not instance["data"]:
            return {
                "message": "ดูข้อมูลที่ตำแหน่งหน้าของ queue โดยไม่ลบออก" + " - " + get_message("traverse_empty"),
                "operation": "front",
                "value": None,
                "error": "empty_queue",
                "instance_name": instance_name
            }
        
        front_value = instance["data"][0]
        return {
            "message": f"ดูข้อมูลที่ตำแหน่งหน้าของ queue โดยไม่ลบออก → คืนค่า {front_value}",
            "operation": "front",
            "value": front_value,
            "instance_name": instance_name
        }
    
    def queue_back(self, instance: Dict[str, Any], instance_name: str = "") -> Dict[str, Any]:
        """Get back value of queue"""
        if not instance["data"]:
            return {
                "message": "ดูข้อมูลที่ตำแหน่งท้ายของ queue โดยไม่ลบออก" + " - " + get_message("traverse_empty"),
                "operation": "back",
                "value": None,
                "error": "empty_queue",
                "instance_name": instance_name
            }
        
        back_value = instance["data"][-1]
        return {
            "message": f"ดูข้อมูลที่ตำแหน่งท้ายของ queue โดยไม่ลบออก → คืนค่า {back_value}",
            "operation": "back",
            "value": back_value,
            "instance_name": instance_name
        }
    
    def queue_size(self, instance: Dict[str, Any], instance_name: str = "") -> Dict[str, Any]:
        """Get queue size"""
        size = len(instance["data"])
        return {
            "message": f"นับจำนวนข้อมูลใน queue → คืนค่า {size}",
            "operation": "size",
            "value": size,
            "instance_name": instance_name
        }
    
    def queue_is_empty(self, instance: Dict[str, Any], instance_name: str = "") -> Dict[str, Any]:
        """Check if queue is empty"""
        is_empty = len(instance["data"]) == 0
        return {
            "message": f"ตรวจสอบว่า queue ว่างเปล่าหรือไม่ → คืนค่า {is_empty}",
            "operation": "is_empty",
            "value": is_empty,
            "instance_name": instance_name
        }
    
    def queue_print(self, instance: Dict[str, Any], instance_name: str = "") -> Dict[str, Any]:
        """Print queue contents"""
        queue_data = instance["data"].copy()
        print_output = str(queue_data)
        self.context["stdout"].append(print_output)
        return {
            "message": f"แสดงข้อมูล {instance_name}: {queue_data}",
            "operation": "printQueue",
            "value": queue_data,
            "instance_name": instance_name,
            "stdout": print_output  # Add print output to result
        }

