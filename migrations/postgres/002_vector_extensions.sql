-- Vector database extensions and tables for Multi-Agent platform

-- Enable vector extension for embeddings
CREATE EXTENSION IF NOT EXISTS vector;

-- Create vector collections table for Qdrant-like functionality
CREATE TABLE vector_collections (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) UNIQUE NOT NULL,
    tenant_id UUID NOT NULL REFERENCES auth.tenants(id) ON DELETE CASCADE,
    dimensions INTEGER NOT NULL DEFAULT 1536,
    distance_metric VARCHAR(20) DEFAULT 'cosine' CHECK (distance_metric IN ('cosine', 'euclidean', 'dot')),
    description TEXT,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Vector entries for semantic search and RAG
CREATE TABLE vector_entries (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    collection_id UUID NOT NULL REFERENCES vector_collections(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES auth.tenants(id) ON DELETE CASCADE,
    user_id UUID REFERENCES auth.users(id) ON DELETE CASCADE,
    vector_data VECTOR(1536) NOT NULL,
    content TEXT NOT NULL,
    content_hash VARCHAR(64) UNIQUE NOT NULL,
    source_type VARCHAR(50) NOT NULL CHECK (source_type IN ('memory', 'document', 'conversation', 'tool_result', 'hypothesis', 'evidence')),
    source_id VARCHAR(255),
    importance_score DECIMAL(3,2) DEFAULT 0.5,
    access_count INTEGER DEFAULT 0,
    last_accessed_at TIMESTAMPTZ DEFAULT NOW(),
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    expires_at TIMESTAMPTZ
);

-- Conversation summaries for context compression
CREATE TABLE conversation_summaries (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    session_id VARCHAR(255) NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES auth.tenants(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    summary_text TEXT NOT NULL,
    summary_vector VECTOR(1536),
    original_message_count INTEGER NOT NULL,
    compression_ratio DECIMAL(5,2),
    summary_type VARCHAR(50) DEFAULT 'progressive' CHECK (summary_type IN ('progressive', 'session_boundary', 'topic_shift')),
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Hypothesis tracking for exploratory understanding
CREATE TABLE hypotheses (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    task_execution_id UUID NOT NULL REFERENCES task_executions(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES auth.tenants(id) ON DELETE CASCADE,
    hypothesis_text TEXT NOT NULL,
    hypothesis_vector VECTOR(1536),
    confidence_score DECIMAL(3,2) DEFAULT 0.5,
    evidence_count INTEGER DEFAULT 0,
    supporting_evidence_count INTEGER DEFAULT 0,
    contradicting_evidence_count INTEGER DEFAULT 0,
    status VARCHAR(50) DEFAULT 'active' CHECK (status IN ('active', 'validated', 'refuted', 'merged')),
    parent_hypothesis_id UUID REFERENCES hypotheses(id) ON DELETE SET NULL,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Evidence entries for hypothesis validation
CREATE TABLE evidence_entries (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    hypothesis_id UUID NOT NULL REFERENCES hypotheses(id) ON DELETE CASCADE,
    task_execution_id UUID NOT NULL REFERENCES task_executions(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES auth.tenants(id) ON DELETE CASCADE,
    evidence_text TEXT NOT NULL,
    evidence_vector VECTOR(1536),
    evidence_type VARCHAR(50) NOT NULL CHECK (evidence_type IN ('supporting', 'contradicting', 'neutral')),
    strength DECIMAL(3,2) DEFAULT 0.5,
    reliability DECIMAL(3,2) DEFAULT 0.5,
    source_tool VARCHAR(255),
    source_agent VARCHAR(255),
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Tool result cache for performance optimization
CREATE TABLE tool_result_cache (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tool_name VARCHAR(255) NOT NULL,
    parameters_hash VARCHAR(64) NOT NULL,
    parameters JSONB NOT NULL,
    result JSONB NOT NULL,
    result_vector VECTOR(1536),
    tenant_id UUID NOT NULL REFERENCES auth.tenants(id) ON DELETE CASCADE,
    cache_tier VARCHAR(20) DEFAULT 'hot' CHECK (cache_tier IN ('hot', 'warm', 'cold')),
    hit_count INTEGER DEFAULT 0,
    last_hit_at TIMESTAMPTZ DEFAULT NOW(),
    cost_usd DECIMAL(10,6) DEFAULT 0,
    execution_time_ms INTEGER,
    ttl_seconds INTEGER DEFAULT 3600,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    expires_at TIMESTAMPTZ DEFAULT (NOW() + INTERVAL '1 hour')
);

-- Pattern library for successful execution patterns
CREATE TABLE execution_patterns (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    pattern_name VARCHAR(255) NOT NULL,
    pattern_type VARCHAR(50) NOT NULL CHECK (pattern_type IN ('task_decomposition', 'tool_sequence', 'model_selection', 'context_assembly')),
    tenant_id UUID REFERENCES auth.tenants(id) ON DELETE CASCADE,
    pattern_data JSONB NOT NULL,
    pattern_vector VECTOR(1536),
    success_rate DECIMAL(5,4) DEFAULT 0.0,
    usage_count INTEGER DEFAULT 0,
    avg_cost_usd DECIMAL(10,6) DEFAULT 0,
    avg_execution_time_ms INTEGER,
    last_used_at TIMESTAMPTZ DEFAULT NOW(),
    is_global BOOLEAN DEFAULT false,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes for vector operations and performance
CREATE INDEX idx_vector_collections_tenant_id ON vector_collections(tenant_id);
CREATE INDEX idx_vector_collections_name ON vector_collections(name);

CREATE INDEX idx_vector_entries_collection_id ON vector_entries(collection_id);
CREATE INDEX idx_vector_entries_tenant_id ON vector_entries(tenant_id);
CREATE INDEX idx_vector_entries_source_type ON vector_entries(source_type);
CREATE INDEX idx_vector_entries_content_hash ON vector_entries(content_hash);
CREATE INDEX idx_vector_entries_importance ON vector_entries(importance_score);

-- Vector similarity search indexes (using HNSW for performance)
CREATE INDEX idx_vector_entries_vector_cosine ON vector_entries USING hnsw (vector_data vector_cosine_ops);
CREATE INDEX idx_vector_entries_vector_l2 ON vector_entries USING hnsw (vector_data vector_l2_ops);

CREATE INDEX idx_conversation_summaries_session_id ON conversation_summaries(session_id);
CREATE INDEX idx_conversation_summaries_tenant_id ON conversation_summaries(tenant_id);
CREATE INDEX idx_conversation_summaries_vector_cosine ON conversation_summaries USING hnsw (summary_vector vector_cosine_ops);

CREATE INDEX idx_hypotheses_task_id ON hypotheses(task_execution_id);
CREATE INDEX idx_hypotheses_tenant_id ON hypotheses(tenant_id);
CREATE INDEX idx_hypotheses_confidence ON hypotheses(confidence_score);
CREATE INDEX idx_hypotheses_status ON hypotheses(status);
CREATE INDEX idx_hypotheses_vector_cosine ON hypotheses USING hnsw (hypothesis_vector vector_cosine_ops);

CREATE INDEX idx_evidence_entries_hypothesis_id ON evidence_entries(hypothesis_id);
CREATE INDEX idx_evidence_entries_task_id ON evidence_entries(task_execution_id);
CREATE INDEX idx_evidence_entries_tenant_id ON evidence_entries(tenant_id);
CREATE INDEX idx_evidence_entries_type ON evidence_entries(evidence_type);
CREATE INDEX idx_evidence_entries_vector_cosine ON evidence_entries USING hnsw (evidence_vector vector_cosine_ops);

CREATE INDEX idx_tool_result_cache_tool_name ON tool_result_cache(tool_name);
CREATE INDEX idx_tool_result_cache_params_hash ON tool_result_cache(parameters_hash);
CREATE INDEX idx_tool_result_cache_tenant_id ON tool_result_cache(tenant_id);
CREATE INDEX idx_tool_result_cache_tier ON tool_result_cache(cache_tier);
CREATE INDEX idx_tool_result_cache_expires_at ON tool_result_cache(expires_at);
CREATE INDEX idx_tool_result_cache_vector_cosine ON tool_result_cache USING hnsw (result_vector vector_cosine_ops);

CREATE INDEX idx_execution_patterns_type ON execution_patterns(pattern_type);
CREATE INDEX idx_execution_patterns_tenant_id ON execution_patterns(tenant_id);
CREATE INDEX idx_execution_patterns_success_rate ON execution_patterns(success_rate);
CREATE INDEX idx_execution_patterns_vector_cosine ON execution_patterns USING hnsw (pattern_vector vector_cosine_ops);

-- Functions for vector similarity search
CREATE OR REPLACE FUNCTION search_similar_vectors(
    collection_name TEXT,
    query_vector VECTOR(1536),
    tenant_uuid UUID,
    similarity_threshold FLOAT DEFAULT 0.8,
    limit_count INTEGER DEFAULT 10
)
RETURNS TABLE (
    id UUID,
    content TEXT,
    similarity FLOAT,
    metadata JSONB,
    created_at TIMESTAMPTZ
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        ve.id,
        ve.content,
        1 - (ve.vector_data <=> query_vector) AS similarity,
        ve.metadata,
        ve.created_at
    FROM vector_entries ve
    JOIN vector_collections vc ON ve.collection_id = vc.id
    WHERE vc.name = collection_name 
        AND ve.tenant_id = tenant_uuid
        AND (1 - (ve.vector_data <=> query_vector)) >= similarity_threshold
    ORDER BY ve.vector_data <=> query_vector
    LIMIT limit_count;
END;
$$ LANGUAGE plpgsql;

-- Function for hypothesis-evidence correlation
CREATE OR REPLACE FUNCTION correlate_evidence_with_hypothesis(
    hypothesis_uuid UUID,
    evidence_text TEXT,
    evidence_vector VECTOR(1536)
)
RETURNS DECIMAL(3,2) AS $$
DECLARE
    hypothesis_vector VECTOR(1536);
    correlation_score DECIMAL(3,2);
BEGIN
    -- Get hypothesis vector
    SELECT h.hypothesis_vector INTO hypothesis_vector
    FROM hypotheses h
    WHERE h.id = hypothesis_uuid;
    
    -- Calculate correlation (cosine similarity)
    correlation_score := 1 - (hypothesis_vector <=> evidence_vector);
    
    RETURN correlation_score;
END;
$$ LANGUAGE plpgsql;

-- Function for automatic cache cleanup
CREATE OR REPLACE FUNCTION cleanup_expired_cache()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM tool_result_cache 
    WHERE expires_at < NOW();
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    
    -- Also cleanup old vector entries with low importance
    DELETE FROM vector_entries 
    WHERE expires_at < NOW() 
        OR (importance_score < 0.1 AND last_accessed_at < NOW() - INTERVAL '30 days');
    
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- Triggers for automatic updates
CREATE TRIGGER update_vector_collections_updated_at 
    BEFORE UPDATE ON vector_collections 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_hypotheses_updated_at 
    BEFORE UPDATE ON hypotheses 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_execution_patterns_updated_at 
    BEFORE UPDATE ON execution_patterns 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Row Level Security for vector tables
ALTER TABLE vector_collections ENABLE ROW LEVEL SECURITY;
ALTER TABLE vector_entries ENABLE ROW LEVEL SECURITY;
ALTER TABLE conversation_summaries ENABLE ROW LEVEL SECURITY;
ALTER TABLE hypotheses ENABLE ROW LEVEL SECURITY;
ALTER TABLE evidence_entries ENABLE ROW LEVEL SECURITY;
ALTER TABLE tool_result_cache ENABLE ROW LEVEL SECURITY;
ALTER TABLE execution_patterns ENABLE ROW LEVEL SECURITY;

-- Initial vector collections
INSERT INTO vector_collections (name, tenant_id, description) VALUES 
('memories', (SELECT id FROM auth.tenants WHERE slug = 'default'), 'Long-term memory storage'),
('context', (SELECT id FROM auth.tenants WHERE slug = 'default'), 'Session context vectors'),
('tools', (SELECT id FROM auth.tenants WHERE slug = 'default'), 'Tool capability vectors'),
('summaries', (SELECT id FROM auth.tenants WHERE slug = 'default'), 'Conversation summaries');