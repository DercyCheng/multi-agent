use anyhow::Result;
use prometheus::{Counter, Histogram, Gauge, Registry, Encoder, TextEncoder};
use std::collections::HashMap;
use std::sync::Arc;
use std::time::Duration;
use tokio::sync::RwLock;
use tracing::{debug, error, info};
use warp::{Filter, Reply};

use crate::execution::CodeLanguage;

/// Metrics collector for the agent core system
pub struct MetricsCollector {
    registry: Registry,
    
    // Execution metrics
    executions_total: Counter,
    execution_duration: Histogram,
    execution_tokens: Histogram,
    execution_success_rate: Gauge,
    
    // Sandbox metrics
    sandbox_instances_active: Gauge,
    sandbox_memory_usage: Histogram,
    sandbox_cpu_usage: Histogram,
    
    // Security metrics
    security_violations_total: Counter,
    policy_evaluations_total: Counter,
    
    // FSM metrics
    fsm_instances_active: Gauge,
    fsm_transitions_total: Counter,
    fsm_state_duration: Histogram,
    
    // Enforcement metrics
    enforcement_checks_total: Counter,
    rate_limit_violations: Counter,
    circuit_breaker_trips: Counter,
    
    // System metrics
    system_memory_usage: Gauge,
    system_cpu_usage: Gauge,
    
    // Runtime statistics
    stats: Arc<RwLock<RuntimeStats>>,
}

/// Runtime statistics
#[derive(Debug, Clone, Default)]
struct RuntimeStats {
    total_executions: u64,
    successful_executions: u64,
    failed_executions: u64,
    total_duration: Duration,
    total_tokens: u64,
}

impl MetricsCollector {
    /// Create a new metrics collector
    pub fn new() -> Result<Self> {
        let registry = Registry::new();
        
        // Initialize execution metrics
        let executions_total = Counter::new(
            "agent_executions_total",
            "Total number of agent code executions"
        )?;
        registry.register(Box::new(executions_total.clone()))?;
        
        let execution_duration = Histogram::with_opts(
            prometheus::HistogramOpts::new(
                "agent_execution_duration_seconds",
                "Duration of agent code executions in seconds"
            ).buckets(vec![0.1, 0.5, 1.0, 2.5, 5.0, 10.0, 30.0, 60.0, 300.0])
        )?;
        registry.register(Box::new(execution_duration.clone()))?;
        
        let execution_tokens = Histogram::with_opts(
            prometheus::HistogramOpts::new(
                "agent_execution_tokens_total",
                "Number of tokens used in agent executions"
            ).buckets(vec![10.0, 50.0, 100.0, 500.0, 1000.0, 5000.0, 10000.0])
        )?;
        registry.register(Box::new(execution_tokens.clone()))?;
        
        let execution_success_rate = Gauge::new(
            "agent_execution_success_rate",
            "Success rate of agent executions (0-1)"
        )?;
        registry.register(Box::new(execution_success_rate.clone()))?;
        
        // Initialize sandbox metrics
        let sandbox_instances_active = Gauge::new(
            "sandbox_instances_active",
            "Number of active sandbox instances"
        )?;
        registry.register(Box::new(sandbox_instances_active.clone()))?;
        
        let sandbox_memory_usage = Histogram::with_opts(
            prometheus::HistogramOpts::new(
                "sandbox_memory_usage_bytes",
                "Memory usage of sandbox instances in bytes"
            ).buckets(vec![
                1024.0 * 1024.0,      // 1MB
                10.0 * 1024.0 * 1024.0, // 10MB
                50.0 * 1024.0 * 1024.0, // 50MB
                100.0 * 1024.0 * 1024.0, // 100MB
                500.0 * 1024.0 * 1024.0, // 500MB
            ])
        )?;
        registry.register(Box::new(sandbox_memory_usage.clone()))?;
        
        let sandbox_cpu_usage = Histogram::with_opts(
            prometheus::HistogramOpts::new(
                "sandbox_cpu_usage_seconds",
                "CPU usage of sandbox instances in seconds"
            ).buckets(vec![0.01, 0.1, 0.5, 1.0, 5.0, 10.0, 30.0])
        )?;
        registry.register(Box::new(sandbox_cpu_usage.clone()))?;
        
        // Initialize security metrics
        let security_violations_total = Counter::new(
            "security_violations_total",
            "Total number of security violations detected"
        )?;
        registry.register(Box::new(security_violations_total.clone()))?;
        
        let policy_evaluations_total = Counter::new(
            "policy_evaluations_total",
            "Total number of policy evaluations performed"
        )?;
        registry.register(Box::new(policy_evaluations_total.clone()))?;
        
        // Initialize FSM metrics
        let fsm_instances_active = Gauge::new(
            "fsm_instances_active",
            "Number of active FSM instances"
        )?;
        registry.register(Box::new(fsm_instances_active.clone()))?;
        
        let fsm_transitions_total = Counter::new(
            "fsm_transitions_total",
            "Total number of FSM state transitions"
        )?;
        registry.register(Box::new(fsm_transitions_total.clone()))?;
        
        let fsm_state_duration = Histogram::with_opts(
            prometheus::HistogramOpts::new(
                "fsm_state_duration_seconds",
                "Duration spent in FSM states in seconds"
            ).buckets(vec![0.1, 1.0, 5.0, 10.0, 30.0, 60.0, 300.0])
        )?;
        registry.register(Box::new(fsm_state_duration.clone()))?;
        
        // Initialize enforcement metrics
        let enforcement_checks_total = Counter::new(
            "enforcement_checks_total",
            "Total number of enforcement checks performed"
        )?;
        registry.register(Box::new(enforcement_checks_total.clone()))?;
        
        let rate_limit_violations = Counter::new(
            "rate_limit_violations_total",
            "Total number of rate limit violations"
        )?;
        registry.register(Box::new(rate_limit_violations.clone()))?;
        
        let circuit_breaker_trips = Counter::new(
            "circuit_breaker_trips_total",
            "Total number of circuit breaker trips"
        )?;
        registry.register(Box::new(circuit_breaker_trips.clone()))?;
        
        // Initialize system metrics
        let system_memory_usage = Gauge::new(
            "system_memory_usage_bytes",
            "System memory usage in bytes"
        )?;
        registry.register(Box::new(system_memory_usage.clone()))?;
        
        let system_cpu_usage = Gauge::new(
            "system_cpu_usage_percent",
            "System CPU usage percentage"
        )?;
        registry.register(Box::new(system_cpu_usage.clone()))?;
        
        Ok(Self {
            registry,
            executions_total,
            execution_duration,
            execution_tokens,
            execution_success_rate,
            sandbox_instances_active,
            sandbox_memory_usage,
            sandbox_cpu_usage,
            security_violations_total,
            policy_evaluations_total,
            fsm_instances_active,
            fsm_transitions_total,
            fsm_state_duration,
            enforcement_checks_total,
            rate_limit_violations,
            circuit_breaker_trips,
            system_memory_usage,
            system_cpu_usage,
            stats: Arc::new(RwLock::new(RuntimeStats::default())),
        })
    }

    /// Record agent execution metrics
    pub fn record_agent_execution(
        &self,
        execution_id: &str,
        language: &CodeLanguage,
        duration: Duration,
        tokens_used: u32,
        success: bool,
    ) {
        debug!("Recording execution metrics for {}: success={}, duration={:?}, tokens={}", 
               execution_id, success, duration, tokens_used);

        // Update counters and histograms
        self.executions_total.with_label_values(&[
            &format!("{:?}", language).to_lowercase(),
            if success { "success" } else { "failure" }
        ]).inc();
        
        self.execution_duration.with_label_values(&[
            &format!("{:?}", language).to_lowercase()
        ]).observe(duration.as_secs_f64());
        
        self.execution_tokens.with_label_values(&[
            &format!("{:?}", language).to_lowercase()
        ]).observe(tokens_used as f64);

        // Update runtime statistics
        tokio::spawn({
            let stats = self.stats.clone();
            async move {
                let mut stats = stats.write().await;
                stats.total_executions += 1;
                if success {
                    stats.successful_executions += 1;
                } else {
                    stats.failed_executions += 1;
                }
                stats.total_duration += duration;
                stats.total_tokens += tokens_used as u64;
                
                // Update success rate gauge
                let success_rate = stats.successful_executions as f64 / stats.total_executions as f64;
                drop(stats); // Release lock before calling gauge
            }
        });
        
        // Update success rate (this is approximate due to async nature)
        tokio::spawn({
            let success_rate_gauge = self.execution_success_rate.clone();
            let stats = self.stats.clone();
            async move {
                let stats = stats.read().await;
                if stats.total_executions > 0 {
                    let rate = stats.successful_executions as f64 / stats.total_executions as f64;
                    success_rate_gauge.set(rate);
                }
            }
        });
    }

    /// Record sandbox metrics
    pub fn record_sandbox_metrics(&self, active_instances: usize, memory_usage: u64, cpu_usage: Duration) {
        self.sandbox_instances_active.set(active_instances as f64);
        self.sandbox_memory_usage.observe(memory_usage as f64);
        self.sandbox_cpu_usage.observe(cpu_usage.as_secs_f64());
    }

    /// Record security violation
    pub fn record_security_violation(&self, violation_type: &str) {
        self.security_violations_total.with_label_values(&[violation_type]).inc();
    }

    /// Record policy evaluation
    pub fn record_policy_evaluation(&self, policy_name: &str, result: bool) {
        self.policy_evaluations_total.with_label_values(&[
            policy_name,
            if result { "allowed" } else { "denied" }
        ]).inc();
    }

    /// Record FSM metrics
    pub fn record_fsm_metrics(&self, active_instances: usize) {
        self.fsm_instances_active.set(active_instances as f64);
    }

    /// Record FSM transition
    pub fn record_fsm_transition(&self, from_state: &str, to_state: &str, duration: Duration) {
        self.fsm_transitions_total.with_label_values(&[from_state, to_state]).inc();
        self.fsm_state_duration.with_label_values(&[from_state]).observe(duration.as_secs_f64());
    }

    /// Record enforcement success
    pub fn record_enforcement_success(&self, task_id: &str) {
        self.enforcement_checks_total.with_label_values(&["success"]).inc();
    }

    /// Record enforcement failure
    pub fn record_enforcement_failure(&self, reason: &str) {
        self.enforcement_checks_total.with_label_values(&["failure"]).inc();
        
        match reason {
            "rate_limit" => self.rate_limit_violations.inc(),
            "circuit_breaker" => self.circuit_breaker_trips.inc(),
            _ => {}
        }
    }

    /// Record task success
    pub fn record_task_success(&self, task_id: &str, duration: Duration, tokens_used: u32) {
        debug!("Task {} completed successfully in {:?} using {} tokens", task_id, duration, tokens_used);
    }

    /// Record task failure
    pub fn record_task_failure(&self, task_id: &str, duration: Duration) {
        debug!("Task {} failed after {:?}", task_id, duration);
    }

    /// Update system metrics
    pub fn update_system_metrics(&self) {
        // Get system memory usage
        if let Ok(memory_info) = sys_info::mem_info() {
            let used_memory = (memory_info.total - memory_info.free) * 1024; // Convert to bytes
            self.system_memory_usage.set(used_memory as f64);
        }

        // Get system CPU usage (simplified)
        if let Ok(load_avg) = sys_info::loadavg() {
            self.system_cpu_usage.set(load_avg.one * 100.0); // Convert to percentage
        }
    }

    /// Get total executions
    pub async fn get_total_executions(&self) -> u64 {
        let stats = self.stats.read().await;
        stats.total_executions
    }

    /// Get success rate
    pub async fn get_success_rate(&self) -> f64 {
        let stats = self.stats.read().await;
        if stats.total_executions > 0 {
            stats.successful_executions as f64 / stats.total_executions as f64
        } else {
            0.0
        }
    }

    /// Get average duration
    pub async fn get_average_duration(&self) -> Duration {
        let stats = self.stats.read().await;
        if stats.total_executions > 0 {
            stats.total_duration / stats.total_executions as u32
        } else {
            Duration::from_secs(0)
        }
    }

    /// Start metrics server
    pub async fn start_server(&self) -> Result<()> {
        let registry = self.registry.clone();
        
        // Create metrics endpoint
        let metrics_route = warp::path("metrics")
            .and(warp::get())
            .map(move || {
                let encoder = TextEncoder::new();
                let metric_families = registry.gather();
                let mut buffer = Vec::new();
                
                match encoder.encode(&metric_families, &mut buffer) {
                    Ok(_) => {
                        String::from_utf8(buffer).unwrap_or_else(|_| "Error encoding metrics".to_string())
                    }
                    Err(e) => {
                        error!("Failed to encode metrics: {}", e);
                        "Error encoding metrics".to_string()
                    }
                }
            })
            .with(warp::reply::with::header("content-type", "text/plain; version=0.0.4"));

        // Health check endpoint
        let health_route = warp::path("health")
            .and(warp::get())
            .map(|| {
                warp::reply::json(&serde_json::json!({
                    "status": "healthy",
                    "timestamp": chrono::Utc::now().to_rfc3339()
                }))
            });

        // Combine routes
        let routes = metrics_route.or(health_route);

        info!("Starting metrics server on 0.0.0.0:2113");
        
        // Start server
        warp::serve(routes)
            .run(([0, 0, 0, 0], 2113))
            .await;

        Ok(())
    }

    /// Get metrics summary
    pub async fn get_metrics_summary(&self) -> MetricsSummary {
        let stats = self.stats.read().await;
        
        MetricsSummary {
            total_executions: stats.total_executions,
            successful_executions: stats.successful_executions,
            failed_executions: stats.failed_executions,
            success_rate: if stats.total_executions > 0 {
                stats.successful_executions as f64 / stats.total_executions as f64
            } else {
                0.0
            },
            average_duration: if stats.total_executions > 0 {
                stats.total_duration / stats.total_executions as u32
            } else {
                Duration::from_secs(0)
            },
            total_tokens: stats.total_tokens,
            average_tokens: if stats.total_executions > 0 {
                stats.total_tokens / stats.total_executions
            } else {
                0
            },
        }
    }
}

/// Summary of key metrics
#[derive(Debug, Clone, serde::Serialize)]
pub struct MetricsSummary {
    pub total_executions: u64,
    pub successful_executions: u64,
    pub failed_executions: u64,
    pub success_rate: f64,
    pub average_duration: Duration,
    pub total_tokens: u64,
    pub average_tokens: u64,
}