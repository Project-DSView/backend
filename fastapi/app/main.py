from fastapi import FastAPI

# Import modules
from .api.routes import api_router
from .api.health import router as health_router
from .api.root import router as root_router
from .core.config import settings
from .core.openapi import get_custom_openapi
from .core.startup import lifespan
from .middleware.cors import setup_cors
from .middleware.rate_limiting import setup_rate_limiting
from .middleware.request_id import add_request_id_middleware
from .middleware.logging import StructuredLoggingMiddleware
from .core.logger import app_logger

# Security (if needed in the future)
# security = HTTPBearer()

# Create FastAPI app
is_production = settings.ENVIRONMENT.lower() == "production"

app = FastAPI(
    title="DSView Backend API",
    description=(
        "Backend API สำหรับระบบ DSView - Stateless Architecture\n"
        "- **Playground**: Code execution with step-by-step visualization (No Auth Required)\n"
        "## Architecture\n"
        "- **Stateless**: No database storage, all data processed in memory\n"
        "- **Playground**: Returns execution steps immediately\n"
    ),
    version="0.0.5-alpha",
    lifespan=lifespan,
    docs_url=None if is_production else "/docs",
    redoc_url=None if is_production else "/redoc",
)

# Setup middleware
setup_cors(app)
setup_rate_limiting(app)
app.middleware("http")(add_request_id_middleware)
app.add_middleware(StructuredLoggingMiddleware)

# Note: Startup/shutdown events are now handled by lifespan parameter

# Custom OpenAPI schema
app.openapi = get_custom_openapi(app)

# Include routers
app.include_router(root_router)
app.include_router(health_router)
app.include_router(api_router, prefix="/api")

# Log application startup
app_logger.info("FastAPI application started - Tracing Enabled", extra={
    "version": "0.0.5-alpha",
    "title": "DSView Backend API"
})
