use anyhow::Result;
use std::sync::Arc;
use tokio::signal;
use tracing::{info, warn};

mod config;
mod enforcement;
mod execution;
mod fsm;
mod grpc;
mod metrics;
mod sandbox;
mod security;

use crate::config::Config;
use crate::enforcement::EnforcementGateway;
use crate::execution::ExecutionEngine;
use crate::fsm::StateMachine;
use crate::grpc::AgentCoreService;
use crate::metrics::MetricsCollector;
use crate::sandbox::WASISandbox;
use crate::security::SecurityManager;

#[tokio::main]
async fn main() -> Result<()> {
    // Initialize tracing
    tracing_subscriber::fmt()
        .with_env_filter(tracing_subscriber::EnvFilter::from_default_env())
        .json()
        .init();

    info!("Starting Multi-Agent Core Execution Engine");

    // Load configuration
    let config = Config::from_env()?;
    info!("Configuration loaded successfully");

    // Initialize metrics collector
    let metrics = Arc::new(MetricsCollector::new()?);
    
    // Initialize security manager
    let security_manager = Arc::new(SecurityManager::new(&config.security).await?);
    
    // Initialize WASI sandbox
    let sandbox = Arc::new(WASISandbox::new(&config.sandbox).await?);
    
    // Initialize enforcement gateway
    let enforcement = Arc::new(EnforcementGateway::new(&config.enforcement, metrics.clone()).await?);
    
    // Initialize state machine
    let state_machine = Arc::new(StateMachine::new(&config.fsm)?);
    
    // Initialize execution engine
    let execution_engine = Arc::new(ExecutionEngine::new(
        sandbox.clone(),
        enforcement.clone(),
        security_manager.clone(),
        state_machine.clone(),
        metrics.clone(),
    )?);

    // Initialize gRPC service
    let grpc_service = AgentCoreService::new(
        execution_engine.clone(),
        metrics.clone(),
    );

    // Start metrics server
    let metrics_handle = tokio::spawn({
        let metrics = metrics.clone();
        async move {
            if let Err(e) = metrics.start_server().await {
                warn!("Metrics server error: {}", e);
            }
        }
    });

    // Start gRPC server
    let grpc_handle = tokio::spawn({
        let service = grpc_service;
        let addr = config.server.grpc_addr.parse()?;
        async move {
            info!("Starting gRPC server on {}", addr);
            tonic::transport::Server::builder()
                .add_service(service.into_service())
                .serve(addr)
                .await
        }
    });

    info!("Multi-Agent Core started successfully");

    // Wait for shutdown signal
    tokio::select! {
        _ = signal::ctrl_c() => {
            info!("Received shutdown signal");
        }
        result = grpc_handle => {
            if let Err(e) = result? {
                warn!("gRPC server error: {}", e);
            }
        }
        result = metrics_handle => {
            if let Err(e) = result? {
                warn!("Metrics server join error: {}", e);
            }
        }
    }

    info!("Shutting down Multi-Agent Core");
    Ok(())
}