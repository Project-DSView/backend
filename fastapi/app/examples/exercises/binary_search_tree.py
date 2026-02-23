example = '''
class BSTNode:
    """
    A class representing a single node in the Binary Search Tree.

    Attributes:
        data: The data stored in the node.
        left: A reference to the left child node.
        right: A reference to the right child node.
    """
    def __init__(self, data):
        """
        Initializes a new BSTNode.

        Args:
            data: The data to be stored in the node.
        """
        self.data = data
        self.left = None
        self.right = None

class BST:
    """
    A class representing the Binary Search Tree.

    Attributes:
        root: The root node of the tree.
    """
    def __init__(self):
        """
        Initializes an empty Binary Search Tree.
        """
        self.root = None

    def is_empty(self):
        """
        Checks if the tree is empty.

        Returns:
            True if the tree is empty, False otherwise.
        """
        return self.root is None

    def insert(self, data):
        """
        Inserts a new data value into the tree.

        Args:
            data: The data to be inserted.
        """
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
        """
        Deletes a node with the specified data from the tree.

        Args:
            data: The data of the node to be deleted.

        Returns:
            The deleted data if successful, None otherwise.
        """
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
        """
        Finds the minimum value in the tree.

        Returns:
            The minimum value, or None if the tree is empty.
        """
        if self.is_empty():
            return None
        current_node = self.root
        while current_node.left is not None:
            current_node = current_node.left
        return current_node.data

    def findMax(self):
        """
        Finds the maximum value in the tree.

        Returns:
            The maximum value, or None if the tree is empty.
        """
        if self.is_empty():
            return None
        current_node = self.root
        while current_node.right is not None:
            current_node = current_node.right
        return current_node.data

    def traverse(self):
        """
        Performs all three traversals (preorder, inorder, postorder) and prints the results.
        """
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
        """
        Performs a preorder traversal (Root -> Left -> Right).
        """
        if root is not None:
            print(f"-> {root.data}", end=" ")
            self.preorder(root.left)
            self.preorder(root.right)

    def inorder(self, root):
        """
        Performs an inorder traversal (Left -> Root -> Right).
        """
        if root is not None:
            self.inorder(root.left)
            print(f"-> {root.data}", end=" ")
            self.inorder(root.right)

    def postorder(self, root):
        """
        Performs a postorder traversal (Left -> Right -> Root).
        """
        if root is not None:
            self.postorder(root.left)
            self.postorder(root.right)
            print(f"-> {root.data}", end=" ")


#
# Example usage based on the lab document
#

myBST = BST()

# --- Test Case 1: Insertion and Traversal ---
print("--- Test Case 1: Insertion and Traversal ---")
myBST.insert(14)
myBST.insert(23)
myBST.insert(7)
myBST.insert(10)
myBST.insert(33)
myBST.traverse()

print("\n--- Test Case 2: Deletion and Finding Min/Max ---")
deleted_data = myBST.delete(14)
print(f"Deleted data: {deleted_data}")
print("Min: ", myBST.findMin())
print("Max: ", myBST.findMax())

# --- Test Case 3: More Deletion Tests ---
print("\n--- Test Case 3: More Deletion Tests ---")
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

print("\nDeleting leaf node (5):")
deleted_data = myBST_del.delete(5)
print(f"Deleted data: {deleted_data}")
myBST_del.traverse()

print("\nDeleting node with two children (14):")
deleted_data = myBST_del.delete(14)
print(f"Deleted data: {deleted_data}")
myBST_del.traverse()

print("\nDeleting node with one child (7):")
deleted_data = myBST_del.delete(7)
print(f"Deleted data: {deleted_data}")
myBST_del.traverse()
'''
