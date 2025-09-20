# Multi-Agent Platform OPA Policy
# This file defines the security policies for the Multi-Agent platform

package multiagent

import future.keywords.in

# ==============================================================================
# DEFAULT POLICY - DENY ALL
# ==============================================================================

# Default deny - all requests are denied unless explicitly allowed
default allow = false

# ==============================================================================
# AUTHENTICATION POLICIES
# ==============================================================================

# Allow health check endpoints without authentication
allow {
    input.path = ["/health"]
    input.method = "GET"
}

allow {
    input.path = ["/metrics"]
    input.method = "GET"
}

# Require authentication for all other endpoints
allow {
    not is_health_endpoint
    user_authenticated
    user_authorized
}

user_authenticated {
    input.user
    input.user.id
    input.user.tenant_id
}

# ==============================================================================
# AUTHORIZATION POLICIES
# ==============================================================================

# User authorization based on role and resource access
user_authorized {
    # Owner can access everything in their tenant
    input.user.role == "owner"
    same_tenant
}

user_authorized {
    # Admin can access most resources in their tenant
    input.user.role == "admin"
    same_tenant
    not restricted_admin_action
}

user_authorized {
    # Regular users can access their own resources
    input.user.role == "user"
    same_tenant
    (user_owns_resource or public_resource)
    not dangerous_action
}

# ==============================================================================
# TENANT ISOLATION
# ==============================================================================

# Ensure users can only access resources in their own tenant
same_tenant {
    input.user.tenant_id == input.resource.tenant_id
}

same_tenant {
    # For session-based requests
    input.user.tenant_id == input.session.tenant_id
}

same_tenant {
    # For task-based requests
    input.user.tenant_id == input.task.tenant_id
}

# ==============================================================================
# RESOURCE OWNERSHIP
# ==============================================================================

# Check if user owns the resource
user_owns_resource {
    input.user.id == input.resource.user_id
}

user_owns_resource {
    # For session resources
    input.user.id == input.session.user_id
}

user_owns_resource {
    # For task resources
    input.user.id == input.task.user_id
}

# Public resources that all authenticated users can access
public_resource {
    input.resource.visibility == "public"
}

public_resource {
    # Model information is generally public within tenant
    input.path = ["/api", "v1", "models"]
    input.method == "GET"
}

public_resource {
    # Tool discovery is public within tenant
    input.path = ["/api", "v1", "tools", "discover"]
    input.method == "POST"
}

# ==============================================================================
# DANGEROUS ACTIONS
# ==============================================================================

# Actions that require admin privileges or special approval
dangerous_action {
    # Code execution requires approval
    input.path = ["/api", "v1", "agents", "execute"]
    input.method == "POST"
    input.request.execution_mode == "privileged"
}

dangerous_action {
    # File system access
    input.request.tools[_] == "file_system"
}

dangerous_action {
    # Network access tools
    input.request.tools[_] == "web_browser"
}

dangerous_action {
    # System commands
    input.request.tools[_] == "shell_command"
}

dangerous_action {
    # Database operations
    input.request.tools[_] == "database_query"
}

# Actions restricted to owners only
restricted_admin_action {
    # Tenant settings modification
    input.path = ["/api", "v1", "tenants", _]
    input.method in ["PUT", "PATCH", "DELETE"]
}

restricted_admin_action {
    # User management (except self)
    input.path = ["/api", "v1", "users", user_id]
    input.method in ["PUT", "PATCH", "DELETE"]
    user_id != input.user.id
}

restricted_admin_action {
    # Budget management
    input.path = ["/api", "v1", "budget", _]
    input.method in ["PUT", "PATCH"]
}

# ==============================================================================
# TOOL EXECUTION POLICIES
# ==============================================================================

# Tool execution authorization
allow_tool_execution {
    input.tool.safety_level == "safe"
    same_tenant
}

allow_tool_execution {
    input.tool.safety_level == "caution"
    same_tenant
    (input.user.role in ["admin", "owner"] or user_has_tool_permission)
}

allow_tool_execution {
    input.tool.safety_level == "dangerous"
    same_tenant
    input.user.role in ["admin", "owner"]
    approval_required
}

# Tool permission checks
user_has_tool_permission {
    input.tool.name in input.user.permissions.tools
}

user_has_tool_permission {
    input.tool.category in input.user.permissions.tool_categories
}

# ==============================================================================
# APPROVAL REQUIREMENTS
# ==============================================================================

approval_required {
    input.request.requires_approval == true
}

approval_required {
    dangerous_action
    not input.approval
}

approval_required {
    # High-cost operations require approval
    input.request.estimated_cost_usd > 1.0
}

approval_required {
    # Large token usage requires approval
    input.request.estimated_tokens > 50000
}

# ==============================================================================
# BUDGET ENFORCEMENT
# ==============================================================================

# Check budget limits
within_budget {
    input.user.budget.remaining_tokens > input.request.estimated_tokens
}

within_budget {
    input.user.budget.remaining_cost_usd > input.request.estimated_cost_usd
}

# Allow if within budget or override is present
allow_budget {
    within_budget
}

allow_budget {
    input.user.role == "owner"
    input.request.override_budget == true
}

# ==============================================================================
# HELPER FUNCTIONS
# ==============================================================================

is_health_endpoint {
    input.path = ["/health"]
}

is_health_endpoint {
    input.path = ["/metrics"]
}

is_health_endpoint {
    input.path = ["/api", "v1", "health"]
}

approval_provided {
    input.approval
    input.approval.approved == true
    input.approval.approver_id
}

# Allow access if user has required permission
allow if {
    input.user_id
    input.tenant_id
    input.resource
    input.action
    
    # Check if user has direct permission
    user_has_permission(input.user_id, input.tenant_id, input.resource, input.action)
}

# Allow access if user has required role
allow if {
    input.user_id
    input.tenant_id
    input.resource
    input.action
    
    # Check if user has role that grants permission
    user_has_role_permission(input.user_id, input.tenant_id, input.resource, input.action)
}

# Tenant access control
tenant_access if {
    input.user_id
    input.tenant_id
    input.action == "access"
    
    # Check if user is member of tenant
    user_is_tenant_member(input.user_id, input.tenant_id)
}

# Admin access - admins can access everything within their tenant
allow if {
    input.user_id
    input.tenant_id
    
    user_has_role(input.user_id, input.tenant_id, "admin")
}

# System admin access - system admins can access everything
allow if {
    input.user_id
    
    user_has_system_role(input.user_id, "system_admin")
}

# Workflow access control
allow if {
    input.user_id
    input.tenant_id
    input.resource == "workflow"
    input.action in ["read", "execute"]
    
    # Users can read/execute workflows they own or have been shared with
    workflow_accessible_to_user(input.workflow_id, input.user_id, input.tenant_id)
}

# Agent access control
allow if {
    input.user_id
    input.tenant_id
    input.resource == "agent"
    input.action in ["read", "execute"]
    
    # Users can read/execute agents they own or have been shared with
    agent_accessible_to_user(input.agent_id, input.user_id, input.tenant_id)
}

# Budget access control
allow if {
    input.user_id
    input.tenant_id
    input.resource == "budget"
    input.action == "read"
    
    # Users can read their own budget information
    input.target_user_id == input.user_id
}

allow if {
    input.user_id
    input.tenant_id
    input.resource == "budget"
    input.action == "manage"
    
    # Only admins and budget managers can manage budgets
    user_has_role(input.user_id, input.tenant_id, "admin")
}

allow if {
    input.user_id
    input.tenant_id
    input.resource == "budget"
    input.action == "manage"
    
    user_has_role(input.user_id, input.tenant_id, "budget_manager")
}

# Session management
allow if {
    input.user_id
    input.tenant_id
    input.resource == "session"
    input.action == "revoke"
    
    # Users can revoke their own sessions
    input.target_user_id == input.user_id
}

allow if {
    input.user_id
    input.tenant_id
    input.resource == "session"
    input.action == "revoke"
    
    # Admins can revoke any session in their tenant
    user_has_role(input.user_id, input.tenant_id, "admin")
}

# Audit log access
allow if {
    input.user_id
    input.tenant_id
    input.resource == "audit"
    input.action == "read"
    
    # Security officers and admins can read audit logs
    user_has_role(input.user_id, input.tenant_id, "security_officer")
}

# Rate limiting policies
rate_limit_exceeded if {
    input.user_id
    input.tenant_id
    
    # Check if user has exceeded rate limits
    user_rate_limit_exceeded(input.user_id, input.tenant_id)
}

# Resource quota policies
quota_exceeded if {
    input.user_id
    input.tenant_id
    input.resource
    
    # Check if tenant has exceeded resource quotas
    tenant_quota_exceeded(input.tenant_id, input.resource)
}

# Time-based access control
time_restricted if {
    input.user_id
    input.tenant_id
    
    # Check if access is restricted based on time
    current_time := time.now_ns()
    user_has_time_restrictions(input.user_id, input.tenant_id, current_time)
}

# IP-based access control
ip_restricted if {
    input.user_id
    input.tenant_id
    input.client_ip
    
    # Check if access is restricted based on IP
    user_has_ip_restrictions(input.user_id, input.tenant_id, input.client_ip)
}

# Helper functions (these would be implemented with external data)

user_has_permission(user_id, tenant_id, resource, action) if {
    # This would query the database or external service
    # For now, return false to force role-based checks
    false
}

user_has_role_permission(user_id, tenant_id, resource, action) if {
    # Check if any of user's roles grant the required permission
    user_roles := data.user_roles[user_id][tenant_id]
    role := user_roles[_]
    role_permissions := data.role_permissions[role]
    permission := role_permissions[_]
    permission.resource == resource
    permission.action == action
}

user_has_role(user_id, tenant_id, role_name) if {
    user_roles := data.user_roles[user_id][tenant_id]
    user_roles[_] == role_name
}

user_has_system_role(user_id, role_name) if {
    system_roles := data.system_roles[user_id]
    system_roles[_] == role_name
}

user_is_tenant_member(user_id, tenant_id) if {
    tenant_members := data.tenant_members[tenant_id]
    tenant_members[_] == user_id
}

workflow_accessible_to_user(workflow_id, user_id, tenant_id) if {
    # Check if user owns the workflow
    workflow_owner := data.workflow_owners[workflow_id]
    workflow_owner == user_id
}

workflow_accessible_to_user(workflow_id, user_id, tenant_id) if {
    # Check if workflow is shared with user
    shared_workflows := data.shared_workflows[user_id][tenant_id]
    shared_workflows[_] == workflow_id
}

agent_accessible_to_user(agent_id, user_id, tenant_id) if {
    # Check if user owns the agent
    agent_owner := data.agent_owners[agent_id]
    agent_owner == user_id
}

agent_accessible_to_user(agent_id, user_id, tenant_id) if {
    # Check if agent is shared with user
    shared_agents := data.shared_agents[user_id][tenant_id]
    shared_agents[_] == agent_id
}

user_rate_limit_exceeded(user_id, tenant_id) if {
    current_requests := data.rate_limits[user_id][tenant_id].current
    max_requests := data.rate_limits[user_id][tenant_id].max
    current_requests >= max_requests
}

tenant_quota_exceeded(tenant_id, resource) if {
    current_usage := data.tenant_quotas[tenant_id][resource].current
    max_quota := data.tenant_quotas[tenant_id][resource].max
    current_usage >= max_quota
}

user_has_time_restrictions(user_id, tenant_id, current_time) if {
    restrictions := data.time_restrictions[user_id][tenant_id]
    restriction := restrictions[_]
    current_time < restriction.start_time
}

user_has_time_restrictions(user_id, tenant_id, current_time) if {
    restrictions := data.time_restrictions[user_id][tenant_id]
    restriction := restrictions[_]
    current_time > restriction.end_time
}

user_has_ip_restrictions(user_id, tenant_id, client_ip) if {
    allowed_ips := data.ip_restrictions[user_id][tenant_id]
    count(allowed_ips) > 0
    not ip_in_list(client_ip, allowed_ips)
}

ip_in_list(ip, ip_list) if {
    ip_list[_] == ip
}

# Policy violation reasons
reason := "insufficient_permissions" if {
    not allow
    input.resource
    input.action
}

reason := "rate_limit_exceeded" if {
    rate_limit_exceeded
}

reason := "quota_exceeded" if {
    quota_exceeded
}

reason := "time_restricted" if {
    time_restricted
}

reason := "ip_restricted" if {
    ip_restricted
}

reason := "tenant_access_denied" if {
    not tenant_access
    input.action == "access"
}