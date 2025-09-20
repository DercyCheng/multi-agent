use anyhow::{Context, Result};
use std::collections::HashMap;
use std::sync::Arc;
use std::time::{Duration, Instant};
use tokio::sync::{Mutex, RwLock};
use tokio::time::sleep;
use tracing::{debug, error, info, warn};

use crate::config::EnforcementConfig;
use crate::metrics::MetricsCollector;

/// Enforcement Gateway - Unified policy execution point
pub struct EnforcementGateway {
    config: EnforcementConfig,
    timeout_enforcer: TimeoutEnforcer,
    rate_limiter: RateLimiter,
    circuit_breaker: CircuitBreaker,
    token_validator: TokenValidator,
    metrics: Arc<MetricsCollector>,
}

/// Request to execute a task
#[derive(Debug, Clone)]
pub struct ExecuteTaskRequest {
    pub user_id: String,
    pub tenant_id: String,
    pub session_id: String,
    pub task_id: String,
    pub estimated_duration: Duration,
    pub estimated_tokens: u32,
    pub priority: TaskPriority,
    pub resource_requirements: ResourceRequirements,
}

/// Task priority levels
#[derive(Debug, Clone, PartialEq)]
pub enum TaskPriority {
    Low,
    Normal,
    High,
    Critical,
}

/// Resource requirements for task execution
#[derive(Debug, Clone)]
pub struct ResourceRequirements {
    pub memory_mb: u64,
    pub cpu_cores: f32,
    pub network_bandwidth_mbps: u32,
    pub storage_mb: u64,
}

/// Enforcement errors
#[derive(Debug, thiserror::Error)]
pub enum EnforcementError {
    #[error("Timeout exceeded: {0:?}")]
    TimeoutExceeded(Duration),
    
    #[error("Token limit exceeded: {current} > {limit}")]
    TokenLimitExceeded { current: u32, limit: u32 },
    
    #[error("Rate limit exceeded for key: {0}")]
    RateLimitExceeded(String),
    
    #[error("Circuit breaker open for key: {0}")]
    CircuitBreakerOpen(String),
    
    #[error("Resource limit exceeded: {resource}")]
    ResourceLimitExceeded { resource: String },
    
    #[error("Policy violation: {0}")]
    PolicyViolation(String),
}

impl EnforcementGateway {
    /// Create a new enforcement gateway
    pub async fn new(
        config: &EnforcementConfig,
        metrics: Arc<MetricsCollector>,
    ) -> Result<Self> {
        info!("Initializing enforcement gateway");

        let timeout_enforcer = TimeoutEnforcer::new(&config.timeout_config);
        let rate_limiter = RateLimiter::new(&config.rate_limit_config).await?;
        let circuit_breaker = CircuitBreaker::new(&config.circuit_breaker_config);
        let token_validator = TokenValidator::new(&config.token_validator_config);

        Ok(Self {
            config: config.clone(),
            timeout_enforcer,
            rate_limiter,
            circuit_breaker,
            token_validator,
            metrics,
        })
    }

    /// Enforce request policies before execution
    pub async fn enforce_request(&self, request: &ExecuteTaskRequest) -> Result<(), EnforcementError> {
        debug!("Enforcing request for task: {}", request.task_id);

        // 1. Timeout enforcement
        self.timeout_enforcer.check_timeout(request.estimated_duration)?;

        // 2. Token limit enforcement
        self.token_validator.validate_tokens(request.estimated_tokens)?;

        // 3. Rate limiting
        let rate_limit_key = format!("user:{}", request.user_id);
        self.rate_limiter.check_rate_limit(&rate_limit_key).await?;

        // 4. Circuit breaker check
        let circuit_key = format!("tenant:{}", request.tenant_id);
        self.circuit_breaker.check_circuit(&circuit_key)?;

        // 5. Resource validation
        self.validate_resources(&request.resource_requirements)?;

        // Record successful enforcement
        self.metrics.record_enforcement_success(&request.task_id);

        info!("Request enforcement passed for task: {}", request.task_id);
        Ok(())
    }

    /// Record execution result for circuit breaker and metrics
    pub async fn record_execution_result(
        &self,
        request: &ExecuteTaskRequest,
        success: bool,
        duration: Duration,
        tokens_used: u32,
    ) {
        let circuit_key = format!("tenant:{}", request.tenant_id);
        
        if success {
            self.circuit_breaker.record_success(&circuit_key);
            self.metrics.record_task_success(&request.task_id, duration, tokens_used);
        } else {
            self.circuit_breaker.record_failure(&circuit_key);
            self.metrics.record_task_failure(&request.task_id, duration);
        }
    }

    /// Validate resource requirements
    fn validate_resources(&self, requirements: &ResourceRequirements) -> Result<(), EnforcementError> {
        // Memory validation
        if requirements.memory_mb > 2048 {
            return Err(EnforcementError::ResourceLimitExceeded {
                resource: format!("memory: {}MB > 2048MB", requirements.memory_mb),
            });
        }

        // CPU validation
        if requirements.cpu_cores > 4.0 {
            return Err(EnforcementError::ResourceLimitExceeded {
                resource: format!("cpu: {} cores > 4.0 cores", requirements.cpu_cores),
            });
        }

        // Network bandwidth validation
        if requirements.network_bandwidth_mbps > 100 {
            return Err(EnforcementError::ResourceLimitExceeded {
                resource: format!("bandwidth: {}Mbps > 100Mbps", requirements.network_bandwidth_mbps),
            });
        }

        // Storage validation
        if requirements.storage_mb > 1024 {
            return Err(EnforcementError::ResourceLimitExceeded {
                resource: format!("storage: {}MB > 1024MB", requirements.storage_mb),
            });
        }

        Ok(())
    }
}

/// Timeout enforcement component
pub struct TimeoutEnforcer {
    max_duration: Duration,
    warning_threshold: Duration,
}

impl TimeoutEnforcer {
    pub fn new(config: &crate::config::TimeoutConfig) -> Self {
        Self {
            max_duration: config.max_duration,
            warning_threshold: config.warning_threshold,
        }
    }

    pub fn check_timeout(&self, estimated_duration: Duration) -> Result<(), EnforcementError> {
        if estimated_duration > self.max_duration {
            return Err(EnforcementError::TimeoutExceeded(estimated_duration));
        }

        if estimated_duration > self.warning_threshold {
            warn!("Task duration approaching limit: {:?}", estimated_duration);
        }

        Ok(())
    }
}

/// Rate limiting component
pub struct RateLimiter {
    requests_per_second: u32,
    burst_size: u32,
    window_size: Duration,
    buckets: Arc<RwLock<HashMap<String, TokenBucket>>>,
}

/// Token bucket for rate limiting
#[derive(Debug, Clone)]
struct TokenBucket {
    tokens: f64,
    last_refill: Instant,
    capacity: f64,
    refill_rate: f64,
}

impl RateLimiter {
    pub async fn new(config: &crate::config::RateLimitConfig) -> Result<Self> {
        Ok(Self {
            requests_per_second: config.requests_per_second,
            burst_size: config.burst_size,
            window_size: config.window_size,
            buckets: Arc::new(RwLock::new(HashMap::new())),
        })
    }

    pub async fn check_rate_limit(&self, key: &str) -> Result<(), EnforcementError> {
        let mut buckets = self.buckets.write().await;
        
        let bucket = buckets.entry(key.to_string()).or_insert_with(|| TokenBucket {
            tokens: self.burst_size as f64,
            last_refill: Instant::now(),
            capacity: self.burst_size as f64,
            refill_rate: self.requests_per_second as f64,
        });

        // Refill tokens based on elapsed time
        let now = Instant::now();
        let elapsed = now.duration_since(bucket.last_refill).as_secs_f64();
        let tokens_to_add = elapsed * bucket.refill_rate;
        
        bucket.tokens = (bucket.tokens + tokens_to_add).min(bucket.capacity);
        bucket.last_refill = now;

        // Check if we can consume a token
        if bucket.tokens >= 1.0 {
            bucket.tokens -= 1.0;
            Ok(())
        } else {
            Err(EnforcementError::RateLimitExceeded(key.to_string()))
        }
    }
}

/// Circuit breaker component
pub struct CircuitBreaker {
    failure_threshold: u32,
    success_threshold: u32,
    timeout: Duration,
    circuits: Arc<Mutex<HashMap<String, CircuitState>>>,
}

/// Circuit breaker state
#[derive(Debug, Clone)]
struct CircuitState {
    state: CircuitStatus,
    failure_count: u32,
    success_count: u32,
    last_failure: Option<Instant>,
}

#[derive(Debug, Clone, PartialEq)]
enum CircuitStatus {
    Closed,
    Open,
    HalfOpen,
}

impl CircuitBreaker {
    pub fn new(config: &crate::config::CircuitBreakerConfig) -> Self {
        Self {
            failure_threshold: config.failure_threshold,
            success_threshold: config.success_threshold,
            timeout: config.timeout,
            circuits: Arc::new(Mutex::new(HashMap::new())),
        }
    }

    pub fn check_circuit(&self, key: &str) -> Result<(), EnforcementError> {
        let circuits = tokio::task::block_in_place(|| {
            tokio::runtime::Handle::current().block_on(self.circuits.lock())
        });
        
        let circuit = circuits.get(key);
        
        match circuit {
            Some(circuit) if circuit.state == CircuitStatus::Open => {
                // Check if timeout has elapsed
                if let Some(last_failure) = circuit.last_failure {
                    if last_failure.elapsed() > self.timeout {
                        // Transition to half-open
                        drop(circuits);
                        self.transition_to_half_open(key);
                        Ok(())
                    } else {
                        Err(EnforcementError::CircuitBreakerOpen(key.to_string()))
                    }
                } else {
                    Err(EnforcementError::CircuitBreakerOpen(key.to_string()))
                }
            }
            _ => Ok(()),
        }
    }

    pub fn record_success(&self, key: &str) {
        tokio::task::spawn({
            let key = key.to_string();
            let circuits = self.circuits.clone();
            let success_threshold = self.success_threshold;
            
            async move {
                let mut circuits = circuits.lock().await;
                let circuit = circuits.entry(key).or_insert_with(|| CircuitState {
                    state: CircuitStatus::Closed,
                    failure_count: 0,
                    success_count: 0,
                    last_failure: None,
                });

                circuit.success_count += 1;
                circuit.failure_count = 0;

                // Transition from half-open to closed if enough successes
                if circuit.state == CircuitStatus::HalfOpen && circuit.success_count >= success_threshold {
                    circuit.state = CircuitStatus::Closed;
                    circuit.success_count = 0;
                }
            }
        });
    }

    pub fn record_failure(&self, key: &str) {
        tokio::task::spawn({
            let key = key.to_string();
            let circuits = self.circuits.clone();
            let failure_threshold = self.failure_threshold;
            
            async move {
                let mut circuits = circuits.lock().await;
                let circuit = circuits.entry(key).or_insert_with(|| CircuitState {
                    state: CircuitStatus::Closed,
                    failure_count: 0,
                    success_count: 0,
                    last_failure: None,
                });

                circuit.failure_count += 1;
                circuit.success_count = 0;
                circuit.last_failure = Some(Instant::now());

                // Transition to open if failure threshold exceeded
                if circuit.failure_count >= failure_threshold {
                    circuit.state = CircuitStatus::Open;
                }
            }
        });
    }

    fn transition_to_half_open(&self, key: &str) {
        tokio::task::spawn({
            let key = key.to_string();
            let circuits = self.circuits.clone();
            
            async move {
                let mut circuits = circuits.lock().await;
                if let Some(circuit) = circuits.get_mut(&key) {
                    circuit.state = CircuitStatus::HalfOpen;
                    circuit.success_count = 0;
                }
            }
        });
    }
}

/// Token validation component
pub struct TokenValidator {
    max_tokens: u32,
    cost_per_token: f64,
}

impl TokenValidator {
    pub fn new(config: &crate::config::TokenValidatorConfig) -> Self {
        Self {
            max_tokens: config.max_tokens,
            cost_per_token: config.cost_per_token,
        }
    }

    pub fn validate_tokens(&self, estimated_tokens: u32) -> Result<(), EnforcementError> {
        if estimated_tokens > self.max_tokens {
            return Err(EnforcementError::TokenLimitExceeded {
                current: estimated_tokens,
                limit: self.max_tokens,
            });
        }

        let estimated_cost = estimated_tokens as f64 * self.cost_per_token;
        if estimated_cost > 10.0 {
            warn!("High token cost estimated: ${:.2}", estimated_cost);
        }

        Ok(())
    }

    pub fn calculate_cost(&self, tokens_used: u32) -> f64 {
        tokens_used as f64 * self.cost_per_token
    }
}