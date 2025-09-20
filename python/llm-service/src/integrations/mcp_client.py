"""
MCP (Model Context Protocol) Client for Multi-Agent LLM Service
Handles tool integration and execution through MCP protocol
"""

import asyncio
import logging
import json
import time
from typing import Dict, List, Optional, Any, Callable
from dataclasses import dataclass, field
import uuid

import websockets
from websockets.exceptions import ConnectionClosed, WebSocketException
import jsonrpc_base
from jsonrpc_websocket import Server

from src.config import MCPConfig
from src.core.models import MCPToolRequest, MCPToolResponse, Tool, ToolFunction

logger = logging.getLogger(__name__)

@dataclass
class MCPTool:
    """MCP tool definition"""
    name: str
    description: str
    parameters: Dict[str, Any]
    handler: Optional[str] = None
    timeout: int = 30
    enabled: bool = True
    metadata: Dict[str, Any] = field(default_factory=dict)

@dataclass
class MCPSession:
    """MCP session information"""
    session_id: str
    user_id: str
    websocket: Optional[websockets.WebSocketServerProtocol] = None
    tools: Dict[str, MCPTool] = field(default_factory=dict)
    created_at: float = field(default_factory=time.time)
    last_activity: float = field(default_factory=time.time)
    is_active: bool = True

class MCPClient:
    """MCP client for tool integration and execution"""
    
    def __init__(self, config: MCPConfig):
        self.config = config
        self.sessions: Dict[str, MCPSession] = {}
        self.available_tools: Dict[str, MCPTool] = {}
        self.tool_registry: Dict[str, Callable] = {}
        self.server: Optional[Server] = None
        self._lock = asyncio.Lock()
        
        # Built-in tools
        self._register_builtin_tools()
        
    async def initialize(self):
        """Initialize MCP client"""
        if not self.config.enabled:
            logger.info("MCP client disabled")
            return
        
        logger.info("Initializing MCP client")
        
        # Start MCP server
        await self._start_mcp_server()
        
        # Load external tools
        await self._load_external_tools()
        
        # Start background tasks
        asyncio.create_task(self._session_cleanup_task())
        asyncio.create_task(self._health_check_task())
        
        logger.info(f"MCP client initialized with {len(self.available_tools)} tools")
    
    def _register_builtin_tools(self):
        """Register built-in tools"""
        
        # Web search tool
        self.available_tools["web_search"] = MCPTool(
            name="web_search",
            description="Search the web for information",
            parameters={
                "type": "object",
                "properties": {
                    "query": {
                        "type": "string",
                        "description": "Search query"
                    },
                    "max_results": {
                        "type": "integer",
                        "description": "Maximum number of results",
                        "default": 5
                    }
                },
                "required": ["query"]
            },
            handler="builtin_web_search"
        )
        
        # Code executor tool
        self.available_tools["code_executor"] = MCPTool(
            name="code_executor",
            description="Execute code in a secure sandbox",
            parameters={
                "type": "object",
                "properties": {
                    "language": {
                        "type": "string",
                        "enum": ["python", "javascript", "bash"],
                        "description": "Programming language"
                    },
                    "code": {
                        "type": "string",
                        "description": "Code to execute"
                    },
                    "timeout": {
                        "type": "integer",
                        "description": "Execution timeout in seconds",
                        "default": 30
                    }
                },
                "required": ["language", "code"]
            },
            handler="builtin_code_executor"
        )
        
        # File operations tool
        self.available_tools["file_operations"] = MCPTool(
            name="file_operations",
            description="Perform file operations (read, write, list)",
            parameters={
                "type": "object",
                "properties": {
                    "operation": {
                        "type": "string",
                        "enum": ["read", "write", "list", "delete"],
                        "description": "File operation to perform"
                    },
                    "path": {
                        "type": "string",
                        "description": "File or directory path"
                    },
                    "content": {
                        "type": "string",
                        "description": "Content to write (for write operation)"
                    }
                },
                "required": ["operation", "path"]
            },
            handler="builtin_file_operations"
        )
        
        # HTTP request tool
        self.available_tools["http_request"] = MCPTool(
            name="http_request",
            description="Make HTTP requests to external APIs",
            parameters={
                "type": "object",
                "properties": {
                    "method": {
                        "type": "string",
                        "enum": ["GET", "POST", "PUT", "DELETE", "PATCH"],
                        "description": "HTTP method"
                    },
                    "url": {
                        "type": "string",
                        "description": "Request URL"
                    },
                    "headers": {
                        "type": "object",
                        "description": "Request headers"
                    },
                    "data": {
                        "type": "object",
                        "description": "Request body data"
                    },
                    "timeout": {
                        "type": "integer",
                        "description": "Request timeout in seconds",
                        "default": 30
                    }
                },
                "required": ["method", "url"]
            },
            handler="builtin_http_request"
        )
        
        # Register handlers
        self.tool_registry.update({
            "builtin_web_search": self._handle_web_search,
            "builtin_code_executor": self._handle_code_executor,
            "builtin_file_operations": self._handle_file_operations,
            "builtin_http_request": self._handle_http_request,
        })
    
    async def _start_mcp_server(self):
        """Start MCP WebSocket server"""
        try:
            async def handle_client(websocket, path):
                await self._handle_mcp_connection(websocket, path)
            
            # Start WebSocket server
            start_server = websockets.serve(
                handle_client,
                self.config.server_host,
                self.config.server_port,
                max_size=10**7,  # 10MB max message size
                ping_interval=20,
                ping_timeout=10
            )
            
            await start_server
            
            logger.info(f"MCP server started on {self.config.server_host}:{self.config.server_port}")
            
        except Exception as e:
            logger.error(f"Failed to start MCP server: {e}")
            raise
    
    async def _handle_mcp_connection(self, websocket, path):
        """Handle MCP WebSocket connection"""
        session_id = str(uuid.uuid4())
        
        try:
            # Create session
            session = MCPSession(
                session_id=session_id,
                user_id="",  # Will be set during authentication
                websocket=websocket
            )
            
            async with self._lock:
                self.sessions[session_id] = session
            
            logger.info(f"MCP session {session_id} connected")
            
            # Handle messages
            async for message in websocket:
                try:
                    await self._handle_mcp_message(session, message)
                except Exception as e:
                    logger.error(f"Error handling MCP message: {e}")
                    await self._send_error_response(websocket, str(e))
                    
        except ConnectionClosed:
            logger.info(f"MCP session {session_id} disconnected")
        except Exception as e:
            logger.error(f"MCP connection error: {e}")
        finally:
            # Cleanup session
            async with self._lock:
                if session_id in self.sessions:
                    self.sessions[session_id].is_active = False
                    del self.sessions[session_id]
    
    async def _handle_mcp_message(self, session: MCPSession, message: str):
        """Handle incoming MCP message"""
        try:
            data = json.loads(message)
            
            # Update session activity
            session.last_activity = time.time()
            
            # Handle different message types
            if data.get("method") == "initialize":
                await self._handle_initialize(session, data)
            elif data.get("method") == "tools/list":
                await self._handle_tools_list(session, data)
            elif data.get("method") == "tools/call":
                await self._handle_tool_call(session, data)
            elif data.get("method") == "ping":
                await self._handle_ping(session, data)
            else:
                await self._send_error_response(
                    session.websocket, 
                    f"Unknown method: {data.get('method')}"
                )
                
        except json.JSONDecodeError as e:
            await self._send_error_response(session.websocket, f"Invalid JSON: {e}")
        except Exception as e:
            await self._send_error_response(session.websocket, f"Message handling error: {e}")
    
    async def _handle_initialize(self, session: MCPSession, data: Dict[str, Any]):
        """Handle MCP initialization"""
        params = data.get("params", {})
        
        # Extract user information
        session.user_id = params.get("user_id", "anonymous")
        
        # Send initialization response
        response = {
            "jsonrpc": "2.0",
            "id": data.get("id"),
            "result": {
                "protocol_version": "1.0",
                "server_info": {
                    "name": "Multi-Agent LLM Service",
                    "version": "1.0.0"
                },
                "capabilities": {
                    "tools": True,
                    "resources": False,
                    "prompts": False
                }
            }
        }
        
        await session.websocket.send(json.dumps(response))
        logger.debug(f"MCP session {session.session_id} initialized for user {session.user_id}")
    
    async def _handle_tools_list(self, session: MCPSession, data: Dict[str, Any]):
        """Handle tools list request"""
        
        # Filter tools based on user permissions (simplified)
        available_tools = []
        
        for tool_name, tool in self.available_tools.items():
            if tool.enabled:
                available_tools.append({
                    "name": tool.name,
                    "description": tool.description,
                    "inputSchema": tool.parameters
                })
        
        response = {
            "jsonrpc": "2.0",
            "id": data.get("id"),
            "result": {
                "tools": available_tools
            }
        }
        
        await session.websocket.send(json.dumps(response))
    
    async def _handle_tool_call(self, session: MCPSession, data: Dict[str, Any]):
        """Handle tool execution request"""
        params = data.get("params", {})
        tool_name = params.get("name")
        arguments = params.get("arguments", {})
        
        if not tool_name or tool_name not in self.available_tools:
            await self._send_error_response(
                session.websocket,
                f"Tool not found: {tool_name}",
                data.get("id")
            )
            return
        
        tool = self.available_tools[tool_name]
        
        try:
            # Execute tool
            result = await self._execute_tool(tool, arguments, session)
            
            response = {
                "jsonrpc": "2.0",
                "id": data.get("id"),
                "result": {
                    "content": [
                        {
                            "type": "text",
                            "text": json.dumps(result) if isinstance(result, dict) else str(result)
                        }
                    ]
                }
            }
            
            await session.websocket.send(json.dumps(response))
            
        except Exception as e:
            await self._send_error_response(
                session.websocket,
                f"Tool execution failed: {e}",
                data.get("id")
            )
    
    async def _handle_ping(self, session: MCPSession, data: Dict[str, Any]):
        """Handle ping request"""
        response = {
            "jsonrpc": "2.0",
            "id": data.get("id"),
            "result": {"status": "pong"}
        }
        
        await session.websocket.send(json.dumps(response))
    
    async def _execute_tool(
        self, 
        tool: MCPTool, 
        arguments: Dict[str, Any], 
        session: MCPSession
    ) -> Any:
        """Execute a tool with given arguments"""
        
        if not tool.handler or tool.handler not in self.tool_registry:
            raise ValueError(f"No handler found for tool: {tool.name}")
        
        handler = self.tool_registry[tool.handler]
        
        # Execute with timeout
        try:
            result = await asyncio.wait_for(
                handler(arguments, session),
                timeout=tool.timeout
            )
            return result
        except asyncio.TimeoutError:
            raise Exception(f"Tool execution timed out after {tool.timeout} seconds")
    
    async def _send_error_response(
        self, 
        websocket: websockets.WebSocketServerProtocol, 
        error_message: str,
        request_id: Optional[str] = None
    ):
        """Send error response"""
        response = {
            "jsonrpc": "2.0",
            "id": request_id,
            "error": {
                "code": -32000,
                "message": error_message
            }
        }
        
        try:
            await websocket.send(json.dumps(response))
        except Exception as e:
            logger.error(f"Failed to send error response: {e}")
    
    # Built-in tool handlers
    
    async def _handle_web_search(self, arguments: Dict[str, Any], session: MCPSession) -> Dict[str, Any]:
        """Handle web search tool"""
        query = arguments.get("query")
        max_results = arguments.get("max_results", 5)
        
        # Mock implementation - replace with actual search API
        results = [
            {
                "title": f"Search result {i+1} for: {query}",
                "url": f"https://example.com/result{i+1}",
                "snippet": f"This is a mock search result snippet for query: {query}"
            }
            for i in range(min(max_results, 3))
        ]
        
        return {
            "query": query,
            "results": results,
            "total_results": len(results)
        }
    
    async def _handle_code_executor(self, arguments: Dict[str, Any], session: MCPSession) -> Dict[str, Any]:
        """Handle code execution tool"""
        language = arguments.get("language")
        code = arguments.get("code")
        timeout = arguments.get("timeout", 30)
        
        # Mock implementation - integrate with Rust execution layer
        if language == "python":
            # This would call the Rust execution layer
            result = {
                "stdout": f"# Executed Python code:\n{code}\n# Output: Hello from Python!",
                "stderr": "",
                "exit_code": 0,
                "execution_time": 0.1
            }
        elif language == "javascript":
            result = {
                "stdout": f"// Executed JavaScript code:\n{code}\n// Output: Hello from JavaScript!",
                "stderr": "",
                "exit_code": 0,
                "execution_time": 0.1
            }
        else:
            result = {
                "stdout": "",
                "stderr": f"Unsupported language: {language}",
                "exit_code": 1,
                "execution_time": 0.0
            }
        
        return result
    
    async def _handle_file_operations(self, arguments: Dict[str, Any], session: MCPSession) -> Dict[str, Any]:
        """Handle file operations tool"""
        operation = arguments.get("operation")
        path = arguments.get("path")
        content = arguments.get("content")
        
        # Mock implementation - add proper security checks
        if operation == "read":
            return {
                "operation": "read",
                "path": path,
                "content": f"Mock file content for: {path}",
                "size": 100
            }
        elif operation == "write":
            return {
                "operation": "write",
                "path": path,
                "bytes_written": len(content) if content else 0
            }
        elif operation == "list":
            return {
                "operation": "list",
                "path": path,
                "files": [
                    {"name": "file1.txt", "size": 100, "type": "file"},
                    {"name": "file2.txt", "size": 200, "type": "file"},
                    {"name": "subdir", "size": 0, "type": "directory"}
                ]
            }
        else:
            raise ValueError(f"Unsupported file operation: {operation}")
    
    async def _handle_http_request(self, arguments: Dict[str, Any], session: MCPSession) -> Dict[str, Any]:
        """Handle HTTP request tool"""
        method = arguments.get("method")
        url = arguments.get("url")
        headers = arguments.get("headers", {})
        data = arguments.get("data")
        timeout = arguments.get("timeout", 30)
        
        # Mock implementation - replace with actual HTTP client
        return {
            "method": method,
            "url": url,
            "status_code": 200,
            "headers": {"content-type": "application/json"},
            "body": {"message": f"Mock response for {method} {url}"},
            "response_time": 0.1
        }
    
    async def _load_external_tools(self):
        """Load external tools from registry"""
        if not self.config.tools_registry_url:
            return
        
        try:
            # Mock implementation - load from external registry
            logger.info("Loading external tools from registry")
            
            # This would fetch tools from external registry
            external_tools = []
            
            for tool_def in external_tools:
                tool = MCPTool(
                    name=tool_def["name"],
                    description=tool_def["description"],
                    parameters=tool_def["parameters"],
                    handler=tool_def.get("handler"),
                    timeout=tool_def.get("timeout", 30)
                )
                self.available_tools[tool.name] = tool
            
            logger.info(f"Loaded {len(external_tools)} external tools")
            
        except Exception as e:
            logger.error(f"Failed to load external tools: {e}")
    
    async def execute_tool_request(self, request: MCPToolRequest) -> MCPToolResponse:
        """Execute tool request (API interface)"""
        start_time = time.time()
        
        try:
            if request.tool_name not in self.available_tools:
                return MCPToolResponse(
                    tool_name=request.tool_name,
                    result=None,
                    success=False,
                    error_message=f"Tool not found: {request.tool_name}",
                    execution_time_ms=int((time.time() - start_time) * 1000)
                )
            
            tool = self.available_tools[request.tool_name]
            
            # Create mock session for API calls
            mock_session = MCPSession(
                session_id=str(uuid.uuid4()),
                user_id=request.user_id
            )
            
            # Execute tool
            result = await self._execute_tool(tool, request.parameters, mock_session)
            
            return MCPToolResponse(
                tool_name=request.tool_name,
                result=result,
                success=True,
                execution_time_ms=int((time.time() - start_time) * 1000)
            )
            
        except Exception as e:
            return MCPToolResponse(
                tool_name=request.tool_name,
                result=None,
                success=False,
                error_message=str(e),
                execution_time_ms=int((time.time() - start_time) * 1000)
            )
    
    async def get_available_tools(self, user_id: str) -> List[Tool]:
        """Get available tools for user"""
        tools = []
        
        for tool_name, mcp_tool in self.available_tools.items():
            if mcp_tool.enabled:
                tool = Tool(
                    type="function",
                    function=ToolFunction(
                        name=mcp_tool.name,
                        description=mcp_tool.description,
                        parameters=mcp_tool.parameters
                    )
                )
                tools.append(tool)
        
        return tools
    
    async def _session_cleanup_task(self):
        """Background task to clean up inactive sessions"""
        while True:
            try:
                await asyncio.sleep(300)  # Check every 5 minutes
                
                current_time = time.time()
                inactive_sessions = []
                
                async with self._lock:
                    for session_id, session in self.sessions.items():
                        # Mark sessions inactive after 1 hour of no activity
                        if current_time - session.last_activity > 3600:
                            inactive_sessions.append(session_id)
                    
                    # Remove inactive sessions
                    for session_id in inactive_sessions:
                        if session_id in self.sessions:
                            session = self.sessions[session_id]
                            session.is_active = False
                            if session.websocket:
                                try:
                                    await session.websocket.close()
                                except Exception:
                                    pass
                            del self.sessions[session_id]
                
                if inactive_sessions:
                    logger.info(f"Cleaned up {len(inactive_sessions)} inactive MCP sessions")
                
            except Exception as e:
                logger.error(f"Session cleanup error: {e}")
    
    async def _health_check_task(self):
        """Background task for health monitoring"""
        while True:
            try:
                await asyncio.sleep(60)  # Check every minute
                
                # Monitor session health
                active_sessions = len([s for s in self.sessions.values() if s.is_active])
                
                logger.debug(f"MCP health check: {active_sessions} active sessions")
                
            except Exception as e:
                logger.error(f"Health check error: {e}")
    
    async def shutdown(self):
        """Shutdown MCP client"""
        logger.info("Shutting down MCP client")
        
        # Close all sessions
        async with self._lock:
            for session in self.sessions.values():
                session.is_active = False
                if session.websocket:
                    try:
                        await session.websocket.close()
                    except Exception:
                        pass
            
            self.sessions.clear()
        
        logger.info("MCP client shutdown completed")