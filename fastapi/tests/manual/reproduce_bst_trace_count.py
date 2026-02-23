
import sys
import os
import json
from datetime import datetime

# Add the project root to sys.path
sys.path.append(os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__)))))

from app.services.simulators.direct_code_executor import DirectCodeExecutor

def reproduce():
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
        # A helper function to find the successor for case 4
        def find_min_node(node):
            current = node
            while current.left is not None:
                current = current.left
            return current

        # A helper function for recursive deletion
        def _delete_recursive(root, key):
            if root is None:
                return root

            if key < root.data:
                root.left = _delete_recursive(root.left, key)
            elif key > root.data:
                root.right = _delete_recursive(root.right, key)
            else:
                # Case 1: Node with only one child or no child
                if root.left is None:
                    return root.right
                elif root.right is None:
                    return root.left

                # Case 2: Node with two children, get the inorder successor
                temp = find_min_node(root.right)
                root.data = temp.data
                root.right = _delete_recursive(root.right, temp.data)
            return root

        # Store the original root data to check if a deletion occurred
        original_root_data = self.root.data if self.root else None
        self.root = _delete_recursive(self.root, data)
        
        # If the root data has changed, it means the old root was deleted and replaced
        # If a deletion occurred, we can return the data. Otherwise, return None.
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
            print("The tree is empty.")
            return

        print("• Preorder: ", end="")
        self.preorder(self.root)
        print()

        print("• Inorder: ", end="")
        self.inorder(self.root)
        print()
        
        print("• Postorder: ", end="")
        self.postorder(self.root)
        print()

    def preorder(self, root):
        if root is not None:
            print(f"-> {root.data}", end=" ")
            self.preorder(root.left)
            self.preorder(root.right)

    def inorder(self, root):
        if root is not None:
            self.inorder(root.left)
            print(f"-> {root.data}", end=" ")
            self.inorder(root.right)

    def postorder(self, root):
        if root is not None:
            self.postorder(root.left)
            self.postorder(root.right)
            print(f"-> {root.data}", end=" ")


myBST = BST()

# --- Test Case 1: Insertion and Traversal ---
print("--- Test Case 1: Insertion and Traversal ---")
myBST.insert(14)
myBST.insert(23)
myBST.insert(7)
myBST.insert(10)
myBST.insert(33)
myBST.traverse()

print("--- Test Case 2: Deletion and Finding Min/Max ---")
deleted_data = myBST.delete(14)
print(f"Deleted data: {deleted_data}")
print("Min: ", myBST.findMin())
print("Max: ", myBST.findMax())

# --- Test Case 3: More Deletion Tests ---
print("--- Test Case 3: More Deletion Tests ---")
myBST_del = BST()
myBST_del.insert(14)
myBST_del.insert(23)
myBST_del.insert(7)
myBST_del.insert(10)
myBST_del.insert(33)
myBST_del.insert(5)
myBST_del.insert(20)
myBST_del.insert(13)
print("Initial tree:")
myBST_del.traverse()

print("Deleting leaf node (5):")
deleted_data = myBST_del.delete(5)
print(f"Deleted data: {deleted_data}")
myBST_del.traverse()

print("Deleting node with two children (14):")
deleted_data = myBST_del.delete(14)
print(f"Deleted data: {deleted_data}")
myBST_del.traverse()

print("Deleting node with one child (7):")
deleted_data = myBST_del.delete(7)
print(f"Deleted data: {deleted_data}")
myBST_del.traverse()
"""

    print("--- Testing BST Step Count with 'binarysearchtree' ---")
    executor = DirectCodeExecutor()
    steps = executor.execute(code, data_structure_type="binarysearchtree")
    
    print(f"Total Steps Generated: {len(steps)}")
    
    # Print sample steps to see line numbers
    print("\nSample Steps (first 50):")
    for i, step in enumerate(steps[:50]):
        print(f"Step {step.stepNumber} [Line {step.line}]: {step.code.strip()}")

    # Print distribution
    line_counts = {}
    for step in steps:
        if step.line not in line_counts:
            line_counts[step.line] = 0
        line_counts[step.line] += 1
    
    print("\nMost frequent lines:")
    sorted_lines = sorted(line_counts.items(), key=lambda x: x[1], reverse=True)
    for line, count in sorted_lines[:10]:
        print(f"Line {line}: {count} times")

if __name__ == "__main__":
    reproduce()
