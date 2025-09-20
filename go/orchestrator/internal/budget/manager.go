package budget

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// Manager handles token budget management and cost control
type Manager struct {
	logger       *zap.Logger
	reservations map[string]*Reservation
	mu           sync.RWMutex
	db           DatabaseClient
	redis        RedisClient
}

// DatabaseClient interface for database operations
type DatabaseClient interface {
	GetUserBudget(ctx context.Context, userID, tenantID string) (*UserBudget, error)
	UpdateBudgetUsage(ctx context.Context, userID, tenantID string, tokens int, costUSD float64) error
	CreateBudgetEntry(ctx context.Context, entry *BudgetEntry) error
	GetBudgetEntries(ctx context.Context, userID, tenantID string, budgetType string, period time.Time) ([]*BudgetEntry, error)
}

// RedisClient interface for Redis operations
type RedisClient interface {
	IncrementCounter(ctx context.Context, key string, value int) error
	GetCounter(ctx context.Context, key string) (int, error)
	SetWithExpiry(ctx context.Context, key string, value interface{}, expiry time.Duration) error
	Get(ctx context.Context, key string) (interface{}, error)
}

// NewManager creates a new budget manager
func NewManager(logger *zap.Logger, db DatabaseClient, redis RedisClient) *Manager {
	return &Manager{
		logger:       logger,
		reservations: make(map[string]*Reservation),
		db:           db,
		redis:        redis,
	}
}

// UserBudget represents user budget information
type UserBudget struct {
	UserID           string    `json:"user_id"`
	TenantID         string    `json:"tenant_id"`
	DailyTokenLimit  int       `json:"daily_token_limit"`
	MonthlyTokenLimit int      `json:"monthly_token_limit"`
	DailyBudgetUSD   float64   `json:"daily_budget_usd"`
	MonthlyBudgetUSD float64   `json:"monthly_budget_usd"`
	TokensUsedToday  int       `json:"tokens_used_today"`
	TokensUsedMonth  int       `json:"tokens_used_month"`
	CostUsedToday    float64   `json:"cost_used_today"`
	CostUsedMonth    float64   `json:"cost_used_month"`
	LastResetDate    time.Time `json:"last_reset_date"`
}

// BudgetEntry represents a budget usage entry
type BudgetEntry struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	TenantID     string    `json:"tenant_id"`
	BudgetType   string    `json:"budget_type"`
	BudgetLimit  float64   `json:"budget_limit"`
	BudgetUsed   float64   `json:"budget_used"`
	TokensLimit  int       `json:"tokens_limit"`
	TokensUsed   int       `json:"tokens_used"`
	PeriodStart  time.Time `json:"period_start"`
	PeriodEnd    time.Time `json:"period_end"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
}

// Reservation represents a budget reservation
type Reservation struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	TenantID    string    `json:"tenant_id"`
	TokensReserved int    `json:"tokens_reserved"`
	TokensUsed  int       `json:"tokens_used"`
	CostUSD     float64   `json:"cost_usd"`
	CreatedAt   time.Time `json:"created_at"`
	ExpiresAt   time.Time `json:"expires_at"`
	Status      string    `json:"status"` // reserved, consumed, released, expired
}

// CheckBudget verifies if the user has sufficient budget for the estimated tokens
func (m *Manager) CheckBudget(ctx context.Context, userID, tenantID string, estimatedTokens int) error {
	m.logger.Debug("Checking budget",
		zap.String("user_id", userID),
		zap.String("tenant_id", tenantID),
		zap.Int("estimated_tokens", estimatedTokens),
	)

	// Get user budget from database
	budget, err := m.db.GetUserBudget(ctx, userID, tenantID)
	if err != nil {
		return fmt.Errorf("failed to get user budget: %w", err)
	}

	// Check daily token limit
	if budget.TokensUsedToday+estimatedTokens > budget.DailyTokenLimit {
		return fmt.Errorf("daily token limit exceeded: used %d + estimated %d > limit %d",
			budget.TokensUsedToday, estimatedTokens, budget.DailyTokenLimit)
	}

	// Check monthly token limit
	if budget.TokensUsedMonth+estimatedTokens > budget.MonthlyTokenLimit {
		return fmt.Errorf("monthly token limit exceeded: used %d + estimated %d > limit %d",
			budget.TokensUsedMonth, estimatedTokens, budget.MonthlyTokenLimit)
	}

	// Estimate cost (using average cost per token)
	estimatedCost := float64(estimatedTokens) * 0.002 // $0.002 per token average

	// Check daily budget
	if budget.CostUsedToday+estimatedCost > budget.DailyBudgetUSD {
		return fmt.Errorf("daily budget exceeded: used $%.4f + estimated $%.4f > limit $%.2f",
			budget.CostUsedToday, estimatedCost, budget.DailyBudgetUSD)
	}

	// Check monthly budget
	if budget.CostUsedMonth+estimatedCost > budget.MonthlyBudgetUSD {
		return fmt.Errorf("monthly budget exceeded: used $%.4f + estimated $%.4f > limit $%.2f",
			budget.CostUsedMonth, estimatedCost, budget.MonthlyBudgetUSD)
	}

	// Check Redis rate limits
	dailyKey := fmt.Sprintf("budget:daily:%s:%s", tenantID, userID)
	dailyUsage, err := m.redis.GetCounter(ctx, dailyKey)
	if err == nil && dailyUsage+estimatedTokens > budget.DailyTokenLimit {
		return fmt.Errorf("rate limit exceeded: daily usage %d + estimated %d > limit %d",
			dailyUsage, estimatedTokens, budget.DailyTokenLimit)
	}

	m.logger.Debug("Budget check passed",
		zap.String("user_id", userID),
		zap.Int("daily_tokens_available", budget.DailyTokenLimit-budget.TokensUsedToday),
		zap.Float64("daily_budget_available", budget.DailyBudgetUSD-budget.CostUsedToday),
	)

	return nil
}

// ReserveBudget reserves budget for a task execution
func (m *Manager) ReserveBudget(ctx context.Context, userID, tenantID string, tokens int) (string, error) {
	m.logger.Debug("Reserving budget",
		zap.String("user_id", userID),
		zap.String("tenant_id", tenantID),
		zap.Int("tokens", tokens),
	)

	// First check if budget is available
	if err := m.CheckBudget(ctx, userID, tenantID, tokens); err != nil {
		return "", fmt.Errorf("budget check failed: %w", err)
	}

	// Create reservation
	reservationID := generateReservationID()
	reservation := &Reservation{
		ID:             reservationID,
		UserID:         userID,
		TenantID:       tenantID,
		TokensReserved: tokens,
		TokensUsed:     0,
		CostUSD:        0,
		CreatedAt:      time.Now(),
		ExpiresAt:      time.Now().Add(time.Hour), // Reservations expire in 1 hour
		Status:         "reserved",
	}

	// Store reservation
	m.mu.Lock()
	m.reservations[reservationID] = reservation
	m.mu.Unlock()

	// Update Redis counters to prevent over-reservation
	dailyKey := fmt.Sprintf("budget:daily:%s:%s", tenantID, userID)
	monthlyKey := fmt.Sprintf("budget:monthly:%s:%s", tenantID, userID)
	
	if err := m.redis.IncrementCounter(ctx, dailyKey, tokens); err != nil {
		m.logger.Warn("Failed to update daily counter", zap.Error(err))
	}
	if err := m.redis.IncrementCounter(ctx, monthlyKey, tokens); err != nil {
		m.logger.Warn("Failed to update monthly counter", zap.Error(err))
	}

	// Set expiry for Redis keys
	m.redis.SetWithExpiry(ctx, dailyKey, nil, 24*time.Hour)
	m.redis.SetWithExpiry(ctx, monthlyKey, nil, 30*24*time.Hour)

	m.logger.Info("Budget reserved",
		zap.String("reservation_id", reservationID),
		zap.String("user_id", userID),
		zap.Int("tokens_reserved", tokens),
	)

	return reservationID, nil
}

// ConsumeBudget consumes the actual budget used
func (m *Manager) ConsumeBudget(ctx context.Context, reservationID string, actualTokens int, costUSD float64) error {
	m.logger.Debug("Consuming budget",
		zap.String("reservation_id", reservationID),
		zap.Int("actual_tokens", actualTokens),
		zap.Float64("cost_usd", costUSD),
	)

	// Get reservation
	m.mu.Lock()
	reservation, exists := m.reservations[reservationID]
	if !exists {
		m.mu.Unlock()
		return fmt.Errorf("reservation not found: %s", reservationID)
	}

	if reservation.Status != "reserved" {
		m.mu.Unlock()
		return fmt.Errorf("reservation already consumed or released: %s", reservationID)
	}

	// Update reservation
	reservation.TokensUsed = actualTokens
	reservation.CostUSD = costUSD
	reservation.Status = "consumed"
	m.mu.Unlock()

	// Update database budget usage
	err := m.db.UpdateBudgetUsage(ctx, reservation.UserID, reservation.TenantID, actualTokens, costUSD)
	if err != nil {
		m.logger.Error("Failed to update database budget usage", zap.Error(err))
		return fmt.Errorf("failed to update budget usage: %w", err)
	}

	// Create budget entry for tracking
	budgetEntry := &BudgetEntry{
		ID:          generateBudgetEntryID(),
		UserID:      reservation.UserID,
		TenantID:    reservation.TenantID,
		BudgetType:  "per_task",
		BudgetUsed:  costUSD,
		TokensUsed:  actualTokens,
		PeriodStart: reservation.CreatedAt,
		PeriodEnd:   time.Now(),
		IsActive:    true,
		CreatedAt:   time.Now(),
	}

	err = m.db.CreateBudgetEntry(ctx, budgetEntry)
	if err != nil {
		m.logger.Warn("Failed to create budget entry", zap.Error(err))
	}

	// Adjust Redis counters if actual usage differs from reservation
	tokenDiff := actualTokens - reservation.TokensReserved
	if tokenDiff != 0 {
		dailyKey := fmt.Sprintf("budget:daily:%s:%s", reservation.TenantID, reservation.UserID)
		monthlyKey := fmt.Sprintf("budget:monthly:%s:%s", reservation.TenantID, reservation.UserID)
		
		if err := m.redis.IncrementCounter(ctx, dailyKey, tokenDiff); err != nil {
			m.logger.Warn("Failed to adjust daily counter", zap.Error(err))
		}
		if err := m.redis.IncrementCounter(ctx, monthlyKey, tokenDiff); err != nil {
			m.logger.Warn("Failed to adjust monthly counter", zap.Error(err))
		}
	}

	m.logger.Info("Budget consumed",
		zap.String("reservation_id", reservationID),
		zap.String("user_id", reservation.UserID),
		zap.Int("tokens_used", actualTokens),
		zap.Float64("cost_usd", costUSD),
	)

	return nil
}

// ReleaseBudget releases an unused budget reservation
func (m *Manager) ReleaseBudget(ctx context.Context, reservationID string) error {
	m.logger.Debug("Releasing budget", zap.String("reservation_id", reservationID))

	// Get reservation
	m.mu.Lock()
	reservation, exists := m.reservations[reservationID]
	if !exists {
		m.mu.Unlock()
		return fmt.Errorf("reservation not found: %s", reservationID)
	}

	if reservation.Status != "reserved" {
		m.mu.Unlock()
		return nil // Already consumed or released
	}

	// Update reservation status
	reservation.Status = "released"
	m.mu.Unlock()

	// Release Redis counters
	dailyKey := fmt.Sprintf("budget:daily:%s:%s", reservation.TenantID, reservation.UserID)
	monthlyKey := fmt.Sprintf("budget:monthly:%s:%s", reservation.TenantID, reservation.UserID)
	
	if err := m.redis.IncrementCounter(ctx, dailyKey, -reservation.TokensReserved); err != nil {
		m.logger.Warn("Failed to release daily counter", zap.Error(err))
	}
	if err := m.redis.IncrementCounter(ctx, monthlyKey, -reservation.TokensReserved); err != nil {
		m.logger.Warn("Failed to release monthly counter", zap.Error(err))
	}

	m.logger.Info("Budget released",
		zap.String("reservation_id", reservationID),
		zap.String("user_id", reservation.UserID),
		zap.Int("tokens_released", reservation.TokensReserved),
	)

	return nil
}

// GetBudgetStatus returns current budget status for a user
func (m *Manager) GetBudgetStatus(ctx context.Context, userID, tenantID string) (*BudgetStatus, error) {
	budget, err := m.db.GetUserBudget(ctx, userID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user budget: %w", err)
	}

	// Get current usage from Redis
	dailyKey := fmt.Sprintf("budget:daily:%s:%s", tenantID, userID)
	monthlyKey := fmt.Sprintf("budget:monthly:%s:%s", tenantID, userID)
	
	dailyUsage, _ := m.redis.GetCounter(ctx, dailyKey)
	monthlyUsage, _ := m.redis.GetCounter(ctx, monthlyKey)

	status := &BudgetStatus{
		UserID:                userID,
		TenantID:              tenantID,
		DailyTokenLimit:       budget.DailyTokenLimit,
		MonthlyTokenLimit:     budget.MonthlyTokenLimit,
		DailyBudgetUSD:        budget.DailyBudgetUSD,
		MonthlyBudgetUSD:      budget.MonthlyBudgetUSD,
		TokensUsedToday:       dailyUsage,
		TokensUsedMonth:       monthlyUsage,
		CostUsedToday:         budget.CostUsedToday,
		CostUsedMonth:         budget.CostUsedMonth,
		DailyTokensRemaining:  budget.DailyTokenLimit - dailyUsage,
		MonthlyTokensRemaining: budget.MonthlyTokenLimit - monthlyUsage,
		DailyBudgetRemaining:  budget.DailyBudgetUSD - budget.CostUsedToday,
		MonthlyBudgetRemaining: budget.MonthlyBudgetUSD - budget.CostUsedMonth,
		LastUpdated:           time.Now(),
	}

	return status, nil
}

// BudgetStatus represents current budget status
type BudgetStatus struct {
	UserID                 string    `json:"user_id"`
	TenantID               string    `json:"tenant_id"`
	DailyTokenLimit        int       `json:"daily_token_limit"`
	MonthlyTokenLimit      int       `json:"monthly_token_limit"`
	DailyBudgetUSD         float64   `json:"daily_budget_usd"`
	MonthlyBudgetUSD       float64   `json:"monthly_budget_usd"`
	TokensUsedToday        int       `json:"tokens_used_today"`
	TokensUsedMonth        int       `json:"tokens_used_month"`
	CostUsedToday          float64   `json:"cost_used_today"`
	CostUsedMonth          float64   `json:"cost_used_month"`
	DailyTokensRemaining   int       `json:"daily_tokens_remaining"`
	MonthlyTokensRemaining int       `json:"monthly_tokens_remaining"`
	DailyBudgetRemaining   float64   `json:"daily_budget_remaining"`
	MonthlyBudgetRemaining float64   `json:"monthly_budget_remaining"`
	LastUpdated            time.Time `json:"last_updated"`
}

// CleanupExpiredReservations removes expired reservations
func (m *Manager) CleanupExpiredReservations(ctx context.Context) {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	expiredCount := 0

	for _, reservation := range m.reservations {
		if now.After(reservation.ExpiresAt) && reservation.Status == "reserved" {
			// Release the reservation
			reservation.Status = "expired"
			
			// Release Redis counters
			dailyKey := fmt.Sprintf("budget:daily:%s:%s", reservation.TenantID, reservation.UserID)
			monthlyKey := fmt.Sprintf("budget:monthly:%s:%s", reservation.TenantID, reservation.UserID)
			
			m.redis.IncrementCounter(ctx, dailyKey, -reservation.TokensReserved)
			m.redis.IncrementCounter(ctx, monthlyKey, -reservation.TokensReserved)
			
			expiredCount++
		}
	}

	if expiredCount > 0 {
		m.logger.Info("Cleaned up expired reservations", zap.Int("count", expiredCount))
	}
}

// Helper functions

func generateReservationID() string {
	return fmt.Sprintf("res_%d", time.Now().UnixNano())
}

func generateBudgetEntryID() string {
	return fmt.Sprintf("budget_%d", time.Now().UnixNano())
}