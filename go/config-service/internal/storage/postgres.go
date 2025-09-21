package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/lib/pq"

	"github.com/multi-agent/go/config-service/internal/config"
	"github.com/multi-agent/go/config-service/internal/flags"
)

// PostgresStorage implements Storage interface with PostgreSQL
type PostgresStorage struct {
	db *sql.DB
}

// NewStorage creates a new storage instance
func NewStorage(cfg config.DatabaseConfig, logger interface{}) (*PostgresStorage, error) {
	dsn := fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.Database, cfg.Username, cfg.Password, cfg.SSLMode)

	db, err := sql.Open(cfg.Driver, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	storage := &PostgresStorage{db: db}

	// Initialize schema
	if err := storage.initSchema(); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return storage, nil
}

// Close closes the database connection
func (s *PostgresStorage) Close() error {
	return s.db.Close()
}

// CreateFlag creates a new feature flag
func (s *PostgresStorage) CreateFlag(ctx context.Context, flag *flags.FeatureFlag) error {
	rulesJSON, _ := json.Marshal(flag.Rules)
	rolloutJSON, _ := json.Marshal(flag.Rollout)
	metadataJSON, _ := json.Marshal(flag.Metadata)

	query := `
		INSERT INTO feature_flags (id, name, description, enabled, rules, rollout, environment, tenant_id, created_by, created_at, updated_at, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	_, err := s.db.ExecContext(ctx, query,
		flag.ID, flag.Name, flag.Description, flag.Enabled,
		rulesJSON, rolloutJSON, flag.Environment, flag.TenantID,
		flag.CreatedBy, flag.CreatedAt, flag.UpdatedAt, metadataJSON)

	return err
}

// GetFlag retrieves a feature flag by ID
func (s *PostgresStorage) GetFlag(ctx context.Context, id string) (*flags.FeatureFlag, error) {
	query := `
		SELECT id, name, description, enabled, rules, rollout, environment, tenant_id, created_by, created_at, updated_at, metadata
		FROM feature_flags WHERE id = $1
	`

	row := s.db.QueryRowContext(ctx, query, id)
	return s.scanFlag(row)
}

// GetFlagByName retrieves a feature flag by name, environment, and tenant
func (s *PostgresStorage) GetFlagByName(ctx context.Context, name, environment, tenantID string) (*flags.FeatureFlag, error) {
	query := `
		SELECT id, name, description, enabled, rules, rollout, environment, tenant_id, created_by, created_at, updated_at, metadata
		FROM feature_flags WHERE name = $1 AND environment = $2 AND tenant_id = $3
	`

	row := s.db.QueryRowContext(ctx, query, name, environment, tenantID)
	return s.scanFlag(row)
}

// UpdateFlag updates a feature flag
func (s *PostgresStorage) UpdateFlag(ctx context.Context, flag *flags.FeatureFlag) error {
	rulesJSON, _ := json.Marshal(flag.Rules)
	rolloutJSON, _ := json.Marshal(flag.Rollout)
	metadataJSON, _ := json.Marshal(flag.Metadata)

	query := `
		UPDATE feature_flags 
		SET name = $2, description = $3, enabled = $4, rules = $5, rollout = $6, 
		    environment = $7, tenant_id = $8, updated_at = $9, metadata = $10
		WHERE id = $1
	`

	_, err := s.db.ExecContext(ctx, query,
		flag.ID, flag.Name, flag.Description, flag.Enabled,
		rulesJSON, rolloutJSON, flag.Environment, flag.TenantID,
		flag.UpdatedAt, metadataJSON)

	return err
}

// DeleteFlag deletes a feature flag
func (s *PostgresStorage) DeleteFlag(ctx context.Context, id string) error {
	query := `DELETE FROM feature_flags WHERE id = $1`
	_, err := s.db.ExecContext(ctx, query, id)
	return err
}

// ListFlags lists feature flags for environment and tenant
func (s *PostgresStorage) ListFlags(ctx context.Context, environment, tenantID string) ([]*flags.FeatureFlag, error) {
	query := `
		SELECT id, name, description, enabled, rules, rollout, environment, tenant_id, created_by, created_at, updated_at, metadata
		FROM feature_flags WHERE environment = $1 AND tenant_id = $2
		ORDER BY created_at DESC
	`

	rows, err := s.db.QueryContext(ctx, query, environment, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var flags []*flags.FeatureFlag
	for rows.Next() {
		flag, err := s.scanFlag(rows)
		if err != nil {
			return nil, err
		}
		flags = append(flags, flag)
	}

	return flags, rows.Err()
}

// LogEvaluation logs flag evaluation
func (s *PostgresStorage) LogEvaluation(ctx context.Context, result *flags.EvaluationResult, evalCtx *flags.EvaluationContext) error {
	resultJSON, _ := json.Marshal(result)
	contextJSON, _ := json.Marshal(evalCtx)

	query := `
		INSERT INTO flag_evaluations (flag_id, user_id, tenant_id, result, context, evaluated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := s.db.ExecContext(ctx, query,
		result.FlagID, evalCtx.UserID, evalCtx.TenantID,
		resultJSON, contextJSON, result.EvaluatedAt)

	return err
}

// LogFlagChange logs flag changes for audit
func (s *PostgresStorage) LogFlagChange(ctx context.Context, flagID, action, changedBy string, before, after interface{}) error {
	beforeJSON, _ := json.Marshal(before)
	afterJSON, _ := json.Marshal(after)

	query := `
		INSERT INTO flag_audit_log (flag_id, action, changed_by, before_value, after_value, changed_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := s.db.ExecContext(ctx, query,
		flagID, action, changedBy, beforeJSON, afterJSON, time.Now())

	return err
}

// scanFlag scans a database row into a FeatureFlag
func (s *PostgresStorage) scanFlag(scanner interface{}) (*flags.FeatureFlag, error) {
	var flag flags.FeatureFlag
	var rulesJSON, rolloutJSON, metadataJSON []byte

	var err error
	switch s := scanner.(type) {
	case *sql.Row:
		err = s.Scan(
			&flag.ID, &flag.Name, &flag.Description, &flag.Enabled,
			&rulesJSON, &rolloutJSON, &flag.Environment, &flag.TenantID,
			&flag.CreatedBy, &flag.CreatedAt, &flag.UpdatedAt, &metadataJSON)
	case *sql.Rows:
		err = s.Scan(
			&flag.ID, &flag.Name, &flag.Description, &flag.Enabled,
			&rulesJSON, &rolloutJSON, &flag.Environment, &flag.TenantID,
			&flag.CreatedBy, &flag.CreatedAt, &flag.UpdatedAt, &metadataJSON)
	default:
		return nil, fmt.Errorf("unsupported scanner type")
	}

	if err != nil {
		return nil, err
	}

	// Unmarshal JSON fields
	if len(rulesJSON) > 0 {
		json.Unmarshal(rulesJSON, &flag.Rules)
	}
	if len(rolloutJSON) > 0 {
		json.Unmarshal(rolloutJSON, &flag.Rollout)
	}
	if len(metadataJSON) > 0 {
		json.Unmarshal(metadataJSON, &flag.Metadata)
	}

	return &flag, nil
}

// initSchema creates the required database tables
func (s *PostgresStorage) initSchema() error {
	schema := `
		CREATE TABLE IF NOT EXISTS feature_flags (
			id VARCHAR(255) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			enabled BOOLEAN NOT NULL DEFAULT false,
			rules JSONB,
			rollout JSONB,
			environment VARCHAR(100) NOT NULL,
			tenant_id VARCHAR(255) NOT NULL,
			created_by VARCHAR(255) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL,
			updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
			metadata JSONB,
			UNIQUE(name, environment, tenant_id)
		);

		CREATE INDEX IF NOT EXISTS idx_feature_flags_env_tenant ON feature_flags(environment, tenant_id);
		CREATE INDEX IF NOT EXISTS idx_feature_flags_name ON feature_flags(name);
		CREATE INDEX IF NOT EXISTS idx_feature_flags_enabled ON feature_flags(enabled);

		CREATE TABLE IF NOT EXISTS flag_evaluations (
			id SERIAL PRIMARY KEY,
			flag_id VARCHAR(255) NOT NULL,
			user_id VARCHAR(255),
			tenant_id VARCHAR(255),
			result JSONB NOT NULL,
			context JSONB NOT NULL,
			evaluated_at TIMESTAMP WITH TIME ZONE NOT NULL
		);

		CREATE INDEX IF NOT EXISTS idx_flag_evaluations_flag_id ON flag_evaluations(flag_id);
		CREATE INDEX IF NOT EXISTS idx_flag_evaluations_user_id ON flag_evaluations(user_id);
		CREATE INDEX IF NOT EXISTS idx_flag_evaluations_evaluated_at ON flag_evaluations(evaluated_at);

		CREATE TABLE IF NOT EXISTS flag_audit_log (
			id SERIAL PRIMARY KEY,
			flag_id VARCHAR(255) NOT NULL,
			action VARCHAR(50) NOT NULL,
			changed_by VARCHAR(255) NOT NULL,
			before_value JSONB,
			after_value JSONB,
			changed_at TIMESTAMP WITH TIME ZONE NOT NULL
		);

		CREATE INDEX IF NOT EXISTS idx_flag_audit_log_flag_id ON flag_audit_log(flag_id);
		CREATE INDEX IF NOT EXISTS idx_flag_audit_log_changed_at ON flag_audit_log(changed_at);

		-- Configuration tables
		CREATE TABLE IF NOT EXISTS configurations (
			id VARCHAR(255) PRIMARY KEY,
			key VARCHAR(255) NOT NULL,
			value JSONB NOT NULL,
			environment VARCHAR(100) NOT NULL,
			tenant_id VARCHAR(255) NOT NULL,
			version INTEGER NOT NULL DEFAULT 1,
			description TEXT,
			created_by VARCHAR(255) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL,
			updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
			metadata JSONB,
			UNIQUE(key, environment, tenant_id)
		);

		CREATE INDEX IF NOT EXISTS idx_configurations_env_tenant ON configurations(environment, tenant_id);
		CREATE INDEX IF NOT EXISTS idx_configurations_key ON configurations(key);

		CREATE TABLE IF NOT EXISTS config_audit_log (
			id SERIAL PRIMARY KEY,
			config_id VARCHAR(255) NOT NULL,
			action VARCHAR(50) NOT NULL,
			changed_by VARCHAR(255) NOT NULL,
			before_value JSONB,
			after_value JSONB,
			changed_at TIMESTAMP WITH TIME ZONE NOT NULL
		);

		CREATE INDEX IF NOT EXISTS idx_config_audit_log_config_id ON config_audit_log(config_id);
		CREATE INDEX IF NOT EXISTS idx_config_audit_log_changed_at ON config_audit_log(changed_at);
	`

	_, err := s.db.Exec(schema)
	return err
}
