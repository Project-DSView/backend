from typing import Protocol
from app.services.simulators.stack_simulator import StackSimulator
from app.services.simulators.linkedlist.singly_linked_list_simulator import SinglyLinkedListSimulator
from app.services.simulators.linkedlist.doubly_linked_list_simulator import DoublyLinkedListSimulator
from app.services.simulators.binarysearchtree.binary_search_tree_simulator import BinarySearchTreeSimulator
from app.services.simulators.graph_simulator import GraphSimulator
from app.services.simulators.graph.undirected_graph_simulator import UndirectedGraphSimulator
from app.services.simulators.graph.directed_graph_simulator import DirectedGraphSimulator
from app.services.simulators.queue_simulator import QueueSimulator


class DataStructureSimulator(Protocol):
    """Protocol for data structure simulators"""
    def execute_code(self, code: str, exec_id: str, created_at) -> list:
        """Execute code and return steps"""
        ...


class SimulatorFactory:
    """Factory class to create appropriate simulator based on data type"""
    
    _simulators = {
        "stack": StackSimulator,
        "singlylinkedlist": SinglyLinkedListSimulator,
        "doublylinkedlist": DoublyLinkedListSimulator,
        "binarysearchtree": BinarySearchTreeSimulator,
        "graph": GraphSimulator,
        "undirectedgraph": UndirectedGraphSimulator,
        "directedgraph": DirectedGraphSimulator,
        "queue": QueueSimulator
    }
    
    @classmethod
    def create_simulator(cls, data_type: str) -> DataStructureSimulator:
        """Create and return appropriate simulator"""
        simulator_class = cls._simulators.get(data_type.lower())
        
        if not simulator_class:
            raise NotImplementedError(f"DataType '{data_type}' not supported yet.")
        
        return simulator_class()
    
    @classmethod
    def get_supported_types(cls) -> list[str]:
        """Get list of supported data types"""
        return [key for key, value in cls._simulators.items() if value is not None]
    
    @classmethod
    def is_supported(cls, data_type: str) -> bool:
        """Check if data type is supported"""
        return data_type.lower() in cls._simulators and cls._simulators[data_type.lower()] is not None