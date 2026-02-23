import os
from pydantic_settings import BaseSettings, SettingsConfigDict
from dotenv import load_dotenv

ENV_FILE = os.getenv("ENV_FILE", ".env")  
load_dotenv(ENV_FILE)

class Settings(BaseSettings):
    model_config = SettingsConfigDict(env_file=ENV_FILE)
    
    # API
    API_KEY_NAME: str
    API_KEY: str
    ENVIRONMENT: str = "development"

    # CORS
    ALLOW_ORIGINS: str | None = None
    
    # Logging
    LOG_LEVEL: str 
    
    # Execution Settings
    MAX_CODE_LENGTH: int
    EXECUTION_TIMEOUT: int
    MAX_LOOPS: int
    MAX_FOR_LOOPS: int 
    MAX_FUNCTIONS: int 

    # Rate Limiting
    RATE_LIMIT_PER_MINUTE: int 
    RATE_LIMIT_PER_SECOND: int

    # Ollama
    OLLAMA_BASE_URL: str = "http://localhost:11434"
    OLLAMA_HOST: str | None = None # Handle potential extra env var

    def get_allow_origins_list(self) -> list[str]:
        value = self.ALLOW_ORIGINS
        if not value:
            return []
        # Support comma-separated values in env
        return [v.strip() for v in value.split(",") if v.strip()]

settings = Settings()
