"""
Simulators module for DSView Backend API.

This module provides data structure simulators and related functionality
for code execution and step-by-step visualization.
"""

from .simulator_factory import SimulatorFactory, DataStructureSimulator
from .common.base_simulator import BaseSimulator, FunctionDefinitionTracker
from .stack_simulator import StackSimulator
from .binarysearchtree.binary_search_tree_simulator import BinarySearchTreeSimulator
from .graph_simulator import GraphSimulator

# Linked list simulators
from .linkedlist.singly_linked_list_simulator import SinglyLinkedListSimulator
from .linkedlist.doubly_linked_list_simulator import DoublyLinkedListSimulator

# Graph simulators
from .graph.undirected_graph_simulator import UndirectedGraphSimulator
from .graph.directed_graph_simulator import DirectedGraphSimulator

__all__ = [
    # Factory and protocol
    "SimulatorFactory",
    "DataStructureSimulator",
    # Common base classes
    "BaseSimulator",
    "FunctionDefinitionTracker",
    # Individual simulators
    "StackSimulator",
    "BinarySearchTreeSimulator", 
    "GraphSimulator",
    # Linked list simulators
    "SinglyLinkedListSimulator",
    "DoublyLinkedListSimulator",
    # Graph simulators
    "UndirectedGraphSimulator",
    "DirectedGraphSimulator",
]
