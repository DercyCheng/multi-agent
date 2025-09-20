package temporal

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	workflowpkg "github.com/multi-agent/go/orchestrator/internal/workflow"
)

// MockAgentClient provides a mock implementation of AgentClient for testing
type MockAgentClient struct {
	logger *zap.Logger
}

// ExecuteAgent executes a mock agent
func (m *MockAgentClient) ExecuteAgent(ctx context.Context, request workflowpkg.AgentExecutionRequest) (*workflowpkg.AgentExecutionResult, error) {
	m.logger.Info("Mock agent execution",
		zap.String("agent_id", request.AgentID),
		zap.String("agent_type", request.AgentType),
		zap.String("query", request.Query),
	)

	// Simulate execution time
	time.Sleep(100 * time.Millisecond)

	// Mock result based on agent type
	var result string
	var confidence float64
	var toolCalls []workflowpkg.ToolCall

	switch request.AgentType {
	case "researcher":
		result = fmt.Sprintf("Research findings for: %s\n- Found relevant information\n- Analyzed data sources\n- Compiled comprehensive report", request.Query)
		confidence = 0.85
		toolCalls = []workflowpkg.ToolCall{
			{
				ToolName:        "search",
				Parameters:      map[string]interface{}{"query": request.Query},
				Result:          "Search results found",
				ExecutionTimeMs: 50,
				CostUSD:         0.001,
				Success:         true,
			},
		}
	case "analyzer":
		result = fmt.Sprintf("Analysis results for: %s\n- Identified key patterns\n- Performed statistical analysis\n- Generated insights", request.Query)
		confidence = 0.90
		toolCalls = []workflowpkg.ToolCall{
			{
				ToolName:        "analyze",
				Parameters:      map[string]interface{}{"data": "research_data"},
				Result:          "Analysis completed",
				ExecutionTimeMs: 75,
				CostUSD:         0.002,
				Success:         true,
			},
		}
	case "planner":
		result = fmt.Sprintf("Execution plan for: %s\n1. Initial assessment\n2. Resource allocation\n3. Implementation steps\n4. Quality assurance", request.Query)
		confidence = 0.80
		toolCalls = []workflowpkg.ToolCall{
			{
				ToolName:        "plan",
				Parameters:      map[string]interface{}{"objective": request.Query},
				Result:          "Plan created",
				ExecutionTimeMs: 60,
				CostUSD:         0.0015,
				Success:         true,
			},
		}
	default:
		result = fmt.Sprintf("General response for: %s\n- Processed request\n- Generated appropriate response", request.Query)
		confidence = 0.75
		toolCalls = []workflowpkg.ToolCall{
			{
				ToolName:        "general_process",
				Parameters:      map[string]interface{}{"input": request.Query},
				Result:          "Processing completed",
				ExecutionTimeMs: 40,
				CostUSD:         0.001,
				Success:         true,
			},
		}
	}

	// Calculate tokens used (mock calculation)
	tokensUsed := len(request.Query)/4 + len(result)/4 + 100 // Rough estimation
	if tokensUsed > request.TokenBudget {
		tokensUsed = request.TokenBudget
	}

	// Calculate cost
	costUSD := float64(tokensUsed) * 0.002 // $0.002 per token

	executionResult := &workflowpkg.AgentExecutionResult{
		AgentID:         request.AgentID,
		Status:          "completed",
		Result:          result,
		TokensUsed:      tokensUsed,
		CostUSD:         costUSD,
		ExecutionTimeMs: 100,
		ToolCalls:       toolCalls,
		Confidence:      confidence,
		Metadata: map[string]interface{}{
			"agent_type": request.AgentType,
			"mock":       true,
		},
	}

	m.logger.Info("Mock agent execution completed",
		zap.String("agent_id", request.AgentID),
		zap.Int("tokens_used", tokensUsed),
		zap.Float64("cost_usd", costUSD),
	)

	return executionResult, nil
}

// GetAgentCapabilities returns mock agent capabilities
func (m *MockAgentClient) GetAgentCapabilities(ctx context.Context, agentType string) (*workflowpkg.AgentCapabilities, error) {
	capabilities := &workflowpkg.AgentCapabilities{
		AgentType:        agentType,
		MaxTokens:        4000,
		CostPerToken:     0.002,
		AverageLatencyMs: 100,
		SuccessRate:      0.95,
	}

	switch agentType {
	case "researcher":
		capabilities.SupportedTools = []string{"search", "web_browse", "document_read"}
	case "analyzer":
		capabilities.SupportedTools = []string{"analyze", "synthesize", "calculate"}
	case "planner":
		capabilities.SupportedTools = []string{"plan", "decompose", "schedule"}
	case "executor":
		capabilities.SupportedTools = []string{"execute", "implement", "code"}
	case "validator":
		capabilities.SupportedTools = []string{"validate", "test", "review"}
	default:
		capabilities.SupportedTools = []string{"general_process", "respond"}
	}

	return capabilities, nil
}