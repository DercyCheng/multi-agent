use anyhow::Result;
use serde::{Deserialize, Serialize};
use std::net::SocketAddr;
use std::path::PathBuf;
use std::time::Duration;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Config {
    pub server: ServerConfig,
    pub sandbox: SandboxConfig,
    pub enforcement: EnforcementConfig,
    pub security: SecurityConfig,
    pub fsm: FSMConfig,
    pub metrics: MetricsConfig,
    pub logging: LoggingConfig,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ServerConfig {
    pub grpc_addr: String,
    pub metrics_addr: String,
    pub max_connections: usize,
    pub request_timeout: Duration,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SandboxConfig {
    pub memory_limit: u64,        // Memory limit in bytes
    pub cpu_limit: u64,           // CPU time limit in nanoseconds
    pub execution_timeout: Duration,
    pub max_file_size: u64,       // Maximum file size in bytes
    pub allowed_hosts: Vec<String>,
    pub blocked_syscalls: Vec<String>,
    pub temp_dir: PathBuf,
    pub max_instances: usize,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct EnforcementConfig {
    pub timeout_config: TimeoutConfig,
    pub rate_limit_config: RateLimitConfig,
    pub circuit_breaker_config: CircuitBreakerConfig,
    pub token_validator_config: TokenValidatorConfig,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TimeoutConfig {
    pub max_duration: Duration,
    pub warning_threshold: Duration,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RateLimitConfig {
    pub requests_per_second: u32,
    pub burst_size: u32,
    pub window_size: Duration,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CircuitBreakerConfig {
    pub failure_threshold: u32,
    pub success_threshold: u32,
    pub timeout: Duration,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TokenValidatorConfig {
    pub max_tokens: u32,
    pub cost_per_token: f64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SecurityConfig {
    pub opa_policy_path: PathBuf,
    pub encryption_key_path: PathBuf,
    pub tls_cert_path: Option<PathBuf>,
    pub tls_key_path: Option<PathBuf>,
    pub enable_audit_log: bool,
    pub audit_log_path: PathBuf,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct FSMConfig {
    pub max_states: usize,
    pub max_transitions: usize,
    pub state_timeout: Duration,
    pub persistence_enabled: bool,
    pub persistence_path: PathBuf,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MetricsConfig {
    pub enabled: bool,
    pub addr: SocketAddr,
    pub path: String,
    pub collection_interval: Duration,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct LoggingConfig {
    pub level: String,
    pub format: String, // json or text
    pub output: String, // stdout, stderr, or file path
}

impl Config {
    pub fn from_env() -> Result<Self> {
        let config = Config {
            server: ServerConfig {
                grpc_addr: std::env::var("GRPC_ADDR").unwrap_or_else(|_| "0.0.0.0:50051".to_string()),
                metrics_addr: std::env::var("METRICS_ADDR").unwrap_or_else(|_| "0.0.0.0:2113".to_string()),
                max_connections: std::env::var("MAX_CONNECTIONS")
                    .unwrap_or_else(|_| "1000".to_string())
                    .parse()
                    .unwrap_or(1000),
                request_timeout: Duration::from_secs(
                    std::env::var("REQUEST_TIMEOUT_SECONDS")
                        .unwrap_or_else(|_| "30".to_string())
                        .parse()
                        .unwrap_or(30),
                ),
            },
            sandbox: SandboxConfig {
                memory_limit: std::env::var("SANDBOX_MEMORY_LIMIT")
                    .unwrap_or_else(|_| "134217728".to_string()) // 128MB
                    .parse()
                    .unwrap_or(134217728),
                cpu_limit: std::env::var("SANDBOX_CPU_LIMIT")
                    .unwrap_or_else(|_| "5000000000".to_string()) // 5 seconds in nanoseconds
                    .parse()
                    .unwrap_or(5000000000),
                execution_timeout: Duration::from_secs(
                    std::env::var("SANDBOX_EXECUTION_TIMEOUT")
                        .unwrap_or_else(|_| "30".to_string())
                        .parse()
                        .unwrap_or(30),
                ),
                max_file_size: std::env::var("SANDBOX_MAX_FILE_SIZE")
                    .unwrap_or_else(|_| "10485760".to_string()) // 10MB
                    .parse()
                    .unwrap_or(10485760),
                allowed_hosts: std::env::var("SANDBOX_ALLOWED_HOSTS")
                    .unwrap_or_else(|_| "localhost,127.0.0.1".to_string())
                    .split(',')
                    .map(|s| s.trim().to_string())
                    .collect(),
                blocked_syscalls: std::env::var("SANDBOX_BLOCKED_SYSCALLS")
                    .unwrap_or_else(|_| "execve,fork,clone".to_string())
                    .split(',')
                    .map(|s| s.trim().to_string())
                    .collect(),
                temp_dir: PathBuf::from(
                    std::env::var("SANDBOX_TEMP_DIR").unwrap_or_else(|_| "/tmp/agent-sandbox".to_string())
                ),
                max_instances: std::env::var("SANDBOX_MAX_INSTANCES")
                    .unwrap_or_else(|_| "100".to_string())
                    .parse()
                    .unwrap_or(100),
            },
            enforcement: EnforcementConfig {
                timeout_config: TimeoutConfig {
                    max_duration: Duration::from_secs(
                        std::env::var("ENFORCEMENT_MAX_DURATION")
                            .unwrap_or_else(|_| "300".to_string())
                            .parse()
                            .unwrap_or(300),
                    ),
                    warning_threshold: Duration::from_secs(
                        std::env::var("ENFORCEMENT_WARNING_THRESHOLD")
                            .unwrap_or_else(|_| "60".to_string())
                            .parse()
                            .unwrap_or(60),
                    ),
                },
                rate_limit_config: RateLimitConfig {
                    requests_per_second: std::env::var("RATE_LIMIT_RPS")
                        .unwrap_or_else(|_| "100".to_string())
                        .parse()
                        .unwrap_or(100),
                    burst_size: std::env::var("RATE_LIMIT_BURST")
                        .unwrap_or_else(|_| "200".to_string())
                        .parse()
                        .unwrap_or(200),
                    window_size: Duration::from_secs(
                        std::env::var("RATE_LIMIT_WINDOW")
                            .unwrap_or_else(|_| "60".to_string())
                            .parse()
                            .unwrap_or(60),
                    ),
                },
                circuit_breaker_config: CircuitBreakerConfig {
                    failure_threshold: std::env::var("CIRCUIT_BREAKER_FAILURE_THRESHOLD")
                        .unwrap_or_else(|_| "5".to_string())
                        .parse()
                        .unwrap_or(5),
                    success_threshold: std::env::var("CIRCUIT_BREAKER_SUCCESS_THRESHOLD")
                        .unwrap_or_else(|_| "3".to_string())
                        .parse()
                        .unwrap_or(3),
                    timeout: Duration::from_secs(
                        std::env::var("CIRCUIT_BREAKER_TIMEOUT")
                            .unwrap_or_else(|_| "60".to_string())
                            .parse()
                            .unwrap_or(60),
                    ),
                },
                token_validator_config: TokenValidatorConfig {
                    max_tokens: std::env::var("TOKEN_VALIDATOR_MAX_TOKENS")
                        .unwrap_or_else(|_| "10000".to_string())
                        .parse()
                        .unwrap_or(10000),
                    cost_per_token: std::env::var("TOKEN_VALIDATOR_COST_PER_TOKEN")
                        .unwrap_or_else(|_| "0.002".to_string())
                        .parse()
                        .unwrap_or(0.002),
                },
            },
            security: SecurityConfig {
                opa_policy_path: PathBuf::from(
                    std::env::var("OPA_POLICY_PATH").unwrap_or_else(|_| "/app/policies".to_string())
                ),
                encryption_key_path: PathBuf::from(
                    std::env::var("ENCRYPTION_KEY_PATH").unwrap_or_else(|_| "/app/keys/encryption.key".to_string())
                ),
                tls_cert_path: std::env::var("TLS_CERT_PATH").ok().map(PathBuf::from),
                tls_key_path: std::env::var("TLS_KEY_PATH").ok().map(PathBuf::from),
                enable_audit_log: std::env::var("ENABLE_AUDIT_LOG")
                    .unwrap_or_else(|_| "true".to_string())
                    .parse()
                    .unwrap_or(true),
                audit_log_path: PathBuf::from(
                    std::env::var("AUDIT_LOG_PATH").unwrap_or_else(|_| "/var/log/agent-audit.log".to_string())
                ),
            },
            fsm: FSMConfig {
                max_states: std::env::var("FSM_MAX_STATES")
                    .unwrap_or_else(|_| "1000".to_string())
                    .parse()
                    .unwrap_or(1000),
                max_transitions: std::env::var("FSM_MAX_TRANSITIONS")
                    .unwrap_or_else(|_| "10000".to_string())
                    .parse()
                    .unwrap_or(10000),
                state_timeout: Duration::from_secs(
                    std::env::var("FSM_STATE_TIMEOUT")
                        .unwrap_or_else(|_| "300".to_string())
                        .parse()
                        .unwrap_or(300),
                ),
                persistence_enabled: std::env::var("FSM_PERSISTENCE_ENABLED")
                    .unwrap_or_else(|_| "true".to_string())
                    .parse()
                    .unwrap_or(true),
                persistence_path: PathBuf::from(
                    std::env::var("FSM_PERSISTENCE_PATH").unwrap_or_else(|_| "/var/lib/agent-fsm".to_string())
                ),
            },
            metrics: MetricsConfig {
                enabled: std::env::var("METRICS_ENABLED")
                    .unwrap_or_else(|_| "true".to_string())
                    .parse()
                    .unwrap_or(true),
                addr: std::env::var("METRICS_ADDR")
                    .unwrap_or_else(|_| "0.0.0.0:2113".to_string())
                    .parse()
                    .unwrap_or_else(|_| "0.0.0.0:2113".parse().unwrap()),
                path: std::env::var("METRICS_PATH").unwrap_or_else(|_| "/metrics".to_string()),
                collection_interval: Duration::from_secs(
                    std::env::var("METRICS_COLLECTION_INTERVAL")
                        .unwrap_or_else(|_| "15".to_string())
                        .parse()
                        .unwrap_or(15),
                ),
            },
            logging: LoggingConfig {
                level: std::env::var("LOG_LEVEL").unwrap_or_else(|_| "info".to_string()),
                format: std::env::var("LOG_FORMAT").unwrap_or_else(|_| "json".to_string()),
                output: std::env::var("LOG_OUTPUT").unwrap_or_else(|_| "stdout".to_string()),
            },
        };

        Ok(config)
    }

    pub fn validate(&self) -> Result<()> {
        // Validate server configuration
        if self.server.max_connections == 0 {
            return Err(anyhow::anyhow!("max_connections must be greater than 0"));
        }

        // Validate sandbox configuration
        if self.sandbox.memory_limit == 0 {
            return Err(anyhow::anyhow!("sandbox memory_limit must be greater than 0"));
        }

        if self.sandbox.max_instances == 0 {
            return Err(anyhow::anyhow!("sandbox max_instances must be greater than 0"));
        }

        // Validate FSM configuration
        if self.fsm.max_states == 0 {
            return Err(anyhow::anyhow!("fsm max_states must be greater than 0"));
        }

        // Validate security paths exist if specified
        if !self.security.opa_policy_path.exists() {
            return Err(anyhow::anyhow!("OPA policy path does not exist: {:?}", self.security.opa_policy_path));
        }

        Ok(())
    }
}