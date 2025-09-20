use anyhow::{Context, Result};
use serde::{Deserialize, Serialize};
use std::collections::{HashMap, HashSet};
use std::sync::Arc;
use std::time::{Duration, Instant};
use tokio::sync::{Mutex, RwLock};
use tracing::{debug, error, info, warn};
use uuid::Uuid;

use crate::config::FSMConfig;

/// Finite State Machine for agent execution control
pub struct StateMachine {
    config: FSMConfig,
    states: Arc<RwLock<HashMap<String, State>>>,
    transitions: Arc<RwLock<HashMap<String, Vec<Transition>>>>,
    active_instances: Arc<Mutex<HashMap<String, StateMachineInstance>>>,
}

/// Represents a state in the FSM
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct State {
    pub id: String,
    pub name: String,
    pub state_type: StateType,
    pub entry_actions: Vec<Action>,
    pub exit_actions: Vec<Action>,
    pub timeout: Option<Duration>,
    pub metadata: HashMap<String, String>,
}

/// Types of states in the FSM
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub enum StateType {
    Initial,
    Processing,
    Waiting,
    Decision,
    Terminal,
    Error,
}

/// Represents a transition between states
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Transition {
    pub id: String,
    pub from_state: String,
    pub to_state: String,
    pub condition: TransitionCondition,
    pub actions: Vec<Action>,
    pub priority: u32,
}

/// Conditions for state transitions
#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum TransitionCondition {
    Always,
    OnEvent(String),
    OnTimeout,
    OnSuccess,
    OnError,
    OnCondition(String), // Expression to evaluate
    Custom(String),      // Custom condition handler
}

/// Actions to execute during state transitions
#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum Action {
    Log(String),
    SetVariable(String, String),
    CallFunction(String, Vec<String>),
    SendEvent(String, HashMap<String, String>),
    UpdateMetrics(String, f64),
    Custom(String, HashMap<String, String>),
}

/// Instance of a running state machine
#[derive(Debug, Clone)]
pub struct StateMachineInstance {
    pub id: String,
    pub current_state: String,
    pub context: StateMachineContext,
    pub created_at: Instant,
    pub last_transition: Instant,
    pub transition_count: u64,
    pub status: InstanceStatus,
}

/// Context for state machine execution
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct StateMachineContext {
    pub variables: HashMap<String, String>,
    pub events: Vec<Event>,
    pub execution_history: Vec<TransitionRecord>,
    pub metadata: HashMap<String, String>,
}

/// Events that can trigger state transitions
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Event {
    pub id: String,
    pub event_type: String,
    pub payload: HashMap<String, String>,
    pub timestamp: chrono::DateTime<chrono::Utc>,
}

/// Record of state transitions
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TransitionRecord {
    pub from_state: String,
    pub to_state: String,
    pub transition_id: String,
    pub timestamp: chrono::DateTime<chrono::Utc>,
    pub duration: Duration,
    pub success: bool,
    pub error_message: Option<String>,
}

/// Status of state machine instance
#[derive(Debug, Clone, PartialEq)]
pub enum InstanceStatus {
    Running,
    Paused,
    Completed,
    Failed,
    Timeout,
}

/// Result of state machine operations
#[derive(Debug, Clone)]
pub struct StateMachineResult {
    pub instance_id: String,
    pub final_state: String,
    pub status: InstanceStatus,
    pub execution_time: Duration,
    pub transition_count: u64,
    pub context: StateMachineContext,
}

impl StateMachine {
    /// Create a new state machine
    pub fn new(config: &FSMConfig) -> Result<Self> {
        info!("Initializing FSM with config: {:?}", config);

        // Create persistence directory if needed
        if config.persistence_enabled && !config.persistence_path.exists() {
            std::fs::create_dir_all(&config.persistence_path)
                .context("Failed to create FSM persistence directory")?;
        }

        let fsm = Self {
            config: config.clone(),
            states: Arc::new(RwLock::new(HashMap::new())),
            transitions: Arc::new(RwLock::new(HashMap::new())),
            active_instances: Arc::new(Mutex::new(HashMap::new())),
        };

        // Load default states and transitions
        fsm.initialize_default_fsm()?;

        Ok(fsm)
    }

    /// Initialize default FSM structure for agent execution
    fn initialize_default_fsm(&self) -> Result<()> {
        // Define default states for agent execution
        let default_states = vec![
            State {
                id: "initial".to_string(),
                name: "Initial".to_string(),
                state_type: StateType::Initial,
                entry_actions: vec![Action::Log("Agent execution started".to_string())],
                exit_actions: vec![],
                timeout: Some(Duration::from_secs(30)),
                metadata: HashMap::new(),
            },
            State {
                id: "analyzing".to_string(),
                name: "Analyzing Task".to_string(),
                state_type: StateType::Processing,
                entry_actions: vec![Action::Log("Starting task analysis".to_string())],
                exit_actions: vec![],
                timeout: Some(Duration::from_secs(60)),
                metadata: HashMap::new(),
            },
            State {
                id: "planning".to_string(),
                name: "Planning Execution".to_string(),
                state_type: StateType::Processing,
                entry_actions: vec![Action::Log("Planning execution strategy".to_string())],
                exit_actions: vec![],
                timeout: Some(Duration::from_secs(45)),
                metadata: HashMap::new(),
            },
            State {
                id: "executing".to_string(),
                name: "Executing Task".to_string(),
                state_type: StateType::Processing,
                entry_actions: vec![Action::Log("Executing task".to_string())],
                exit_actions: vec![],
                timeout: Some(Duration::from_secs(300)),
                metadata: HashMap::new(),
            },
            State {
                id: "validating".to_string(),
                name: "Validating Results".to_string(),
                state_type: StateType::Processing,
                entry_actions: vec![Action::Log("Validating execution results".to_string())],
                exit_actions: vec![],
                timeout: Some(Duration::from_secs(30)),
                metadata: HashMap::new(),
            },
            State {
                id: "completed".to_string(),
                name: "Completed".to_string(),
                state_type: StateType::Terminal,
                entry_actions: vec![Action::Log("Task completed successfully".to_string())],
                exit_actions: vec![],
                timeout: None,
                metadata: HashMap::new(),
            },
            State {
                id: "failed".to_string(),
                name: "Failed".to_string(),
                state_type: StateType::Error,
                entry_actions: vec![Action::Log("Task execution failed".to_string())],
                exit_actions: vec![],
                timeout: None,
                metadata: HashMap::new(),
            },
        ];

        // Define default transitions
        let default_transitions = vec![
            Transition {
                id: "init_to_analyzing".to_string(),
                from_state: "initial".to_string(),
                to_state: "analyzing".to_string(),
                condition: TransitionCondition::Always,
                actions: vec![],
                priority: 1,
            },
            Transition {
                id: "analyzing_to_planning".to_string(),
                from_state: "analyzing".to_string(),
                to_state: "planning".to_string(),
                condition: TransitionCondition::OnSuccess,
                actions: vec![],
                priority: 1,
            },
            Transition {
                id: "planning_to_executing".to_string(),
                from_state: "planning".to_string(),
                to_state: "executing".to_string(),
                condition: TransitionCondition::OnSuccess,
                actions: vec![],
                priority: 1,
            },
            Transition {
                id: "executing_to_validating".to_string(),
                from_state: "executing".to_string(),
                to_state: "validating".to_string(),
                condition: TransitionCondition::OnSuccess,
                actions: vec![],
                priority: 1,
            },
            Transition {
                id: "validating_to_completed".to_string(),
                from_state: "validating".to_string(),
                to_state: "completed".to_string(),
                condition: TransitionCondition::OnSuccess,
                actions: vec![],
                priority: 1,
            },
            // Error transitions
            Transition {
                id: "any_to_failed".to_string(),
                from_state: "*".to_string(), // Wildcard for any state
                to_state: "failed".to_string(),
                condition: TransitionCondition::OnError,
                actions: vec![Action::Log("Transitioning to failed state".to_string())],
                priority: 10, // High priority for error handling
            },
            // Timeout transitions
            Transition {
                id: "timeout_to_failed".to_string(),
                from_state: "*".to_string(),
                to_state: "failed".to_string(),
                condition: TransitionCondition::OnTimeout,
                actions: vec![Action::Log("State timeout occurred".to_string())],
                priority: 9,
            },
        ];

        // Add states and transitions
        tokio::task::block_in_place(|| {
            tokio::runtime::Handle::current().block_on(async {
                self.add_states(default_states).await?;
                self.add_transitions(default_transitions).await
            })
        })
    }

    /// Add states to the FSM
    pub async fn add_states(&self, states: Vec<State>) -> Result<()> {
        let mut state_map = self.states.write().await;
        
        for state in states {
            if state_map.len() >= self.config.max_states {
                return Err(anyhow::anyhow!("Maximum number of states exceeded"));
            }
            
            debug!("Adding state: {}", state.id);
            state_map.insert(state.id.clone(), state);
        }
        
        Ok(())
    }

    /// Add transitions to the FSM
    pub async fn add_transitions(&self, transitions: Vec<Transition>) -> Result<()> {
        let mut transition_map = self.transitions.write().await;
        
        for transition in transitions {
            let total_transitions: usize = transition_map.values().map(|v| v.len()).sum();
            if total_transitions >= self.config.max_transitions {
                return Err(anyhow::anyhow!("Maximum number of transitions exceeded"));
            }
            
            debug!("Adding transition: {} -> {}", transition.from_state, transition.to_state);
            
            transition_map
                .entry(transition.from_state.clone())
                .or_insert_with(Vec::new)
                .push(transition);
        }
        
        Ok(())
    }

    /// Create a new state machine instance
    pub async fn create_instance(&self, initial_context: StateMachineContext) -> Result<String> {
        let instance_id = Uuid::new_v4().to_string();
        
        let instance = StateMachineInstance {
            id: instance_id.clone(),
            current_state: "initial".to_string(),
            context: initial_context,
            created_at: Instant::now(),
            last_transition: Instant::now(),
            transition_count: 0,
            status: InstanceStatus::Running,
        };

        let mut instances = self.active_instances.lock().await;
        instances.insert(instance_id.clone(), instance);
        
        info!("Created FSM instance: {}", instance_id);
        
        // Execute entry actions for initial state
        self.execute_state_entry_actions(&instance_id, "initial").await?;
        
        Ok(instance_id)
    }

    /// Trigger an event on a state machine instance
    pub async fn trigger_event(&self, instance_id: &str, event: Event) -> Result<()> {
        debug!("Triggering event {} on instance {}", event.event_type, instance_id);
        
        let mut instances = self.active_instances.lock().await;
        let instance = instances.get_mut(instance_id)
            .ok_or_else(|| anyhow::anyhow!("Instance not found: {}", instance_id))?;

        // Add event to context
        instance.context.events.push(event.clone());
        
        // Check for applicable transitions
        let current_state = instance.current_state.clone();
        drop(instances); // Release lock before async operations
        
        self.check_transitions(instance_id, &current_state, Some(&event)).await
    }

    /// Check and execute applicable transitions
    async fn check_transitions(
        &self,
        instance_id: &str,
        current_state: &str,
        event: Option<&Event>,
    ) -> Result<()> {
        let transitions = self.transitions.read().await;
        
        // Get transitions from current state and wildcard transitions
        let mut applicable_transitions = Vec::new();
        
        if let Some(state_transitions) = transitions.get(current_state) {
            applicable_transitions.extend(state_transitions.iter());
        }
        
        if let Some(wildcard_transitions) = transitions.get("*") {
            applicable_transitions.extend(wildcard_transitions.iter());
        }
        
        // Sort by priority (higher priority first)
        applicable_transitions.sort_by(|a, b| b.priority.cmp(&a.priority));
        
        // Find first applicable transition
        for transition in applicable_transitions {
            if self.evaluate_transition_condition(&transition.condition, event).await? {
                self.execute_transition(instance_id, transition).await?;
                break;
            }
        }
        
        Ok(())
    }

    /// Evaluate transition condition
    async fn evaluate_transition_condition(
        &self,
        condition: &TransitionCondition,
        event: Option<&Event>,
    ) -> Result<bool> {
        match condition {
            TransitionCondition::Always => Ok(true),
            TransitionCondition::OnEvent(event_type) => {
                Ok(event.map_or(false, |e| e.event_type == *event_type))
            }
            TransitionCondition::OnTimeout => {
                // This would be handled by timeout monitoring
                Ok(false)
            }
            TransitionCondition::OnSuccess => {
                Ok(event.map_or(false, |e| e.event_type == "success"))
            }
            TransitionCondition::OnError => {
                Ok(event.map_or(false, |e| e.event_type == "error"))
            }
            TransitionCondition::OnCondition(_expr) => {
                // Would evaluate expression against context
                Ok(false)
            }
            TransitionCondition::Custom(_handler) => {
                // Would call custom condition handler
                Ok(false)
            }
        }
    }

    /// Execute a state transition
    async fn execute_transition(&self, instance_id: &str, transition: &Transition) -> Result<()> {
        let start_time = Instant::now();
        
        debug!(
            "Executing transition {} -> {} for instance {}",
            transition.from_state, transition.to_state, instance_id
        );

        // Execute exit actions for current state
        self.execute_state_exit_actions(instance_id, &transition.from_state).await?;
        
        // Execute transition actions
        for action in &transition.actions {
            self.execute_action(instance_id, action).await?;
        }
        
        // Update instance state
        let mut instances = self.active_instances.lock().await;
        if let Some(instance) = instances.get_mut(instance_id) {
            let transition_record = TransitionRecord {
                from_state: transition.from_state.clone(),
                to_state: transition.to_state.clone(),
                transition_id: transition.id.clone(),
                timestamp: chrono::Utc::now(),
                duration: start_time.elapsed(),
                success: true,
                error_message: None,
            };
            
            instance.context.execution_history.push(transition_record);
            instance.current_state = transition.to_state.clone();
            instance.last_transition = Instant::now();
            instance.transition_count += 1;
            
            // Check if reached terminal state
            let states = self.states.read().await;
            if let Some(state) = states.get(&transition.to_state) {
                if state.state_type == StateType::Terminal || state.state_type == StateType::Error {
                    instance.status = if state.state_type == StateType::Terminal {
                        InstanceStatus::Completed
                    } else {
                        InstanceStatus::Failed
                    };
                }
            }
        }
        drop(instances);
        
        // Execute entry actions for new state
        self.execute_state_entry_actions(instance_id, &transition.to_state).await?;
        
        Ok(())
    }

    /// Execute state entry actions
    async fn execute_state_entry_actions(&self, instance_id: &str, state_id: &str) -> Result<()> {
        let states = self.states.read().await;
        if let Some(state) = states.get(state_id) {
            for action in &state.entry_actions {
                self.execute_action(instance_id, action).await?;
            }
        }
        Ok(())
    }

    /// Execute state exit actions
    async fn execute_state_exit_actions(&self, instance_id: &str, state_id: &str) -> Result<()> {
        let states = self.states.read().await;
        if let Some(state) = states.get(state_id) {
            for action in &state.exit_actions {
                self.execute_action(instance_id, action).await?;
            }
        }
        Ok(())
    }

    /// Execute an action
    async fn execute_action(&self, instance_id: &str, action: &Action) -> Result<()> {
        match action {
            Action::Log(message) => {
                info!("FSM[{}]: {}", instance_id, message);
            }
            Action::SetVariable(key, value) => {
                let mut instances = self.active_instances.lock().await;
                if let Some(instance) = instances.get_mut(instance_id) {
                    instance.context.variables.insert(key.clone(), value.clone());
                }
            }
            Action::CallFunction(func_name, args) => {
                debug!("Calling function {} with args: {:?}", func_name, args);
                // Would call registered function handlers
            }
            Action::SendEvent(event_type, payload) => {
                let event = Event {
                    id: Uuid::new_v4().to_string(),
                    event_type: event_type.clone(),
                    payload: payload.clone(),
                    timestamp: chrono::Utc::now(),
                };
                // Would send event to external systems
                debug!("Sending event: {:?}", event);
            }
            Action::UpdateMetrics(metric_name, value) => {
                debug!("Updating metric {} = {}", metric_name, value);
                // Would update metrics system
            }
            Action::Custom(handler_name, params) => {
                debug!("Executing custom action {} with params: {:?}", handler_name, params);
                // Would call custom action handlers
            }
        }
        Ok(())
    }

    /// Get state machine instance
    pub async fn get_instance(&self, instance_id: &str) -> Result<StateMachineInstance> {
        let instances = self.active_instances.lock().await;
        instances.get(instance_id)
            .cloned()
            .ok_or_else(|| anyhow::anyhow!("Instance not found: {}", instance_id))
    }

    /// Complete state machine instance
    pub async fn complete_instance(&self, instance_id: &str) -> Result<StateMachineResult> {
        let mut instances = self.active_instances.lock().await;
        let instance = instances.remove(instance_id)
            .ok_or_else(|| anyhow::anyhow!("Instance not found: {}", instance_id))?;

        let execution_time = instance.created_at.elapsed();
        
        Ok(StateMachineResult {
            instance_id: instance.id,
            final_state: instance.current_state,
            status: instance.status,
            execution_time,
            transition_count: instance.transition_count,
            context: instance.context,
        })
    }

    /// Get FSM statistics
    pub async fn get_stats(&self) -> FSMStats {
        let instances = self.active_instances.lock().await;
        let states = self.states.read().await;
        let transitions = self.transitions.read().await;
        
        FSMStats {
            active_instances: instances.len(),
            total_states: states.len(),
            total_transitions: transitions.values().map(|v| v.len()).sum(),
            max_states: self.config.max_states,
            max_transitions: self.config.max_transitions,
        }
    }
}

/// FSM statistics
#[derive(Debug, Clone)]
pub struct FSMStats {
    pub active_instances: usize,
    pub total_states: usize,
    pub total_transitions: usize,
    pub max_states: usize,
    pub max_transitions: usize,
}