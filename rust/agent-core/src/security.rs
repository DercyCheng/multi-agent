use anyhow::{Context, Result};
use std::collections::HashMap;
use std::path::PathBuf;
use std::sync::Arc;
use tokio::fs;
use tracing::{debug, error, info, warn};

use crate::config::SecurityConfig;

/// Security manager for code validation and policy enforcement
pub struct SecurityManager {
    config: SecurityConfig,
    policy_engine: PolicyEngine,
    code_analyzer: CodeAnalyzer,
    audit_logger: AuditLogger,
}

/// Result of security validation
#[derive(Debug, Clone)]
pub struct SecurityValidationResult {
    pub is_safe: bool,
    pub violations: Vec<String>,
    pub risk_score: f64,
    pub recommendations: Vec<String>,
}

/// Policy engine for OPA integration
struct PolicyEngine {
    policies: HashMap<String, String>,
}

/// Code analyzer for static analysis
struct CodeAnalyzer {
    dangerous_patterns: Vec<DangerousPattern>,
    allowed_imports: Vec<String>,
    blocked_functions: Vec<String>,
}

/// Dangerous code patterns
#[derive(Debug, Clone)]
struct DangerousPattern {
    pattern: regex::Regex,
    severity: Severity,
    description: String,
}

/// Security severity levels
#[derive(Debug, Clone, PartialEq)]
enum Severity {
    Low,
    Medium,
    High,
    Critical,
}

/// Audit logger for security events
struct AuditLogger {
    enabled: bool,
    log_path: PathBuf,
}

/// Security audit event
#[derive(Debug, Clone, serde::Serialize)]
struct AuditEvent {
    timestamp: chrono::DateTime<chrono::Utc>,
    user_id: String,
    event_type: String,
    severity: String,
    description: String,
    metadata: HashMap<String, String>,
}

impl SecurityManager {
    /// Create a new security manager
    pub async fn new(config: &SecurityConfig) -> Result<Self> {
        info!("Initializing security manager");

        let policy_engine = PolicyEngine::new(&config.opa_policy_path).await?;
        let code_analyzer = CodeAnalyzer::new().await?;
        let audit_logger = AuditLogger::new(config.enable_audit_log, &config.audit_log_path)?;

        Ok(Self {
            config: config.clone(),
            policy_engine,
            code_analyzer,
            audit_logger,
        })
    }

    /// Validate code for security issues
    pub async fn validate_code(&self, code: &str, user_id: &str) -> Result<SecurityValidationResult> {
        debug!("Validating code for user: {}", user_id);

        let mut violations = Vec::new();
        let mut risk_score = 0.0;
        let mut recommendations = Vec::new();

        // 1. Static code analysis
        let analysis_result = self.code_analyzer.analyze_code(code).await?;
        violations.extend(analysis_result.violations);
        risk_score += analysis_result.risk_score;
        recommendations.extend(analysis_result.recommendations);

        // 2. Policy evaluation
        let policy_result = self.policy_engine.evaluate_code_policy(code, user_id).await?;
        if !policy_result.allowed {
            violations.push(format!("Policy violation: {}", policy_result.reason));
            risk_score += 50.0; // High penalty for policy violations
        }

        // 3. Determine if code is safe
        let is_safe = violations.is_empty() && risk_score < 70.0;

        // 4. Log audit event
        self.audit_logger.log_validation_event(
            user_id,
            is_safe,
            risk_score,
            &violations,
        ).await?;

        Ok(SecurityValidationResult {
            is_safe,
            violations,
            risk_score,
            recommendations,
        })
    }

    /// Validate network access request
    pub async fn validate_network_access(
        &self,
        host: &str,
        port: u16,
        user_id: &str,
    ) -> Result<bool> {
        debug!("Validating network access to {}:{} for user {}", host, port, user_id);

        // Check against allowed hosts
        let allowed = self.is_host_allowed(host) && self.is_port_allowed(port);

        // Log audit event
        self.audit_logger.log_network_access_event(
            user_id,
            host,
            port,
            allowed,
        ).await?;

        Ok(allowed)
    }

    /// Check if host is allowed
    fn is_host_allowed(&self, host: &str) -> bool {
        // Allow localhost and specific whitelisted hosts
        if host == "localhost" || host == "127.0.0.1" || host == "::1" {
            return true;
        }

        // Check against configuration (would be loaded from config)
        let allowed_hosts = vec![
            "api.openai.com",
            "api.anthropic.com",
            "api.cohere.ai",
            "httpbin.org", // For testing
        ];

        allowed_hosts.iter().any(|&allowed| host.contains(allowed))
    }

    /// Check if port is allowed
    fn is_port_allowed(&self, port: u16) -> bool {
        // Allow standard HTTP/HTTPS ports and some common API ports
        matches!(port, 80 | 443 | 8000..=8999)
    }

    /// Encrypt sensitive data
    pub fn encrypt_data(&self, data: &str) -> Result<String> {
        // Simple encryption implementation (would use proper crypto in production)
        use ring::aead::{Aad, LessSafeKey, Nonce, UnboundKey, AES_256_GCM};
        use ring::rand::{SecureRandom, SystemRandom};

        let rng = SystemRandom::new();
        let mut key_bytes = [0u8; 32];
        rng.fill(&mut key_bytes)?;

        let unbound_key = UnboundKey::new(&AES_256_GCM, &key_bytes)?;
        let key = LessSafeKey::new(unbound_key);

        let mut nonce_bytes = [0u8; 12];
        rng.fill(&mut nonce_bytes)?;
        let nonce = Nonce::assume_unique_for_key(nonce_bytes);

        let mut in_out = data.as_bytes().to_vec();
        key.seal_in_place_append_tag(nonce, Aad::empty(), &mut in_out)?;

        // Combine nonce and ciphertext
        let mut result = nonce_bytes.to_vec();
        result.extend_from_slice(&in_out);

        Ok(base64::encode(result))
    }

    /// Decrypt sensitive data
    pub fn decrypt_data(&self, encrypted_data: &str) -> Result<String> {
        // Simple decryption implementation
        let data = base64::decode(encrypted_data)?;
        
        if data.len() < 12 {
            return Err(anyhow::anyhow!("Invalid encrypted data"));
        }

        // In a real implementation, we would properly manage keys
        // For now, just return a placeholder
        Ok("decrypted_data".to_string())
    }
}

impl PolicyEngine {
    async fn new(policy_path: &PathBuf) -> Result<Self> {
        let mut policies = HashMap::new();

        // Load policy files
        if policy_path.exists() {
            let mut entries = fs::read_dir(policy_path).await?;
            while let Some(entry) = entries.next_entry().await? {
                let path = entry.path();
                if path.extension().and_then(|s| s.to_str()) == Some("rego") {
                    let policy_name = path.file_stem()
                        .and_then(|s| s.to_str())
                        .unwrap_or("unknown")
                        .to_string();
                    let policy_content = fs::read_to_string(&path).await?;
                    policies.insert(policy_name, policy_content);
                }
            }
        }

        Ok(Self { policies })
    }

    async fn evaluate_code_policy(&self, code: &str, user_id: &str) -> Result<PolicyResult> {
        // Simplified policy evaluation
        // In a real implementation, this would use OPA's Rego engine

        // Check for dangerous imports
        if code.contains("import os") && code.contains("system") {
            return Ok(PolicyResult {
                allowed: false,
                reason: "System command execution not allowed".to_string(),
            });
        }

        // Check for file system access
        if code.contains("open(") && (code.contains("'w'") || code.contains("'a'")) {
            return Ok(PolicyResult {
                allowed: false,
                reason: "File write access not allowed".to_string(),
            });
        }

        // Check for network access
        if code.contains("requests.") || code.contains("urllib") || code.contains("socket") {
            return Ok(PolicyResult {
                allowed: false,
                reason: "Direct network access not allowed".to_string(),
            });
        }

        Ok(PolicyResult {
            allowed: true,
            reason: "Code passed policy evaluation".to_string(),
        })
    }
}

#[derive(Debug)]
struct PolicyResult {
    allowed: bool,
    reason: String,
}

impl CodeAnalyzer {
    async fn new() -> Result<Self> {
        let dangerous_patterns = vec![
            DangerousPattern {
                pattern: regex::Regex::new(r"eval\s*\(")?,
                severity: Severity::Critical,
                description: "Use of eval() function".to_string(),
            },
            DangerousPattern {
                pattern: regex::Regex::new(r"exec\s*\(")?,
                severity: Severity::Critical,
                description: "Use of exec() function".to_string(),
            },
            DangerousPattern {
                pattern: regex::Regex::new(r"__import__\s*\(")?,
                severity: Severity::High,
                description: "Dynamic import usage".to_string(),
            },
            DangerousPattern {
                pattern: regex::Regex::new(r"subprocess\.")?,
                severity: Severity::High,
                description: "Subprocess execution".to_string(),
            },
            DangerousPattern {
                pattern: regex::Regex::new(r"os\.system")?,
                severity: Severity::Critical,
                description: "System command execution".to_string(),
            },
            DangerousPattern {
                pattern: regex::Regex::new(r"pickle\.loads")?,
                severity: Severity::High,
                description: "Unsafe deserialization".to_string(),
            },
        ];

        let allowed_imports = vec![
            "json".to_string(),
            "math".to_string(),
            "datetime".to_string(),
            "re".to_string(),
            "collections".to_string(),
            "itertools".to_string(),
            "functools".to_string(),
        ];

        let blocked_functions = vec![
            "eval".to_string(),
            "exec".to_string(),
            "compile".to_string(),
            "__import__".to_string(),
        ];

        Ok(Self {
            dangerous_patterns,
            allowed_imports,
            blocked_functions,
        })
    }

    async fn analyze_code(&self, code: &str) -> Result<CodeAnalysisResult> {
        let mut violations = Vec::new();
        let mut risk_score = 0.0;
        let mut recommendations = Vec::new();

        // Check for dangerous patterns
        for pattern in &self.dangerous_patterns {
            if pattern.pattern.is_match(code) {
                let severity_score = match pattern.severity {
                    Severity::Low => 10.0,
                    Severity::Medium => 25.0,
                    Severity::High => 50.0,
                    Severity::Critical => 100.0,
                };

                violations.push(format!("{}: {}", pattern.severity_str(), pattern.description));
                risk_score += severity_score;
                
                recommendations.push(format!("Remove or replace: {}", pattern.description));
            }
        }

        // Check for blocked functions
        for func in &self.blocked_functions {
            if code.contains(func) {
                violations.push(format!("Blocked function usage: {}", func));
                risk_score += 30.0;
            }
        }

        // Check import statements
        for line in code.lines() {
            let trimmed = line.trim();
            if trimmed.starts_with("import ") || trimmed.starts_with("from ") {
                if !self.is_import_allowed(trimmed) {
                    violations.push(format!("Disallowed import: {}", trimmed));
                    risk_score += 20.0;
                }
            }
        }

        Ok(CodeAnalysisResult {
            violations,
            risk_score,
            recommendations,
        })
    }

    fn is_import_allowed(&self, import_line: &str) -> bool {
        // Simple import validation
        for allowed in &self.allowed_imports {
            if import_line.contains(allowed) {
                return true;
            }
        }
        false
    }
}

#[derive(Debug)]
struct CodeAnalysisResult {
    violations: Vec<String>,
    risk_score: f64,
    recommendations: Vec<String>,
}

impl Severity {
    fn severity_str(&self) -> &'static str {
        match self {
            Severity::Low => "LOW",
            Severity::Medium => "MEDIUM",
            Severity::High => "HIGH",
            Severity::Critical => "CRITICAL",
        }
    }
}

impl AuditLogger {
    fn new(enabled: bool, log_path: &PathBuf) -> Result<Self> {
        if enabled {
            // Ensure log directory exists
            if let Some(parent) = log_path.parent() {
                std::fs::create_dir_all(parent)?;
            }
        }

        Ok(Self {
            enabled,
            log_path: log_path.clone(),
        })
    }

    async fn log_validation_event(
        &self,
        user_id: &str,
        is_safe: bool,
        risk_score: f64,
        violations: &[String],
    ) -> Result<()> {
        if !self.enabled {
            return Ok(());
        }

        let event = AuditEvent {
            timestamp: chrono::Utc::now(),
            user_id: user_id.to_string(),
            event_type: "code_validation".to_string(),
            severity: if is_safe { "INFO".to_string() } else { "WARNING".to_string() },
            description: format!("Code validation result: safe={}, risk_score={:.2}", is_safe, risk_score),
            metadata: HashMap::from([
                ("risk_score".to_string(), risk_score.to_string()),
                ("violations_count".to_string(), violations.len().to_string()),
                ("violations".to_string(), violations.join("; ")),
            ]),
        };

        self.write_audit_event(&event).await
    }

    async fn log_network_access_event(
        &self,
        user_id: &str,
        host: &str,
        port: u16,
        allowed: bool,
    ) -> Result<()> {
        if !self.enabled {
            return Ok(());
        }

        let event = AuditEvent {
            timestamp: chrono::Utc::now(),
            user_id: user_id.to_string(),
            event_type: "network_access".to_string(),
            severity: if allowed { "INFO".to_string() } else { "WARNING".to_string() },
            description: format!("Network access request: {}:{} - {}", host, port, if allowed { "ALLOWED" } else { "DENIED" }),
            metadata: HashMap::from([
                ("host".to_string(), host.to_string()),
                ("port".to_string(), port.to_string()),
                ("allowed".to_string(), allowed.to_string()),
            ]),
        };

        self.write_audit_event(&event).await
    }

    async fn write_audit_event(&self, event: &AuditEvent) -> Result<()> {
        let log_line = serde_json::to_string(event)? + "\n";
        
        tokio::fs::OpenOptions::new()
            .create(true)
            .append(true)
            .open(&self.log_path)
            .await?
            .write_all(log_line.as_bytes())
            .await?;

        Ok(())
    }
}