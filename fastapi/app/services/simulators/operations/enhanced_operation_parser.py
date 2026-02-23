import re
from typing import List, Dict, Any
from app.schemas.playground import ExecutionStepSchema
from app.services.simulators.operations.node_manager import NodeManager
from app.services.simulators.operations.print_handler import PrintHandler
from app.services.simulators.operations.behavior_analyzer import BehaviorAnalyzer


class EnhancedOperationParser:
    """Enhanced parser that uses behavior analysis instead of method names"""
    
    def __init__(self, context: Dict[str, Any]):
        self.context = context
        self.node_manager = NodeManager(context)
        self.print_handler = PrintHandler(context)
        
        # Initialize behavior analyzer and parse any existing classes
        self.behavior_analyzer = BehaviorAnalyzer(context)
        
        # If context has classes defined, we need to parse them for behavior analysis
        self._initialize_behavior_analysis()
    
    def _initialize_behavior_analysis(self):
        """Initialize behavior analysis for existing classes in context"""
        classes = self.context.get("classes", {})
        if isinstance(classes, dict) and classes:
            # Try to reconstruct class code for behavior analysis if possible
            # This is a simplified approach - in practice you might want to store
            # the original code or pass it explicitly
            pass
    
    def set_class_code(self, code: str):
        """Set the full class code for behavior analysis"""
        self.behavior_analyzer.parse_class_methods(code)
    
    def parse_and_execute(self, line: str, line_number: int, step_number: int, 
                         steps: List[ExecutionStepSchema], 
                         create_step_func) -> bool:
        """Parse and execute a single operation using behavior analysis"""
        
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
        
        # Method calls with behavior analysis
        if self._handle_method_calls_with_behavior(line, line_number, step_number, steps, create_step_func):
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
        
        # Ensure classes context exists
        if "classes" not in self.context or not isinstance(self.context["classes"], dict):
            self.context["classes"] = {}
        
        if class_name not in self.context["classes"]:
            return False
        
        try:
            message = self.node_manager.create_instance(var_name, class_name)
            
            # Create step with behavior information
            step_detail = {
                "operation": "instantiate",
                "class_name": class_name,
                "instance_name": var_name,
                "available_behaviors": self._get_class_behaviors(class_name)
            }
            
            steps.append(create_step_func(step_number, line_number, line, message, None, {"step_detail": step_detail}))
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
        
        # Add step_detail for progressive node rendering
        step_detail = {
            "operation": "node_creation",
            "node_variable": var_name,
            "node_value": node_name,
            "class_name": "DataNode",
            "is_connected": False  # Node created but not yet connected to list
        }
        steps.append(create_step_func(step_number, line_number, line, message, None, {"step_detail": step_detail}))
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
        
        # Ensure required contexts exist
        instances = self.context.get("instances", {})
        variables = self.context.get("variables", {})
        
        if not isinstance(instances, dict) or instance_name not in instances:
            return False
        if not isinstance(variables, dict) or value_var not in variables:
            return False
        
        message = self.node_manager.set_attribute(instance_name, attribute, value_var)
        
        # Add step_detail for pointer assignment
        step_detail = {
            "operation": "pointer_assignment",
            "source_var": value_var,
            "target_instance": instance_name,
            "target_attribute": attribute,
            "is_head_assignment": attribute == "head",
            "creates_connection": True
        }
        steps.append(create_step_func(step_number, line_number, line, message, None, {"step_detail": step_detail}))
        return True
    
    def _handle_chained_assignment(self, line: str, line_number: int, step_number: int, 
                                 steps: List[ExecutionStepSchema], create_step_func) -> bool:
        """Handle chained attribute assignment: mylist.head.next = pNew or pNew.prev = mylist.head"""
        
        # Handle mylist.head.next = pNew
        match = re.match(r"(\w+)\.head\.next\s*=\s*(\w+)", line)
        if match:
            instance_name = match.group(1)
            value_var = match.group(2)
            
            # Ensure required contexts exist
            instances = self.context.get("instances", {})
            variables = self.context.get("variables", {})
            
            if not isinstance(instances, dict) or instance_name not in instances:
                return False
            if not isinstance(variables, dict) or value_var not in variables:
                return False
            
            message = self.node_manager.set_chained_attribute(instance_name, value_var)
            
            # Add step_detail for chained pointer assignment (e.g., mylist.head.next = pNew)
            step_detail = {
                "operation": "chained_pointer_assignment",
                "source_var": value_var,
                "target_instance": instance_name,
                "target_chain": "head.next",
                "creates_connection": True
            }
            steps.append(create_step_func(step_number, line_number, line, message, None, {"step_detail": step_detail}))
            return True
        
        # Handle pNew.prev = mylist.head (for doubly linked list)
        match = re.match(r"(\w+)\.prev\s*=\s*(\w+)\.head", line)
        if match:
            node_var = match.group(1)
            instance_name = match.group(2)
            
            # Ensure required contexts exist
            instances = self.context.get("instances", {})
            variables = self.context.get("variables", {})
            
            if not isinstance(instances, dict) or instance_name not in instances:
                return False
            if not isinstance(variables, dict) or node_var not in variables:
                return False
            
            message = self.node_manager.set_prev_attribute(node_var, instance_name)
            
            # Add step_detail for prev pointer assignment (doubly linked list)
            step_detail = {
                "operation": "prev_pointer_assignment",
                "node_var": node_var,
                "target_instance": instance_name,
                "creates_connection": True
            }
            steps.append(create_step_func(step_number, line_number, line, message, None, {"step_detail": step_detail}))
            return True
        
        return False
    
    def _handle_method_calls_with_behavior(self, line: str, line_number: int, step_number: int, 
                                         steps: List[ExecutionStepSchema], create_step_func) -> bool:
        """Handle method calls using behavior analysis instead of method names"""
        method_match = re.match(r"(\w+)\.(\w+)\((.*?)\)", line)
        if not method_match:
            return False
        
        instance_name = method_match.group(1)
        method_name = method_match.group(2)
        params = method_match.group(3).strip()
        
        # Ensure instances context exists
        instances = self.context.get("instances", {})
        if not isinstance(instances, dict) or instance_name not in instances:
            raise ValueError(f"Instance '{instance_name}' not found")
        
        instance = instances[instance_name]
        self.context["active_instance"] = instance_name
        
        # Use behavior analyzer to execute method
        try:
            message = self.behavior_analyzer.execute_method_by_behavior(instance_name, method_name, params)
            
            # Create step with behavior analysis details
            method_info = self._get_method_behavior_info(instance.get("class_type"), method_name)
            
            step_detail = {
                "operation": "method_call",
                "method_name": method_name,
                "detected_behavior": method_info.get("behavior_type", "unknown") if method_info else "unknown",
                "parameters": params,
                "behavior_analysis": method_info
            }
            
            steps.append(create_step_func(step_number, line_number, line, message, None, {"step_detail": step_detail}))
            return True
            
        except Exception as e:
            raise ValueError(f"Error executing method {method_name}: {str(e)}")
    
    def _handle_legacy_operations(self, line: str, line_number: int, step_number: int, 
                                steps: List[ExecutionStepSchema], create_step_func) -> bool:
        """Handle legacy linkedlist operations for backward compatibility"""
        if line == "linkedlist = []":
            if "linkedlist" not in self.context or not isinstance(self.context["linkedlist"], list):
                self.context["linkedlist"] = []
            else:
                self.context["linkedlist"] = []
            steps.append(create_step_func(step_number, line_number, line, "Initialized empty linkedlist"))
            return True
        
        return False
    
    def _get_class_behaviors(self, class_name: str) -> Dict[str, str]:
        """Get all available behaviors for a class"""
        classes = self.context.get("classes", {})
        if (isinstance(classes, dict) and 
            class_name in classes and 
            isinstance(classes[class_name], dict) and
            "methods" in classes[class_name] and
            isinstance(classes[class_name]["methods"], dict)):
            
            methods = classes[class_name]["methods"]
            return {
                method_name: method_info.get("behavior_type", "unknown")
                for method_name, method_info in methods.items()
                if isinstance(method_info, dict)
            }
        return {}
    
    def _get_method_behavior_info(self, class_name: str, method_name: str) -> Dict[str, Any]:
        """Get detailed behavior information for a method"""
        if not class_name:
            return None
            
        method_key = f"{class_name}.{method_name}"
        if (hasattr(self.behavior_analyzer, 'method_definitions') and 
            isinstance(self.behavior_analyzer.method_definitions, dict) and
            method_key in self.behavior_analyzer.method_definitions):
            return self.behavior_analyzer.method_definitions[method_key]
        return None