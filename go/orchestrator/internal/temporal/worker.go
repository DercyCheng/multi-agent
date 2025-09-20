package temporal

import (
	"context"
	"fmt"
	"time"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.uber.org/zap"

	"github.com/multi-agent/go/orchestrator/internal/budget"
	workflowpkg "github.com/multi-agent/go/orchestrator/internal/workflow"
)

// Config holds Temporal configuration
type Config struct {
	HostPort   string `yaml:"host_port" env:"TEMPORAL_HOST_PORT" default:"localhost:7233"`
	Namespace  string `yaml:"namespace" env:"TEMPORAL_NAMESPACE" default:"default"`
	TaskQueue  string `yaml:"task_queue" env:"TEMPORAL_TASK_QUEUE" default:"multi-agent-task-queue"`
	WorkerName string `yaml:"worker_name" env:"TEMPORAL_WORKER_NAME" default:"multi-agent-worker"`
}

// Worker wraps Temporal worker functionality
type Worker struct {
	client       client.Client
	worker       worker.Worker
	config       Config
	logger       *zap.Logger
	dagEngine    *workflowpkg.DAGEngine
	activities   *workflowpkg.Activities
	budgetMgr    *budget.Manager
}

// NewWorker creates a new Temporal worker
func NewWorker(config Config, logger *zap.Logger, budgetMgr *budget.Manager) (*Worker, error) {
	// Create Temporal client
	c, err := client.Dial(client.Options{
		HostPort:  config.HostPort,
		Namespace: config.Namespace,
		Logger:    NewTemporalLogger(logger),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Temporal client: %w", err)
	}

	// Create worker
	w := worker.New(c, config.TaskQueue, worker.Options{
		MaxConcurrentActivityExecutionSize:   10,
		MaxConcurrentWorkflowTaskExecutionSize: 5,
		EnableLoggingInReplay:             false,
	})

	// Create DAG engine and activities
	dagEngine := workflowpkg.NewDAGEngine(logger)
	
	// Mock implementations for now - these would be replaced with actual implementations
	agentClient := &MockAgentClient{logger: logger}
	cacheManager := &MockCacheManager{logger: logger}
	
	activities := workflowpkg.NewActivities(logger, agentClient, budgetMgr, cacheManager)

	return &Worker{
		client:     c,
		worker:     w,
		config:     config,
		logger:     logger,
		dagEngine:  dagEngine,
		activities: activities,
		budgetMgr:  budgetMgr,
	}, nil
}

// Start starts the Temporal worker
func (w *Worker) Start(ctx context.Context) error {
	w.logger.Info("Starting Temporal worker",
		zap.String("host_port", w.config.HostPort),
		zap.String("namespace", w.config.Namespace),
		zap.String("task_queue", w.config.TaskQueue),
	)

	// Register workflows
	w.worker.RegisterWorkflow(w.dagEngine.ExecuteDAGWorkflow)
	w.worker.RegisterWorkflow(w.dagEngine.ExploratoryCoordinationWorkflow)
	w.worker.RegisterWorkflow(w.dagEngine.P2PCoordinationWorkflow)

	// Register activities
	w.worker.RegisterActivity(w.activities.AnalyzeComplexity)
	w.worker.RegisterActivity(w.activities.DecomposeTask)
	w.worker.RegisterActivity(w.activities.ExecuteDAG)
	w.worker.RegisterActivity(w.activities.SynthesizeResults)

	// Register exploratory activities
	w.worker.RegisterActivity(w.GenerateHypotheses)
	w.worker.RegisterActivity(w.TestHypothesis)
	w.worker.RegisterActivity(w.UpdateBeliefState)
	w.worker.RegisterActivity(w.ShouldContinueExploration)
	w.worker.RegisterActivity(w.SynthesizeExploratoryResult)

	// Register P2P activities
	w.worker.RegisterActivity(w.InitializeWorkspace)
	w.worker.RegisterActivity(w.SpawnPeerAgents)
	w.worker.RegisterActivity(w.CoordinatePeerExecution)
	w.worker.RegisterActivity(w.AggregateP2PResults)
	w.worker.RegisterActivity(w.CleanupWorkspace)

	// Start worker
	err := w.worker.Start()
	if err != nil {
		return fmt.Errorf("failed to start Temporal worker: %w", err)
	}

	w.logger.Info("Temporal worker started successfully")

	// Start cleanup routine
	go w.startCleanupRoutine(ctx)

	return nil
}

// Stop stops the Temporal worker
func (w *Worker) Stop() {
	w.logger.Info("Stopping Temporal worker")
	w.worker.Stop()
	w.client.Close()
}

// ExecuteWorkflow executes a workflow
func (w *Worker) ExecuteWorkflow(ctx context.Context, workflowType string, input workflowpkg.TaskInput) (*workflowpkg.TaskResult, error) {
	w.logger.Info("Executing workflow",
		zap.String("workflow_type", workflowType),
		zap.String("workflow_id", input.WorkflowID),
	)

	options := client.StartWorkflowOptions{
		ID:                 input.WorkflowID,
		TaskQueue:          w.config.TaskQueue,
		WorkflowRunTimeout: time.Duration(input.TimeoutSeconds) * time.Second,
	}

	var workflowFunc interface{}
	switch workflowType {
	case "dag":
		workflowFunc = w.dagEngine.ExecuteDAGWorkflow
	case "exploratory":
		workflowFunc = w.dagEngine.ExploratoryCoordinationWorkflow
	case "p2p":
		workflowFunc = w.dagEngine.P2PCoordinationWorkflow
	default:
		return nil, fmt.Errorf("unknown workflow type: %s", workflowType)
	}

	we, err := w.client.ExecuteWorkflow(ctx, options, workflowFunc, input)
	if err != nil {
		return nil, fmt.Errorf("failed to execute workflow: %w", err)
	}

	var result workflowpkg.TaskResult
	err = we.Get(ctx, &result)
	if err != nil {
		return nil, fmt.Errorf("workflow execution failed: %w", err)
	}

	w.logger.Info("Workflow execution completed",
		zap.String("workflow_id", input.WorkflowID),
		zap.String("status", result.Status),
		zap.Int64("duration_ms", result.DurationMs),
	)

	return &result, nil
}

// GetWorkflowStatus gets the status of a running workflow
func (w *Worker) GetWorkflowStatus(ctx context.Context, workflowID string) (*WorkflowStatus, error) {
	we := w.client.GetWorkflow(ctx, workflowID, "")
	
	// Try to get workflow result
	var result workflowpkg.TaskResult
	err := we.Get(ctx, &result)
	
	status := &WorkflowStatus{
		WorkflowID: workflowID,
	}
	
	if err != nil {
		// Workflow might still be running or failed
		status.Status = "Running"
	} else {
		// Workflow completed successfully
		status.Status = "Completed"
		status.Result = &result
	}

	return status, nil
}

// WorkflowStatus represents workflow execution status
type WorkflowStatus struct {
	WorkflowID    string                    `json:"workflow_id"`
	Status        string                    `json:"status"`
	StartTime     *time.Time                `json:"start_time"`
	CloseTime     *time.Time                `json:"close_time"`
	ExecutionTime *time.Time                `json:"execution_time"`
	Result        *workflowpkg.TaskResult   `json:"result,omitempty"`
}

// Exploratory workflow activities

// GenerateHypotheses generates competing hypotheses for exploratory understanding
func (w *Worker) GenerateHypotheses(ctx context.Context, input workflowpkg.TaskInput) ([]workflowpkg.Hypothesis, error) {
	logger := w.logger.With(zap.String("activity", "GenerateHypotheses"))
	logger.Info("Generating hypotheses", zap.String("query", input.Query))

	// Mock implementation - in real system, this would call LLM service
	hypotheses := []workflowpkg.Hypothesis{
		{
			ID:         "hyp_1",
			Text:       fmt.Sprintf("Hypothesis 1: %s requires approach A", input.Query),
			Confidence: 0.6,
			Evidence:   []string{},
			Context:    input.Context,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		},
		{
			ID:         "hyp_2",
			Text:       fmt.Sprintf("Hypothesis 2: %s requires approach B", input.Query),
			Confidence: 0.5,
			Evidence:   []string{},
			Context:    input.Context,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		},
		{
			ID:         "hyp_3",
			Text:       fmt.Sprintf("Hypothesis 3: %s requires hybrid approach", input.Query),
			Confidence: 0.7,
			Evidence:   []string{},
			Context:    input.Context,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		},
	}

	logger.Info("Generated hypotheses", zap.Int("count", len(hypotheses)))
	return hypotheses, nil
}

// TestHypothesis tests a hypothesis and collects evidence
func (w *Worker) TestHypothesis(ctx context.Context, hypothesis workflowpkg.Hypothesis, input workflowpkg.TaskInput) (workflowpkg.Evidence, error) {
	logger := w.logger.With(zap.String("activity", "TestHypothesis"))
	logger.Info("Testing hypothesis", zap.String("hypothesis_id", hypothesis.ID))

	// Mock implementation - in real system, this would execute agents to gather evidence
	evidence := workflowpkg.Evidence{
		ID:           fmt.Sprintf("ev_%s_%d", hypothesis.ID, time.Now().Unix()),
		HypothesisID: hypothesis.ID,
		Text:         fmt.Sprintf("Evidence for %s: Supporting data found", hypothesis.Text),
		Type:         "supporting",
		Strength:     0.8,
		Reliability:  0.9,
		Source:       "test_agent",
		Context:      input.Context,
		CreatedAt:    time.Now(),
	}

	logger.Info("Hypothesis tested", zap.String("evidence_id", evidence.ID))
	return evidence, nil
}

// UpdateBeliefState updates the belief state with new evidence
func (w *Worker) UpdateBeliefState(ctx context.Context, hypotheses []workflowpkg.Hypothesis, evidence []workflowpkg.Evidence) (workflowpkg.BeliefState, error) {
	logger := w.logger.With(zap.String("activity", "UpdateBeliefState"))
	logger.Info("Updating belief state", zap.Int("hypotheses", len(hypotheses)), zap.Int("evidence", len(evidence)))

	// Update hypothesis confidence based on evidence
	for i := range hypotheses {
		for _, ev := range evidence {
			if ev.HypothesisID == hypotheses[i].ID {
				if ev.Type == "supporting" {
					hypotheses[i].Confidence += ev.Strength * ev.Reliability * 0.1
				} else if ev.Type == "contradicting" {
					hypotheses[i].Confidence -= ev.Strength * ev.Reliability * 0.1
				}
				hypotheses[i].Evidence = append(hypotheses[i].Evidence, ev.ID)
			}
		}
		// Clamp confidence between 0 and 1
		if hypotheses[i].Confidence > 1.0 {
			hypotheses[i].Confidence = 1.0
		}
		if hypotheses[i].Confidence < 0.0 {
			hypotheses[i].Confidence = 0.0
		}
	}

	// Find best hypothesis
	bestHypothesis := hypotheses[0]
	for _, h := range hypotheses {
		if h.Confidence > bestHypothesis.Confidence {
			bestHypothesis = h
		}
	}

	beliefState := workflowpkg.BeliefState{
		Hypotheses:     hypotheses,
		Evidence:       evidence,
		BestHypothesis: bestHypothesis,
		Confidence:     bestHypothesis.Confidence,
		Contradictions: []string{}, // Would detect contradictions in real implementation
		UpdatedAt:      time.Now(),
	}

	logger.Info("Belief state updated", zap.Float64("best_confidence", bestHypothesis.Confidence))
	return beliefState, nil
}

// ShouldContinueExploration determines if more exploration is needed
func (w *Worker) ShouldContinueExploration(ctx context.Context, beliefState workflowpkg.BeliefState) (bool, error) {
	logger := w.logger.With(zap.String("activity", "ShouldContinueExploration"))
	
	confidenceThreshold := 0.8
	shouldContinue := beliefState.BestHypothesis.Confidence < confidenceThreshold
	
	logger.Info("Exploration decision",
		zap.Float64("confidence", beliefState.BestHypothesis.Confidence),
		zap.Float64("threshold", confidenceThreshold),
		zap.Bool("continue", shouldContinue),
	)
	
	return shouldContinue, nil
}

// SynthesizeExploratoryResult synthesizes the final exploratory result
func (w *Worker) SynthesizeExploratoryResult(ctx context.Context, beliefState workflowpkg.BeliefState, evidence []workflowpkg.Evidence) (string, error) {
	logger := w.logger.With(zap.String("activity", "SynthesizeExploratoryResult"))
	logger.Info("Synthesizing exploratory result")

	result := fmt.Sprintf("Exploratory Analysis Results:\n\n")
	result += fmt.Sprintf("Best Hypothesis: %s\n", beliefState.BestHypothesis.Text)
	result += fmt.Sprintf("Confidence: %.2f\n\n", beliefState.BestHypothesis.Confidence)
	result += fmt.Sprintf("Supporting Evidence:\n")
	
	for _, ev := range evidence {
		if ev.HypothesisID == beliefState.BestHypothesis.ID && ev.Type == "supporting" {
			result += fmt.Sprintf("- %s (Strength: %.2f, Reliability: %.2f)\n", ev.Text, ev.Strength, ev.Reliability)
		}
	}

	logger.Info("Exploratory result synthesized")
	return result, nil
}

// P2P workflow activities

// InitializeWorkspace initializes a P2P workspace
func (w *Worker) InitializeWorkspace(ctx context.Context, input workflowpkg.TaskInput) (string, error) {
	logger := w.logger.With(zap.String("activity", "InitializeWorkspace"))
	workspaceID := fmt.Sprintf("workspace_%s_%d", input.TenantID, time.Now().Unix())
	logger.Info("Initializing workspace", zap.String("workspace_id", workspaceID))
	return workspaceID, nil
}

// SpawnPeerAgents spawns peer agents for P2P coordination
func (w *Worker) SpawnPeerAgents(ctx context.Context, workspaceID string, input workflowpkg.TaskInput) ([]string, error) {
	logger := w.logger.With(zap.String("activity", "SpawnPeerAgents"))
	logger.Info("Spawning peer agents", zap.String("workspace_id", workspaceID))

	agentIDs := make([]string, input.MaxAgents)
	for i := 0; i < input.MaxAgents; i++ {
		agentIDs[i] = fmt.Sprintf("agent_%s_%d", workspaceID, i)
	}

	logger.Info("Peer agents spawned", zap.Int("count", len(agentIDs)))
	return agentIDs, nil
}

// CoordinatePeerExecution coordinates peer agent execution
func (w *Worker) CoordinatePeerExecution(ctx context.Context, workspaceID string, agentIDs []string, input workflowpkg.TaskInput) (map[string]interface{}, error) {
	logger := w.logger.With(zap.String("activity", "CoordinatePeerExecution"))
	logger.Info("Coordinating peer execution", zap.String("workspace_id", workspaceID))

	result := map[string]interface{}{
		"workspace_id": workspaceID,
		"agent_results": make(map[string]string),
		"coordination_success": true,
	}

	// Mock peer coordination
	agentResults := result["agent_results"].(map[string]string)
	for _, agentID := range agentIDs {
		agentResults[agentID] = fmt.Sprintf("Result from %s for query: %s", agentID, input.Query)
	}

	logger.Info("Peer execution coordinated", zap.Int("agents", len(agentIDs)))
	return result, nil
}

// AggregateP2PResults aggregates P2P execution results
func (w *Worker) AggregateP2PResults(ctx context.Context, coordinationResult map[string]interface{}, input workflowpkg.TaskInput) (string, error) {
	logger := w.logger.With(zap.String("activity", "AggregateP2PResults"))
	logger.Info("Aggregating P2P results")

	result := fmt.Sprintf("P2P Coordination Results for: %s\n\n", input.Query)
	
	if agentResults, ok := coordinationResult["agent_results"].(map[string]string); ok {
		for agentID, agentResult := range agentResults {
			result += fmt.Sprintf("Agent %s: %s\n", agentID, agentResult)
		}
	}

	logger.Info("P2P results aggregated")
	return result, nil
}

// CleanupWorkspace cleans up P2P workspace
func (w *Worker) CleanupWorkspace(ctx context.Context, workspaceID string) error {
	logger := w.logger.With(zap.String("activity", "CleanupWorkspace"))
	logger.Info("Cleaning up workspace", zap.String("workspace_id", workspaceID))
	// Mock cleanup
	return nil
}

// startCleanupRoutine starts background cleanup routines
func (w *Worker) startCleanupRoutine(ctx context.Context) {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.budgetMgr.CleanupExpiredReservations(ctx)
		}
	}
}