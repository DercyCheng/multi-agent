package workflow

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"go.temporal.io/sdk/activity"
	"go.uber.org/zap"
)

// Activities contains all workflow activities
type Activities struct {
	logger        *zap.Logger
	agentClient   AgentClient
	budgetManager BudgetManager
	cacheManager  CacheManager
}

// NewActivities creates a new activities instance
func NewActivities(logger *zap.Logger, agentClient AgentClient, budgetManager BudgetManager, cacheManager CacheManager) *Activities {
	return &Activities{
		logger:        logger,
		agentClient:   agentClient,
		budgetManager: budgetManager,
		cacheManager:  cacheManager,
	}
}

// AgentClient interface for agent communication
type AgentClient interface {
	ExecuteAgent(ctx context.Context, request AgentExecutionRequest) (*AgentExecutionResult, error)
	GetAgentCapabilities(ctx context.Context, agentType string) (*AgentCapabilities, error)
}

// BudgetManager interface for token budget management
type BudgetManager interface {
	CheckBudget(ctx context.Context, userID, tenantID string, estimatedTokens int) error
	ReserveBudget(ctx context.Context, userID, tenantID string, tokens int) (string, error)
	ConsumeBudget(ctx context.Context, reservationID string, actualTokens int, costUSD float64) error
	ReleaseBudget(ctx context.Context, reservationID string) error
}

// CacheManager interface for caching operations
type CacheManager interface {
	Get(ctx context.Context, key string) (interface{}, error)
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	GetPattern(ctx context.Context, pattern string) ([]interface{}, error)
	SetPattern(ctx context.Context, pattern string, value interface{}, ttl time.Duration) error
}

// AgentExecutionRequest represents a request to execute an agent
type AgentExecutionRequest struct {
	AgentID     string                 `json:"agent_id"`
	AgentType   string                 `json:"agent_type"`
	Query       string                 `json:"query"`
	Context     map[string]interface{} `json:"context"`
	Tools       []string               `json:"tools"`
	TokenBudget int                    `json:"token_budget"`
	UserID      string                 `json:"user_id"`
	TenantID    string                 `json:"tenant_id"`
	SessionID   string                 `json:"session_id"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// AgentExecutionResult represents the result of agent execution
type AgentExecutionResult struct {
	AgentID         string                 `json:"agent_id"`
	Status          string                 `json:"status"`
	Result          string                 `json:"result"`
	ErrorMessage    string                 `json:"error_message"`
	TokensUsed      int                    `json:"tokens_used"`
	CostUSD         float64                `json:"cost_usd"`
	ExecutionTimeMs int64                  `json:"execution_time_ms"`
	ToolCalls       []ToolCall             `json:"tool_calls"`
	Confidence      float64                `json:"confidence"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// AgentCapabilities represents agent capabilities
type AgentCapabilities struct {
	AgentType        string   `json:"agent_type"`
	SupportedTools   []string `json:"supported_tools"`
	MaxTokens        int      `json:"max_tokens"`
	CostPerToken     float64  `json:"cost_per_token"`
	AverageLatencyMs int64    `json:"average_latency_ms"`
	SuccessRate      float64  `json:"success_rate"`
}

// ToolCall represents a tool call made by an agent
type ToolCall struct {
	ToolName        string                 `json:"tool_name"`
	Parameters      map[string]interface{} `json:"parameters"`
	Result          interface{}            `json:"result"`
	ExecutionTimeMs int64                  `json:"execution_time_ms"`
	CostUSD         float64                `json:"cost_usd"`
	Success         bool                   `json:"success"`
}

// AnalyzeComplexity analyzes the complexity of a task
func (a *Activities) AnalyzeComplexity(ctx context.Context, input TaskInput) (ComplexityAnalysis, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Analyzing task complexity", "query", input.Query)

	// Check cache first
	cacheKey := fmt.Sprintf("complexity:%s", hashString(input.Query))
	if cached, err := a.cacheManager.Get(ctx, cacheKey); err == nil {
		if analysis, ok := cached.(ComplexityAnalysis); ok {
			logger.Info("Using cached complexity analysis")
			return analysis, nil
		}
	}

	analysis := ComplexityAnalysis{
		Factors:  make(map[string]float64),
		Metadata: make(map[string]interface{}),
	}

	// Factor 1: Query length and structure
	queryLength := len(strings.Fields(input.Query))
	lengthFactor := math.Min(float64(queryLength)/50.0, 1.0)
	analysis.Factors["query_length"] = lengthFactor

	// Factor 2: Question complexity indicators
	complexityKeywords := []string{
		"analyze", "compare", "evaluate", "synthesize", "research",
		"multiple", "various", "different", "complex", "detailed",
		"comprehensive", "thorough", "in-depth", "extensive",
	}
	
	keywordCount := 0
	queryLower := strings.ToLower(input.Query)
	for _, keyword := range complexityKeywords {
		if strings.Contains(queryLower, keyword) {
			keywordCount++
		}
	}
	
	keywordFactor := math.Min(float64(keywordCount)/5.0, 1.0)
	analysis.Factors["complexity_keywords"] = keywordFactor

	// Factor 3: Multi-step indicators
	multiStepKeywords := []string{
		"first", "then", "next", "after", "finally", "step",
		"process", "workflow", "sequence", "order", "stages",
	}
	
	multiStepCount := 0
	for _, keyword := range multiStepKeywords {
		if strings.Contains(queryLower, keyword) {
			multiStepCount++
		}
	}
	
	multiStepFactor := math.Min(float64(multiStepCount)/3.0, 1.0)
	analysis.Factors["multi_step"] = multiStepFactor

	// Factor 4: Domain complexity
	technicalKeywords := []string{
		"algorithm", "architecture", "implementation", "optimization",
		"performance", "scalability", "security", "integration",
		"database", "api", "framework", "protocol", "specification",
	}
	
	technicalCount := 0
	for _, keyword := range technicalKeywords {
		if strings.Contains(queryLower, keyword) {
			technicalCount++
		}
	}
	
	technicalFactor := math.Min(float64(technicalCount)/5.0, 1.0)
	analysis.Factors["technical_complexity"] = technicalFactor

	// Factor 5: Context complexity
	contextFactor := 0.0
	if len(input.Context) > 0 {
		contextFactor = math.Min(float64(len(input.Context))/10.0, 0.5)
	}
	analysis.Factors["context_complexity"] = contextFactor

	// Calculate overall complexity score (weighted average)
	weights := map[string]float64{
		"query_length":         0.2,
		"complexity_keywords":  0.3,
		"multi_step":          0.2,
		"technical_complexity": 0.2,
		"context_complexity":   0.1,
	}

	totalScore := 0.0
	for factor, value := range analysis.Factors {
		if weight, exists := weights[factor]; exists {
			totalScore += value * weight
		}
	}

	analysis.Score = totalScore

	// Determine recommended mode and estimates
	switch {
	case analysis.Score < 0.3:
		analysis.RecommendedMode = "simple"
		analysis.EstimatedAgents = 1
		analysis.EstimatedTokens = 1000
		analysis.Reasoning = "Simple query requiring single agent execution"
	case analysis.Score < 0.6:
		analysis.RecommendedMode = "standard"
		analysis.EstimatedAgents = 2
		analysis.EstimatedTokens = 3000
		analysis.Reasoning = "Moderate complexity requiring coordinated execution"
	case analysis.Score < 0.8:
		analysis.RecommendedMode = "complex"
		analysis.EstimatedAgents = 3
		analysis.EstimatedTokens = 6000
		analysis.Reasoning = "High complexity requiring multi-agent coordination"
	default:
		analysis.RecommendedMode = "exploratory"
		analysis.EstimatedAgents = 5
		analysis.EstimatedTokens = 10000
		analysis.Reasoning = "Very high complexity requiring exploratory understanding"
	}

	// Cache the analysis
	a.cacheManager.Set(ctx, cacheKey, analysis, time.Hour)

	logger.Info("Complexity analysis completed",
		"score", analysis.Score,
		"mode", analysis.RecommendedMode,
		"estimated_agents", analysis.EstimatedAgents,
	)

	return analysis, nil
}

// DecomposeTask breaks down a complex task into agent tasks
func (a *Activities) DecomposeTask(ctx context.Context, input TaskInput, complexity ComplexityAnalysis) (TaskDecomposition, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Decomposing task", "mode", complexity.RecommendedMode)

	decomposition := TaskDecomposition{
		Mode:            complexity.RecommendedMode,
		ComplexityScore: complexity.Score,
		AgentTasks:      make([]AgentTask, 0),
		EstimatedTokens: complexity.EstimatedTokens,
	}

	switch complexity.RecommendedMode {
	case "simple":
		decomposition.AgentTasks = a.createSimpleDecomposition(input)
	case "standard":
		decomposition.AgentTasks = a.createStandardDecomposition(input)
	case "complex":
		decomposition.AgentTasks = a.createComplexDecomposition(input)
	case "exploratory":
		decomposition.AgentTasks = a.createExploratoryDecomposition(input)
	default:
		return decomposition, fmt.Errorf("unknown execution mode: %s", complexity.RecommendedMode)
	}

	// Build DAG structure
	engine := NewDAGEngine(a.logger)
	decomposition.DAG = engine.BuildDAGStructure(decomposition.AgentTasks)

	// Calculate cost estimate
	decomposition.EstimatedCost = a.calculateEstimatedCost(decomposition.AgentTasks)

	logger.Info("Task decomposition completed",
		"agent_count", len(decomposition.AgentTasks),
		"estimated_cost", decomposition.EstimatedCost,
	)

	return decomposition, nil
}

// ExecuteDAG executes the task DAG
func (a *Activities) ExecuteDAG(ctx context.Context, decomposition TaskDecomposition, input TaskInput) (map[string]interface{}, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Executing DAG", "agent_count", len(decomposition.AgentTasks))

	// Reserve budget
	reservationID, err := a.budgetManager.ReserveBudget(ctx, input.UserID, input.TenantID, decomposition.EstimatedTokens)
	if err != nil {
		return nil, fmt.Errorf("failed to reserve budget: %w", err)
	}
	defer a.budgetManager.ReleaseBudget(ctx, reservationID)

	results := make(map[string]interface{})
	metrics := map[string]interface{}{
		"total_tokens":     0,
		"total_cost_usd":   0.0,
		"tool_calls_count": 0,
		"agent_results":    make(map[string]interface{}),
	}

	// Execute tasks level by level
	levels := a.organizeLevels(decomposition.DAG)
	
	for level, taskIDs := range levels {
		logger.Info("Executing level", "level", level, "task_count", len(taskIDs))
		
		levelResults := make(map[string]*AgentExecutionResult)
		
		// Execute tasks in parallel within the same level
		for _, taskID := range taskIDs {
			task := a.findTaskByID(decomposition.AgentTasks, taskID)
			if task == nil {
				continue
			}

			// Prepare agent execution request
			request := AgentExecutionRequest{
				AgentID:     task.ID,
				AgentType:   task.AgentType,
				Query:       task.Query,
				Context:     a.mergeContext(task.Context, results),
				Tools:       task.Tools,
				TokenBudget: task.TokenBudget,
				UserID:      input.UserID,
				TenantID:    input.TenantID,
				SessionID:   input.SessionID,
				Metadata:    task.Metadata,
			}

			// Execute agent
			result, err := a.agentClient.ExecuteAgent(ctx, request)
			if err != nil {
				logger.Error("Agent execution failed", "agent_id", task.ID, "error", err)
				result = &AgentExecutionResult{
					AgentID:      task.ID,
					Status:       "failed",
					ErrorMessage: err.Error(),
				}
			}

			levelResults[taskID] = result
			
			// Update metrics
			if tokens, ok := metrics["total_tokens"].(int); ok {
				metrics["total_tokens"] = tokens + result.TokensUsed
			}
			if cost, ok := metrics["total_cost_usd"].(float64); ok {
				metrics["total_cost_usd"] = cost + result.CostUSD
			}
			if toolCalls, ok := metrics["tool_calls_count"].(int); ok {
				metrics["tool_calls_count"] = toolCalls + len(result.ToolCalls)
			}
			
			// Store agent result
			if agentResults, ok := metrics["agent_results"].(map[string]interface{}); ok {
				agentResults[task.ID] = result
			}
		}
		
		// Merge level results into overall results
		for taskID, result := range levelResults {
			results[taskID] = result.Result
		}
	}

	// Consume actual budget
	if totalTokens, ok := metrics["total_tokens"].(int); ok {
		if totalCost, ok := metrics["total_cost_usd"].(float64); ok {
			err = a.budgetManager.ConsumeBudget(ctx, reservationID, totalTokens, totalCost)
			if err != nil {
				logger.Warn("Failed to consume budget", "error", err)
			}
		}
	}

	results["metrics"] = metrics

	logger.Info("DAG execution completed",
		"total_tokens", metrics["total_tokens"],
		"total_cost_usd", metrics["total_cost_usd"],
	)

	return results, nil
}

// SynthesizeResults combines agent results into a final result
func (a *Activities) SynthesizeResults(ctx context.Context, executionResult map[string]interface{}, input TaskInput) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Synthesizing results")

	// Extract agent results
	agentResults := make(map[string]*AgentExecutionResult)
	if metrics, ok := executionResult["metrics"].(map[string]interface{}); ok {
		if results, ok := metrics["agent_results"].(map[string]interface{}); ok {
			for agentID, result := range results {
				if agentResult, ok := result.(*AgentExecutionResult); ok {
					agentResults[agentID] = agentResult
				}
			}
		}
	}

	// Simple synthesis strategy - combine all successful results
	var successfulResults []string
	for _, result := range agentResults {
		if result.Status == "completed" && result.Result != "" {
			successfulResults = append(successfulResults, result.Result)
		}
	}

	if len(successfulResults) == 0 {
		return "", fmt.Errorf("no successful agent results to synthesize")
	}

	// For now, use simple concatenation with formatting
	synthesized := fmt.Sprintf("Task: %s\n\nResults:\n", input.Query)
	for i, result := range successfulResults {
		synthesized += fmt.Sprintf("\n%d. %s\n", i+1, result)
	}

	logger.Info("Results synthesis completed", "result_count", len(successfulResults))

	return synthesized, nil
}

// Helper methods

func (a *Activities) createSimpleDecomposition(input TaskInput) []AgentTask {
	return []AgentTask{
		{
			ID:          "simple_agent",
			AgentType:   "general",
			Query:       input.Query,
			Context:     input.Context,
			Dependencies: []string{},
			Priority:    1,
			TokenBudget: input.TokenBudget,
			Tools:       []string{"search", "analyze"},
			Metadata:    map[string]interface{}{"mode": "simple"},
		},
	}
}

func (a *Activities) createStandardDecomposition(input TaskInput) []AgentTask {
	tokenBudgetPerAgent := input.TokenBudget / 2
	
	return []AgentTask{
		{
			ID:          "research_agent",
			AgentType:   "researcher",
			Query:       fmt.Sprintf("Research and gather information about: %s", input.Query),
			Context:     input.Context,
			Dependencies: []string{},
			Priority:    1,
			TokenBudget: tokenBudgetPerAgent,
			Tools:       []string{"search", "web_browse", "document_read"},
			Metadata:    map[string]interface{}{"role": "researcher"},
		},
		{
			ID:          "analyzer_agent",
			AgentType:   "analyzer",
			Query:       fmt.Sprintf("Analyze the research findings and provide insights for: %s", input.Query),
			Context:     input.Context,
			Dependencies: []string{"research_agent"},
			Priority:    2,
			TokenBudget: tokenBudgetPerAgent,
			Tools:       []string{"analyze", "synthesize"},
			Metadata:    map[string]interface{}{"role": "analyzer"},
		},
	}
}

func (a *Activities) createComplexDecomposition(input TaskInput) []AgentTask {
	tokenBudgetPerAgent := input.TokenBudget / 3
	
	return []AgentTask{
		{
			ID:          "planner_agent",
			AgentType:   "planner",
			Query:       fmt.Sprintf("Create a plan to address: %s", input.Query),
			Context:     input.Context,
			Dependencies: []string{},
			Priority:    1,
			TokenBudget: tokenBudgetPerAgent,
			Tools:       []string{"plan", "decompose"},
			Metadata:    map[string]interface{}{"role": "planner"},
		},
		{
			ID:          "executor_agent",
			AgentType:   "executor",
			Query:       fmt.Sprintf("Execute the plan for: %s", input.Query),
			Context:     input.Context,
			Dependencies: []string{"planner_agent"},
			Priority:    2,
			TokenBudget: tokenBudgetPerAgent,
			Tools:       []string{"execute", "implement", "code"},
			Metadata:    map[string]interface{}{"role": "executor"},
		},
		{
			ID:          "validator_agent",
			AgentType:   "validator",
			Query:       fmt.Sprintf("Validate and review the execution for: %s", input.Query),
			Context:     input.Context,
			Dependencies: []string{"executor_agent"},
			Priority:    3,
			TokenBudget: tokenBudgetPerAgent,
			Tools:       []string{"validate", "test", "review"},
			Metadata:    map[string]interface{}{"role": "validator"},
		},
	}
}

func (a *Activities) createExploratoryDecomposition(input TaskInput) []AgentTask {
	tokenBudgetPerAgent := input.TokenBudget / 5
	
	return []AgentTask{
		{
			ID:          "hypothesis_generator",
			AgentType:   "generator",
			Query:       fmt.Sprintf("Generate hypotheses for: %s", input.Query),
			Context:     input.Context,
			Dependencies: []string{},
			Priority:    1,
			TokenBudget: tokenBudgetPerAgent,
			Tools:       []string{"generate", "hypothesize"},
			Metadata:    map[string]interface{}{"role": "hypothesis_generator"},
		},
		{
			ID:          "evidence_collector_1",
			AgentType:   "collector",
			Query:       fmt.Sprintf("Collect evidence for hypothesis 1 regarding: %s", input.Query),
			Context:     input.Context,
			Dependencies: []string{"hypothesis_generator"},
			Priority:    2,
			TokenBudget: tokenBudgetPerAgent,
			Tools:       []string{"search", "collect", "verify"},
			Metadata:    map[string]interface{}{"role": "evidence_collector", "hypothesis_index": 1},
		},
		{
			ID:          "evidence_collector_2",
			AgentType:   "collector",
			Query:       fmt.Sprintf("Collect evidence for hypothesis 2 regarding: %s", input.Query),
			Context:     input.Context,
			Dependencies: []string{"hypothesis_generator"},
			Priority:    2,
			TokenBudget: tokenBudgetPerAgent,
			Tools:       []string{"search", "collect", "verify"},
			Metadata:    map[string]interface{}{"role": "evidence_collector", "hypothesis_index": 2},
		},
		{
			ID:          "belief_updater",
			AgentType:   "updater",
			Query:       fmt.Sprintf("Update beliefs based on evidence for: %s", input.Query),
			Context:     input.Context,
			Dependencies: []string{"evidence_collector_1", "evidence_collector_2"},
			Priority:    3,
			TokenBudget: tokenBudgetPerAgent,
			Tools:       []string{"update", "correlate", "synthesize"},
			Metadata:    map[string]interface{}{"role": "belief_updater"},
		},
		{
			ID:          "conclusion_synthesizer",
			AgentType:   "synthesizer",
			Query:       fmt.Sprintf("Synthesize conclusions for: %s", input.Query),
			Context:     input.Context,
			Dependencies: []string{"belief_updater"},
			Priority:    4,
			TokenBudget: tokenBudgetPerAgent,
			Tools:       []string{"synthesize", "conclude"},
			Metadata:    map[string]interface{}{"role": "conclusion_synthesizer"},
		},
	}
}

func (a *Activities) calculateEstimatedCost(tasks []AgentTask) float64 {
	totalCost := 0.0
	costPerToken := 0.002 // Average cost per token
	
	for _, task := range tasks {
		totalCost += float64(task.TokenBudget) * costPerToken
	}
	
	return totalCost
}

func (a *Activities) organizeLevels(dag DAGStructure) map[int][]string {
	levels := make(map[int][]string)
	
	for _, node := range dag.Nodes {
		if levels[node.Level] == nil {
			levels[node.Level] = make([]string, 0)
		}
		levels[node.Level] = append(levels[node.Level], node.TaskID)
	}
	
	return levels
}

func (a *Activities) findTaskByID(tasks []AgentTask, taskID string) *AgentTask {
	for i := range tasks {
		if tasks[i].ID == taskID {
			return &tasks[i]
		}
	}
	return nil
}

func (a *Activities) mergeContext(taskContext map[string]interface{}, results map[string]interface{}) map[string]interface{} {
	merged := make(map[string]interface{})
	
	// Copy task context
	for k, v := range taskContext {
		merged[k] = v
	}
	
	// Add previous results
	merged["previous_results"] = results
	
	return merged
}

func hashString(s string) string {
	// Simple hash function for caching
	hash := 0
	for _, c := range s {
		hash = hash*31 + int(c)
	}
	return fmt.Sprintf("%x", hash)
}