package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/multi-agent/go/security-service/internal/auth"
	"github.com/multi-agent/go/security-service/internal/rbac"
	"github.com/multi-agent/go/security-service/internal/tenant"
)

// Security-related database operations

// User operations
func (c *Client) GetUserByUsername(ctx context.Context, username string) (*auth.User, error) {
	query := `
		SELECT id, username, email, password_hash, first_name, last_name, 
		       status, last_login_at, created_at, updated_at
		FROM auth.users 
		WHERE username = $1 AND status != 'deleted'`

	var user auth.User
	var lastLoginAt sql.NullTime
	
	err := c.db.QueryRowContext(ctx, query, username).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.FirstName, &user.LastName, &user.Status, &lastLoginAt,
		&user.CreatedAt, &user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if lastLoginAt.Valid {
		user.LastLoginAt = &lastLoginAt.Time
	}

	return &user, nil
}

func (c *Client) GetUser(ctx context.Context, userID string) (*auth.User, error) {
	query := `
		SELECT id, username, email, password_hash, first_name, last_name, 
		       status, last_login_at, created_at, updated_at
		FROM auth.users 
		WHERE id = $1 AND status != 'deleted'`

	var user auth.User
	var lastLoginAt sql.NullTime
	
	err := c.db.QueryRowContext(ctx, query, userID).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.FirstName, &user.LastName, &user.Status, &lastLoginAt,
		&user.CreatedAt, &user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if lastLoginAt.Valid {
		user.LastLoginAt = &lastLoginAt.Time
	}

	return &user, nil
}

func (c *Client) UpdateUserLastLogin(ctx context.Context, userID string, loginTime time.Time) error {
	query := `UPDATE auth.users SET last_login_at = $1, updated_at = NOW() WHERE id = $2`
	_, err := c.db.ExecContext(ctx, query, loginTime, userID)
	return err
}

// Session operations
func (c *Client) CreateSession(ctx context.Context, session *auth.Session) error {
	query := `
		INSERT INTO auth.sessions (id, user_id, tenant_id, access_token, refresh_token, 
		                          expires_at, created_at, last_used_at, ip_address, 
		                          user_agent, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	_, err := c.db.ExecContext(ctx, query,
		session.ID, session.UserID, session.TenantID, session.AccessToken,
		session.RefreshToken, session.ExpiresAt, session.CreatedAt,
		session.LastUsedAt, session.IPAddress, session.UserAgent, session.Status,
	)

	return err
}

func (c *Client) GetSessionByRefreshToken(ctx context.Context, refreshToken string) (*auth.Session, error) {
	query := `
		SELECT id, user_id, tenant_id, access_token, refresh_token, expires_at, 
		       created_at, last_used_at, ip_address, user_agent, status
		FROM auth.sessions 
		WHERE refresh_token = $1 AND status = 'active'`

	var session auth.Session
	err := c.db.QueryRowContext(ctx, query, refreshToken).Scan(
		&session.ID, &session.UserID, &session.TenantID, &session.AccessToken,
		&session.RefreshToken, &session.ExpiresAt, &session.CreatedAt,
		&session.LastUsedAt, &session.IPAddress, &session.UserAgent, &session.Status,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("session not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return &session, nil
}

func (c *Client) GetSession(ctx context.Context, sessionID string) (*auth.Session, error) {
	query := `
		SELECT id, user_id, tenant_id, access_token, refresh_token, expires_at, 
		       created_at, last_used_at, ip_address, user_agent, status
		FROM auth.sessions 
		WHERE id = $1`

	var session auth.Session
	err := c.db.QueryRowContext(ctx, query, sessionID).Scan(
		&session.ID, &session.UserID, &session.TenantID, &session.AccessToken,
		&session.RefreshToken, &session.ExpiresAt, &session.CreatedAt,
		&session.LastUsedAt, &session.IPAddress, &session.UserAgent, &session.Status,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("session not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return &session, nil
}

func (c *Client) UpdateSession(ctx context.Context, session *auth.Session) error {
	query := `
		UPDATE auth.sessions 
		SET access_token = $1, last_used_at = $2 
		WHERE id = $3`

	_, err := c.db.ExecContext(ctx, query, session.AccessToken, session.LastUsedAt, session.ID)
	return err
}

func (c *Client) InvalidateSession(ctx context.Context, sessionID string) error {
	query := `UPDATE auth.sessions SET status = 'revoked' WHERE id = $1`
	_, err := c.db.ExecContext(ctx, query, sessionID)
	return err
}

func (c *Client) GetActiveSessions(ctx context.Context, userID string) ([]auth.Session, error) {
	query := `
		SELECT id, user_id, tenant_id, access_token, refresh_token, expires_at, 
		       created_at, last_used_at, ip_address, user_agent, status
		FROM auth.sessions 
		WHERE user_id = $1 AND status = 'active' AND expires_at > NOW()
		ORDER BY last_used_at DESC`

	rows, err := c.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query sessions: %w", err)
	}
	defer rows.Close()

	var sessions []auth.Session
	for rows.Next() {
		var session auth.Session
		err := rows.Scan(
			&session.ID, &session.UserID, &session.TenantID, &session.AccessToken,
			&session.RefreshToken, &session.ExpiresAt, &session.CreatedAt,
			&session.LastUsedAt, &session.IPAddress, &session.UserAgent, &session.Status,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}
		sessions = append(sessions, session)
	}

	return sessions, nil
}

// Tenant operations
func (c *Client) CreateTenant(ctx context.Context, tenant *tenant.Tenant) error {
	settingsJSON, _ := json.Marshal(tenant.Settings)
	limitsJSON, _ := json.Marshal(tenant.Limits)

	query := `
		INSERT INTO auth.tenants (id, name, domain, status, settings, limits, 
		                         created_at, updated_at, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err := c.db.ExecContext(ctx, query,
		tenant.ID, tenant.Name, tenant.Domain, tenant.Status,
		settingsJSON, limitsJSON, tenant.CreatedAt, tenant.UpdatedAt, tenant.CreatedBy,
	)

	return err
}

func (c *Client) GetTenant(ctx context.Context, tenantID string) (*tenant.Tenant, error) {
	query := `
		SELECT id, name, domain, status, settings, limits, created_at, updated_at, created_by
		FROM auth.tenants 
		WHERE id = $1 AND status != 'deleted'`

	var t tenant.Tenant
	var settingsJSON, limitsJSON []byte

	err := c.db.QueryRowContext(ctx, query, tenantID).Scan(
		&t.ID, &t.Name, &t.Domain, &t.Status, &settingsJSON, &limitsJSON,
		&t.CreatedAt, &t.UpdatedAt, &t.CreatedBy,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("tenant not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	json.Unmarshal(settingsJSON, &t.Settings)
	json.Unmarshal(limitsJSON, &t.Limits)

	return &t, nil
}

func (c *Client) ListTenants(ctx context.Context, page, limit int, status string) ([]tenant.Tenant, int, error) {
	offset := (page - 1) * limit
	
	whereClause := "WHERE status != 'deleted'"
	args := []interface{}{limit, offset}
	argIndex := 3

	if status != "" {
		whereClause += " AND status = $" + fmt.Sprintf("%d", argIndex)
		args = append(args, status)
		argIndex++
	}

	// Get total count
	countQuery := "SELECT COUNT(*) FROM auth.tenants " + whereClause
	var total int
	err := c.db.QueryRowContext(ctx, countQuery, args[2:]...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count tenants: %w", err)
	}

	// Get tenants
	query := `
		SELECT id, name, domain, status, settings, limits, created_at, updated_at, created_by
		FROM auth.tenants ` + whereClause + `
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`

	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query tenants: %w", err)
	}
	defer rows.Close()

	var tenants []tenant.Tenant
	for rows.Next() {
		var t tenant.Tenant
		var settingsJSON, limitsJSON []byte

		err := rows.Scan(
			&t.ID, &t.Name, &t.Domain, &t.Status, &settingsJSON, &limitsJSON,
			&t.CreatedAt, &t.UpdatedAt, &t.CreatedBy,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan tenant: %w", err)
		}

		json.Unmarshal(settingsJSON, &t.Settings)
		json.Unmarshal(limitsJSON, &t.Limits)
		tenants = append(tenants, t)
	}

	return tenants, total, nil
}

func (c *Client) UpdateTenant(ctx context.Context, tenant *tenant.Tenant) error {
	settingsJSON, _ := json.Marshal(tenant.Settings)
	limitsJSON, _ := json.Marshal(tenant.Limits)

	query := `
		UPDATE auth.tenants 
		SET name = $1, domain = $2, status = $3, settings = $4, limits = $5, updated_at = $6
		WHERE id = $7`

	_, err := c.db.ExecContext(ctx, query,
		tenant.Name, tenant.Domain, tenant.Status, settingsJSON, limitsJSON,
		tenant.UpdatedAt, tenant.ID,
	)

	return err
}

func (c *Client) DeleteTenant(ctx context.Context, tenantID string) error {
	query := `UPDATE auth.tenants SET status = 'deleted', updated_at = NOW() WHERE id = $1`
	result, err := c.db.ExecContext(ctx, query, tenantID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("tenant not found")
	}

	return nil
}

func (c *Client) AddUserToTenant(ctx context.Context, tenantID, userID, role string) error {
	query := `
		INSERT INTO auth.tenant_users (tenant_id, user_id, role, created_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (tenant_id, user_id) DO UPDATE SET role = $3, updated_at = NOW()`

	_, err := c.db.ExecContext(ctx, query, tenantID, userID, role)
	return err
}

func (c *Client) RemoveUserFromTenant(ctx context.Context, tenantID, userID string) error {
	query := `DELETE FROM auth.tenant_users WHERE tenant_id = $1 AND user_id = $2`
	_, err := c.db.ExecContext(ctx, query, tenantID, userID)
	return err
}

// RBAC operations
func (c *Client) CreateRole(ctx context.Context, role *rbac.Role) error {
	query := `
		INSERT INTO auth.roles (id, name, description, tenant_id, created_at, updated_at, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := c.db.ExecContext(ctx, query,
		role.ID, role.Name, role.Description, role.TenantID,
		role.CreatedAt, role.UpdatedAt, role.CreatedBy,
	)

	return err
}

func (c *Client) GetRole(ctx context.Context, roleID string) (*rbac.Role, error) {
	query := `
		SELECT id, name, description, tenant_id, created_at, updated_at, created_by
		FROM auth.roles 
		WHERE id = $1`

	var role rbac.Role
	err := c.db.QueryRowContext(ctx, query, roleID).Scan(
		&role.ID, &role.Name, &role.Description, &role.TenantID,
		&role.CreatedAt, &role.UpdatedAt, &role.CreatedBy,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("role not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get role: %w", err)
	}

	return &role, nil
}

func (c *Client) ListRoles(ctx context.Context, page, limit int, tenantID string) ([]rbac.Role, int, error) {
	offset := (page - 1) * limit
	
	whereClause := "WHERE 1=1"
	args := []interface{}{limit, offset}
	argIndex := 3

	if tenantID != "" {
		whereClause += " AND tenant_id = $" + fmt.Sprintf("%d", argIndex)
		args = append(args, tenantID)
		argIndex++
	}

	// Get total count
	countQuery := "SELECT COUNT(*) FROM auth.roles " + whereClause
	var total int
	err := c.db.QueryRowContext(ctx, countQuery, args[2:]...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count roles: %w", err)
	}

	// Get roles
	query := `
		SELECT id, name, description, tenant_id, created_at, updated_at, created_by
		FROM auth.roles ` + whereClause + `
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`

	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query roles: %w", err)
	}
	defer rows.Close()

	var roles []rbac.Role
	for rows.Next() {
		var role rbac.Role
		err := rows.Scan(
			&role.ID, &role.Name, &role.Description, &role.TenantID,
			&role.CreatedAt, &role.UpdatedAt, &role.CreatedBy,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan role: %w", err)
		}
		roles = append(roles, role)
	}

	return roles, total, nil
}

func (c *Client) UpdateRole(ctx context.Context, role *rbac.Role) error {
	query := `
		UPDATE auth.roles 
		SET name = $1, description = $2, updated_at = $3
		WHERE id = $4`

	_, err := c.db.ExecContext(ctx, query, role.Name, role.Description, role.UpdatedAt, role.ID)
	return err
}

func (c *Client) DeleteRole(ctx context.Context, roleID string) error {
	query := `DELETE FROM auth.roles WHERE id = $1`
	result, err := c.db.ExecContext(ctx, query, roleID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("role not found")
	}

	return nil
}

func (c *Client) CreatePermission(ctx context.Context, permission *rbac.Permission) error {
	query := `
		INSERT INTO auth.permissions (id, name, resource, action, description, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := c.db.ExecContext(ctx, query,
		permission.ID, permission.Name, permission.Resource, permission.Action,
		permission.Description, permission.CreatedAt,
	)

	return err
}

func (c *Client) ListPermissions(ctx context.Context, page, limit int, resource string) ([]rbac.Permission, int, error) {
	offset := (page - 1) * limit
	
	whereClause := "WHERE 1=1"
	args := []interface{}{limit, offset}
	argIndex := 3

	if resource != "" {
		whereClause += " AND resource = $" + fmt.Sprintf("%d", argIndex)
		args = append(args, resource)
		argIndex++
	}

	// Get total count
	countQuery := "SELECT COUNT(*) FROM auth.permissions " + whereClause
	var total int
	err := c.db.QueryRowContext(ctx, countQuery, args[2:]...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count permissions: %w", err)
	}

	// Get permissions
	query := `
		SELECT id, name, resource, action, description, created_at
		FROM auth.permissions ` + whereClause + `
		ORDER BY resource, action
		LIMIT $1 OFFSET $2`

	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query permissions: %w", err)
	}
	defer rows.Close()

	var permissions []rbac.Permission
	for rows.Next() {
		var perm rbac.Permission
		err := rows.Scan(
			&perm.ID, &perm.Name, &perm.Resource, &perm.Action,
			&perm.Description, &perm.CreatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan permission: %w", err)
		}
		permissions = append(permissions, perm)
	}

	return permissions, total, nil
}

func (c *Client) AssignPermissionsToRole(ctx context.Context, roleID string, permissionNames []string) error {
	// First, remove existing permissions
	_, err := c.db.ExecContext(ctx, "DELETE FROM auth.role_permissions WHERE role_id = $1", roleID)
	if err != nil {
		return fmt.Errorf("failed to remove existing permissions: %w", err)
	}

	// Then add new permissions
	for _, permName := range permissionNames {
		query := `
			INSERT INTO auth.role_permissions (role_id, permission_id)
			SELECT $1, id FROM auth.permissions WHERE name = $2`
		
		_, err := c.db.ExecContext(ctx, query, roleID, permName)
		if err != nil {
			return fmt.Errorf("failed to assign permission %s: %w", permName, err)
		}
	}

	return nil
}

func (c *Client) UpdateRolePermissions(ctx context.Context, roleID string, permissionNames []string) error {
	return c.AssignPermissionsToRole(ctx, roleID, permissionNames)
}

func (c *Client) AssignRoleToUser(ctx context.Context, userRole *rbac.UserRole) error {
	query := `
		INSERT INTO auth.user_roles (id, user_id, role_id, tenant_id, assigned_at, assigned_by, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := c.db.ExecContext(ctx, query,
		userRole.ID, userRole.UserID, userRole.RoleID, userRole.TenantID,
		userRole.AssignedAt, userRole.AssignedBy, userRole.ExpiresAt,
	)

	return err
}

func (c *Client) RevokeRoleFromUser(ctx context.Context, userID, roleID string) error {
	query := `DELETE FROM auth.user_roles WHERE user_id = $1 AND role_id = $2`
	_, err := c.db.ExecContext(ctx, query, userID, roleID)
	return err
}

func (c *Client) GetUserPermissions(ctx context.Context, userID, tenantID string) ([]rbac.Permission, error) {
	query := `
		SELECT DISTINCT p.id, p.name, p.resource, p.action, p.description, p.created_at
		FROM auth.permissions p
		JOIN auth.role_permissions rp ON p.id = rp.permission_id
		JOIN auth.user_roles ur ON rp.role_id = ur.role_id
		WHERE ur.user_id = $1 AND ur.tenant_id = $2 
		      AND (ur.expires_at IS NULL OR ur.expires_at > NOW())`

	rows, err := c.db.QueryContext(ctx, query, userID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to query user permissions: %w", err)
	}
	defer rows.Close()

	var permissions []rbac.Permission
	for rows.Next() {
		var perm rbac.Permission
		err := rows.Scan(
			&perm.ID, &perm.Name, &perm.Resource, &perm.Action,
			&perm.Description, &perm.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan permission: %w", err)
		}
		permissions = append(permissions, perm)
	}

	return permissions, nil
}

func (c *Client) CheckUserPermission(ctx context.Context, userID, tenantID, resource, action string) (bool, error) {
	query := `
		SELECT COUNT(*)
		FROM auth.permissions p
		JOIN auth.role_permissions rp ON p.id = rp.permission_id
		JOIN auth.user_roles ur ON rp.role_id = ur.role_id
		WHERE ur.user_id = $1 AND ur.tenant_id = $2 
		      AND p.resource = $3 AND p.action = $4
		      AND (ur.expires_at IS NULL OR ur.expires_at > NOW())`

	var count int
	err := c.db.QueryRowContext(ctx, query, userID, tenantID, resource, action).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check user permission: %w", err)
	}

	return count > 0, nil
}

func (c *Client) CheckUserRole(ctx context.Context, userID, tenantID, roleName string) (bool, error) {
	query := `
		SELECT COUNT(*)
		FROM auth.user_roles ur
		JOIN auth.roles r ON ur.role_id = r.id
		WHERE ur.user_id = $1 AND ur.tenant_id = $2 AND r.name = $3
		      AND (ur.expires_at IS NULL OR ur.expires_at > NOW())`

	var count int
	err := c.db.QueryRowContext(ctx, query, userID, tenantID, roleName).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check user role: %w", err)
	}

	return count > 0, nil
}

func (c *Client) GetUserRoles(ctx context.Context, userID, tenantID string) ([]rbac.Role, error) {
	query := `
		SELECT r.id, r.name, r.description, r.tenant_id, r.created_at, r.updated_at, r.created_by
		FROM auth.roles r
		JOIN auth.user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = $1 AND ur.tenant_id = $2 
		      AND (ur.expires_at IS NULL OR ur.expires_at > NOW())`

	rows, err := c.db.QueryContext(ctx, query, userID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to query user roles: %w", err)
	}
	defer rows.Close()

	var roles []rbac.Role
	for rows.Next() {
		var role rbac.Role
		err := rows.Scan(
			&role.ID, &role.Name, &role.Description, &role.TenantID,
			&role.CreatedAt, &role.UpdatedAt, &role.CreatedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan role: %w", err)
		}
		roles = append(roles, role)
	}

	return roles, nil
}

// Audit operations
func (c *Client) CreateAuditEvent(ctx context.Context, event *auth.AuditEvent) error {
	detailsJSON, _ := json.Marshal(event.Details)

	query := `
		INSERT INTO auth.audit_events (id, user_id, tenant_id, session_id, event_type, 
		                              resource, action, result, details, ip_address, 
		                              user_agent, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

	_, err := c.db.ExecContext(ctx, query,
		event.ID, event.UserID, event.TenantID, event.SessionID, event.EventType,
		event.Resource, event.Action, event.Result, detailsJSON, event.IPAddress,
		event.UserAgent, event.CreatedAt,
	)

	return err
}

func (c *Client) GetAuditEvents(ctx context.Context, tenantID string, page, limit int, eventType, userID string) ([]auth.AuditEvent, int, error) {
	offset := (page - 1) * limit
	
	whereClause := "WHERE tenant_id = $3"
	args := []interface{}{limit, offset, tenantID}
	argIndex := 4

	if eventType != "" {
		whereClause += " AND event_type = $" + fmt.Sprintf("%d", argIndex)
		args = append(args, eventType)
		argIndex++
	}

	if userID != "" {
		whereClause += " AND user_id = $" + fmt.Sprintf("%d", argIndex)
		args = append(args, userID)
		argIndex++
	}

	// Get total count
	countQuery := "SELECT COUNT(*) FROM auth.audit_events " + whereClause
	var total int
	err := c.db.QueryRowContext(ctx, countQuery, args[2:]...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count audit events: %w", err)
	}

	// Get events
	query := `
		SELECT id, user_id, tenant_id, session_id, event_type, resource, action, 
		       result, details, ip_address, user_agent, created_at
		FROM auth.audit_events ` + whereClause + `
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`

	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query audit events: %w", err)
	}
	defer rows.Close()

	var events []auth.AuditEvent
	for rows.Next() {
		var event auth.AuditEvent
		var detailsJSON []byte

		err := rows.Scan(
			&event.ID, &event.UserID, &event.TenantID, &event.SessionID,
			&event.EventType, &event.Resource, &event.Action, &event.Result,
			&detailsJSON, &event.IPAddress, &event.UserAgent, &event.CreatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan audit event: %w", err)
		}

		json.Unmarshal(detailsJSON, &event.Details)
		events = append(events, event)
	}

	return events, total, nil
}