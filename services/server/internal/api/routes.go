package api

import (
	"github.com/labstack/echo/v4"
	"github.com/tunnelkit/services/server/internal/config"
	"github.com/tunnelkit/services/server/internal/tunnel"
)

func SetupRoutes(e *echo.Echo, deps *Dependencies) {
	// Health check
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"status": "ok"})
	})

	// API v1
	v1 := e.Group("/api/v1")

	// Auth (no JWT required)
	auth := v1.Group("/auth")
	auth.POST("/login", deps.Auth.Login)
	auth.POST("/register", deps.Auth.Register)
	auth.POST("/logout", func(c echo.Context) error { return c.JSON(200, map[string]string{"status": "ok"}) })

	// Protected routes
	protected := v1.Group("")
	protected.Use(deps.JWTMiddleware)

	// Auth protected
	protected.GET("/auth/me", deps.Auth.GetMe)
	protected.POST("/auth/refresh", func(c echo.Context) error { return c.JSON(200, map[string]string{"token": "refreshed"}) })

	// Tunnels
	tunnels := protected.Group("/tunnels")
	tunnels.GET("", deps.Tunnel.List)
	tunnels.POST("", deps.Tunnel.Create)
	tunnels.GET("/:id", deps.Tunnel.GetByID)
	tunnels.PATCH("/:id", func(c echo.Context) error { return c.JSON(501, map[string]string{"status": "not implemented"}) })
	tunnels.DELETE("/:id", deps.Tunnel.Delete)

	// API Keys
	apiKeys := protected.Group("/api-keys")
	apiKeys.GET("", deps.APIKey.List)
	apiKeys.POST("", deps.APIKey.Create)
	apiKeys.DELETE("/:id", deps.APIKey.Revoke)

	// WebSocket endpoint for tunnel agents
	e.GET("/ws/agent", deps.TunnelHub.HandleWS, deps.JWTMiddleware)
}

type Dependencies struct {
	Config        *config.Config
	Auth          *AuthHandler
	Tunnel        *TunnelHandler
	APIKey        *APIKeyHandler
	TunnelHub     *tunnel.TunnelHub
	JWTMiddleware echo.MiddlewareFunc
}
