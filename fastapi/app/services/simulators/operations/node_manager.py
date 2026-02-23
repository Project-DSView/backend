from typing import Dict, Any
from app.utils.messages_th import (
    get_instance_created_message, get_node_created_message, get_insert_message,
    get_delete_message, get_error_message, get_message
)


class NodeManager:
    """Manages node creation and linked list instance operations with enhanced tracking"""
    
    def __init__(self, context: Dict[str, Any]):
        self.context = context
    
    def create_instance_data(self, class_name: str) -> Dict[str, Any]:
        """Create new instance data structure"""
        return {
            "data": [],
            "head": None,
            "count": 0,
            "class_type": class_name,
            "attributes": {},
            "initialized": True
        }
    
    def create_node_data(self, class_name: str, name: str = "") -> Dict[str, Any]:
        """Create node data structure"""
        node_data = {
            "class_type": class_name,
            "name": name,
            "next": None,
            "attributes": {"name": name},
            "created_at": None
        }
        # Add prev for doubly linked list nodes
        if class_name == "DataNode":
            node_data["prev"] = None
        return node_data
    
    def create_instance(self, var_name: str, class_name: str) -> str:
        """Create a new instance of a class with proper initialization tracking"""
        # Ensure instances dict exists
        if "instances" not in self.context:
            self.context["instances"] = {}
        
        if class_name in ["SinglyLinkedList", "DoublyLinkedList"]:
            # Create instance with explicit initialization
            instance_data = self.create_instance_data(class_name)
            # Add tail for doubly linked list
            if class_name == "DoublyLinkedList":
                instance_data["tail"] = None
            self.context["instances"][var_name] = instance_data
            self.context["active_instance"] = var_name
            
            # Log the initialization step
            if class_name == "DoublyLinkedList":
                return get_instance_created_message(class_name, var_name, is_doubly=True)
            else:
                return get_instance_created_message(class_name, var_name, is_doubly=False)
            
        elif class_name == "DataNode":
            # This shouldn't happen with DataNode() - it needs a parameter
            if "nodes" not in self.context:
                self.context["nodes"] = {}
            if "variables" not in self.context:
                self.context["variables"] = {}
                
            node_id = f"node_{len(self.context['nodes'])}"
            self.context["nodes"][node_id] = self.create_node_data(class_name)
            self.context["variables"][var_name] = node_id
            return get_instance_created_message(class_name, var_name)
        else:
            raise ValueError(f"Unknown class: {class_name}")
    
    def create_node(self, var_name: str, node_name: str) -> str:
        """Create a new DataNode with a name"""
        # Ensure required dicts exist  
        if "nodes" not in self.context:
            self.context["nodes"] = {}
        if "variables" not in self.context:
            self.context["variables"] = {}
            
        node_id = f"node_{len(self.context['nodes'])}"
        node_data = self.create_node_data("DataNode", node_name)
        self.context["nodes"][node_id] = node_data
        self.context["variables"][var_name] = node_id
        return get_node_created_message(var_name, node_name)
    
    def set_attribute(self, instance_name: str, attribute: str, value_var: str) -> str:
        """Set an attribute of an instance"""
        instances = self.context.get("instances", {})
        variables = self.context.get("variables", {})
        
        if not isinstance(instances, dict) or instance_name not in instances:
            raise ValueError(get_error_message("instance_not_found", instance_name=instance_name))
        if not isinstance(variables, dict) or value_var not in variables:
            raise ValueError(get_error_message("variable_not_found", var_name=value_var))
        
        instance = instances[instance_name]
        node_id = variables[value_var]
        
        if attribute == "head":
            old_head = instance.get("head")
            instance["head"] = node_id
            self.context["active_instance"] = instance_name
            
            if old_head is None:
                return get_message("attribute_set_first", instance_name=instance_name, attribute=attribute, value_var=value_var)
            else:
                return get_message("attribute_set", instance_name=instance_name, attribute=attribute, value_var=value_var)
        
        return get_message("attribute_set_generic", instance_name=instance_name, attribute=attribute, value_var=value_var)
    
    def set_chained_attribute(self, instance_name: str, value_var: str) -> str:
        """Set a chained attribute like mylist.head.next"""
        instances = self.context.get("instances", {})
        variables = self.context.get("variables", {})
        nodes = self.context.get("nodes", {})
        
        if not isinstance(instances, dict) or instance_name not in instances:
            raise ValueError(get_error_message("instance_not_found", instance_name=instance_name))
        if not isinstance(variables, dict) or value_var not in variables:
            raise ValueError(get_error_message("variable_not_found", var_name=value_var))
        if not isinstance(nodes, dict):
            raise ValueError(get_error_message("nodes_corrupted"))
        
        instance = instances[instance_name]
        if instance["head"] and instance["head"] in nodes:
            head_node = nodes[instance["head"]]
            if isinstance(head_node, dict):
                head_node["next"] = variables[value_var]
                return get_message("chained_attribute_set", instance_name=instance_name, value_var=value_var)
        
        return get_message("chained_attribute_failed", instance_name=instance_name, value_var=value_var)
    
    def set_prev_attribute(self, node_var: str, instance_name: str) -> str:
        """Set prev attribute of a node to instance's head (for doubly linked list)"""
        instances = self.context.get("instances", {})
        variables = self.context.get("variables", {})
        nodes = self.context.get("nodes", {})
        
        if not isinstance(instances, dict) or instance_name not in instances:
            raise ValueError(get_error_message("instance_not_found", instance_name=instance_name))
        if not isinstance(variables, dict) or node_var not in variables:
            raise ValueError(get_error_message("variable_not_found", var_name=node_var))
        if not isinstance(nodes, dict):
            raise ValueError(get_error_message("nodes_corrupted"))
        
        instance = instances[instance_name]
        node_id = variables[node_var]
        head_id = instance.get("head")
        
        if node_id in nodes and isinstance(nodes[node_id], dict):
            nodes[node_id]["prev"] = head_id
            return get_message("prev_attribute_set", node_var=node_var, instance_name=instance_name)
        
        return get_message("prev_attribute_failed", node_var=node_var, instance_name=instance_name)
    
    def insert_front(self, instance: Dict[str, Any], data: str) -> str:
        """Insert a node at the front of the linked list"""
        if not isinstance(instance, dict):
            raise ValueError(get_error_message("invalid_instance"))
            
        # Ensure nodes dict exists
        if "nodes" not in self.context:
            self.context["nodes"] = {}
        
        # Create new node
        node_id = f"node_{len(self.context['nodes'])}"
        self.context["nodes"][node_id] = self.create_node_data("DataNode", data)
        
        # Insert at front
        old_head = instance.get("head")
        if old_head is None:
            instance["head"] = node_id
            instance["count"] = 1
            return get_insert_message("first", data)
        else:
            self.context["nodes"][node_id]["next"] = old_head
            instance["head"] = node_id
            instance["count"] += 1
            return get_insert_message("front", data, instance["count"])
    
    def insert_last(self, instance: Dict[str, Any], data: str) -> str:
        """Insert a node at the end of the linked list"""
        if not isinstance(instance, dict):
            raise ValueError(get_error_message("invalid_instance"))
            
        # Ensure nodes dict exists
        if "nodes" not in self.context:
            self.context["nodes"] = {}
        
        node_id = f"node_{len(self.context['nodes'])}"
        self.context["nodes"][node_id] = self.create_node_data("DataNode", data)
        
        if instance["head"] is None:
            instance["head"] = node_id
            instance["count"] = 1
            return get_insert_message("first", data)
        else:
            # Find last node
            current = instance["head"]
            nodes = self.context.get("nodes", {})
            while current and isinstance(nodes.get(current), dict) and nodes[current]["next"]:
                current = nodes[current]["next"]
            if current and isinstance(nodes.get(current), dict):
                nodes[current]["next"] = node_id
            
            instance["count"] += 1
            return get_insert_message("last", data, instance["count"])
    
    def insert_before(self, instance: Dict[str, Any], target_name: str, new_data: str) -> str:
        """Insert a node before the target node"""
        if not isinstance(instance, dict):
            raise ValueError(get_error_message("invalid_instance"))
            
        # Ensure nodes dict exists
        if "nodes" not in self.context:
            self.context["nodes"] = {}
        
        node_id = f"node_{len(self.context['nodes'])}"
        self.context["nodes"][node_id] = self.create_node_data("DataNode", new_data)
        
        nodes = self.context.get("nodes", {})
        if not isinstance(nodes, dict):
            return get_error_message("nodes_corrupted")
        
        # If inserting before head
        if (instance["head"] and instance["head"] in nodes and 
            isinstance(nodes[instance["head"]], dict) and
            nodes[instance["head"]]["name"] == target_name):
            nodes[node_id]["next"] = instance["head"]
            instance["head"] = node_id
            instance["count"] += 1
            return get_insert_message("before", new_data, instance["count"], target_name)
        else:
            # Find the node before target
            current = instance["head"]
            while current and current in nodes and isinstance(nodes[current], dict):
                next_node_id = nodes[current]["next"]
                if (next_node_id and next_node_id in nodes and 
                    isinstance(nodes[next_node_id], dict) and
                    nodes[next_node_id]["name"] == target_name):
                    nodes[node_id]["next"] = next_node_id
                    nodes[current]["next"] = node_id
                    instance["count"] += 1
                    return get_insert_message("before", new_data, instance["count"], target_name)
                current = next_node_id
            
            return get_message("target_not_found", target_name=target_name)
    
    def delete_node(self, instance: Dict[str, Any], target_name: str) -> str:
        """Delete a node with the given name"""
        if not isinstance(instance, dict):
            raise ValueError(get_error_message("invalid_instance"))
            
        nodes = self.context.get("nodes", {})
        if not isinstance(nodes, dict):
            return get_error_message("nodes_corrupted")
        
        if not instance["head"]:
            return get_message("delete_empty_list", target_name=target_name)
        elif (instance["head"] in nodes and isinstance(nodes[instance["head"]], dict) and
              nodes[instance["head"]]["name"] == target_name):
            # Delete head
            old_head = instance["head"]
            instance["head"] = nodes[old_head]["next"]
            instance["count"] = max(0, instance["count"] - 1)
            return get_delete_message(target_name, instance["count"], from_head=True)
        else:
            # Find and delete node
            current = instance["head"]
            while current and current in nodes and isinstance(nodes[current], dict):
                next_node_id = nodes[current]["next"]
                if (next_node_id and next_node_id in nodes and 
                    isinstance(nodes[next_node_id], dict) and
                    nodes[next_node_id]["name"] == target_name):
                    nodes[current]["next"] = nodes[next_node_id]["next"]
                    instance["count"] = max(0, instance["count"] - 1)
                    return get_delete_message(target_name, instance["count"], from_head=False)
                current = next_node_id
            
            return get_message("delete_target_not_found", target_name=target_name)