
import sys
import os
import time
from datetime import datetime

# Add the project root to sys.path
sys.path.append(os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__)))))

from app.services.simulators.direct_code_executor import DirectCodeExecutor

def measure_perf():
    code = """
class BSTNode:
    def __init__(self, data):
        self.data = data
        self.left = None
        self.right = None

class BST:
    def __init__(self):
        self.root = None

    def is_empty(self):
        return self.root is None

    def insert(self, data):
        new_node = BSTNode(data)
        if self.is_empty():
            self.root = new_node
        else:
            current_node = self.root
            while True:
                if data < current_node.data:
                    # Go left
                    if current_node.left is None:
                        current_node.left = new_node
                        return
                    current_node = current_node.left
                else:
                    # Go right
                    if current_node.right is None:
                        current_node.right = new_node
                        return
                    current_node = current_node.right

    def delete(self, data):
        def find_min_node(node):
            current = node
            while current.left is not None:
                current = current.left
            return current

        def _delete_recursive(root, key):
            if root is None:
                return root

            if key < root.data:
                root.left = _delete_recursive(root.left, key)
            elif key > root.data:
                root.right = _delete_recursive(root.right, key)
            else:
                if root.left is None:
                    return root.right
                elif root.right is None:
                    return root.left

                temp = find_min_node(root.right)
                root.data = temp.data
                root.right = _delete_recursive(root.right, temp.data)
            return root

        original_root_data = self.root.data if self.root else None
        self.root = _delete_recursive(self.root, data)
        
        if self.root is None or self.root.data != original_root_data:
            return data
        elif self.root.data == original_root_data and _delete_recursive(self.root, data) != self.root:
            return data
        else:
            return None

    def findMin(self):
        if self.is_empty():
            return None
        current_node = self.root
        while current_node.left is not None:
            current_node = current_node.left
        return current_node.data

    def findMax(self):
        if self.is_empty():
            return None
        current_node = self.root
        while current_node.right is not None:
            current_node = current_node.right
        return current_node.data

    def traverse(self):
        if self.is_empty():
            return
        self.preorder(self.root)
        self.inorder(self.root)
        self.postorder(self.root)

    def preorder(self, root):
        if root is not None:
            self.preorder(root.left)
            self.preorder(root.right)

    def inorder(self, root):
        if root is not None:
            self.inorder(root.left)
            self.inorder(root.right)

    def postorder(self, root):
        if root is not None:
            self.postorder(root.left)
            self.postorder(root.right)

myBST = BST()
# Similar workload to user's test
myBST.insert(14)
myBST.insert(23)
myBST.insert(7)
myBST.insert(10)
myBST.insert(33)
myBST.traverse()
myBST.delete(14)
myBST.findMin()
myBST.findMax()

myBST_del = BST()
myBST_del.insert(14)
myBST_del.insert(23)
myBST_del.insert(7)
myBST_del.insert(10)
myBST_del.insert(33)
myBST_del.insert(5)
myBST_del.insert(20)
myBST_del.insert(13)
myBST_del.traverse()
myBST_del.delete(5)
myBST_del.delete(14)
myBST_del.delete(7)
"""

    print("--- Measuring BST Performance ---")
    start_time = time.time()
    
    executor = DirectCodeExecutor()
    # Force full tracing for BST
    steps = executor.execute(code, data_structure_type="binarysearchtree")
    
    end_time = time.time()
    duration = end_time - start_time
    
    print(f"Total Steps: {len(steps)}")
    print(f"Execution Time: {duration:.4f} seconds")
    print(f"Steps/Second: {len(steps)/duration:.2f}")

if __name__ == "__main__":
    measure_perf()
