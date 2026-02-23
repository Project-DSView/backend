from typing import Dict, Any, List
from app.services.simulators.operations.node_manager import NodeManager
from app.utils.messages_th import get_bst_message, get_message


class BSTNodeManager(NodeManager):
    """Enhanced BST-specific node manager for Binary Search Tree operations"""
    
    def create_instance_data(self, class_name: str) -> Dict[str, Any]:
        """Create new instance data structure for BST classes"""
        if class_name == "BST":
            return {
                "root": None,
                "class_type": class_name,
                "attributes": {},
                "history": []  # Track operation history
            }
        elif class_name == "BSTNode":
            return {
                "data": None,
                "left": None,
                "right": None,
                "class_type": class_name,
                "attributes": {},
                "history": []
            }
        return super().create_instance_data(class_name)
    
    def create_instance(self, var_name: str, class_name: str) -> str:
        """Create a new instance of a class"""
        if class_name == "BST":
            self.context["instances"][var_name] = self.create_instance_data(class_name)
            self.context["active_instance"] = var_name
            return get_message("instance_created_generic", class_name=class_name, var_name=var_name)
        elif class_name == "BSTNode":
            self.context["instances"][var_name] = self.create_instance_data(class_name)
            self.context["active_instance"] = var_name
            return get_message("instance_created_generic", class_name=class_name, var_name=var_name)
        else:
            return super().create_instance(var_name, class_name)
    
    def create_node_instance(self, var_name: str, data_value: Any) -> str:
        """Create a new BSTNode instance with data"""
        node_data = self.create_instance_data("BSTNode")
        node_data["data"] = data_value
        self.context["instances"][var_name] = node_data
        self.context["active_instance"] = var_name
        return get_message("node_created", var_name=var_name, node_name=str(data_value))
    
    def bst_insert(self, instance: Dict[str, Any], value: Any, instance_name: str = "") -> Dict[str, Any]:
        """Insert value into BST with detailed tracking"""
        old_root = self._deep_copy_tree(instance["root"])
        
        # Create new node
        new_node = {
            "data": value,
            "left": None,
            "right": None,
            "class_type": "BSTNode"
        }
        
        if instance["root"] is None:
            # Tree is empty
            instance["root"] = new_node
            path = ["root"]
        else:
            # Find insertion point
            path = self._find_insertion_path(instance["root"], value)
            current = instance["root"]
            
            # Navigate to insertion point
            for direction in path[1:-1]:  # Skip 'root' and final direction
                if direction == "left":
                    current = current["left"]
                else:
                    current = current["right"]
            
            # Insert the new node
            final_direction = path[-1]
            if final_direction == "left":
                current["left"] = new_node
            else:
                current["right"] = new_node
        
        new_root = self._deep_copy_tree(instance["root"])
        
        # Record operation in history
        instance["history"].append({
            "operation": "insert",
            "value": value,
            "before": old_root,
            "after": new_root,
            "path": path
        })
        
        return {
            "message": get_bst_message("insert", str(value)) + f" → เส้นทาง: {' → '.join(path)}",
            "operation": "insert",
            "value": value,
            "inserted_node": str(value),  # Explicitly indicate which node was inserted
            "current_node": str(value),  # Current node for animation
            "path": path,
            "before_tree": old_root,
            "after_tree": new_root,
            "instance_name": instance_name
        }
    
    def bst_delete(self, instance: Dict[str, Any], value: Any, instance_name: str = "") -> Dict[str, Any]:
        """Delete value from BST with detailed tracking"""
        old_root = self._deep_copy_tree(instance["root"])
        
        if instance["root"] is None:
            return {
                "message": get_bst_message("delete", str(value)) + " - " + get_message("traverse_empty"),
                "operation": "delete",
                "value": None,
                "error": "empty_tree",
                "instance_name": instance_name
            }
        
        # Find the node to delete
        deleted_value, new_root = self._delete_recursive(instance["root"], value)
        instance["root"] = new_root
        
        if deleted_value is None:
            return {
                "message": get_bst_message("delete", str(value)) + " - " + get_message("target_not_found", target_name=str(value)),
                "operation": "delete",
                "value": None,
                "error": "value_not_found",
                "instance_name": instance_name
            }
        
        new_root_copy = self._deep_copy_tree(instance["root"])
        
        # Record operation in history
        instance["history"].append({
            "operation": "delete",
            "value": deleted_value,
            "before": old_root,
            "after": new_root_copy
        })
        
        return {
            "message": get_bst_message("delete", str(value)) + f" → ลบ {deleted_value} สำเร็จ",
            "operation": "delete",
            "value": deleted_value,
            "current_node": str(deleted_value),  # Current node (deleted node) for animation
            "deleted_node": str(deleted_value),  # Explicitly indicate which node was deleted
            "before_tree": old_root,
            "after_tree": new_root_copy,
            "instance_name": instance_name
        }
    
    def bst_is_empty(self, instance: Dict[str, Any], instance_name: str = "") -> Dict[str, Any]:
        """Check if BST is empty"""
        is_empty = instance["root"] is None
        return {
            "message": get_bst_message("search", "") + f" → คืนค่า {is_empty}",
            "operation": "is_empty",
            "value": is_empty,
            "instance_name": instance_name
        }
    
    def bst_find_min(self, instance: Dict[str, Any], instance_name: str = "") -> Dict[str, Any]:
        """Find minimum value in BST"""
        if instance["root"] is None:
            return {
                "message": get_bst_message("search", "") + " - " + get_message("traverse_empty"),
                "operation": "findMin",
                "value": None,
                "error": "empty_tree",
                "instance_name": instance_name
            }
        
        current = instance["root"]
        path = ["root"]
        
        while current["left"] is not None:
            current = current["left"]
            path.append("left")
        
        min_value = current["data"]
        return {
            "message": get_bst_message("search", str(min_value)) + f" → คืนค่า {min_value} (เส้นทาง: {' → '.join(path)})",
            "operation": "findMin",
            "value": min_value,
            "current_node": str(min_value),  # Current node (found node) for animation
            "path": path,
            "instance_name": instance_name
        }
    
    def bst_find_max(self, instance: Dict[str, Any], instance_name: str = "") -> Dict[str, Any]:
        """Find maximum value in BST"""
        if instance["root"] is None:
            return {
                "message": get_bst_message("search", "") + " - " + get_message("traverse_empty"),
                "operation": "findMax",
                "value": None,
                "error": "empty_tree",
                "instance_name": instance_name
            }
        
        current = instance["root"]
        path = ["root"]
        
        while current["right"] is not None:
            current = current["right"]
            path.append("right")
        
        max_value = current["data"]
        return {
            "message": get_bst_message("search", str(max_value)) + f" → คืนค่า {max_value} (เส้นทาง: {' → '.join(path)})",
            "operation": "findMax",
            "value": max_value,
            "current_node": str(max_value),  # Current node (found node) for animation
            "path": path,
            "instance_name": instance_name
        }
    
    def bst_traverse(self, instance: Dict[str, Any], instance_name: str = "") -> Dict[str, Any]:
        """Perform all three traversals"""
        if instance["root"] is None:
            self.context["stdout"].append("The tree is empty.")
            return {
                "message": get_bst_message("traverse_inorder") + " - " + get_message("traverse_empty"),
                "operation": "traverse",
                "value": {"preorder": [], "inorder": [], "postorder": []},
                "instance_name": instance_name
            }
        
        preorder_result = []
        inorder_result = []
        postorder_result = []
        
        self._preorder_traversal(instance["root"], preorder_result)
        self._inorder_traversal(instance["root"], inorder_result)
        self._postorder_traversal(instance["root"], postorder_result)
        
        # Add to print output
        self.context["stdout"].append(f"• Preorder: -> {' '.join(map(str, preorder_result))}")
        self.context["stdout"].append(f"• Inorder: -> {' '.join(map(str, inorder_result))}")
        self.context["stdout"].append(f"• Postorder: -> {' '.join(map(str, postorder_result))}")
        
        return {
            "message": get_bst_message("traverse_inorder") + " → ดำเนินการทั้ง 3 แบบ",
            "operation": "traverse",
            "value": {
                "preorder": preorder_result,
                "inorder": inorder_result,
                "postorder": postorder_result
            },
            "instance_name": instance_name
        }
    
    def bst_preorder(self, instance: Dict[str, Any], root_node: Dict[str, Any], instance_name: str = "") -> Dict[str, Any]:
        """Perform preorder traversal"""
        result = []
        if root_node is not None:
            self._preorder_traversal(root_node, result)
        
        return {
            "message": get_bst_message("traverse_preorder") + f" → {result}",
            "operation": "preorder",
            "value": result,
            "instance_name": instance_name
        }
    
    def bst_inorder(self, instance: Dict[str, Any], root_node: Dict[str, Any], instance_name: str = "") -> Dict[str, Any]:
        """Perform inorder traversal"""
        result = []
        if root_node is not None:
            self._inorder_traversal(root_node, result)
        
        return {
            "message": get_bst_message("traverse_inorder") + f" → {result}",
            "operation": "inorder",
            "value": result,
            "instance_name": instance_name
        }
    
    def bst_postorder(self, instance: Dict[str, Any], root_node: Dict[str, Any], instance_name: str = "") -> Dict[str, Any]:
        """Perform postorder traversal"""
        result = []
        if root_node is not None:
            self._postorder_traversal(root_node, result)
        
        return {
            "message": get_bst_message("traverse_postorder") + f" → {result}",
            "operation": "postorder",
            "value": result,
            "instance_name": instance_name
        }
    
    def _find_insertion_path(self, root: Dict[str, Any], value: Any) -> List[str]:
        """Find the path where a value should be inserted"""
        path = ["root"]
        current = root
        
        while True:
            if value < current["data"]:
                path.append("left")
                if current["left"] is None:
                    break
                current = current["left"]
            else:
                path.append("right")
                if current["right"] is None:
                    break
                current = current["right"]
        
        return path
    
    def _delete_recursive(self, root: Dict[str, Any], key: Any):
        """Recursively delete a node and return (deleted_value, new_root)"""
        if root is None:
            return None, None
        
        if key < root["data"]:
            deleted_value, new_left = self._delete_recursive(root["left"], key)
            root["left"] = new_left
            return deleted_value, root
        elif key > root["data"]:
            deleted_value, new_right = self._delete_recursive(root["right"], key)
            root["right"] = new_right
            return deleted_value, root
        else:
            # Found the node to delete
            deleted_value = root["data"]
            
            # Case 1: Node with only right child or no child
            if root["left"] is None:
                return deleted_value, root["right"]
            # Case 2: Node with only left child
            elif root["right"] is None:
                return deleted_value, root["left"]
            
            # Case 3: Node with two children
            # Find inorder successor (minimum in right subtree)
            successor = self._find_min_node(root["right"])
            root["data"] = successor["data"]
            
            # Delete the successor
            _, new_right = self._delete_recursive(root["right"], successor["data"])
            root["right"] = new_right
            
            return deleted_value, root
    
    def _find_min_node(self, node: Dict[str, Any]) -> Dict[str, Any]:
        """Find the node with minimum value"""
        current = node
        while current["left"] is not None:
            current = current["left"]
        return current
    
    def _preorder_traversal(self, node: Dict[str, Any], result: List):
        """Perform preorder traversal: Root -> Left -> Right"""
        if node is not None:
            result.append(node["data"])
            self._preorder_traversal(node["left"], result)
            self._preorder_traversal(node["right"], result)
    
    def _inorder_traversal(self, node: Dict[str, Any], result: List):
        """Perform inorder traversal: Left -> Root -> Right"""
        if node is not None:
            self._inorder_traversal(node["left"], result)
            result.append(node["data"])
            self._inorder_traversal(node["right"], result)
    
    def _postorder_traversal(self, node: Dict[str, Any], result: List):
        """Perform postorder traversal: Left -> Right -> Root"""
        if node is not None:
            self._postorder_traversal(node["left"], result)
            self._postorder_traversal(node["right"], result)
            result.append(node["data"])
    
    def _deep_copy_tree(self, node: Dict[str, Any]) -> Dict[str, Any]:
        """Create a deep copy of a tree node"""
        if node is None:
            return None
        
        return {
            "data": node["data"],
            "left": self._deep_copy_tree(node["left"]),
            "right": self._deep_copy_tree(node["right"]),
            "class_type": node.get("class_type", "BSTNode")
        }