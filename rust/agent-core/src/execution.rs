use anyhow::{Context, Result};
use std::sync::Arc;
use std::time::{Duration, Instant};
use tokio::sync::Mutex;
use tracing::{debug, error, info, warn};
use uuid::Uuid;

use crate::enforcement::{EnforcementGateway, ExecuteTaskRequest, TaskPriority, ResourceRequirements};
use crate::fsm::{StateMachine, StateMachineContext, Event};
use crate::metrics::MetricsCollector;
use crate::sandbox::{WASISandbox, ExecutionContext, ExecutionResult, ExecutionStatus};
use crate::security::SecurityManager;

/// Main execution engine that coordinates all components
pub struct ExecutionEngine {
    sandbox: Arc<WASISandbox>,
    enforcement: Arc<EnforcementGateway>,
    security: Arc<SecurityManager>,
    state_machine: Arc<StateMachine>,
    metrics: Arc<MetricsCollector>,
    active_executions: Arc<Mutex<std::collections::HashMap<String, ActiveExecution>>>,
}

/// Represents an active execution
#[derive(Debug, Clone)]
struct ActiveExecution {
    execution_id: String,
    user_id: String,
    tenant_id: String,
    session_id: String,
    fsm_instance_id: String,
    started_at: Instant,
    status: ExecutionEngineStatus,
}

/// Status of execution in the engine
#[derive(Debug, Clone, PartialEq)]
enum ExecutionEngineStatus {
    Initializing,
    PolicyCheck,
    Executing,
    Validating,
    Completed,
    Failed,
}

/// Request to execute agent code
#[derive(Debug, Clone)]
pub struct AgentExecutionRequest {
    pub user_id: String,
    pub tenant_id: String,
    pub session_id: String,
    pub code: String,
    pub language: CodeLanguage,
    pub timeout: Duration,
    pub memory_limit: u64,
    pub cpu_limit: u64,
    pub environment: std::collections::HashMap<String, String>,
    pub allowed_hosts: Vec<String>,
}

/// Supported code languages
#[derive(Debug, Clone, PartialEq)]
pub enum CodeLanguage {
    Python,
    JavaScript,
    WebAssembly,
}

/// Result of agent execution
#[derive(Debug, Clone)]
pub struct AgentExecutionResult {
    pub execution_id: String,
    pub status: ExecutionStatus,
    pub output: String,
    pub error_message: Option<String>,
    pub execution_time: Duration,
    pub tokens_used: u32,
    pub cost_usd: f64,
    pub security_violations: Vec<String>,
    pub fsm_result: Option<crate::fsm::StateMachineResult>,
}

impl ExecutionEngine {
    /// Create a new execution engine
    pub fn new(
        sandbox: Arc<WASISandbox>,
        enforcement: Arc<EnforcementGateway>,
        security: Arc<SecurityManager>,
        state_machine: Arc<StateMachine>,
        metrics: Arc<MetricsCollector>,
    ) -> Result<Self> {
        info!("Initializing execution engine");

        Ok(Self {
            sandbox,
            enforcement,
            security,
            state_machine,
            metrics,
            active_executions: Arc::new(Mutex::new(std::collections::HashMap::new())),
        })
    }

    /// Execute agent code with full security and monitoring
    pub async fn execute_agent_code(
        &self,
        request: AgentExecutionRequest,
    ) -> Result<AgentExecutionResult> {
        let execution_id = Uuid::new_v4().to_string();
        let start_time = Instant::now();
        
        info!("Starting agent execution: {}", execution_id);

        // Create active execution record
        let active_execution = ActiveExecution {
            execution_id: execution_id.clone(),
            user_id: request.user_id.clone(),
            tenant_id: request.tenant_id.clone(),
            session_id: request.session_id.clone(),
            fsm_instance_id: String::new(), // Will be set later
            started_at: start_time,
            status: ExecutionEngineStatus::Initializing,
        };

        {
            let mut executions = self.active_executions.lock().await;
            executions.insert(execution_id.clone(), active_execution);
        }

        // Execute with comprehensive error handling
        let result = self.execute_with_monitoring(&execution_id, request).await;
        
        // Clean up active execution
        {
            let mut executions = self.active_executions.lock().await;
            executions.remove(&execution_id);
        }

        result
    }

    /// Execute with full monitoring and state management
    async fn execute_with_monitoring(
        &self,
        execution_id: &str,
        request: AgentExecutionRequest,
    ) -> Result<AgentExecutionResult> {
        let start_time = Instant::now();

        // Step 1: Security validation
        self.update_execution_status(execution_id, ExecutionEngineStatus::PolicyCheck).await;
        
        let security_result = self.security.validate_code(&request.code, &request.user_id).await?;
        if !security_result.is_safe {
            return Ok(AgentExecutionResult {
                execution_id: execution_id.to_string(),
                status: ExecutionStatus::SecurityViolation,
                output: String::new(),
                error_message: Some("Security validation failed".to_string()),
                execution_time: start_time.elapsed(),
                tokens_used: 0,
                cost_usd: 0.0,
                security_violations: security_result.violations,
                fsm_result: None,
            });
        }

        // Step 2: Create enforcement request
        let enforcement_request = ExecuteTaskRequest {
            user_id: request.user_id.clone(),
            tenant_id: request.tenant_id.clone(),
            session_id: request.session_id.clone(),
            task_id: execution_id.to_string(),
            estimated_duration: request.timeout,
            estimated_tokens: self.estimate_tokens(&request.code),
            priority: TaskPriority::Normal,
            resource_requirements: ResourceRequirements {
                memory_mb: request.memory_limit / 1024 / 1024,
                cpu_cores: 1.0,
                network_bandwidth_mbps: 10,
                storage_mb: 100,
            },
        };

        // Step 3: Policy enforcement
        if let Err(e) = self.enforcement.enforce_request(&enforcement_request).await {
            return Ok(AgentExecutionResult {
                execution_id: execution_id.to_string(),
                status: ExecutionStatus::SecurityViolation,
                output: String::new(),
                error_message: Some(format!("Policy enforcement failed: {}", e)),
                execution_time: start_time.elapsed(),
                tokens_used: 0,
                cost_usd: 0.0,
                security_violations: vec![format!("Policy violation: {}", e)],
                fsm_result: None,
            });
        }

        // Step 4: Create FSM instance for execution tracking
        let fsm_context = StateMachineContext {
            variables: std::collections::HashMap::from([
                ("execution_id".to_string(), execution_id.to_string()),
                ("user_id".to_string(), request.user_id.clone()),
                ("language".to_string(), format!("{:?}", request.language)),
            ]),
            events: vec![],
            execution_history: vec![],
            metadata: std::collections::HashMap::new(),
        };

        let fsm_instance_id = self.state_machine.create_instance(fsm_context).await?;
        
        // Update active execution with FSM instance ID
        {
            let mut executions = self.active_executions.lock().await;
            if let Some(execution) = executions.get_mut(execution_id) {
                execution.fsm_instance_id = fsm_instance_id.clone();
            }
        }

        // Step 5: Execute code in sandbox
        self.update_execution_status(execution_id, ExecutionEngineStatus::Executing).await;
        
        // Trigger FSM transition to analyzing state
        let analyzing_event = Event {
            id: Uuid::new_v4().to_string(),
            event_type: "start_analysis".to_string(),
            payload: std::collections::HashMap::new(),
            timestamp: chrono::Utc::now(),
        };
        self.state_machine.trigger_event(&fsm_instance_id, analyzing_event).await?;

        // Create execution context for sandbox
        let execution_context = ExecutionContext {
            user_id: request.user_id.clone(),
            tenant_id: request.tenant_id.clone(),
            session_id: request.session_id.clone(),
            execution_id: execution_id.to_string(),
            memory_limit: request.memory_limit,
            cpu_limit: request.cpu_limit as u64,
            timeout: request.timeout,
            allowed_hosts: request.allowed_hosts,
            environment: request.environment,
        };

        // Execute based on language
        let sandbox_result = match request.language {
            CodeLanguage::Python => {
                self.sandbox.execute_python(&request.code, execution_context).await?
            }
            CodeLanguage::JavaScript => {
                self.sandbox.execute_javascript(&request.code, execution_context).await?
            }
            CodeLanguage::WebAssembly => {
                // Would implement WASM execution
                return Err(anyhow::anyhow!("WebAssembly execution not yet implemented"));
            }
        };

        // Step 6: Validate results
        self.update_execution_status(execution_id, ExecutionEngineStatus::Validating).await;
        
        // Trigger FSM events based on execution result
        let result_event = if sandbox_result.status == ExecutionStatus::Success {
            Event {
                id: Uuid::new_v4().to_string(),
                event_type: "success".to_string(),
                payload: std::collections::HashMap::from([
                    ("output_length".to_string(), sandbox_result.output.len().to_string()),
                ]),
                timestamp: chrono::Utc::now(),
            }
        } else {
            Event {
                id: Uuid::new_v4().to_string(),
                event_type: "error".to_string(),
                payload: std::collections::HashMap::from([
                    ("error".to_string(), sandbox_result.error_message.clone().unwrap_or_default()),
                ]),
                timestamp: chrono::Utc::now(),
            }
        };
        
        self.state_machine.trigger_event(&fsm_instance_id, result_event).await?;

        // Step 7: Record execution metrics
        let execution_success = sandbox_result.status == ExecutionStatus::Success;
        let tokens_used = self.calculate_actual_tokens(&sandbox_result);
        let cost_usd = self.calculate_cost(tokens_used);

        self.enforcement.record_execution_result(
            &enforcement_request,
            execution_success,
            sandbox_result.duration,
            tokens_used,
        ).await;

        // Step 8: Complete FSM instance
        let fsm_result = self.state_machine.complete_instance(&fsm_instance_id).await?;
        
        // Step 9: Final status update
        let final_status = if execution_success {
            ExecutionEngineStatus::Completed
        } else {
            ExecutionEngineStatus::Failed
        };
        self.update_execution_status(execution_id, final_status).await;

        // Record final metrics
        self.metrics.record_agent_execution(
            execution_id,
            &request.language,
            sandbox_result.duration,
            tokens_used,
            execution_success,
        );

        Ok(AgentExecutionResult {
            execution_id: execution_id.to_string(),
            status: sandbox_result.status,
            output: sandbox_result.output,
            error_message: sandbox_result.error_message,
            execution_time: start_time.elapsed(),
            tokens_used,
            cost_usd,
            security_violations: security_result.violations,
            fsm_result: Some(fsm_result),
        })
    }

    /// Update execution status
    async fn update_execution_status(&self, execution_id: &str, status: ExecutionEngineStatus) {
        let mut executions = self.active_executions.lock().await;
        if let Some(execution) = executions.get_mut(execution_id) {
            execution.status = status;
            debug!("Execution {} status updated to {:?}", execution_id, execution.status);
        }
    }

    /// Estimate tokens for code execution
    fn estimate_tokens(&self, code: &str) -> u32 {
        // Simple estimation based on code length
        // In a real implementation, this would use more sophisticated analysis
        let base_tokens = 100; // Base overhead
        let code_tokens = code.len() as u32 / 4; // Rough approximation
        base_tokens + code_tokens
    }

    /// Calculate actual tokens used based on execution result
    fn calculate_actual_tokens(&self, result: &ExecutionResult) -> u32 {
        // Calculate based on execution metrics and output
        let base_tokens = 50;
        let output_tokens = result.output.len() as u32 / 4;
        let cpu_tokens = (result.metrics.cpu_time.as_millis() / 100) as u32;
        
        base_tokens + output_tokens + cpu_tokens
    }

    /// Calculate cost in USD
    fn calculate_cost(&self, tokens_used: u32) -> f64 {
        // Simple cost calculation - would be more sophisticated in production
        tokens_used as f64 * 0.002 // $0.002 per token
    }

    /// Get active executions
    pub async fn get_active_executions(&self) -> Vec<ActiveExecution> {
        let executions = self.active_executions.lock().await;
        executions.values().cloned().collect()
    }

    /// Get execution statistics
    pub async fn get_execution_stats(&self) -> ExecutionStats {
        let executions = self.active_executions.lock().await;
        
        ExecutionStats {
            active_executions: executions.len(),
            total_executions: self.metrics.get_total_executions().await,
            success_rate: self.metrics.get_success_rate().await,
            average_duration: self.metrics.get_average_duration().await,
        }
    }
}

/// Execution engine statistics
#[derive(Debug, Clone)]
pub struct ExecutionStats {
    pub active_executions: usize,
    pub total_executions: u64,
    pub success_rate: f64,
    pub average_duration: Duration,
}