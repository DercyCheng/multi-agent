"""
Context Engineering Engine for Multi-Agent LLM Service
Intelligent context management with knowledge injection and memory retrieval
"""

import asyncio
import logging
import json
import hashlib
from typing import Dict, List, Optional, Any, Tuple
from dataclasses import dataclass
import time

import asyncpg
from qdrant_client import QdrantClient
from qdrant_client.models import Distance, VectorParams, PointStruct, Filter, FieldCondition, MatchValue
from sentence_transformers import SentenceTransformer

from src.config import ContextConfig
from src.core.models import ContextRequest, EngineeredContext, Message, Tool, MessageRole

logger = logging.getLogger(__name__)

@dataclass
class KnowledgeChunk:
    """Knowledge chunk with metadata"""
    id: str
    content: str
    source: str
    relevance_score: float
    metadata: Dict[str, Any]
    embedding: Optional[List[float]] = None

@dataclass
class MemoryEntry:
    """Memory entry for conversation context"""
    session_id: str
    user_id: str
    content: str
    timestamp: float
    importance_score: float
    access_count: int = 0

class ContextEngine:
    """Advanced context engineering with RAG and memory management"""
    
    def __init__(self, config: ContextConfig):
        self.config = config
        self.db_pool: Optional[asyncpg.Pool] = None
        self.vector_client: Optional[QdrantClient] = None
        self.embedding_model: Optional[SentenceTransformer] = None
        self.template_cache: Dict[str, str] = {}
        self.memory_cache: Dict[str, List[MemoryEntry]] = {}
        self._lock = asyncio.Lock()
        
    async def initialize(self):
        """Initialize context engine"""
        logger.info("Initializing context engine")
        
        # Initialize database connection
        await self._init_database()
        
        # Initialize vector database
        await self._init_vector_db()
        
        # Initialize embedding model
        await self._init_embedding_model()
        
        # Start background tasks
        asyncio.create_task(self._memory_cleanup_task())
        asyncio.create_task(self._template_cache_cleanup())
        
        logger.info("Context engine initialized successfully")
    
    async def _init_database(self):
        """Initialize PostgreSQL connection"""
        try:
            self.db_pool = await asyncpg.create_pool(
                host="localhost",
                port=5432,
                database="multiagent",
                user="postgres",
                password="password",
                min_size=5,
                max_size=20
            )
            logger.info("Database connection pool created")
        except Exception as e:
            logger.error(f"Failed to initialize database: {e}")
            raise
    
    async def _init_vector_db(self):
        """Initialize Qdrant vector database"""
        try:
            self.vector_client = QdrantClient(
                host="localhost",
                port=6333,
                timeout=30
            )
            
            # Ensure collections exist
            collections = ["knowledge_base", "conversation_memory"]
            for collection in collections:
                try:
                    await self._ensure_collection_exists(collection)
                except Exception as e:
                    logger.warning(f"Collection {collection} setup issue: {e}")
            
            logger.info("Vector database initialized")
        except Exception as e:
            logger.error(f"Failed to initialize vector database: {e}")
            raise
    
    async def _ensure_collection_exists(self, collection_name: str):
        """Ensure vector collection exists"""
        try:
            collections = self.vector_client.get_collections().collections
            collection_names = [c.name for c in collections]
            
            if collection_name not in collection_names:
                self.vector_client.create_collection(
                    collection_name=collection_name,
                    vectors_config=VectorParams(
                        size=384,  # sentence-transformers/all-MiniLM-L6-v2 dimension
                        distance=Distance.COSINE
                    )
                )
                logger.info(f"Created vector collection: {collection_name}")
        except Exception as e:
            logger.error(f"Error ensuring collection {collection_name}: {e}")
    
    async def _init_embedding_model(self):
        """Initialize embedding model"""
        try:
            # Use lightweight model for fast inference
            self.embedding_model = SentenceTransformer('all-MiniLM-L6-v2')
            logger.info("Embedding model loaded")
        except Exception as e:
            logger.error(f"Failed to load embedding model: {e}")
            raise
    
    async def engineer_context(self, request: ContextRequest) -> EngineeredContext:
        """Engineer context for the request"""
        async with self._lock:
            start_time = time.time()
            
            try:
                # Generate context components
                system_instructions = await self._generate_system_instructions(request)
                knowledge = await self._retrieve_knowledge(request)
                tools = await self._select_tools(request)
                memory = await self._retrieve_memory(request)
                
                # Calculate token count
                total_content = f"{system_instructions}\n{knowledge}\n{json.dumps(memory)}"
                token_count = len(total_content) // 4  # Rough estimation
                
                # Apply compression if needed
                compression_ratio = None
                if token_count > self.config.max_context_length:
                    compressed_content = await self._compress_context(
                        system_instructions, knowledge, memory, request
                    )
                    system_instructions, knowledge, memory = compressed_content
                    new_token_count = len(f"{system_instructions}\n{knowledge}\n{json.dumps(memory)}") // 4
                    compression_ratio = new_token_count / token_count
                    token_count = new_token_count
                
                # Create engineered context
                context = EngineeredContext(
                    system_instructions=system_instructions,
                    knowledge=knowledge,
                    tools=tools,
                    memory=memory,
                    metadata={
                        "request_id": request.session_id,
                        "task_type": request.task_type,
                        "processing_time_ms": int((time.time() - start_time) * 1000),
                        "knowledge_sources": len(knowledge.split('\n\n')) if knowledge else 0,
                        "memory_entries": len(memory),
                    },
                    token_count=token_count,
                    compression_ratio=compression_ratio
                )
                
                # Store context for future reference
                await self._store_context_usage(request, context)
                
                logger.debug(f"Context engineered in {context.metadata['processing_time_ms']}ms")
                
                return context
                
            except Exception as e:
                logger.error(f"Context engineering failed: {e}")
                raise
    
    async def _generate_system_instructions(self, request: ContextRequest) -> str:
        """Generate system instructions based on task type and user preferences"""
        
        # Check template cache
        cache_key = f"system_{request.task_type}_{hash(str(request.user_preferences))}"
        if cache_key in self.template_cache:
            return self.template_cache[cache_key]
        
        # Base system instruction
        base_instruction = """You are an intelligent AI assistant in a multi-agent platform. 
Your role is to provide helpful, accurate, and contextually relevant responses."""
        
        # Task-specific instructions
        task_instructions = {
            "code_generation": """
Focus on writing clean, efficient, and well-documented code.
Consider best practices, security implications, and maintainability.
Provide explanations for complex logic and suggest improvements when appropriate.
""",
            "data_analysis": """
Analyze data systematically and provide clear insights.
Use appropriate statistical methods and visualizations.
Explain your methodology and highlight key findings.
""",
            "creative_writing": """
Be creative and engaging while maintaining coherence.
Adapt your writing style to the requested genre or format.
Focus on originality and compelling narrative structure.
""",
            "problem_solving": """
Break down complex problems into manageable components.
Consider multiple approaches and evaluate trade-offs.
Provide step-by-step solutions with clear reasoning.
""",
            "research": """
Provide comprehensive and well-researched information.
Cite relevant sources and distinguish between facts and opinions.
Organize information logically and highlight key points.
""",
        }
        
        # Get task-specific instruction
        task_instruction = task_instructions.get(request.task_type, "")
        
        # User preference adaptations
        preferences = request.user_preferences
        preference_adaptations = []
        
        if preferences.get("communication_style") == "formal":
            preference_adaptations.append("Use formal language and professional tone.")
        elif preferences.get("communication_style") == "casual":
            preference_adaptations.append("Use conversational and approachable language.")
        
        if preferences.get("detail_level") == "high":
            preference_adaptations.append("Provide detailed explanations and comprehensive coverage.")
        elif preferences.get("detail_level") == "low":
            preference_adaptations.append("Keep responses concise and focus on key points.")
        
        if preferences.get("expertise_level") == "beginner":
            preference_adaptations.append("Explain concepts clearly and avoid technical jargon.")
        elif preferences.get("expertise_level") == "expert":
            preference_adaptations.append("Use technical terminology and assume advanced knowledge.")
        
        # Combine instructions
        full_instruction = base_instruction
        if task_instruction:
            full_instruction += f"\n\nTask-specific guidance:\n{task_instruction}"
        if preference_adaptations:
            full_instruction += f"\n\nUser preferences:\n" + "\n".join(f"- {pref}" for pref in preference_adaptations)
        
        # Cache the result
        if len(self.template_cache) < self.config.template_cache_size:
            self.template_cache[cache_key] = full_instruction
        
        return full_instruction
    
    async def _retrieve_knowledge(self, request: ContextRequest) -> str:
        """Retrieve relevant knowledge from knowledge base"""
        if not self.config.knowledge_injection_enabled:
            return ""
        
        try:
            # Generate query embedding
            query_embedding = self.embedding_model.encode(request.query).tolist()
            
            # Search vector database
            search_results = self.vector_client.search(
                collection_name="knowledge_base",
                query_vector=query_embedding,
                limit=10,
                score_threshold=0.7,
                with_payload=True
            )
            
            # Process results
            knowledge_chunks = []
            total_tokens = 0
            
            for result in search_results:
                if total_tokens >= request.knowledge_budget:
                    break
                
                chunk = KnowledgeChunk(
                    id=str(result.id),
                    content=result.payload.get("content", ""),
                    source=result.payload.get("source", "unknown"),
                    relevance_score=result.score,
                    metadata=result.payload.get("metadata", {})
                )
                
                chunk_tokens = len(chunk.content) // 4
                if total_tokens + chunk_tokens <= request.knowledge_budget:
                    knowledge_chunks.append(chunk)
                    total_tokens += chunk_tokens
            
            # Format knowledge for context
            if not knowledge_chunks:
                return ""
            
            knowledge_sections = []
            for chunk in knowledge_chunks:
                section = f"Source: {chunk.source} (Relevance: {chunk.relevance_score:.2f})\n{chunk.content}"
                knowledge_sections.append(section)
            
            return "\n\n".join(knowledge_sections)
            
        except Exception as e:
            logger.error(f"Knowledge retrieval failed: {e}")
            return ""
    
    async def _select_tools(self, request: ContextRequest) -> List[Tool]:
        """Select appropriate tools for the request"""
        if not request.available_tools:
            return []
        
        # Simple tool selection based on task type and query
        selected_tools = []
        
        # Task-type based tool selection
        task_tool_mapping = {
            "code_generation": ["code_executor", "syntax_checker", "documentation_generator"],
            "data_analysis": ["data_processor", "chart_generator", "statistical_analyzer"],
            "research": ["web_search", "document_reader", "citation_formatter"],
            "creative_writing": ["grammar_checker", "style_analyzer", "thesaurus"],
        }
        
        recommended_tools = task_tool_mapping.get(request.task_type, [])
        
        # Filter available tools
        for tool_name in recommended_tools:
            if tool_name in request.available_tools:
                # Create tool definition (simplified)
                tool = Tool(
                    type="function",
                    function={
                        "name": tool_name,
                        "description": f"Tool for {tool_name.replace('_', ' ')}",
                        "parameters": {
                            "type": "object",
                            "properties": {},
                            "required": []
                        }
                    }
                )
                selected_tools.append(tool)
        
        return selected_tools
    
    async def _retrieve_memory(self, request: ContextRequest) -> Dict[str, Any]:
        """Retrieve relevant conversation memory"""
        if not self.config.memory_retrieval_enabled:
            return {}
        
        try:
            # Check cache first
            cache_key = f"{request.user_id}:{request.session_id}"
            if cache_key in self.memory_cache:
                memory_entries = self.memory_cache[cache_key]
            else:
                # Retrieve from database
                memory_entries = await self._fetch_memory_from_db(request)
                self.memory_cache[cache_key] = memory_entries
            
            # Process memory entries
            memory = {
                "recent_interactions": [],
                "important_facts": [],
                "user_preferences": {},
                "conversation_summary": ""
            }
            
            # Sort by importance and recency
            sorted_entries = sorted(
                memory_entries,
                key=lambda x: (x.importance_score, x.timestamp),
                reverse=True
            )
            
            for entry in sorted_entries[:10]:  # Limit to top 10 entries
                if entry.importance_score > 0.7:
                    memory["important_facts"].append({
                        "content": entry.content,
                        "timestamp": entry.timestamp,
                        "importance": entry.importance_score
                    })
                else:
                    memory["recent_interactions"].append({
                        "content": entry.content,
                        "timestamp": entry.timestamp
                    })
            
            return memory
            
        except Exception as e:
            logger.error(f"Memory retrieval failed: {e}")
            return {}
    
    async def _fetch_memory_from_db(self, request: ContextRequest) -> List[MemoryEntry]:
        """Fetch memory entries from database"""
        if not self.db_pool:
            return []
        
        try:
            async with self.db_pool.acquire() as conn:
                rows = await conn.fetch("""
                    SELECT session_id, user_id, content, created_at, importance_score, access_count
                    FROM conversation_memory 
                    WHERE user_id = $1 AND session_id = $2
                    ORDER BY created_at DESC
                    LIMIT 50
                """, request.user_id, request.session_id)
                
                entries = []
                for row in rows:
                    entry = MemoryEntry(
                        session_id=row['session_id'],
                        user_id=row['user_id'],
                        content=row['content'],
                        timestamp=row['created_at'].timestamp(),
                        importance_score=row['importance_score'],
                        access_count=row['access_count']
                    )
                    entries.append(entry)
                
                return entries
                
        except Exception as e:
            logger.error(f"Database memory fetch failed: {e}")
            return []
    
    async def _compress_context(
        self, 
        system_instructions: str, 
        knowledge: str, 
        memory: Dict[str, Any], 
        request: ContextRequest
    ) -> Tuple[str, str, Dict[str, Any]]:
        """Compress context to fit within token limits"""
        
        target_tokens = int(self.config.max_context_length * self.config.compression_threshold)
        
        # Prioritize system instructions (keep full)
        compressed_system = system_instructions
        
        # Compress knowledge (keep most relevant)
        compressed_knowledge = ""
        if knowledge:
            knowledge_sections = knowledge.split('\n\n')
            knowledge_tokens = 0
            knowledge_budget = target_tokens // 2  # Allocate half to knowledge
            
            for section in knowledge_sections:
                section_tokens = len(section) // 4
                if knowledge_tokens + section_tokens <= knowledge_budget:
                    compressed_knowledge += section + '\n\n'
                    knowledge_tokens += section_tokens
                else:
                    break
        
        # Compress memory (keep most important)
        compressed_memory = {}
        if memory:
            memory_budget = target_tokens // 4  # Allocate quarter to memory
            memory_tokens = 0
            
            # Prioritize important facts
            if "important_facts" in memory:
                compressed_memory["important_facts"] = []
                for fact in memory["important_facts"]:
                    fact_tokens = len(str(fact)) // 4
                    if memory_tokens + fact_tokens <= memory_budget:
                        compressed_memory["important_facts"].append(fact)
                        memory_tokens += fact_tokens
                    else:
                        break
            
            # Add recent interactions if space allows
            if "recent_interactions" in memory and memory_tokens < memory_budget:
                compressed_memory["recent_interactions"] = []
                remaining_budget = memory_budget - memory_tokens
                
                for interaction in memory["recent_interactions"]:
                    interaction_tokens = len(str(interaction)) // 4
                    if interaction_tokens <= remaining_budget:
                        compressed_memory["recent_interactions"].append(interaction)
                        remaining_budget -= interaction_tokens
                    else:
                        break
        
        return compressed_system, compressed_knowledge.strip(), compressed_memory
    
    async def _store_context_usage(self, request: ContextRequest, context: EngineeredContext):
        """Store context usage for analytics"""
        if not self.db_pool:
            return
        
        try:
            async with self.db_pool.acquire() as conn:
                await conn.execute("""
                    INSERT INTO context_usage (
                        user_id, session_id, task_type, token_count, 
                        compression_ratio, processing_time_ms, created_at
                    ) VALUES ($1, $2, $3, $4, $5, $6, NOW())
                """, 
                    request.user_id,
                    request.session_id,
                    request.task_type,
                    context.token_count,
                    context.compression_ratio,
                    context.metadata.get("processing_time_ms", 0)
                )
        except Exception as e:
            logger.error(f"Failed to store context usage: {e}")
    
    async def store_conversation_memory(
        self, 
        user_id: str, 
        session_id: str, 
        messages: List[Message]
    ):
        """Store conversation in memory for future context"""
        try:
            for message in messages:
                # Calculate importance score
                importance_score = await self._calculate_importance_score(message)
                
                # Store in database
                if self.db_pool:
                    async with self.db_pool.acquire() as conn:
                        await conn.execute("""
                            INSERT INTO conversation_memory (
                                user_id, session_id, content, importance_score, created_at
                            ) VALUES ($1, $2, $3, $4, NOW())
                        """, user_id, session_id, message.content, importance_score)
                
                # Update cache
                cache_key = f"{user_id}:{session_id}"
                if cache_key not in self.memory_cache:
                    self.memory_cache[cache_key] = []
                
                entry = MemoryEntry(
                    session_id=session_id,
                    user_id=user_id,
                    content=message.content,
                    timestamp=time.time(),
                    importance_score=importance_score
                )
                self.memory_cache[cache_key].append(entry)
                
                # Limit cache size
                if len(self.memory_cache[cache_key]) > 100:
                    self.memory_cache[cache_key] = self.memory_cache[cache_key][-50:]
                
        except Exception as e:
            logger.error(f"Failed to store conversation memory: {e}")
    
    async def _calculate_importance_score(self, message: Message) -> float:
        """Calculate importance score for a message"""
        # Simple heuristic-based scoring
        score = 0.5  # Base score
        
        content = message.content.lower()
        
        # Boost for questions
        if any(word in content for word in ['?', 'how', 'what', 'why', 'when', 'where']):
            score += 0.2
        
        # Boost for preferences or personal information
        if any(word in content for word in ['prefer', 'like', 'dislike', 'always', 'never']):
            score += 0.3
        
        # Boost for important keywords
        important_keywords = ['important', 'critical', 'urgent', 'remember', 'note']
        if any(keyword in content for keyword in important_keywords):
            score += 0.4
        
        # Reduce for very short messages
        if len(content) < 10:
            score -= 0.2
        
        # Boost for longer, detailed messages
        if len(content) > 100:
            score += 0.1
        
        return max(0.0, min(1.0, score))
    
    async def _memory_cleanup_task(self):
        """Background task to clean up old memory entries"""
        while True:
            try:
                await asyncio.sleep(3600)  # Run every hour
                
                # Clean database
                if self.db_pool:
                    async with self.db_pool.acquire() as conn:
                        # Delete entries older than 30 days with low importance
                        await conn.execute("""
                            DELETE FROM conversation_memory 
                            WHERE created_at < NOW() - INTERVAL '30 days' 
                            AND importance_score < 0.3
                        """)
                        
                        # Delete entries older than 90 days regardless of importance
                        await conn.execute("""
                            DELETE FROM conversation_memory 
                            WHERE created_at < NOW() - INTERVAL '90 days'
                        """)
                
                # Clean cache
                current_time = time.time()
                for cache_key in list(self.memory_cache.keys()):
                    entries = self.memory_cache[cache_key]
                    # Keep only entries from last 24 hours in cache
                    recent_entries = [
                        entry for entry in entries 
                        if current_time - entry.timestamp < 86400
                    ]
                    if recent_entries:
                        self.memory_cache[cache_key] = recent_entries
                    else:
                        del self.memory_cache[cache_key]
                
                logger.debug("Memory cleanup completed")
                
            except Exception as e:
                logger.error(f"Memory cleanup error: {e}")
    
    async def _template_cache_cleanup(self):
        """Background task to clean up template cache"""
        while True:
            try:
                await asyncio.sleep(1800)  # Run every 30 minutes
                
                # Simple LRU-like cleanup - remove random entries if cache is full
                if len(self.template_cache) > self.config.template_cache_size:
                    # Remove 20% of entries
                    remove_count = len(self.template_cache) // 5
                    keys_to_remove = list(self.template_cache.keys())[:remove_count]
                    for key in keys_to_remove:
                        del self.template_cache[key]
                
                logger.debug("Template cache cleanup completed")
                
            except Exception as e:
                logger.error(f"Template cache cleanup error: {e}")
    
    async def shutdown(self):
        """Shutdown context engine"""
        logger.info("Shutting down context engine")
        
        if self.db_pool:
            await self.db_pool.close()
        
        if self.vector_client:
            self.vector_client.close()
        
        logger.info("Context engine shutdown completed")