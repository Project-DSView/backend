import pytest
from unittest.mock import AsyncMock, patch
from fastapi.testclient import TestClient
from app.main import app
from app.api.controllers.complexity_controller import ComplexityController

client = TestClient(app)

@pytest.fixture
def mock_ollama_service():
    with patch("app.api.controllers.complexity_controller.OllamaService") as MockService:
        mock_instance = MockService.return_value
        # Set up async mock for analyze_complexity
        mock_instance.analyze_complexity = AsyncMock(return_value={
            "complexity": "O(n)",
            "explanation": "Mock explanation"
        })
        yield mock_instance

def test_analyze_performance_endpoint():
    request_data = {
        "code": "def foo(n): return n"
    }
    response = client.post("/api/complexity/performance", json=request_data)
    assert response.status_code == 200
    data = response.json()
    assert "time_complexity" in data
    assert "space_complexity" in data
    assert "function_complexities" in data

def test_analyze_llm_endpoint(mock_ollama_service):
    request_data = {
        "code": "def foo(n): return n",
        "model": "mistral"
    }
    # We need to mock the controller's ollama service specifically or patch it where it's used
    with patch.object(ComplexityController, 'analyze_with_llm', new_callable=AsyncMock) as mock_analyze:
        mock_analyze.return_value = {
            "complexity": "O(n)",
            "explanation": "Mock explanation"
        }
        
        response = client.post("/api/complexity/llm", json=request_data)
        
        # Note: Since the real app uses dependency injection or instantiation, testing FastAPIs with deep mocks can be tricky.
        # But here we are just testing that the route invokes the controller.
        # If the above patch doesn't propagate to the route handler's instance, we might need a different strategy.
        # However, for now let's try this standard approach.
        
        # Actually, let's look at the implementation:
        # router.py instantiates controller = ComplexityController() at module level.
        # So mocking app.api.routes.complexity.controller might be better.
        pass

@patch("app.api.routes.complexity.controller")
def test_analyze_llm_route_mocked_controller(mock_controller):
    # Mock the analyze_with_llm method of the controller instance used by the router
    mock_controller.analyze_with_llm = AsyncMock(return_value={
        "complexity": "O(log n)",
        "explanation": "Logarithmic time"
    })
    
    request_data = {
        "code": "for i in range(10): pass",
        "model": "mistral"
    }
    response = client.post("/api/complexity/llm", json=request_data)
    assert response.status_code == 200
    data = response.json()
    assert data["complexity"] == "O(log n)"
    assert data["explanation"] == "Logarithmic time"

