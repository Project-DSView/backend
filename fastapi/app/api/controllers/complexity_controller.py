from app.services.llm.ollama_service import OllamaService
from app.services.simulators.operations.complexity_analyzer import ComplexityAnalyzer, ComplexityResult
from app.schemas.complexity import (
    ComplexityAnalysisRequest, 
    ComplexityAnalysisResponse,
    PerformanceAnalysisResponse,
    FunctionComplexitySchema
)

class ComplexityController:
    def __init__(self):
        self.ollama_service = OllamaService()
        self.analyzer = ComplexityAnalyzer()

    def _generate_ast_explanation(self, result: ComplexityResult) -> str:
        """Generate a technical explanation summary from AST result"""
        parts = []
        
        # Structure the technical context clearly for the LLM
        parts.append("Technical Analysis Results (Source of Truth):")
        parts.append(f"Overall Time Complexity: {result.time_complexity}")
        parts.append(f"Overall Space Complexity: {result.space_complexity}")
        
        if result.analysis_details.get("has_recursion"):
            parts.append(f"Recursion Detected: {result.analysis_details.get('recursion_type')}")

        if result.function_complexities:
            parts.append("\nDetected Functions & Their Complexities (Use this list):")
            for fc in result.function_complexities:
                parts.append(f"- Function '{fc.function_name}': Time={fc.time_complexity}, Space={fc.space_complexity}")
                reason = []
                if fc.loop_count > 0: reason.append(f"Contains {fc.loop_count} loop(s)")
                if fc.max_nesting > 0: reason.append(f"Max nesting depth: {fc.max_nesting}")
                if fc.has_recursion: reason.append("Uses recursion")
                
                if reason:
                    parts.append(f"  Reasoning Base: {', '.join(reason)}")
                else:
                    parts.append("  Reasoning Base: Constant time operations only")
            
        return "\n".join(parts)

    async def analyze_with_llm(self, request: ComplexityAnalysisRequest) -> ComplexityAnalysisResponse:
        """
        Analyze code complexity using LLM (Ollama).
        """
        # 1. Run AST Analysis first (Fast & Accurate for Big O label)
        ast_result = self.analyzer.analyze(request.code)
        
        # 2. Prepare context from AST result
        ast_context = self._generate_ast_explanation(ast_result)
        
        # 3. Run LLM Analysis for Explanation (Detailed) with context
        llm_result = await self.ollama_service.analyze_complexity(
            code=request.code, 
            model=request.model,
            language="th",
            ast_context=ast_context
        )
        
        # 3. Merge: Use AST for the short label, LLM for the explanation
        return ComplexityAnalysisResponse(
            complexity=ast_result.time_complexity, # Trust AST for the label e.g. O(n)
            explanation=llm_result.get("explanation", "")
        )

    async def analyze_performance(self, request: ComplexityAnalysisRequest) -> PerformanceAnalysisResponse:
        """
        Analyze performance (time/space) using AST based ComplexityAnalyzer.
        This represents the 'legacy' or 'fast' method.
        """
        # Use the existing analyzer logic
        result = self.analyzer.analyze(request.code)
        
        return PerformanceAnalysisResponse(
            time_complexity=result.time_complexity,
            space_complexity=result.space_complexity,
            time_explanation=result.time_explanation,
            space_explanation=result.space_explanation,
            analysis_details=result.analysis_details,
            function_complexities=[
                FunctionComplexitySchema(
                    function_name=fc.function_name,
                    time_complexity=fc.time_complexity,
                    space_complexity=fc.space_complexity,
                    line_start=fc.line_start,
                    line_end=fc.line_end,
                    time_complexity_rank=fc.time_complexity_rank
                )
                for fc in result.function_complexities
            ]
        )
