package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"

	"github.com/tunnelkit/services/server/internal/api"
	"github.com/tunnelkit/services/server/internal/config"
	"github.com/tunnelkit/services/server/internal/repository"
	"github.com/tunnelkit/services/server/internal/tunnel"
)

func main() {
	// Load config
	cfg := config.Load()

	// Logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Database
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer pool.Close()

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		log.Fatal("Failed to ping database:", err)
	}
	logger.Info("Database connected")

	// Repositories
	userRepo := repository.NewPostgresUserRepo(pool)
	tunnelRepo := repository.NewPostgresTunnelRepo(pool)
	apiKeyRepo := repository.NewPostgresAPIKeyRepo(pool)
	auditRepo := repository.NewPostgresAuditLogRepo(pool)

	// Handlers
	authHandler := api.NewAuthHandler(userRepo, auditRepo, cfg)
	tunnelHandler := api.NewTunnelHandler(tunnelRepo, auditRepo, cfg)
	apiKeyHandler := api.NewAPIKeyHandler(apiKeyRepo, auditRepo, cfg)

	// Tunnel hub
	tunnelHub := tunnel.NewTunnelHub()

	// Subdomain router
	subdomainRouter := tunnel.NewSubdomainRouter(tunnelHub, tunnelRepo, cfg.BaseDomain)
	_ = subdomainRouter // TODO: integrate with ingress

	// Echo setup
	e := echo.New()
	e.HideBanner = true

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{echo.GET, echo.POST, echo.PUT, echo.PATCH, echo.DELETE},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
	}))

	// Routes
	api.SetupRoutes(e, &api.Dependencies{
		Config:        cfg,
		Auth:          authHandler,
		Tunnel:        tunnelHandler,
		APIKey:        apiKeyHandler,
		TunnelHub:     tunnelHub,
		JWTMiddleware: api.JWTAuth(cfg),
	})

	// Start server
	go func() {
		addr := fmt.Sprintf(":%s", cfg.Port)
		logger.Info("Server starting", zap.String("addr", addr))
		if err := e.Start(addr); err != nil {
			logger.Error("Server error", zap.Error(err))
		}
	}()

	// Wait for interrupt
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited")
}
