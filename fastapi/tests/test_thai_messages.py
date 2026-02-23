#!/usr/bin/env python3
"""
Test script to demonstrate Thai messages using frontend drag & drop style
"""

import sys
import os
sys.path.append(os.path.dirname(os.path.abspath(__file__)))

from app.utils.messages_th import (
    get_class_defined_message,
    get_instance_created_message,
    get_insert_message,
    get_delete_message,
    get_traverse_message,
    get_stack_message,
    get_bst_message,
    get_graph_message
)

def test_linked_list_messages():
    """Test singly linked list messages"""
    print("=== Singly Linked List Messages ===")
    print(f"Class definition: {get_class_defined_message('DataNode')}")
    print(f"Class definition: {get_class_defined_message('SinglyLinkedList')}")
    print(f"Instance creation: {get_instance_created_message('SinglyLinkedList', 'mylist')}")
    print(f"Insert front: {get_insert_message('front', 'Tony', 1)}")
    print(f"Insert last: {get_insert_message('last', 'Saori', 2)}")
    print(f"Insert before: {get_insert_message('before', 'Ako', 3, 'John')}")
    print(f"Delete: {get_delete_message('John', 2)}")
    print(f"Traverse: {get_traverse_message('mylist', 'traverse', '-> Tony -> Saori')}")
    print()

def test_stack_messages():
    """Test stack messages"""
    print("=== Stack Messages ===")
    print(f"Push: {get_stack_message('push', '1')}")
    print(f"Pop: {get_stack_message('pop')}")
    print(f"Peek: {get_stack_message('peek')}")
    print(f"Is Empty: {get_stack_message('is_empty')}")
    print(f"Size: {get_stack_message('size')}")
    print()

def test_bst_messages():
    """Test BST messages"""
    print("=== Binary Search Tree Messages ===")
    print(f"Insert: {get_bst_message('insert', '5')}")
    print(f"Delete: {get_bst_message('delete', '5')}")
    print(f"Search: {get_bst_message('search', '5')}")
    print(f"Inorder: {get_bst_message('traverse_inorder')}")
    print(f"Preorder: {get_bst_message('traverse_preorder')}")
    print(f"Postorder: {get_bst_message('traverse_postorder')}")
    print()

def test_graph_messages():
    """Test graph messages"""
    print("=== Graph Messages ===")
    print(f"Add Vertex: {get_graph_message('add_vertex', 'A')}")
    print(f"Add Edge: {get_graph_message('add_edge', from_vertex='A', to_vertex='B')}")
    print(f"Remove Vertex: {get_graph_message('remove_vertex', 'A')}")
    print(f"Remove Edge: {get_graph_message('remove_edge', from_vertex='A', to_vertex='B')}")
    print(f"DFS: {get_graph_message('traversal_dfs', start_vertex='A')}")
    print(f"BFS: {get_graph_message('traversal_bfs', start_vertex='A')}")
    print(f"Shortest Path: {get_graph_message('shortest_path', start_vertex='A', end_vertex='B')}")
    print()

if __name__ == "__main__":
    print("Testing Thai Messages with Frontend Drag & Drop Style")
    print("=" * 60)
    
    test_linked_list_messages()
    test_stack_messages()
    test_bst_messages()
    test_graph_messages()
    
    print("All tests completed successfully!")
