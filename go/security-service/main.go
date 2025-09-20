package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/multi-agent/go/security-service/internal/auth"
	"github.com/multi-agent/go/security-service/internal/config"
	"github.com/multi-agent/go/security-service/internal/database"
	"github.com/multi-agent/go/security-service/internal/middleware"
	"github.com/multi-agent/go/security-service/internal/opa"
	"github.com/multi-agent/go/security-service/internal/rbac"
	"github.com/multi-agent/go/security-service/internal/tenant"
)

func main() {
	// Initialize logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("Failed to load configuration", zap.Error(err))
	}

	// Initialize database
	db, err := database.NewClient(cfg.Database, logger)
	if err != nil {
		logger.Fatal("Failed to initialize database", zap.Error(err))
	}
	defer db.Close()

	// Initialize OPA policy engine
	opaEngine, err := opa.NewEngine(cfg.OPA, logger)
	if err != nil {
		logger.Fatal("Failed to initialize OPA engine", zap.Error(err))
	}

	// Initialize services
	tenantService := tenant.NewService(db, logger)
	rbacService := rbac.NewService(db, opaEngine, logger)
	authService := auth.NewService(db, rbacService, cfg.JWT, logger)

	// Initialize Gin router
	if cfg.Server.Mode == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.Logger(logger))
	router.Use(middleware.CORS())
	router.Use(middleware.Security())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   "security-service",
			"timestamp": time.Now().UTC(),
		})
	})

	// API routes
	v1 := router.Group("/api/v1")
	{
		// Authentication routes
		auth := v1.Group("/auth")
		{
			auth.POST("/login", authService.Login)
			auth.POST("/refresh", authService.RefreshToken)
			auth.POST("/logout", middleware.RequireAuth(authService), authService.Logout)
			auth.GET("/profile", middleware.RequireAuth(authService), authService.GetProfile)
		}

		// Tenant management routes (admin only)
		tenants := v1.Group("/tenants")
		tenants.Use(middleware.RequireAuth(authService))
		tenants.Use(middleware.RequireRole(rbacService, "admin"))
		{
			tenants.POST("", tenantService.CreateTenant)
			tenants.GET("", tenantService.ListTenants)
			tenants.GET("/:id", tenantService.GetTenant)
			tenants.PUT("/:id", tenantService.UpdateTenant)
			tenants.DELETE("/:id", tenantService.DeleteTenant)
			tenants.POST("/:id/users", tenantService.AddUserToTenant)
			tenants.DELETE("/:id/users/:user_id", tenantService.RemoveUserFromTenant)
		}

		// RBAC management routes
		rbac := v1.Group("/rbac")
		rbac.Use(middleware.RequireAuth(authService))
		rbac.Use(middleware.RequirePermission(rbacService, "rbac:manage"))
		{
			rbac.POST("/roles", rbacService.CreateRole)
			rbac.GET("/roles", rbacService.ListRoles)
			rbac.PUT("/roles/:id", rbacService.UpdateRole)
			rbac.DELETE("/roles/:id", rbacService.DeleteRole)
			
			rbac.POST("/permissions", rbacService.CreatePermission)
			rbac.GET("/permissions", rbacService.ListPermissions)
			
			rbac.POST("/users/:user_id/roles", rbacService.AssignRole)
			rbac.DELETE("/users/:user_id/roles/:role_id", rbacService.RevokeRole)
			rbac.GET("/users/:user_id/permissions", rbacService.GetUserPermissions)
		}

		// Policy management routes
		policies := v1.Group("/policies")
		policies.Use(middleware.RequireAuth(authService))
		policies.Use(middleware.RequirePermission(rbacService, "policy:manage"))
		{
			policies.POST("", opaEngine.CreatePolicy)
			policies.GET("", opaEngine.ListPolicies)
			policies.GET("/:id", opaEngine.GetPolicy)
			policies.PUT("/:id", opaEngine.UpdatePolicy)
			policies.DELETE("/:id", opaEngine.DeletePolicy)
			policies.POST("/evaluate", opaEngine.EvaluatePolicy)
		}

		// Security audit routes
		audit := v1.Group("/audit")
		audit.Use(middleware.RequireAuth(authService))
		audit.Use(middleware.RequirePermission(rbacService, "audit:read"))
		{
			audit.GET("/events", authService.GetAuditEvents)
			audit.GET("/sessions", authService.GetActiveSessions)
			audit.POST("/sessions/:id/revoke", authService.RevokeSession)
		}
	}

	// Start server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.Server.IdleTimeout) * time.Second,
	}

	// Graceful shutdown
	go func() {
		logger.Info("Starting security service",
			zap.Int("port", cfg.Server.Port),
			zap.String("mode", cfg.Server.Mode),
		)
		
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down security service...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Security service stopped")
}