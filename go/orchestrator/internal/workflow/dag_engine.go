package workflow

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

// DAGEngine handles task decomposition and execution orchestration
type DAGEngine struct {
	logger *zap.Logger
}

// NewDAGEngine creates a new DAG engine
func NewDAGEngine(logger *zap.Logger) *DAGEngine {
	return &DAGEngine{
		logger: logger,
	}
}

// TaskInput represents input for task execution
type TaskInput struct {
	WorkflowID     string                 `json:"workflow_id"`
	UserID         string                 `json:"user_id"`
	TenantID       string                 `json:"tenant_id"`
	SessionID      string                 `json:"session_id"`
	Query          string                 `json:"query"`
	Mode           string                 `json:"mode"`
	Context        map[string]interface{} `json:"context"`
	TokenBudget    int                    `json:"token_budget"`
	MaxAgents      int                    `json:"max_agents"`
	TimeoutSeconds int                    `json:"timeout_seconds"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// TaskResult represents the result of task execution
type TaskResult struct {
	WorkflowID      string                 `json:"workflow_id"`
	Status          string                 `json:"status"`
	Result          string                 `json:"result"`
	ErrorMessage    string                 `json:"error_message"`
	TotalTokens     int                    `json:"total_tokens"`
	TotalCostUSD    float64                `json:"total_cost_usd"`
	DurationMs      int64                  `json:"duration_ms"`
	AgentCount      int                    `json:"agent_count"`
	ToolCallsCount  int                    `json:"tool_calls_count"`
	ComplexityScore float64                `json:"complexity_score"`
	Metadata        map[string]interface{} `json:"metadata"`
	CreatedAt       time.Time              `json:"created_at"`
	CompletedAt     time.Time              `json:"completed_at"`
}

// TaskDecomposition represents the breakdown of a complex task
type TaskDecomposition struct {
	Mode            string      `json:"mode"`
	ComplexityScore float64     `json:"complexity_score"`
	AgentTasks      []AgentTask `json:"agent_tasks"`
	DAG             DAGStructure `json:"dag"`
	EstimatedTokens int         `json:"estimated_tokens"`
	EstimatedCost   float64     `json:"estimated_cost"`
}

// AgentTask represents a task for a specific agent
type AgentTask struct {
	ID           string                 `json:"id"`
	AgentType    string                 `json:"agent_type"`
	Query        string                 `json:"query"`
	Context      map[string]interface{} `json:"context"`
	Dependencies []string               `json:"dependencies"`
	Priority     int                    `json:"priority"`
	TokenBudget  int                    `json:"token_budget"`
	Tools        []string               `json:"tools"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// DAGStructure represents the execution graph
type DAGStructure struct {
	Nodes []DAGNode `json:"nodes"`
	Edges []DAGEdge `json:"edges"`
}

// DAGNode represents a node in the execution graph
type DAGNode struct {
	ID       string `json:"id"`
	TaskID   string `json:"task_id"`
	Level    int    `json:"level"`
	Parallel bool   `json:"parallel"`
}

// DAGEdge represents a dependency edge
type DAGEdge struct {
	From string `json:"from"`
	To   string `json:"to"`
	Type string `json:"type"`
}

// ComplexityAnalysis represents task complexity analysis
type ComplexityAnalysis struct {
	Score           float64                `json:"score"`
	Factors         map[string]float64     `json:"factors"`
	RecommendedMode string                 `json:"recommended_mode"`
	EstimatedAgents int                    `json:"estimated_agents"`
	EstimatedTokens int                    `json:"estimated_tokens"`
	Reasoning       string                 `json:"reasoning"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// ExecuteDAGWorkflow is the main workflow for DAG execution
func (d *DAGEngine) ExecuteDAGWorkflow(ctx workflow.Context, input TaskInput) (TaskResult, error) {
	logger := workflow.GetLogger(ctx)
	startTime := workflow.Now(ctx)
	
	logger.Info("Starting DAG workflow execution",
		"workflow_id", input.WorkflowID,
		"user_id", input.UserID,
		"tenant_id", input.TenantID,
		"query", input.Query,
		"mode", input.Mode,
	)

	// Set workflow timeout
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: time.Duration(input.TimeoutSeconds) * time.Second,
	})

	result := TaskResult{
		WorkflowID: input.WorkflowID,
		Status:     "running",
		CreatedAt:  startTime,
		Metadata:   make(map[string]interface{}),
	}

	// Step 1: Analyze task complexity
	var complexity ComplexityAnalysis
	err := workflow.ExecuteActivity(ctx, "AnalyzeComplexity", input).Get(ctx, &complexity)
	if err != nil {
		logger.Error("Failed to analyze complexity", "error", err)
		result.Status = "failed"
		result.ErrorMessage = fmt.Sprintf("Complexity analysis failed: %v", err)
		result.CompletedAt = workflow.Now(ctx)
		return result, err
	}

	result.ComplexityScore = complexity.Score
	logger.Info("Complexity analysis completed", "score", complexity.Score, "mode", complexity.RecommendedMode)

	// Step 2: Task decomposition
	var decomposition TaskDecomposition
	err = workflow.ExecuteActivity(ctx, "DecomposeTask", input, complexity).Get(ctx, &decomposition)
	if err != nil {
		logger.Error("Failed to decompose task", "error", err)
		result.Status = "failed"
		result.ErrorMessage = fmt.Sprintf("Task decomposition failed: %v", err)
		result.CompletedAt = workflow.Now(ctx)
		return result, err
	}

	result.AgentCount = len(decomposition.AgentTasks)
	logger.Info("Task decomposition completed", "agent_count", result.AgentCount)

	// Step 3: Execute DAG
	var executionResult map[string]interface{}
	err = workflow.ExecuteActivity(ctx, "ExecuteDAG", decomposition, input).Get(ctx, &executionResult)
	if err != nil {
		logger.Error("Failed to execute DAG", "error", err)
		result.Status = "failed"
		result.ErrorMessage = fmt.Sprintf("DAG execution failed: %v", err)
		result.CompletedAt = workflow.Now(ctx)
		return result, err
	}

	// Step 4: Synthesize results
	var finalResult string
	err = workflow.ExecuteActivity(ctx, "SynthesizeResults", executionResult, input).Get(ctx, &finalResult)
	if err != nil {
		logger.Error("Failed to synthesize results", "error", err)
		result.Status = "failed"
		result.ErrorMessage = fmt.Sprintf("Result synthesis failed: %v", err)
		result.CompletedAt = workflow.Now(ctx)
		return result, err
	}

	// Complete the workflow
	result.Status = "completed"
	result.Result = finalResult
	result.CompletedAt = workflow.Now(ctx)
	result.DurationMs = result.CompletedAt.Sub(startTime).Milliseconds()

	// Extract metrics from execution result
	if metrics, ok := executionResult["metrics"].(map[string]interface{}); ok {
		if tokens, ok := metrics["total_tokens"].(float64); ok {
			result.TotalTokens = int(tokens)
		}
		if cost, ok := metrics["total_cost_usd"].(float64); ok {
			result.TotalCostUSD = cost
		}
		if toolCalls, ok := metrics["tool_calls_count"].(float64); ok {
			result.ToolCallsCount = int(toolCalls)
		}
	}

	logger.Info("DAG workflow execution completed",
		"workflow_id", input.WorkflowID,
		"status", result.Status,
		"duration_ms", result.DurationMs,
		"total_tokens", result.TotalTokens,
		"total_cost_usd", result.TotalCostUSD,
	)

	return result, nil
}

// ExploratoryCoordinationWorkflow implements exploratory understanding coordination
func (d *DAGEngine) ExploratoryCoordinationWorkflow(ctx workflow.Context, input TaskInput) (TaskResult, error) {
	logger := workflow.GetLogger(ctx)
	startTime := workflow.Now(ctx)
	
	logger.Info("Starting exploratory coordination workflow",
		"workflow_id", input.WorkflowID,
		"query", input.Query,
	)

	result := TaskResult{
		WorkflowID: input.WorkflowID,
		Status:     "running",
		CreatedAt:  startTime,
		Metadata:   make(map[string]interface{}),
	}

	// Step 1: Generate competing hypotheses
	var hypotheses []Hypothesis
	err := workflow.ExecuteActivity(ctx, "GenerateHypotheses", input).Get(ctx, &hypotheses)
	if err != nil {
		logger.Error("Failed to generate hypotheses", "error", err)
		result.Status = "failed"
		result.ErrorMessage = fmt.Sprintf("Hypothesis generation failed: %v", err)
		return result, err
	}

	logger.Info("Generated hypotheses", "count", len(hypotheses))

	// Step 2: Parallel hypothesis testing
	var evidenceResults []Evidence
	futures := make([]workflow.Future, len(hypotheses))
	
	for i, hypothesis := range hypotheses {
		futures[i] = workflow.ExecuteActivity(ctx, "TestHypothesis", hypothesis, input)
	}

	// Wait for all hypothesis tests to complete
	for i, future := range futures {
		var evidence Evidence
		err := future.Get(ctx, &evidence)
		if err != nil {
			logger.Warn("Hypothesis testing failed", "hypothesis_id", hypotheses[i].ID, "error", err)
			continue
		}
		evidenceResults = append(evidenceResults, evidence)
	}

	// Step 3: Update belief state
	var beliefState BeliefState
	err = workflow.ExecuteActivity(ctx, "UpdateBeliefState", hypotheses, evidenceResults).Get(ctx, &beliefState)
	if err != nil {
		logger.Error("Failed to update belief state", "error", err)
		result.Status = "failed"
		result.ErrorMessage = fmt.Sprintf("Belief state update failed: %v", err)
		return result, err
	}

	// Step 4: Check if more exploration is needed
	var shouldContinue bool
	err = workflow.ExecuteActivity(ctx, "ShouldContinueExploration", beliefState).Get(ctx, &shouldContinue)
	if err != nil {
		logger.Error("Failed to check exploration continuation", "error", err)
		result.Status = "failed"
		result.ErrorMessage = fmt.Sprintf("Exploration check failed: %v", err)
		return result, err
	}

	// Step 5: Continue exploration if needed (recursive call with updated context)
	if shouldContinue && result.DurationMs < int64(float64(input.TimeoutSeconds*1000)*0.8) {
		logger.Info("Continuing exploration with updated context")
		
		// Update input context with current findings
		updatedInput := input
		updatedInput.Context["previous_hypotheses"] = hypotheses
		updatedInput.Context["evidence"] = evidenceResults
		updatedInput.Context["belief_state"] = beliefState
		
		// Recursive call
		return d.ExploratoryCoordinationWorkflow(ctx, updatedInput)
	}

	// Step 6: Synthesize final result
	var finalResult string
	err = workflow.ExecuteActivity(ctx, "SynthesizeExploratoryResult", beliefState, evidenceResults).Get(ctx, &finalResult)
	if err != nil {
		logger.Error("Failed to synthesize exploratory result", "error", err)
		result.Status = "failed"
		result.ErrorMessage = fmt.Sprintf("Result synthesis failed: %v", err)
		return result, err
	}

	result.Status = "completed"
	result.Result = finalResult
	result.CompletedAt = workflow.Now(ctx)
	result.DurationMs = result.CompletedAt.Sub(startTime).Milliseconds()

	logger.Info("Exploratory coordination workflow completed",
		"workflow_id", input.WorkflowID,
		"final_confidence", beliefState.BestHypothesis.Confidence,
		"evidence_count", len(evidenceResults),
	)

	return result, nil
}

// Hypothesis represents a hypothesis in exploratory understanding
type Hypothesis struct {
	ID          string                 `json:"id"`
	Text        string                 `json:"text"`
	Confidence  float64                `json:"confidence"`
	Evidence    []string               `json:"evidence"`
	Context     map[string]interface{} `json:"context"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// Evidence represents evidence for or against a hypothesis
type Evidence struct {
	ID           string                 `json:"id"`
	HypothesisID string                 `json:"hypothesis_id"`
	Text         string                 `json:"text"`
	Type         string                 `json:"type"` // supporting, contradicting, neutral
	Strength     float64                `json:"strength"`
	Reliability  float64                `json:"reliability"`
	Source       string                 `json:"source"`
	Context      map[string]interface{} `json:"context"`
	CreatedAt    time.Time              `json:"created_at"`
}

// BeliefState represents the current belief state
type BeliefState struct {
	Hypotheses      []Hypothesis `json:"hypotheses"`
	Evidence        []Evidence   `json:"evidence"`
	BestHypothesis  Hypothesis   `json:"best_hypothesis"`
	Confidence      float64      `json:"confidence"`
	Contradictions  []string     `json:"contradictions"`
	UpdatedAt       time.Time    `json:"updated_at"`
}

// P2PCoordinationWorkflow implements peer-to-peer agent coordination
func (d *DAGEngine) P2PCoordinationWorkflow(ctx workflow.Context, input TaskInput) (TaskResult, error) {
	logger := workflow.GetLogger(ctx)
	startTime := workflow.Now(ctx)
	
	logger.Info("Starting P2P coordination workflow",
		"workflow_id", input.WorkflowID,
		"max_agents", input.MaxAgents,
	)

	result := TaskResult{
		WorkflowID: input.WorkflowID,
		Status:     "running",
		CreatedAt:  startTime,
		Metadata:   make(map[string]interface{}),
	}

	// Step 1: Initialize workspace
	var workspaceID string
	err := workflow.ExecuteActivity(ctx, "InitializeWorkspace", input).Get(ctx, &workspaceID)
	if err != nil {
		logger.Error("Failed to initialize workspace", "error", err)
		result.Status = "failed"
		result.ErrorMessage = fmt.Sprintf("Workspace initialization failed: %v", err)
		return result, err
	}

	// Step 2: Spawn peer agents
	var agentIDs []string
	err = workflow.ExecuteActivity(ctx, "SpawnPeerAgents", workspaceID, input).Get(ctx, &agentIDs)
	if err != nil {
		logger.Error("Failed to spawn peer agents", "error", err)
		result.Status = "failed"
		result.ErrorMessage = fmt.Sprintf("Agent spawning failed: %v", err)
		return result, err
	}

	result.AgentCount = len(agentIDs)
	logger.Info("Spawned peer agents", "count", len(agentIDs))

	// Step 3: Coordinate peer execution
	var coordinationResult map[string]interface{}
	err = workflow.ExecuteActivity(ctx, "CoordinatePeerExecution", workspaceID, agentIDs, input).Get(ctx, &coordinationResult)
	if err != nil {
		logger.Error("Failed to coordinate peer execution", "error", err)
		result.Status = "failed"
		result.ErrorMessage = fmt.Sprintf("Peer coordination failed: %v", err)
		return result, err
	}

	// Step 4: Aggregate results
	var aggregatedResult string
	err = workflow.ExecuteActivity(ctx, "AggregateP2PResults", coordinationResult, input).Get(ctx, &aggregatedResult)
	if err != nil {
		logger.Error("Failed to aggregate P2P results", "error", err)
		result.Status = "failed"
		result.ErrorMessage = fmt.Sprintf("Result aggregation failed: %v", err)
		return result, err
	}

	// Step 5: Cleanup workspace
	err = workflow.ExecuteActivity(ctx, "CleanupWorkspace", workspaceID).Get(ctx, nil)
	if err != nil {
		logger.Warn("Failed to cleanup workspace", "workspace_id", workspaceID, "error", err)
	}

	result.Status = "completed"
	result.Result = aggregatedResult
	result.CompletedAt = workflow.Now(ctx)
	result.DurationMs = result.CompletedAt.Sub(startTime).Milliseconds()

	logger.Info("P2P coordination workflow completed",
		"workflow_id", input.WorkflowID,
		"agent_count", result.AgentCount,
		"duration_ms", result.DurationMs,
	)

	return result, nil
}

// SelectExecutionMode determines the appropriate execution mode based on complexity
func (d *DAGEngine) SelectExecutionMode(complexity ComplexityAnalysis) string {
	switch {
	case complexity.Score < 0.3:
		return "simple"
	case complexity.Score < 0.7:
		return "standard"
	case complexity.Score < 0.9:
		return "complex"
	default:
		return "exploratory"
	}
}

// BuildDAGStructure creates a DAG structure from agent tasks
func (d *DAGEngine) BuildDAGStructure(tasks []AgentTask) DAGStructure {
	dag := DAGStructure{
		Nodes: make([]DAGNode, 0, len(tasks)),
		Edges: make([]DAGEdge, 0),
	}

	// Create nodes
	for i, task := range tasks {
		node := DAGNode{
			ID:       fmt.Sprintf("node_%d", i),
			TaskID:   task.ID,
			Level:    0,
			Parallel: len(task.Dependencies) == 0,
		}
		dag.Nodes = append(dag.Nodes, node)
	}

	// Create edges based on dependencies
	for i, task := range tasks {
		for _, depID := range task.Dependencies {
			for j, depTask := range tasks {
				if depTask.ID == depID {
					edge := DAGEdge{
						From: fmt.Sprintf("node_%d", j),
						To:   fmt.Sprintf("node_%d", i),
						Type: "dependency",
					}
					dag.Edges = append(dag.Edges, edge)
				}
			}
		}
	}

	// Calculate levels for topological ordering
	d.calculateLevels(&dag)

	return dag
}

// calculateLevels assigns levels to nodes for topological ordering
func (d *DAGEngine) calculateLevels(dag *DAGStructure) {
	// Build adjacency list
	inDegree := make(map[string]int)
	adjList := make(map[string][]string)

	for _, node := range dag.Nodes {
		inDegree[node.ID] = 0
		adjList[node.ID] = make([]string, 0)
	}

	for _, edge := range dag.Edges {
		adjList[edge.From] = append(adjList[edge.From], edge.To)
		inDegree[edge.To]++
	}

	// Topological sort with level assignment
	queue := make([]string, 0)
	levels := make(map[string]int)

	// Find nodes with no dependencies
	for nodeID, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, nodeID)
			levels[nodeID] = 0
		}
	}

	// Process nodes level by level
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		for _, neighbor := range adjList[current] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				levels[neighbor] = levels[current] + 1
				queue = append(queue, neighbor)
			}
		}
	}

	// Update node levels
	for i := range dag.Nodes {
		dag.Nodes[i].Level = levels[dag.Nodes[i].ID]
	}
}