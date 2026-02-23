"""
Pytest configuration and fixtures for DSView FastAPI tests
"""
import pytest
from fastapi.testclient import TestClient
from unittest.mock import patch
import os

# Set environment variables before importing the app
os.environ["API_KEY"] = "test-api-key"
os.environ["API_KEY_NAME"] = "dsview-api-key"
os.environ["LOG_LEVEL"] = "DEBUG"
os.environ["MAX_CODE_LENGTH"] = "1000"
os.environ["EXECUTION_TIMEOUT"] = "10"
os.environ["MAX_LOOPS"] = "5"
os.environ["MAX_FOR_LOOPS"] = "10"
os.environ["MAX_FUNCTIONS"] = "5"
os.environ["RATE_LIMIT_PER_MINUTE"] = "100"
os.environ["RATE_LIMIT_PER_SECOND"] = "10"

# Import app after setting environment variables
from app.main import app

@pytest.fixture
def client():
    """Create test client"""
    return TestClient(app)

@pytest.fixture
def valid_api_key():
    """Valid API key for testing"""
    return "test-api-key"

@pytest.fixture
def invalid_api_key():
    """Invalid API key for testing"""
    return "invalid-api-key"

@pytest.fixture
def valid_headers(valid_api_key):
    """Valid headers with API key"""
    return {"dsview-api-key": valid_api_key}

@pytest.fixture
def invalid_headers(invalid_api_key):
    """Invalid headers with wrong API key"""
    return {"dsview-api-key": invalid_api_key}

@pytest.fixture
def sample_code():
    """Sample code for testing"""
    return {
        "code": "print('Hello World')",
        "dataType": "stack"
    }

@pytest.fixture
def stack_code():
    """Stack implementation code for testing"""
    return {
        "code": """class Stack:
    def __init__(self):
        self.items = []
    
    def push(self, item):
        self.items.append(item)
    
    def pop(self):
        return self.items.pop() if self.items else None

s = Stack()
s.push(1)
s.push(2)
print(s.pop())""",
        "dataType": "stack"
    }

@pytest.fixture
def dangerous_code():
    """Dangerous code that should be rejected"""
    return {
        "code": "import os; os.system('rm -rf /')",
        "dataType": "stack"
    }

@pytest.fixture
def empty_code():
    """Empty code that should be rejected"""
    return {
        "code": "",
        "dataType": "stack"
    }

@pytest.fixture
def syntax_error_code_missing_parenthesis():
    """Code with syntax error - missing closing parenthesis"""
    return {
        "code": """class DataNode:
    def __init__(self, name):
        self.name = name
        self.next = None

class SinglyLinkedList:
    def __init__(self):
        self.count = 0
        self.head = None

mylist = SinglyLinkedList()
mylist.traverse(""",
        "dataType": "singlylinkedlist"
    }

@pytest.fixture
def syntax_error_code_missing_colon():
    """Code with syntax error - missing colon"""
    return {
        "code": """class DataNode
    def __init__(self, name):
        self.name = name""",
        "dataType": "singlylinkedlist"
    }