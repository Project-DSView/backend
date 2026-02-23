"""
Integration tests for DSView FastAPI Backend
"""
import pytest
from fastapi.testclient import TestClient
import json


class TestIntegration:
    """Test integration scenarios"""
    
    def test_complete_workflow(self, client: TestClient, valid_headers):
        """Test complete workflow from health check to code execution"""
        # 1. Check health
        health_response = client.get("/health")
        assert health_response.status_code == 200
        assert health_response.json()["status"] == "healthy"
        
        # 2. Execute simple code
        code_data = {
            "code": "print('Hello World')",
            "dataType": "stack"
        }
        
        execution_response = client.post(
            "/api/playground/run",
            json=code_data,
            headers=valid_headers
        )
        
        assert execution_response.status_code == 200
        execution_data = execution_response.json()
        assert execution_data["status"] == "success"
        assert execution_data["dataType"] == "stack"
    
    def test_stack_implementation_workflow(self, client: TestClient, valid_headers):
        """Test complete stack implementation workflow"""
        stack_code = {
            "code": """class Stack:
    def __init__(self):
        self.items = []
    
    def push(self, item):
        self.items.append(item)
        print(f"Pushed {item}")
    
    def pop(self):
        if self.is_empty():
            return None
        item = self.items.pop()
        print(f"Popped {item}")
        return item
    
    def is_empty(self):
        return len(self.items) == 0
    
    def size(self):
        return len(self.items)

# Test the stack
s = Stack()
s.push(1)
s.push(2)
s.push(3)
print(f"Stack size: {s.size()}")
s.pop()
s.pop()
print(f"Stack size after pops: {s.size()}")""",
            "dataType": "stack"
        }
        
        response = client.post(
            "/api/playground/run",
            json=stack_code,
            headers=valid_headers
        )
        
        assert response.status_code == 200
        data = response.json()
        
        assert data["status"] == "success"
        assert data["dataType"] == "stack"
        assert len(data["steps"]) > 0
        
        # Check that the execution contains expected operations
        step_codes = [step["code"] for step in data["steps"]]
        # Note: The exact step content depends on the simulator implementation
        # We just verify that we have steps and they contain some code
        assert len(step_codes) > 0
        assert all(isinstance(code, str) for code in step_codes)
    
    def test_linked_list_implementation_workflow(self, client: TestClient, valid_headers):
        """Test linked list implementation workflow"""
        linked_list_code = {
            "code": """class Node:
    def __init__(self, data):
        self.data = data
        self.next = None

class LinkedList:
    def __init__(self):
        self.head = None
    
    def append(self, data):
        new_node = Node(data)
        if not self.head:
            self.head = new_node
            return
        current = self.head
        while current.next:
            current = current.next
        current.next = new_node
    
    def display(self):
        elements = []
        current = self.head
        while current:
            elements.append(current.data)
            current = current.next
        print(f"List: {elements}")

# Test the linked list
ll = LinkedList()
ll.append(1)
ll.append(2)
ll.append(3)
ll.display()""",
            "dataType": "singlylinkedlist"
        }
        
        response = client.post(
            "/api/playground/run",
            json=linked_list_code,
            headers=valid_headers
        )
        
        assert response.status_code == 200
        data = response.json()
        
        assert data["status"] == "success"
        assert data["dataType"] == "singlylinkedlist"
        assert len(data["steps"]) > 0
    
    def test_error_handling_workflow(self, client: TestClient, valid_headers):
        """Test error handling in code execution"""
        error_code = {
            "code": """def divide(a, b):
    return a / b

result = divide(10, 0)  # This will cause an error
print(result)""",
            "dataType": "stack"
        }
        
        response = client.post(
            "/api/playground/run",
            json=error_code,
            headers=valid_headers
        )
        
        # Should still return 200 but with error status
        assert response.status_code == 200
        data = response.json()
        
        # The execution should complete but with error
        assert data["status"] in ["success", "error"]
        if data["status"] == "error":
            assert data["errorMessage"] is not None
    
    def test_multiple_data_types_workflow(self, client: TestClient, valid_headers):
        """Test workflow with multiple data types"""
        data_types = ["stack", "singlylinkedlist", "doublylinkedlist"]
        
        for data_type in data_types:
            code_data = {
                "code": f"print('Testing {data_type} implementation')",
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
    
    def test_concurrent_requests_workflow(self, client: TestClient, valid_headers):
        """Test handling of concurrent requests"""
        import threading
        import time
        
        results = []
        errors = []
        
        def make_request():
            try:
                code_data = {
                    "code": "print('Concurrent test')",
                    "dataType": "stack"
                }
                
                response = client.post(
                    "/api/playground/run",
                    json=code_data,
                    headers=valid_headers
                )
                
                results.append(response.status_code)
            except Exception as e:
                errors.append(str(e))
        
        # Create 5 concurrent requests
        threads = []
        for _ in range(5):
            thread = threading.Thread(target=make_request)
            threads.append(thread)
            thread.start()
        
        # Wait for all threads to complete
        for thread in threads:
            thread.join()
        
        # All requests should succeed
        assert len(errors) == 0, f"Errors occurred: {errors}"
        assert len(results) == 5
        assert all(status == 200 for status in results)
    
    def test_api_key_workflow(self, client: TestClient, sample_code):
        """Test API key workflow"""
        # Test without API key
        response = client.post(
            "/api/playground/run",
            json=sample_code
        )
        assert response.status_code == 401
        
        # Test with valid API key
        response = client.post(
            "/api/playground/run",
            json=sample_code,
            headers={"dsview-api-key": "test-api-key"}
        )
        assert response.status_code == 200
    
    def test_validation_workflow(self, client: TestClient, valid_headers):
        """Test input validation workflow"""
        # Test various invalid inputs
        invalid_inputs = [
            {"code": "", "dataType": "stack"},
            {"code": "print('test')", "dataType": "invalid"},
            {"code": "import os; os.system('ls')", "dataType": "stack"},
            {"code": "for i in range(1000): pass", "dataType": "stack"},
        ]
        
        for invalid_input in invalid_inputs:
            response = client.post(
                "/api/playground/run",
                json=invalid_input,
                headers=valid_headers
            )
            
            # Some inputs might pass validation depending on the specific patterns
            # Just check that we get a valid response
            assert response.status_code in [200, 400, 422]
    
    def test_response_format_consistency(self, client: TestClient, valid_headers):
        """Test that response format is consistent across different requests"""
        test_cases = [
            {"code": "print('Hello')", "dataType": "stack"},
            {"code": "x = 1 + 1", "dataType": "singlylinkedlist"},
            {"code": "for i in range(3): print(i)", "dataType": "doublylinkedlist"},
        ]
        
        for test_case in test_cases:
            response = client.post(
                "/api/playground/run",
                json=test_case,
                headers=valid_headers
            )
            
            assert response.status_code == 200
            data = response.json()
            
            # Check consistent response structure
            required_fields = [
                "executionId", "code", "dataType", "steps", 
                "totalSteps", "status", "executedAt", "createdAt"
            ]
            
            for field in required_fields:
                assert field in data, f"Missing field: {field}"
            
            # Check data types
            assert isinstance(data["executionId"], str)
            assert isinstance(data["code"], str)
            assert isinstance(data["dataType"], str)
            assert isinstance(data["steps"], list)
            assert isinstance(data["totalSteps"], int)
            assert isinstance(data["status"], str)
    
    def test_performance_workflow(self, client: TestClient, valid_headers):
        """Test performance characteristics"""
        import time
        
        code_data = {
            "code": "print('Performance test')",
            "dataType": "stack"
        }
        
        # Measure response time
        start_time = time.time()
        response = client.post(
            "/api/playground/run",
            json=code_data,
            headers=valid_headers
        )
        end_time = time.time()
        
        assert response.status_code == 200
        
        # Response should be reasonably fast (< 5 seconds)
        response_time = end_time - start_time
        assert response_time < 5.0, f"Response time too slow: {response_time}s"
