"""
Health check endpoint tests
"""
import pytest
from fastapi.testclient import TestClient


class TestHealthCheck:
    """Test health check endpoints"""
    
    def test_health_endpoint(self, client: TestClient):
        """Test basic health check endpoint"""
        response = client.get("/health")
        
        assert response.status_code == 200
        data = response.json()
        
        # Check response structure
        assert "status" in data
        assert "timestamp" in data
        assert "version" in data
        assert "service" in data
        assert "uptime" in data
        
        # Check values
        assert data["status"] == "healthy"
        assert data["service"] == "DSView Backend API"
        assert data["version"] == "0.0.5-alpha"
        assert data["uptime"] == "running"
        
        # Check timestamp format (ISO format)
        assert "T" in data["timestamp"]  # ISO format contains T
        # Note: UTC timezone indicator might be "+00:00" or "Z" depending on system
    
    def test_health_endpoint_methods(self, client: TestClient):
        """Test health endpoint with different HTTP methods"""
        # GET should work
        response = client.get("/health")
        assert response.status_code == 200
        
        # POST should not work (method not allowed)
        response = client.post("/health")
        assert response.status_code == 405
        
        # PUT should not work (method not allowed)
        response = client.put("/health")
        assert response.status_code == 405
    
    def test_health_endpoint_headers(self, client: TestClient):
        """Test health endpoint response headers"""
        response = client.get("/health")
        
        assert response.status_code == 200
        
        # Check content type
        assert response.headers["content-type"] == "application/json"
        
        # Check CORS headers (if configured)
        # Note: CORS headers might not be present in test environment
        # but they should be present in production
    
    def test_health_endpoint_performance(self, client: TestClient):
        """Test health endpoint response time"""
        import time
        
        start_time = time.time()
        response = client.get("/health")
        end_time = time.time()
        
        assert response.status_code == 200
        
        # Health check should be very fast (< 100ms)
        response_time = (end_time - start_time) * 1000  # Convert to milliseconds
        assert response_time < 100, f"Health check took {response_time}ms, expected < 100ms"
    
    def test_health_endpoint_concurrent(self, client: TestClient):
        """Test health endpoint with concurrent requests"""
        import threading
        import time
        
        results = []
        errors = []
        
        def make_request():
            try:
                response = client.get("/health")
                results.append(response.status_code)
            except Exception as e:
                errors.append(str(e))
        
        # Create 10 concurrent requests
        threads = []
        for _ in range(10):
            thread = threading.Thread(target=make_request)
            threads.append(thread)
            thread.start()
        
        # Wait for all threads to complete
        for thread in threads:
            thread.join()
        
        # All requests should succeed
        assert len(errors) == 0, f"Errors occurred: {errors}"
        assert len(results) == 10
        assert all(status == 200 for status in results)
    
    def test_health_endpoint_with_query_params(self, client: TestClient):
        """Test health endpoint with query parameters (should be ignored)"""
        response = client.get("/health?param1=value1&param2=value2")
        
        assert response.status_code == 200
        data = response.json()
        assert data["status"] == "healthy"
    
    def test_health_endpoint_with_headers(self, client: TestClient):
        """Test health endpoint with various headers"""
        headers = {
            "User-Agent": "TestAgent/1.0",
            "Accept": "application/json",
            "X-Test-Header": "test-value"
        }
        
        response = client.get("/health", headers=headers)
        
        assert response.status_code == 200
        data = response.json()
        assert data["status"] == "healthy"
