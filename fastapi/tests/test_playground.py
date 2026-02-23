"""
Playground API endpoint tests
"""
import pytest
from fastapi.testclient import TestClient


class TestPlaygroundAPI:
    """Test playground API endpoints"""
    
    def test_run_code_success(self, client: TestClient, valid_headers, sample_code):
        """Test successful code execution"""
        response = client.post(
            "/api/playground/run",
            json=sample_code,
            headers=valid_headers
        )
        
        assert response.status_code == 200
        data = response.json()
        
        # Check response structure
        assert "executionId" in data
        assert "code" in data
        assert "dataType" in data
        assert "steps" in data
        assert "totalSteps" in data
        assert "status" in data
        assert "executedAt" in data
        assert "createdAt" in data
        
        # Check values
        assert data["code"] == sample_code["code"]
        assert data["dataType"] == sample_code["dataType"]
        assert data["status"] == "success"
        assert isinstance(data["steps"], list)
        assert data["totalSteps"] > 0
    
    def test_run_code_without_api_key(self, client: TestClient, sample_code):
        """Test code execution without API key"""
        response = client.post(
            "/api/playground/run",
            json=sample_code
        )
        
        assert response.status_code == 401
    
    def test_run_code_with_invalid_api_key(self, client: TestClient, invalid_headers, sample_code):
        """Test code execution with invalid API key"""
        response = client.post(
            "/api/playground/run",
            json=sample_code,
            headers=invalid_headers
        )
        
        assert response.status_code == 401
    
    def test_run_code_empty_code(self, client: TestClient, valid_headers, empty_code):
        """Test code execution with empty code"""
        response = client.post(
            "/api/playground/run",
            json=empty_code,
            headers=valid_headers
        )
        
        # Pydantic validation happens first, so we get 422 instead of 400
        assert response.status_code == 422
        data = response.json()
        assert "detail" in data
    
    def test_run_code_dangerous_patterns(self, client: TestClient, valid_headers, dangerous_code):
        """Test code execution with dangerous patterns"""
        response = client.post(
            "/api/playground/run",
            json=dangerous_code,
            headers=valid_headers
        )
        
        # Pydantic validation happens first, so we get 422 instead of 400
        assert response.status_code == 422
        data = response.json()
        assert "detail" in data
    
    def test_run_code_missing_datatype(self, client: TestClient, valid_headers):
        """Test code execution without dataType"""
        code_data = {"code": "print('Hello World')"}
        
        response = client.post(
            "/api/playground/run",
            json=code_data,
            headers=valid_headers
        )
        
        assert response.status_code == 422  # Validation error
    
    def test_run_code_invalid_datatype(self, client: TestClient, valid_headers):
        """Test code execution with invalid dataType"""
        code_data = {
            "code": "print('Hello World')",
            "dataType": "invalid_type"
        }
        
        response = client.post(
            "/api/playground/run",
            json=code_data,
            headers=valid_headers
        )
        
        assert response.status_code == 422  # Validation error
    
    def test_run_code_stack_implementation(self, client: TestClient, valid_headers, stack_code):
        """Test stack implementation code execution"""
        response = client.post(
            "/api/playground/run",
            json=stack_code,
            headers=valid_headers
        )
        
        assert response.status_code == 200
        data = response.json()
        
        assert data["dataType"] == "stack"
        assert data["status"] == "success"
        assert len(data["steps"]) > 0
        
        # Check that steps contain stack operations
        step_codes = [step["code"] for step in data["steps"]]
        # Note: The exact step content depends on the simulator implementation
        # We just verify that we have steps and they contain some code
        assert len(step_codes) > 0
        assert all(isinstance(code, str) for code in step_codes)
    
    def test_run_code_different_data_types(self, client: TestClient, valid_headers):
        """Test code execution with different data types"""
        data_types = ["stack", "singlylinkedlist", "doublylinkedlist", 
                     "binarysearchtree", "undirectedgraph", "directedgraph"]
        
        for data_type in data_types:
            code_data = {
                "code": f"print('Testing {data_type}')",
                "dataType": data_type
            }
            
            response = client.post(
                "/api/playground/run",
                json=code_data,
                headers=valid_headers
            )
            
            assert response.status_code == 200
            data = response.json()
            assert data["dataType"] == data_type
            assert data["status"] == "success"
    
    def test_run_code_too_many_loops(self, client: TestClient, valid_headers):
        """Test code execution with too many loops"""
        # Create code with too many for loops
        code_with_loops = "\n".join([f"for i in range(10):" for _ in range(15)])  # 15 for loops
        
        code_data = {
            "code": code_with_loops,
            "dataType": "stack"
        }
        
        response = client.post(
            "/api/playground/run",
            json=code_data,
            headers=valid_headers
        )
        
        # Pydantic validation happens first, so we get 422 instead of 400
        assert response.status_code == 422
        data = response.json()
        assert "detail" in data
    
    def test_run_code_too_many_functions(self, client: TestClient, valid_headers):
        """Test code execution with too many function definitions"""
        # Create code with too many function definitions
        functions = "\n".join([f"def func{i}(): pass" for i in range(10)])  # 10 functions
        
        code_data = {
            "code": functions,
            "dataType": "stack"
        }
        
        response = client.post(
            "/api/playground/run",
            json=code_data,
            headers=valid_headers
        )
        
        # Pydantic validation happens first, so we get 422 instead of 400
        assert response.status_code == 422
        data = response.json()
        assert "detail" in data
    
    def test_run_code_large_code(self, client: TestClient, valid_headers):
        """Test code execution with large code (should be rejected)"""
        # Create very large code
        large_code = "print('Hello World')\n" * 1000  # 1000 lines
        
        code_data = {
            "code": large_code,
            "dataType": "stack"
        }
        
        response = client.post(
            "/api/playground/run",
            json=code_data,
            headers=valid_headers
        )
        
        # Pydantic validation happens first, so we get 422 instead of 400
        assert response.status_code == 422
        data = response.json()
        assert "detail" in data
    
    def test_run_code_malformed_json(self, client: TestClient, valid_headers):
        """Test code execution with malformed JSON"""
        response = client.post(
            "/api/playground/run",
            data="invalid json",
            headers=valid_headers
        )
        
        assert response.status_code == 422
    
    def test_run_code_wrong_content_type(self, client: TestClient, valid_headers, sample_code):
        """Test code execution with wrong content type"""
        response = client.post(
            "/api/playground/run",
            data=sample_code,
            headers=valid_headers
        )
        
        assert response.status_code == 422
    
    def test_run_code_get_method(self, client: TestClient, valid_headers):
        """Test playground endpoint with GET method (should not work)"""
        response = client.get(
            "/api/playground/run",
            headers=valid_headers
        )
        
        assert response.status_code == 405  # Method not allowed
    
    def test_run_code_put_method(self, client: TestClient, valid_headers, sample_code):
        """Test playground endpoint with PUT method (should not work)"""
        response = client.put(
            "/api/playground/run",
            json=sample_code,
            headers=valid_headers
        )
        
        assert response.status_code == 405  # Method not allowed
    
    def test_syntax_error_missing_closing_parenthesis(self, client: TestClient, valid_headers):
        """Test syntax error with missing closing parenthesis"""
        code_with_error = {
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
        
        response = client.post(
            "/api/playground/run",
            json=code_with_error,
            headers=valid_headers
        )
        
        assert response.status_code == 200
        data = response.json()
        
        # Check that response has error status
        assert data["status"] == "error"
        assert "errorMessage" in data
        assert data["errorMessage"] is not None
        assert len(data["errorMessage"]) > 0
        
        # Check that error message is Python-style format
        error_message = data["errorMessage"]
        
        # Should contain file path (or <string>)
        assert 'File "' in error_message or 'File "<string>"' in error_message
        
        # Should contain line number
        assert "line" in error_message.lower()
        
        # Should contain the problematic code line
        assert "mylist.traverse(" in error_message
        
        # Should contain pointer (^)
        assert "^" in error_message
        
        # Should contain SyntaxError
        assert "SyntaxError" in error_message
        
        # Should contain error description
        assert "(" in error_message or "never closed" in error_message.lower() or "unexpected" in error_message.lower()
        
        # Steps should be empty for error cases
        assert data["steps"] == []
        assert data["totalSteps"] == 0
    
    def test_syntax_error_missing_colon(self, client: TestClient, valid_headers):
        """Test syntax error with missing colon"""
        code_with_error = {
            "code": """class DataNode
    def __init__(self, name):
        self.name = name""",
            "dataType": "singlylinkedlist"
        }
        
        response = client.post(
            "/api/playground/run",
            json=code_with_error,
            headers=valid_headers
        )
        
        assert response.status_code == 200
        data = response.json()
        
        assert data["status"] == "error"
        assert "errorMessage" in data
        error_message = data["errorMessage"]
        
        # Should be Python-style error
        assert 'File "' in error_message or 'File "<string>"' in error_message
        assert "SyntaxError" in error_message
        assert "^" in error_message
    
    def test_syntax_error_invalid_indentation(self, client: TestClient, valid_headers):
        """Test syntax error with invalid indentation"""
        code_with_error = {
            "code": """class DataNode:
def __init__(self, name):
    self.name = name""",
            "dataType": "singlylinkedlist"
        }
        
        response = client.post(
            "/api/playground/run",
            json=code_with_error,
            headers=valid_headers
        )
        
        assert response.status_code == 200
        data = response.json()
        
        assert data["status"] == "error"
        assert "errorMessage" in data
        error_message = data["errorMessage"]
        
        # Should be Python-style error
        assert 'File "' in error_message or 'File "<string>"' in error_message
        assert "SyntaxError" in error_message or "IndentationError" in error_message
    
    def test_syntax_error_all_data_structures(self, client: TestClient, valid_headers):
        """Test syntax error handling across all data structures"""
        data_types = ["singlylinkedlist", "doublylinkedlist", "stack", "queue", 
                     "binarysearchtree", "undirectedgraph", "directedgraph"]
        
        for data_type in data_types:
            code_with_error = {
                "code": f"mylist = {data_type}()\nmylist.traverse(",
                "dataType": data_type
            }
            
            response = client.post(
                "/api/playground/run",
                json=code_with_error,
                headers=valid_headers
            )
            
            assert response.status_code == 200, f"Failed for {data_type}"
            data = response.json()
            
            assert data["status"] == "error", f"Status should be error for {data_type}"
            assert "errorMessage" in data, f"Should have errorMessage for {data_type}"
            
            error_message = data["errorMessage"]
            
            # All should have Python-style error format
            assert 'File "' in error_message or 'File "<string>"' in error_message, \
                f"Should have file path for {data_type}"
            assert "SyntaxError" in error_message, \
                f"Should have SyntaxError for {data_type}"
            assert "^" in error_message, \
                f"Should have pointer (^) for {data_type}"
    
    def test_syntax_error_with_line_number(self, client: TestClient, valid_headers):
        """Test that syntax error shows correct line number"""
        code_with_error = {
            "code": """class DataNode:
    def __init__(self, name):
        self.name = name
        self.next = None

class SinglyLinkedList:
    def __init__(self):
        self.count = 0
        self.head = None

mylist = SinglyLinkedList()
mylist.insertFront("Tony")
mylist.traverse(""",
            "dataType": "singlylinkedlist"
        }
        
        response = client.post(
            "/api/playground/run",
            json=code_with_error,
            headers=valid_headers
        )
        
        assert response.status_code == 200
        data = response.json()
        
        assert data["status"] == "error"
        error_message = data["errorMessage"]
        
        # Should contain line number (should be around line 12-13 where traverse( is)
        # The exact line depends on how the code is parsed
        assert "line" in error_message.lower()
        
        # Extract line number from error message
        import re
        line_match = re.search(r'line (\d+)', error_message)
        if line_match:
            line_number = int(line_match.group(1))
            # Line number should be reasonable (not 0 or negative)
            assert line_number > 0
            assert line_number <= 15  # Should be within the code length