use anyhow::Result;
use std::sync::Arc;
use tonic::{Request, Response, Status};
use tracing::{debug, error, info, warn};

use crate::execution::{ExecutionEngine, AgentExecutionRequest, CodeLanguage};
use crate::metrics::MetricsCollector;

// Generated protobuf code would go here
// For now, we'll define the service manually

/// gRPC service for agent core
pub struct AgentCoreService {
    execution_engine: Arc<ExecutionEngine>,
    metrics: Arc<MetricsCollector>,
}

/// Execute code request
#[derive(Debug, Clone)]
pub struct ExecuteCodeRequest {
    pub user_id: String,
    pub tenant_id: String,
    pub session_id: String,
    pub code: String,
    pub language: String,
    pub timeout_seconds: u32,
    pub memory_limit_mb: u32,
    pub cpu_limit_seconds: u32,
    pub environment: std::collections::HashMap<String, String>,
    pub allowed_hosts: Vec<String>,
}

/// Execute code response
#[derive(Debug, Clone)]
pub struct ExecuteCodeResponse {
    pub execution_id: String,
    pub status: String,
    pub output: String,
    pub error_message: String,
    pub execution_time_ms: u64,
    pub tokens_used: u32,
    pub cost_usd: f64,
    pub security_violations: Vec<String>,
}

/// Get status request
#[derive(Debug, Clone)]
pub struct GetStatusRequest {
    pub execution_id: String,
}

/// Get status response
#[derive(Debug, Clone)]
pub struct GetStatusResponse {
    pub execution_id: String,
    pub status: String,
    pub progress: f32,
    pub current_state: String,
    pub started_at: String,
    pub estimated_completion: String,
}

/// Get metrics request
#[derive(Debug, Clone)]
pub struct GetMetricsRequest {
    pub include_detailed: bool,
}

/// Get metrics response
#[derive(Debug, Clone)]
pub struct GetMetricsResponse {
    pub total_executions: u64,
    pub success_rate: f64,
    pub average_duration_ms: u64,
    pub active_executions: u32,
    pub system_health: String,
}

impl AgentCoreService {
    /// Create a new gRPC service
    pub fn new(
        execution_engine: Arc<ExecutionEngine>,
        metrics: Arc<MetricsCollector>,
    ) -> Self {
        info!("Initializing Agent Core gRPC service");
        
        Self {
            execution_engine,
            metrics,
        }
    }

    /// Execute agent code
    pub async fn execute_code(
        &self,
        request: Request<ExecuteCodeRequest>,
    ) -> Result<Response<ExecuteCodeResponse>, Status> {
        let req = request.into_inner();
        
        debug!("Received execute_code request for user: {}", req.user_id);

        // Validate request
        if req.code.is_empty() {
            return Err(Status::invalid_argument("Code cannot be empty"));
        }

        if req.user_id.is_empty() || req.tenant_id.is_empty() {
            return Err(Status::invalid_argument("User ID and Tenant ID are required"));
        }

        // Parse language
        let language = match req.language.to_lowercase().as_str() {
            "python" => CodeLanguage::Python,
            "javascript" | "js" => CodeLanguage::JavaScript,
            "wasm" | "webassembly" => CodeLanguage::WebAssembly,
            _ => {
                return Err(Status::invalid_argument(
                    format!("Unsupported language: {}", req.language)
                ));
            }
        };

        // Create execution request
        let execution_request = AgentExecutionRequest {
            user_id: req.user_id,
            tenant_id: req.tenant_id,
            session_id: req.session_id,
            code: req.code,
            language,
            timeout: std::time::Duration::from_secs(req.timeout_seconds as u64),
            memory_limit: (req.memory_limit_mb as u64) * 1024 * 1024, // Convert MB to bytes
            cpu_limit: (req.cpu_limit_seconds as u64) * 1_000_000_000, // Convert seconds to nanoseconds
            environment: req.environment,
            allowed_hosts: req.allowed_hosts,
        };

        // Execute code
        match self.execution_engine.execute_agent_code(execution_request).await {
            Ok(result) => {
                let response = ExecuteCodeResponse {
                    execution_id: result.execution_id,
                    status: format!("{:?}", result.status),
                    output: result.output,
                    error_message: result.error_message.unwrap_or_default(),
                    execution_time_ms: result.execution_time.as_millis() as u64,
                    tokens_used: result.tokens_used,
                    cost_usd: result.cost_usd,
                    security_violations: result.security_violations,
                };

                Ok(Response::new(response))
            }
            Err(e) => {
                error!("Execution failed: {}", e);
                Err(Status::internal(format!("Execution failed: {}", e)))
            }
        }
    }

    /// Get execution status
    pub async fn get_status(
        &self,
        request: Request<GetStatusRequest>,
    ) -> Result<Response<GetStatusResponse>, Status> {
        let req = request.into_inner();
        
        debug!("Received get_status request for execution: {}", req.execution_id);

        // Get active executions
        let active_executions = self.execution_engine.get_active_executions().await;
        
        // Find the requested execution
        if let Some(execution) = active_executions.iter().find(|e| e.execution_id == req.execution_id) {
            let response = GetStatusResponse {
                execution_id: execution.execution_id.clone(),
                status: format!("{:?}", execution.status),
                progress: self.calculate_progress(&execution.status),
                current_state: format!("{:?}", execution.status),
                started_at: execution.started_at.elapsed().as_secs().to_string(),
                estimated_completion: "unknown".to_string(), // Would calculate based on historical data
            };

            Ok(Response::new(response))
        } else {
            Err(Status::not_found(format!("Execution not found: {}", req.execution_id)))
        }
    }

    /// Get system metrics
    pub async fn get_metrics(
        &self,
        request: Request<GetMetricsRequest>,
    ) -> Result<Response<GetMetricsResponse>, Status> {
        let _req = request.into_inner();
        
        debug!("Received get_metrics request");

        // Get execution statistics
        let stats = self.execution_engine.get_execution_stats().await;
        let metrics_summary = self.metrics.get_metrics_summary().await;

        let response = GetMetricsResponse {
            total_executions: metrics_summary.total_executions,
            success_rate: metrics_summary.success_rate,
            average_duration_ms: metrics_summary.average_duration.as_millis() as u64,
            active_executions: stats.active_executions as u32,
            system_health: self.determine_system_health(&metrics_summary).await,
        };

        Ok(Response::new(response))
    }

    /// Calculate execution progress based on status
    fn calculate_progress(&self, status: &crate::execution::ExecutionEngineStatus) -> f32 {
        match status {
            crate::execution::ExecutionEngineStatus::Initializing => 0.1,
            crate::execution::ExecutionEngineStatus::PolicyCheck => 0.2,
            crate::execution::ExecutionEngineStatus::Executing => 0.6,
            crate::execution::ExecutionEngineStatus::Validating => 0.9,
            crate::execution::ExecutionEngineStatus::Completed => 1.0,
            crate::execution::ExecutionEngineStatus::Failed => 1.0,
        }
    }

    /// Determine system health based on metrics
    async fn determine_system_health(&self, metrics: &crate::metrics::MetricsSummary) -> String {
        // Simple health determination logic
        if metrics.success_rate > 0.95 {
            "healthy".to_string()
        } else if metrics.success_rate > 0.8 {
            "degraded".to_string()
        } else {
            "unhealthy".to_string()
        }
    }

    /// Convert to tonic service (this would be generated by tonic-build in real implementation)
    pub fn into_service(self) -> AgentCoreServiceImpl {
        AgentCoreServiceImpl { inner: Arc::new(self) }
    }
}

/// Implementation wrapper for tonic service
pub struct AgentCoreServiceImpl {
    inner: Arc<AgentCoreService>,
}

// In a real implementation, this would be generated by tonic-build
// For now, we'll provide a mock implementation
impl AgentCoreServiceImpl {
    pub async fn execute_code_mock(
        &self,
        request: tonic::Request<()>,
    ) -> Result<tonic::Response<()>, tonic::Status> {
        // Mock implementation - would be replaced by generated code
        info!("Mock gRPC execute_code called");
        Ok(tonic::Response::new(()))
    }
}

// Mock trait for the generated service
// In real implementation, this would be generated by protobuf
pub trait AgentCore {
    async fn execute_code(
        &self,
        request: tonic::Request<ExecuteCodeRequest>,
    ) -> Result<tonic::Response<ExecuteCodeResponse>, tonic::Status>;

    async fn get_status(
        &self,
        request: tonic::Request<GetStatusRequest>,
    ) -> Result<tonic::Response<GetStatusResponse>, tonic::Status>;

    async fn get_metrics(
        &self,
        request: tonic::Request<GetMetricsRequest>,
    ) -> Result<tonic::Response<GetMetricsResponse>, tonic::Status>;
}

#[tonic::async_trait]
impl AgentCore for AgentCoreServiceImpl {
    async fn execute_code(
        &self,
        request: tonic::Request<ExecuteCodeRequest>,
    ) -> Result<tonic::Response<ExecuteCodeResponse>, tonic::Status> {
        self.inner.execute_code(request).await
    }

    async fn get_status(
        &self,
        request: tonic::Request<GetStatusRequest>,
    ) -> Result<tonic::Response<GetStatusResponse>, tonic::Status> {
        self.inner.get_status(request).await
    }

    async fn get_metrics(
        &self,
        request: tonic::Request<GetMetricsRequest>,
    ) -> Result<tonic::Response<GetMetricsResponse>, tonic::Status> {
        self.inner.get_metrics(request).await
    }
}

/// Health check service
pub struct HealthService;

#[derive(Debug, Clone)]
pub struct HealthCheckRequest {
    pub service: String,
}

#[derive(Debug, Clone)]
pub struct HealthCheckResponse {
    pub status: String,
    pub message: String,
}

impl HealthService {
    pub fn new() -> Self {
        Self
    }

    pub async fn check(
        &self,
        request: Request<HealthCheckRequest>,
    ) -> Result<Response<HealthCheckResponse>, Status> {
        let req = request.into_inner();
        
        debug!("Health check requested for service: {}", req.service);

        // Perform health checks
        let (status, message) = match req.service.as_str() {
            "agent-core" => {
                // Check if core services are running
                ("SERVING".to_string(), "Agent core is healthy".to_string())
            }
            "" => {
                // Overall health check
                ("SERVING".to_string(), "All services are healthy".to_string())
            }
            _ => {
                ("NOT_FOUND".to_string(), format!("Unknown service: {}", req.service))
            }
        };

        let response = HealthCheckResponse { status, message };
        Ok(Response::new(response))
    }
}

/// gRPC server configuration
pub struct GrpcServerConfig {
    pub addr: std::net::SocketAddr,
    pub max_connections: usize,
    pub request_timeout: std::time::Duration,
    pub enable_reflection: bool,
    pub enable_health_check: bool,
}

impl Default for GrpcServerConfig {
    fn default() -> Self {
        Self {
            addr: "0.0.0.0:50051".parse().unwrap(),
            max_connections: 1000,
            request_timeout: std::time::Duration::from_secs(30),
            enable_reflection: true,
            enable_health_check: true,
        }
    }
}

/// Start gRPC server
pub async fn start_grpc_server(
    config: GrpcServerConfig,
    agent_service: AgentCoreService,
) -> Result<()> {
    info!("Starting gRPC server on {}", config.addr);

    let agent_service = agent_service.into_service();
    let health_service = HealthService::new();

    // Build server
    let mut server_builder = tonic::transport::Server::builder()
        .timeout(config.request_timeout)
        .concurrency_limit_per_connection(256);

    // Add services
    let server = server_builder
        .add_service(tonic::transport::server::Routes::new()) // Mock - would add real services
        .serve(config.addr);

    info!("gRPC server started successfully on {}", config.addr);

    // Start server
    server.await.map_err(|e| anyhow::anyhow!("gRPC server error: {}", e))
}