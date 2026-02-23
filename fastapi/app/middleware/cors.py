"""
CORS middleware configuration.
"""

from fastapi.middleware.cors import CORSMiddleware
from app.core.config import settings


def setup_cors(app):
    """Setup CORS middleware for the FastAPI app"""
    # Get allowed origins from environment variables
    allowed_origins = settings.get_allow_origins_list()
    
    # Add default localhost origins if none specified
    if not allowed_origins:
        allowed_origins = [
            "http://localhost:3000",
            "https://localhost:3000",
        ]
    
    # Add production Vercel domains and local testing domains
    production_origins = [
        "https://dsview-frontend-pj1-dev.vercel.app",
        "https://dsview-frontend-pj1-gvmct2qt9-guests-projects-61264a98.vercel.app",
        "http://dsview.lvh.me",
        "https://dsview.lvh.me"
    ]
    
    # Combine all origins
    all_origins = allowed_origins + production_origins
    
    app.add_middleware(
        CORSMiddleware,
        allow_origins=all_origins,
        # Allow GitHub Codespaces (*.app.github.dev) and Vercel (*.vercel.app)
        allow_origin_regex=r"https://.*(\.app\.github\.dev|\.vercel\.app)",
        allow_credentials=True,
        allow_methods=["GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"],
        allow_headers=["*"],
    )
