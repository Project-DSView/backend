
from app.services.simulators.stack_simulator import StackSimulator
from app.services.simulators.queue_simulator import QueueSimulator
from app.services.simulators.binarysearchtree_simulator import BinarySearchTreeSimulator
from app.services.simulators.linkedlist.singly_linked_list_simulator import SinglyLinkedListSimulator
from datetime import datetime
import json

def test_simulator(sim_cls, name, code):
    print(f"Testing {name}...")
    sim = sim_cls()
    steps = sim.execute_code(code, "test_id", datetime.now())
    
    has_stdout = False
    for step in steps:
        if "stdout" in step.state:
            has_stdout = True
            print(f"  Step {step.stepNumber} has stdout: {step.state['stdout']}")
        if "print_output" in step.state:
            print(f"  WARNING: Step {step.stepNumber} still has print_output: {step.state['print_output']}")
            
    if has_stdout:
        print(f"SUCCESS: {name} produced stdout.")
    else:
        print(f"FAILURE: {name} did NOT produce stdout.")

if __name__ == "__main__":
    current_time = datetime.now()
    
    # Stack Test
    stack_code = """
class ArrayStack:
    def __init__(self):
        self.data = []
    def push(self, item):
        self.data.append(item)
    def pop(self):
        return self.data.pop()

stack = ArrayStack()
stack.push(1)
print(stack.pop())
"""
    try:
        test_simulator(StackSimulator, "StackSimulator", stack_code)
    except Exception as e:
        print(f"StackSimulator crashed: {e}")

    # Queue Test
    queue_code = """
class ArrayQueue:
    def __init__(self):
        self.data = []
    def enqueue(self, item):
        self.data.append(item)
    def dequeue(self):
        return self.data.pop(0)

q = ArrayQueue()
q.enqueue(1)
print(q.dequeue())
"""
    try:
        test_simulator(QueueSimulator, "QueueSimulator", queue_code)
    except Exception as e:
        print(f"QueueSimulator crashed: {e}")
        import traceback
        traceback.print_exc()

    # BST Test
    bst_code = """
class BSTNode:
    def __init__(self, data):
        self.data = data
        self.left = None
        self.right = None

class BST:
    def __init__(self):
        self.root = None
    def insert(self, data):
        if not self.root:
            self.root = BSTNode(data)
            
bst = BST()
bst.insert(5)
print("BST Inserted")
"""
    try:
        test_simulator(BinarySearchTreeSimulator, "BinarySearchTreeSimulator", bst_code)
    except Exception as e:
        print(f"BSTSimulator crashed: {e}")

    # LinkedList Test
    ll_code = """
class Node:
    def __init__(self, data):
        self.data = data
        self.next = None

class LinkedList:
    def __init__(self):
        self.head = None
    def append(self, data):
        if not self.head:
            self.head = Node(data)

ll = LinkedList()
ll.append(10)
print("LL Appended")
"""
    try:
        test_simulator(SinglyLinkedListSimulator, "SinglyLinkedListSimulator", ll_code)
    except Exception as e:
        print(f"SinglyLinkedListSimulator crashed: {e}")
