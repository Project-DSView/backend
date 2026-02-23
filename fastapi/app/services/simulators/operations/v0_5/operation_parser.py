import re
from typing import List, Dict, Any
from app.schemas.playground import ExecutionStepSchema
from app.services.simulators.operations.node_manager import NodeManager
from app.services.simulators.operations.print_handler import PrintHandler


class OperationParser:
    """Handles parsing and execution of individual operations"""
    
    def __init__(self, context: Dict[str, Any]):
        self.context = context
        self.node_manager = NodeManager(context)
        self.print_handler = PrintHandler(context)
    
    def parse_and_execute(self, line: str, line_number: int, step_number: int, 
                         steps: List[ExecutionStepSchema], 
                         create_step_func) -> bool:
        """Parse and execute a single operation. Returns True if handled."""
        
        # Handle print statements first
        if self.print_handler.handle_print_statement(line, line_number, step_number, steps, create_step_func):
            return True
        
        # Class instantiation
        if self._handle_class_instantiation(line, line_number, step_number, steps, create_step_func):
            return True
        
        # Node creation with parameter
        if self._handle_node_creation(line, line_number, step_number, steps, create_step_func):
            return True
        
        # Direct attribute assignment
        if self._handle_attribute_assignment(line, line_number, step_number, steps, create_step_func):
            return True
        
        # Chained attribute assignment
        if self._handle_chained_assignment(line, line_number, step_number, steps, create_step_func):
            return True
        
        # Method calls
        if self._handle_method_calls(line, line_number, step_number, steps, create_step_func):
            return True
        
        # Legacy operations
        if self._handle_legacy_operations(line, line_number, step_number, steps, create_step_func):
            return True
        
        return False
    
    def _handle_class_instantiation(self, line: str, line_number: int, step_number: int, 
                                   steps: List[ExecutionStepSchema], create_step_func) -> bool:
        """Handle class instantiation: var = ClassName()"""
        match = re.match(r"(\w+)\s*=\s*(\w+)\(\)", line)
        if not match:
            return False
        
        var_name = match.group(1)
        class_name = match.group(2)
        
        if class_name not in self.context["classes"]:
            return False
        
        try:
            message = self.node_manager.create_instance(var_name, class_name)
            steps.append(create_step_func(step_number, line_number, line, message))
            return True
        except ValueError as e:
            steps.append(create_step_func(step_number, line_number, line, error=str(e)))
            raise e
    
    def _handle_node_creation(self, line: str, line_number: int, step_number: int, 
                             steps: List[ExecutionStepSchema], create_step_func) -> bool:
        """Handle node creation: pNew = DataNode("John")"""
        match = re.match(r"(\w+)\s*=\s*DataNode\([\"']([^\"']*)[\"']\)", line)
        if not match:
            return False
        
        var_name = match.group(1)
        node_name = match.group(2)
        
        message = self.node_manager.create_node(var_name, node_name)
        steps.append(create_step_func(step_number, line_number, line, message))
        return True
    
    def _handle_attribute_assignment(self, line: str, line_number: int, step_number: int, 
                                   steps: List[ExecutionStepSchema], create_step_func) -> bool:
        """Handle direct attribute assignment: mylist.head = pNew"""
        match = re.match(r"(\w+)\.(\w+)\s*=\s*(\w+)", line)
        if not match:
            return False
        
        instance_name = match.group(1)
        attribute = match.group(2)
        value_var = match.group(3)
        
        if instance_name not in self.context["instances"] or value_var not in self.context["variables"]:
            return False
        
        message = self.node_manager.set_attribute(instance_name, attribute, value_var)
        steps.append(create_step_func(step_number, line_number, line, message))
        return True
    
    def _handle_chained_assignment(self, line: str, line_number: int, step_number: int, 
                                 steps: List[ExecutionStepSchema], create_step_func) -> bool:
        """Handle chained attribute assignment: mylist.head.next = pNew or pNew.prev = mylist.head"""
        
        # Handle mylist.head.next = pNew
        match = re.match(r"(\w+)\.head\.next\s*=\s*(\w+)", line)
        if match:
            instance_name = match.group(1)
            value_var = match.group(2)
            
            if instance_name not in self.context["instances"] or value_var not in self.context["variables"]:
                return False
            
            message = self.node_manager.set_chained_attribute(instance_name, value_var)
            steps.append(create_step_func(step_number, line_number, line, message))
            return True
        
        # Handle pNew.prev = mylist.head (for doubly linked list)
        match = re.match(r"(\w+)\.prev\s*=\s*(\w+)\.head", line)
        if match:
            node_var = match.group(1)
            instance_name = match.group(2)
            
            if instance_name not in self.context["instances"] or node_var not in self.context["variables"]:
                return False
            
            message = self.node_manager.set_prev_attribute(node_var, instance_name)
            steps.append(create_step_func(step_number, line_number, line, message))
            return True
        
        return False
    
    def _handle_method_calls(self, line: str, line_number: int, step_number: int, 
                           steps: List[ExecutionStepSchema], create_step_func) -> bool:
        """Handle method calls: mylist.traverse(), mylist.insertFront("data"), etc."""
        method_match = re.match(r"(\w+)\.(\w+)\((.*?)\)", line)
        if not method_match:
            return False
        
        instance_name = method_match.group(1)
        method_name = method_match.group(2)
        params = method_match.group(3).strip()
        
        if instance_name not in self.context["instances"]:
            raise ValueError(f"Instance '{instance_name}' not found")
        
        instance = self.context["instances"][instance_name]
        self.context["active_instance"] = instance_name
        
        message = self._execute_method(instance, instance_name, method_name, params)
        steps.append(create_step_func(step_number, line_number, line, message))
        return True
    
    def _execute_method(self, instance: Dict[str, Any], instance_name: str, 
                       method_name: str, params: str) -> str:
        """Execute a specific method on an instance"""
        if method_name == "traverse":
            display_data = self._get_instance_display(instance)
            if not display_data:
                # Simulate the print output for empty list
                self.context["print_output"].append("Empty list")
                return f"{instance_name}.traverse(): empty list"
            else:
                # Simulate the print output that traverse() would generate
                traverse_output = "-> " + "-> ".join(display_data)
                self.context["print_output"].append(traverse_output)
                display_str = " -> ".join(display_data)
                return f"{instance_name}.traverse(): -> {display_str}"
        
        elif method_name == "insertFront":
            if params:
                data = params.strip('"\'')
                message = self.node_manager.insert_front(instance, data)
                return f"{message} of {instance_name}"
            return "insertFront requires data parameter"
        
        elif method_name == "insertLast":
            if params:
                data = params.strip('"\'')
                message = self.node_manager.insert_last(instance, data)
                return f"{message} of {instance_name}"
            return "insertLast requires data parameter"
        
        elif method_name == "insertBefore":
            param_parts = [p.strip().strip('"\'') for p in params.split(',')]
            if len(param_parts) == 2:
                target_name, new_data = param_parts
                message = self.node_manager.insert_before(instance, target_name, new_data)
                return f"{message} in {instance_name}"
            return "insertBefore requires target and data parameters"
        
        elif method_name == "delete":
            if params:
                target_name = params.strip('"\'')
                message = self.node_manager.delete_node(instance, target_name)
                return f"{message} from {instance_name}"
            return "delete requires target parameter"
        
        elif method_name == "insert" and params:
            # Legacy support
            try:
                value = int(params)
                instance["data"].append(value)
                self.context["linkedlist"] = instance["data"].copy()
                return f"Inserted {value} into '{instance_name}'"
            except ValueError:
                raise ValueError(f"Invalid value for insert: {params}")
        
        else:
            raise ValueError(f"Unsupported method '{method_name}' for instance '{instance_name}'")
    
    def _handle_legacy_operations(self, line: str, line_number: int, step_number: int, 
                                steps: List[ExecutionStepSchema], create_step_func) -> bool:
        """Handle legacy linkedlist operations for backward compatibility"""
        if line == "linkedlist = []":
            self.context["linkedlist"] = []
            steps.append(create_step_func(step_number, line_number, line, "Initialized empty linkedlist"))
            return True
        
        return False
    
    def _get_instance_display(self, instance: Dict[str, Any]) -> List[str]:
        """Get display representation of an instance"""
        if instance.get("class_type") in ["SinglyLinkedList", "DoublyLinkedList"]:
            return self._traverse_linked_list(instance)
        return instance.get("data", [])
    
    def _traverse_linked_list(self, instance: Dict[str, Any]) -> List[str]:
        """Traverse a linked list instance and return display data"""
        result = []
        if instance.get("head") is None:
            return []
        
        current_node_id = instance["head"]
        visited = set()  # Prevent infinite loops
        
        while current_node_id is not None and current_node_id not in visited:
            visited.add(current_node_id)
            if current_node_id in self.context["nodes"]:
                node = self.context["nodes"][current_node_id]
                result.append(node["name"])
                current_node_id = node["next"]
            else:
                break
                
        return result