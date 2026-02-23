import httpx
from app.core.config import settings
from app.core.logger import app_logger

class OllamaService:
    def __init__(self, base_url: str = "http://ollama:11434"):
        # Use config if available, otherwise default
        self.base_url = getattr(settings, "OLLAMA_BASE_URL", base_url)

    async def analyze_complexity(self, code: str, model: str = "qwen2.5-coder:1.5b", language: str = "th", ast_context: str = None) -> dict:
        """
        Analyze the Time and Space Complexity (Big O) using Ollama.
        Uses AST context for faster analysis when available.
        """
        # If we have AST context, use it directly (faster, less tokens)
        # Otherwise fall back to analyzing code
        analysis_data = ast_context if ast_context else f"Code:\n{code}"

        if language == "th":
            prompt = f"""คุณคือผู้เชี่ยวชาญด้าน Algorithm ตอบเป็นภาษาไทยเท่านั้น (ห้ามใช้ภาษาอังกฤษในคำอธิบาย)

ข้อมูล:
{analysis_data}

ตอบตาม format นี้ (ภาษาไทยทั้งหมด):

### รายฟังก์ชัน
* **ชื่อฟังก์ชัน**: O(...) - มีลูป 1 ชั้น / ทำงานคงที่ / มีลูปซ้อน 2 ชั้น
* **ชื่อฟังก์ชัน**: O(...) - เหตุผลสั้นๆ

### สรุป
**Time**: O(...) | **Space**: O(...)
**คำอธิบาย**: [1-2 ประโยค อธิบายว่าทำไมถึงได้ค่านี้]"""
        else:
            prompt = f"""
            Analyze the Time and Space Complexity (Big O) of the following Python code.
            Provide a detailed explanation of why it has that complexity.
            
            {analysis_data}
            
            Code:
            ```python
            {code}
            ```
            
            Format your response exactly as follows (Markdown is supported):
            
            **Time Complexity**: O(...)
            **Space Complexity**: O(...)
            
            **Explanation**:
            [Detailed explanation here]
            """

        try:
            async with httpx.AsyncClient() as client:
                response = await client.post(
                    f"{self.base_url}/api/generate",
                    json={
                        "model": model,
                        "prompt": prompt,
                        "stream": False,
                        "options": {
                            "num_ctx": 2048,      # Smaller context window
                            "num_predict": 512,   # Limit output length
                            "temperature": 0.3    # Less randomness, faster
                        }
                    },
                    timeout=240.0  # Reduced timeout since it should be faster
                )
                response.raise_for_status()
                result = response.json()
                response_text = result.get("response", "")

                return {
                    "complexity": "See explanation",
                    "explanation": response_text
                }
        except Exception as e:
            app_logger.error(f"Failed to call Ollama: {str(e)}")
            return {
                "complexity": "Error",
                "explanation": f"Failed to analyze complexity via LLM: {str(e)}"
            }