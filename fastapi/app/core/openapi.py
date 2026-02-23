"""
OpenAPI schema configuration.
"""
from fastapi.openapi.utils import get_openapi


def custom_openapi(app):
    """Create custom OpenAPI schema with proper security"""
    if app.openapi_schema:
        return app.openapi_schema
    
    openapi_schema = get_openapi(
        title=app.title,
        version=app.version,
        description=app.description,
        routes=app.routes,
    )
    
    # Add API key security scheme
    openapi_schema["components"]["securitySchemes"] = {
        "ApiKeyAuth": {
            "type": "apiKey",
            "in": "header",
            "name": "dsview-api-key",
            "description": "API Key for authentication"
        }
    }
    
    # Add security requirement to all endpoints
    openapi_schema["security"] = [{"ApiKeyAuth": []}]
    
    app.openapi_schema = openapi_schema
    return app.openapi_schema


def get_custom_openapi(app):
    """Factory function that returns a callable without arguments"""
    def openapi():
        return custom_openapi(app)
    return openapi