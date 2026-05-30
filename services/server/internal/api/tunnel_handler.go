package api

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/tunnelkit/services/server/internal/domain"
	"github.com/tunnelkit/services/server/internal/config"
)

type TunnelHandler struct {
	tunnelRepo domain.TunnelRepository
	auditRepo  domain.AuditLogRepository
	config     *config.Config
}

func NewTunnelHandler(tunnelRepo domain.TunnelRepository, auditRepo domain.AuditLogRepository, cfg *config.Config) *TunnelHandler {
	return &TunnelHandler{
		tunnelRepo: tunnelRepo,
		auditRepo:  auditRepo,
		config:     cfg,
	}
}

type CreateTunnelRequest struct {
	Name        string   `json:"name" validate:"required"`
	Protocol    string   `json:"protocol" validate:"required,oneof=http https tcp"`
	Subdomain   string   `json:"subdomain"`
	RemotePort  int      `json:"remote_port"`
	AuthType    string   `json:"auth_type" validate:"omitempty,oneof=basic token"`
	IPAllowlist []string `json:"ip_allowlist"`
}

func (h *TunnelHandler) Create(c echo.Context) error {
	var req CreateTunnelRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	userID := c.Get("user_id").(string)
	uid, _ := uuid.Parse(userID)

	subdomain := &req.Subdomain
	if req.Subdomain == "" {
		// Auto-generate unique subdomain
		random := uuid.New().String()[:8]
		autoSubdomain := req.Name + "-" + random
		subdomain = &autoSubdomain
	}

	tunnel := &domain.Tunnel{
		ID:          uuid.New(),
		UserID:      uid,
		Name:        req.Name,
		Protocol:    req.Protocol,
		Subdomain:   subdomain,
		RemotePort:  &req.RemotePort,
		AuthType:    req.AuthType,
		IPAllowlist: req.IPAllowlist,
		Status:      "active",
		CreatedAt:   time.Now(),
	}

	if err := h.tunnelRepo.Create(c.Request().Context(), tunnel); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create tunnel")
	}

	go h.auditRepo.Create(c.Request().Context(), &domain.AuditLog{
		ActorID:   uid,
		Action:    "tunnel.create",
		Resource:  "tunnels:" + tunnel.ID.String(),
		IPAddress: c.RealIP(),
		Metadata:  map[string]any{"name": tunnel.Name, "protocol": tunnel.Protocol},
		CreatedAt: time.Now(),
	})

	return c.JSON(http.StatusCreated, tunnel)
}

func (h *TunnelHandler) List(c echo.Context) error {
	userID := c.Get("user_id").(string)
	uid, _ := uuid.Parse(userID)

	tunnels, err := h.tunnelRepo.List(c.Request().Context(), uid, 0, 100)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list tunnels")
	}

	return c.JSON(http.StatusOK, tunnels)
}

func (h *TunnelHandler) GetByID(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid tunnel id")
	}

	tunnel, err := h.tunnelRepo.GetByID(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "tunnel not found")
	}

	userID := c.Get("user_id").(string)
	if tunnel.UserID.String() != userID {
		return echo.NewHTTPError(http.StatusForbidden, "access denied")
	}

	return c.JSON(http.StatusOK, tunnel)
}

func (h *TunnelHandler) Update(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid tunnel id")
	}

	tunnel, err := h.tunnelRepo.GetByID(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "tunnel not found")
	}

	userID := c.Get("user_id").(string)
	if tunnel.UserID.String() != userID {
		return echo.NewHTTPError(http.StatusForbidden, "access denied")
	}

	var req struct {
		Name        string   `json:"name"`
		Subdomain   string   `json:"subdomain"`
		AuthType    string   `json:"auth_type"`
		IPAllowlist []string `json:"ip_allowlist"`
		Status      string   `json:"status"`
	}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	if req.Name != "" {
		tunnel.Name = req.Name
	}
	if req.Subdomain != "" {
		tunnel.Subdomain = &req.Subdomain
	}
	if req.AuthType != "" {
		tunnel.AuthType = req.AuthType
	}
	if req.IPAllowlist != nil {
		tunnel.IPAllowlist = req.IPAllowlist
	}
	if req.Status != "" {
		tunnel.Status = req.Status
	}

	if err := h.tunnelRepo.Update(c.Request().Context(), tunnel); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update tunnel")
	}

	go h.auditRepo.Create(c.Request().Context(), &domain.AuditLog{
		ActorID:   uuid.MustParse(userID),
		Action:    "tunnel.update",
		Resource:  "tunnels:" + tunnel.ID.String(),
		IPAddress: c.RealIP(),
		CreatedAt: time.Now(),
	})

	return c.JSON(http.StatusOK, tunnel)
}

func (h *TunnelHandler) Delete(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid tunnel id")
	}

	tunnel, err := h.tunnelRepo.GetByID(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "tunnel not found")
	}

	userID := c.Get("user_id").(string)
	if tunnel.UserID.String() != userID {
		return echo.NewHTTPError(http.StatusForbidden, "access denied")
	}

	if err := h.tunnelRepo.Delete(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to delete tunnel")
	}

	go h.auditRepo.Create(c.Request().Context(), &domain.AuditLog{
		ActorID:   tunnel.UserID,
		Action:    "tunnel.delete",
		Resource:  "tunnels:" + id.String(),
		IPAddress: c.RealIP(),
		CreatedAt: time.Now(),
	})

	return c.JSON(http.StatusOK, map[string]string{"status": "deleted"})
}
