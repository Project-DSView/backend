from fastapi import APIRouter, HTTPException, Body, Request, Depends
from fastapi.responses import Response
from slowapi import Limiter
from slowapi.util import get_remote_address
import logging
import traceback
from asyncio import TimeoutError

from app.api.controllers.playground import ExecuteController
from app.schemas.playground import ExecutionCreateRequest, ExecutionResponse
from app.core.security import verify_api_key
from app.core.config import settings

from app.examples.exercises import single_linklist, double_linklist, stack, binary_search_tree, undirected_graph, directed_graph, queue

router = APIRouter()
logger = logging.getLogger(__name__)
limiter = Limiter(key_func=get_remote_address)

@router.post("/run", response_model=ExecutionResponse, dependencies=[Depends(verify_api_key)])
@limiter.limit(f"{settings.RATE_LIMIT_PER_MINUTE}/minute")  
async def run_code_guest(
    request: Request,
    request_body: ExecutionCreateRequest = Body(
        ...,
        openapi_examples={
            "singlylinkedlist": {
                "summary": "Singly Linked List Example",
                "description": "Example code for singly linked list operations",
                "value": {
                    "code": single_linklist.example,
                    "dataType": "singlylinkedlist"
                }
            },
            "doublylinkedlist": {
                "summary": "Doubly Linked List Example",
                "description": "Example code for doubly linked list operations with forward and reverse traversal",
                "value": {
                    "code": double_linklist.example,
                    "dataType": "doublylinkedlist"
                }
            },
            "stack": {
                "summary": "Stack Example", 
                "description": "Example code for stack operations",
                "value": {
                    "code": stack.example,
                    "dataType": "stack"
                }
            },
            "binarysearchtree": {
                "summary": "Binary Search Tree Example",
                "description": "Example code for binary search tree operations",
                "value": {
                    "code": binary_search_tree.example,
                    "dataType": "binarysearchtree"
                }
            },
            "undirectedgraph": {
                "summary": "Undirected Graph Example",
                "description": "Example code for undirected graph operations with cycle detection and connectivity",
                "value": {
                    "code": undirected_graph.example,
                    "dataType": "undirectedgraph"
                }
            },
            "directedgraph": {
                "summary": "Directed Graph Example",
                "description": "Example code for directed graph operations with topological sort and cycle detection",
                "value": {
                    "code": directed_graph.example,
                    "dataType": "directedgraph"
                }
            },
            "queue": {
                "summary": "Queue Example",
                "description": "Example code for queue operations (FIFO - First In First Out)",
                "value": {
                    "code": queue.example,
                    "dataType": "queue"
                }
            },
        }
    )
):
    """Execute code for guest users (no authentication, no database storage)"""
    try:
        logger.info(f"Guest code execution - DataType: {request_body.dataType}")
        
        # Validate request
        if not request_body.code or not request_body.code.strip():
            raise HTTPException(status_code=400, detail="Code cannot be empty")
            
        if not request_body.dataType:
            raise HTTPException(status_code=400, detail="DataType is required")
            
        controller = ExecuteController()
        result = await controller.run_code_guest(request_body)
        
        logger.info(f"Guest code executed successfully: execution_id={result.executionId}")
        return result
        
    except ValueError as e:
        logger.warning(f"Validation error: {e}")
        raise HTTPException(status_code=400, detail="Invalid input provided")
    except NotImplementedError as e:
        logger.warning(f"Not implemented error: {e}")
        raise HTTPException(status_code=501, detail="Feature not implemented")
    except TimeoutError as e:
        logger.warning(f"Timeout error: {e}")
        raise HTTPException(status_code=408, detail="Request timeout")
    except HTTPException as e:
        logger.warning(f"HTTP exception: {e.status_code} - {e.detail}")
        raise e
    except Exception as e:
        tb = traceback.format_exc()
        logger.error(f"Unexpected error in run_code_guest: {e}")
        logger.error(f"Full traceback: {tb}")
        
        # Don't expose internal error details to client
        raise HTTPException(
            status_code=500, 
            detail="Internal server error occurred"
        )

