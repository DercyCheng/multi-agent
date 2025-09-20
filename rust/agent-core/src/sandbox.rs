use anyhow::{Context, Result};
use std::collections::HashMap;
use std::path::PathBuf;
use std::sync::Arc;
use std::time::{Duration, Instant};
use tokio::sync::{Mutex, Semaphore};
use tracing::{debug, error, info, warn};
use uuid::Uuid;

use wasmtime::*;
use wasmtime_wasi::{WasiCtx, WasiCtxBuilder};

use crate::config::SandboxConfig;

/// WASI Sandbox provides secure execution environment for agent code
pub struct WASISandbox {
    config: SandboxConfig,
    engine: Engine,
    instances: Arc<Mutex<HashMap<String, SandboxInstance>>>,
    semaphore: Arc<Semaphore>,
}

/// Represents a single sandbox instance
pub struct SandboxInstance {
    pub id: String,
    pub store: Store<WasiCtx>,
    pub instance: Instance,
    pub created_at: Instant,
    pub last_used: Instant,
    pub execution_count: u64,
}

/// Execution context for sandbox operations
#[derive(Debug, Clone)]
pub struct ExecutionContext {
    pub user_id: String,
    pub tenant_id: String,
    pub session_id: String,
    pub execution_id: String,
    pub memory_limit: u64,
    pub cpu_limit: u64,
    pub timeout: Duration,
    pub allowed_hosts: Vec<String>,
    pub environment: HashMap<String, String>,
}

/// Result of code execution in sandbox
#[derive(Debug, Clone)]
pub struct ExecutionResult {
    pub execution_id: String,
    pub status: ExecutionStatus,
    pub output: String,
    pub error_message: Option<String>,
    pub metrics: ExecutionMetrics,
    pub duration: Duration,
}

/// Execution status enumeration
#[derive(Debug, Clone, PartialEq)]
pub enum ExecutionStatus {
    Success,
    Timeout,
    MemoryLimit,
    CpuLimit,
    SecurityViolation,
    RuntimeError,
    CompilationError,
}

/// Execution metrics
#[derive(Debug, Clone)]
pub struct ExecutionMetrics {
    pub memory_used: u64,
    pub cpu_time: Duration,
    pub syscalls_count: u64,
    pub file_operations: u64,
    pub network_requests: u64,
}

impl WASISandbox {
    /// Create a new WASI sandbox
    pub async fn new(config: &SandboxConfig) -> Result<Self> {
        info!("Initializing WASI sandbox with config: {:?}", config);

        // Create Wasmtime engine with security configurations
        let mut engine_config = Config::new();
        engine_config.wasm_component_model(true);
        engine_config.async_support(true);
        
        // Security configurations
        engine_config.consume_fuel(true);
        engine_config.epoch_interruption(true);
        engine_config.max_wasm_stack(1024 * 1024); // 1MB stack limit
        
        // Memory configurations
        engine_config.static_memory_maximum_size(config.memory_limit);
        engine_config.dynamic_memory_guard_size(65536); // 64KB guard
        
        let engine = Engine::new(&engine_config)
            .context("Failed to create Wasmtime engine")?;

        // Create temp directory if it doesn't exist
        if !config.temp_dir.exists() {
            std::fs::create_dir_all(&config.temp_dir)
                .context("Failed to create sandbox temp directory")?;
        }

        let semaphore = Arc::new(Semaphore::new(config.max_instances));

        Ok(Self {
            config: config.clone(),
            engine,
            instances: Arc::new(Mutex::new(HashMap::new())),
            semaphore,
        })
    }

    /// Execute Python code in WASI sandbox
    pub async fn execute_python(
        &self,
        code: &str,
        context: ExecutionContext,
    ) -> Result<ExecutionResult> {
        let start_time = Instant::now();
        
        debug!("Executing Python code in sandbox: {}", context.execution_id);

        // Acquire semaphore permit
        let _permit = self.semaphore.acquire().await
            .context("Failed to acquire sandbox permit")?;

        // Create WASI context with restrictions
        let wasi_ctx = self.create_wasi_context(&context)?;
        
        // Create store with resource limits
        let mut store = Store::new(&self.engine, wasi_ctx);
        
        // Set fuel limit (CPU time approximation)
        store.set_fuel(context.cpu_limit)
            .context("Failed to set fuel limit")?;
        
        // Set epoch deadline for timeout
        store.set_epoch_deadline(1);
        
        // Create Python WASM module (this would be a pre-compiled Python interpreter)
        let python_wasm = self.get_python_wasm_module().await?;
        
        // Instantiate the module
        let instance = Instance::new_async(&mut store, &python_wasm, &[]).await
            .context("Failed to instantiate Python WASM module")?;

        // Execute the code
        let result = self.execute_in_instance(
            &mut store,
            &instance,
            code,
            &context,
            start_time,
        ).await;

        // Clean up and return result
        self.cleanup_instance(&context.execution_id).await;
        
        result
    }

    /// Execute JavaScript code in WASI sandbox
    pub async fn execute_javascript(
        &self,
        code: &str,
        context: ExecutionContext,
    ) -> Result<ExecutionResult> {
        let start_time = Instant::now();
        
        debug!("Executing JavaScript code in sandbox: {}", context.execution_id);

        // Similar implementation to Python but with JavaScript runtime
        let _permit = self.semaphore.acquire().await
            .context("Failed to acquire sandbox permit")?;

        let wasi_ctx = self.create_wasi_context(&context)?;
        let mut store = Store::new(&self.engine, wasi_ctx);
        
        store.set_fuel(context.cpu_limit)
            .context("Failed to set fuel limit")?;
        store.set_epoch_deadline(1);

        // Get JavaScript WASM module (QuickJS or similar)
        let js_wasm = self.get_javascript_wasm_module().await?;
        
        let instance = Instance::new_async(&mut store, &js_wasm, &[]).await
            .context("Failed to instantiate JavaScript WASM module")?;

        let result = self.execute_in_instance(
            &mut store,
            &instance,
            code,
            &context,
            start_time,
        ).await;

        self.cleanup_instance(&context.execution_id).await;
        result
    }

    /// Create WASI context with security restrictions
    fn create_wasi_context(&self, context: &ExecutionContext) -> Result<WasiCtx> {
        let mut builder = WasiCtxBuilder::new();
        
        // Set up stdio
        builder.inherit_stdio();
        
        // Set up environment variables (filtered)
        for (key, value) in &context.environment {
            if self.is_safe_env_var(key) {
                builder.env(key, value)?;
            }
        }
        
        // Set up file system access (restricted)
        let sandbox_dir = self.config.temp_dir.join(&context.execution_id);
        std::fs::create_dir_all(&sandbox_dir)
            .context("Failed to create execution directory")?;
            
        builder.preopened_dir(
            &sandbox_dir,
            "/sandbox",
            cap_std::fs::DirPerms::all(),
            cap_std::fs::FilePerms::all(),
        )?;
        
        // Add read-only access to system libraries if needed
        builder.preopened_dir(
            "/usr/lib",
            "/usr/lib",
            cap_std::fs::DirPerms::READ,
            cap_std::fs::FilePerms::READ,
        )?;

        Ok(builder.build())
    }

    /// Execute code within a WASM instance
    async fn execute_in_instance(
        &self,
        store: &mut Store<WasiCtx>,
        instance: &Instance,
        code: &str,
        context: &ExecutionContext,
        start_time: Instant,
    ) -> Result<ExecutionResult> {
        let execution_id = context.execution_id.clone();
        
        // Start epoch thread for timeout handling
        let engine = store.engine().clone();
        let timeout_handle = tokio::spawn(async move {
            tokio::time::sleep(context.timeout).await;
            engine.increment_epoch();
        });

        // Get the main execution function
        let main_func = instance
            .get_typed_func::<(i32, i32), i32>(store, "execute_code")
            .context("Failed to get execute_code function")?;

        // Prepare code input (this is simplified - real implementation would handle memory management)
        let code_ptr = self.allocate_string_in_wasm(store, instance, code).await?;
        let code_len = code.len() as i32;

        // Execute the code
        let execution_result = main_func.call_async(store, (code_ptr, code_len)).await;
        
        // Cancel timeout
        timeout_handle.abort();
        
        let duration = start_time.elapsed();
        
        // Process execution result
        match execution_result {
            Ok(result_code) => {
                let output = self.get_execution_output(store, instance, result_code).await?;
                
                let metrics = ExecutionMetrics {
                    memory_used: self.get_memory_usage(store)?,
                    cpu_time: self.get_cpu_time(store)?,
                    syscalls_count: 0, // Would be tracked by WASI implementation
                    file_operations: 0,
                    network_requests: 0,
                };

                Ok(ExecutionResult {
                    execution_id,
                    status: if result_code == 0 { ExecutionStatus::Success } else { ExecutionStatus::RuntimeError },
                    output,
                    error_message: None,
                    metrics,
                    duration,
                })
            }
            Err(trap) => {
                let (status, error_message) = self.classify_trap(&trap);
                
                Ok(ExecutionResult {
                    execution_id,
                    status,
                    output: String::new(),
                    error_message: Some(error_message),
                    metrics: ExecutionMetrics {
                        memory_used: self.get_memory_usage(store).unwrap_or(0),
                        cpu_time: duration,
                        syscalls_count: 0,
                        file_operations: 0,
                        network_requests: 0,
                    },
                    duration,
                })
            }
        }
    }

    /// Get Python WASM module (placeholder - would load actual compiled module)
    async fn get_python_wasm_module(&self) -> Result<Module> {
        // In a real implementation, this would load a pre-compiled Python interpreter
        // For now, we'll create a simple mock module
        let wat = r#"
            (module
                (import "wasi_snapshot_preview1" "fd_write" (func $fd_write (param i32 i32 i32 i32) (result i32)))
                (memory (export "memory") 1)
                (func (export "execute_code") (param $code_ptr i32) (param $code_len i32) (result i32)
                    i32.const 0
                )
            )
        "#;
        
        Module::new(&self.engine, wat)
            .context("Failed to create Python WASM module")
    }

    /// Get JavaScript WASM module (placeholder)
    async fn get_javascript_wasm_module(&self) -> Result<Module> {
        // Similar to Python module but for JavaScript runtime
        let wat = r#"
            (module
                (import "wasi_snapshot_preview1" "fd_write" (func $fd_write (param i32 i32 i32 i32) (result i32)))
                (memory (export "memory") 1)
                (func (export "execute_code") (param $code_ptr i32) (param $code_len i32) (result i32)
                    i32.const 0
                )
            )
        "#;
        
        Module::new(&self.engine, wat)
            .context("Failed to create JavaScript WASM module")
    }

    /// Allocate string in WASM memory (simplified implementation)
    async fn allocate_string_in_wasm(
        &self,
        _store: &mut Store<WasiCtx>,
        _instance: &Instance,
        _s: &str,
    ) -> Result<i32> {
        // Simplified - real implementation would manage WASM memory
        Ok(0)
    }

    /// Get execution output from WASM instance
    async fn get_execution_output(
        &self,
        _store: &mut Store<WasiCtx>,
        _instance: &Instance,
        _result_code: i32,
    ) -> Result<String> {
        // Simplified - real implementation would read from WASM memory
        Ok("Execution completed successfully".to_string())
    }

    /// Get memory usage from store
    fn get_memory_usage(&self, store: &Store<WasiCtx>) -> Result<u64> {
        // Get memory usage from store data
        Ok(store.data().memory_consumed() as u64)
    }

    /// Get CPU time from store
    fn get_cpu_time(&self, store: &Store<WasiCtx>) -> Result<Duration> {
        // Calculate CPU time based on fuel consumed
        let fuel_consumed = store.fuel_consumed().unwrap_or(0);
        Ok(Duration::from_nanos(fuel_consumed))
    }

    /// Classify trap error
    fn classify_trap(&self, trap: &Trap) -> (ExecutionStatus, String) {
        match trap {
            Trap::OutOfFuel => (ExecutionStatus::CpuLimit, "CPU time limit exceeded".to_string()),
            Trap::Interrupt => (ExecutionStatus::Timeout, "Execution timeout".to_string()),
            Trap::MemoryOutOfBounds => (ExecutionStatus::MemoryLimit, "Memory limit exceeded".to_string()),
            _ => (ExecutionStatus::RuntimeError, format!("Runtime error: {}", trap)),
        }
    }

    /// Check if environment variable is safe to pass to sandbox
    fn is_safe_env_var(&self, key: &str) -> bool {
        const SAFE_VARS: &[&str] = &[
            "PATH", "HOME", "USER", "LANG", "LC_ALL", "TZ",
            "PYTHONPATH", "NODE_PATH",
        ];
        
        SAFE_VARS.contains(&key) || key.starts_with("AGENT_")
    }

    /// Clean up sandbox instance
    async fn cleanup_instance(&self, execution_id: &str) {
        let mut instances = self.instances.lock().await;
        instances.remove(execution_id);
        
        // Clean up temp directory
        let sandbox_dir = self.config.temp_dir.join(execution_id);
        if sandbox_dir.exists() {
            if let Err(e) = std::fs::remove_dir_all(&sandbox_dir) {
                warn!("Failed to cleanup sandbox directory: {}", e);
            }
        }
    }

    /// Get sandbox statistics
    pub async fn get_stats(&self) -> SandboxStats {
        let instances = self.instances.lock().await;
        
        SandboxStats {
            active_instances: instances.len(),
            max_instances: self.config.max_instances,
            total_executions: instances.values().map(|i| i.execution_count).sum(),
            memory_limit: self.config.memory_limit,
            cpu_limit: self.config.cpu_limit,
        }
    }
}

/// Sandbox statistics
#[derive(Debug, Clone)]
pub struct SandboxStats {
    pub active_instances: usize,
    pub max_instances: usize,
    pub total_executions: u64,
    pub memory_limit: u64,
    pub cpu_limit: u64,
}