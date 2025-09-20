package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"go.uber.org/zap"

	"github.com/multi-agent/go/orchestrator/internal/budget"
)

// Config holds database configuration
type Config struct {
	Host         string `yaml:"host" env:"DB_HOST" default:"localhost"`
	Port         int    `yaml:"port" env:"DB_PORT" default:"5432"`
	Database     string `yaml:"database" env:"DB_NAME" default:"multiagent"`
	Username     string `yaml:"username" env:"DB_USER" default:"postgres"`
	Password     string `yaml:"password" env:"DB_PASSWORD" default:""`
	SSLMode      string `yaml:"ssl_mode" env:"DB_SSL_MODE" default:"disable"`
	MaxOpenConns int    `yaml:"max_open_conns" env:"DB_MAX_OPEN_CONNS" default:"25"`
	MaxIdleConns int    `yaml:"max_idle_conns" env:"DB_MAX_IDLE_CONNS" default:"5"`
	MaxLifetime  string `yaml:"max_lifetime" env:"DB_MAX_LIFETIME" default:"5m"`
}

// Client wraps database operations
type Client struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewClient creates a new database client
func NewClient(config Config, logger *zap.Logger) (*Client, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.Username, config.Password, config.Database, config.SSLMode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	
	if maxLifetime, err := time.ParseDuration(config.MaxLifetime); err == nil {
		db.SetConnMaxLifetime(maxLifetime)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("Database connection established",
		zap.String("host", config.Host),
		zap.Int("port", config.Port),
		zap.String("database", config.Database),
	)

	return &Client{
		db:     db,
		logger: logger,
	}, nil
}

// Close closes the database connection
func (c *Client) Close() error {
	return c.db.Close()
}

// GetUserBudget retrieves user budget information
func (c *Client) GetUserBudget(ctx context.Context, userID, tenantID string) (*budget.UserBudget, error) {
	query := `
		SELECT user_id, tenant_id, daily_limit_tokens, daily_limit_usd, 
		       monthly_limit_tokens, monthly_limit_usd, daily_used_tokens, 
		       daily_used_usd, monthly_used_tokens, monthly_used_usd, 
		       last_reset_date, created_at, updated_at
		FROM budget_limits 
		WHERE user_id = $1 AND tenant_id = $2`

	var ub budget.UserBudget
	err := c.db.QueryRowContext(ctx, query, userID, tenantID).Scan(
		&ub.UserID, &ub.TenantID, &ub.DailyTokenLimit, &ub.DailyBudgetUSD,
		&ub.MonthlyTokenLimit, &ub.MonthlyBudgetUSD, &ub.TokensUsedToday,
		&ub.CostUsedToday, &ub.TokensUsedMonth, &ub.CostUsedMonth,
		&ub.LastResetDate,
	)

	if err == sql.ErrNoRows {
		// Return default budget if not found
		return &budget.UserBudget{
			UserID:              userID,
			TenantID:            tenantID,
			DailyTokenLimit:     100000,
			DailyBudgetUSD:      10.0,
			MonthlyTokenLimit:   3000000,
			MonthlyBudgetUSD:    300.0,
			TokensUsedToday:     0,
			CostUsedToday:       0.0,
			TokensUsedMonth:     0,
			CostUsedMonth:       0.0,
			LastResetDate:       time.Now(),
		}, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get user budget: %w", err)
	}

	return &ub, nil
}

// UpdateBudgetUsage updates budget usage
func (c *Client) UpdateBudgetUsage(ctx context.Context, userID, tenantID string, tokens int, costUSD float64) error {
	query := `
		INSERT INTO budget_limits (user_id, tenant_id, tokens_used_today, cost_used_today, 
		                          tokens_used_month, cost_used_month, last_reset_date)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
		ON CONFLICT (user_id, tenant_id)
		DO UPDATE SET
			tokens_used_today = budget_limits.tokens_used_today + $3,
			cost_used_today = budget_limits.cost_used_today + $4,
			tokens_used_month = budget_limits.tokens_used_month + $5,
			cost_used_month = budget_limits.cost_used_month + $6,
			last_reset_date = NOW()`

	_, err := c.db.ExecContext(ctx, query, userID, tenantID, tokens, costUSD, tokens, costUSD)
	if err != nil {
		return fmt.Errorf("failed to update budget usage: %w", err)
	}

	return nil
}

// CreateBudgetEntry creates a new budget entry
func (c *Client) CreateBudgetEntry(ctx context.Context, entry *budget.BudgetEntry) error {
	query := `
		INSERT INTO budget_entries (id, user_id, tenant_id, budget_type, 
		                           budget_limit, budget_used, tokens_limit, tokens_used, 
		                           period_start, period_end, is_active, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

	_, err := c.db.ExecContext(ctx, query,
		entry.ID, entry.UserID, entry.TenantID, entry.BudgetType,
		entry.BudgetLimit, entry.BudgetUsed, entry.TokensLimit, entry.TokensUsed,
		entry.PeriodStart, entry.PeriodEnd, entry.IsActive, entry.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create budget entry: %w", err)
	}

	return nil
}

// GetBudgetEntries retrieves budget entries
func (c *Client) GetBudgetEntries(ctx context.Context, userID, tenantID string, budgetType string, period time.Time) ([]*budget.BudgetEntry, error) {
	query := `
		SELECT id, user_id, tenant_id, budget_type, budget_limit, budget_used, 
		       tokens_limit, tokens_used, period_start, period_end, is_active, created_at
		FROM budget_entries 
		WHERE user_id = $1 AND tenant_id = $2 AND budget_type = $3 
		      AND created_at >= $4
		ORDER BY created_at DESC`

	rows, err := c.db.QueryContext(ctx, query, userID, tenantID, budgetType, period)
	if err != nil {
		return nil, fmt.Errorf("failed to query budget entries: %w", err)
	}
	defer rows.Close()

	var entries []*budget.BudgetEntry
	for rows.Next() {
		var entry budget.BudgetEntry
		err := rows.Scan(
			&entry.ID, &entry.UserID, &entry.TenantID, &entry.BudgetType,
			&entry.BudgetLimit, &entry.BudgetUsed, &entry.TokensLimit, &entry.TokensUsed,
			&entry.PeriodStart, &entry.PeriodEnd, &entry.IsActive, &entry.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan budget entry: %w", err)
		}
		entries = append(entries, &entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate budget entries: %w", err)
	}

	return entries, nil
}

// Health checks database health
func (c *Client) Health(ctx context.Context) error {
	return c.db.PingContext(ctx)
}