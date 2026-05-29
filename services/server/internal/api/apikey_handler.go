package api

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
	"github.com/tunnelkit/services/server/internal/domain"
	"github.com/tunnelkit/services/server/internal/config"
)

type APIKeyHandler struct {
	apiKeyRepo domain.APIKeyRepository
	auditRepo  domain.AuditLogRepository
	config     *config.Config
}

func NewAPIKeyHandler(apiKeyRepo domain.APIKeyRepository, auditRepo domain.AuditLogRepository, cfg *config.Config) *APIKeyHandler {
	return &APIKeyHandler{
		apiKeyRepo: apiKeyRepo,
		auditRepo:  auditRepo,
		config:     cfg,
	}
}

type CreateAPIKeyRequest struct {
	Name   string   `json:"name" validate:"required"`
	Scopes []string `json:"scopes" validate:"required"`
}

type CreateAPIKeyResponse struct {
	*domain.APIKey
	PlainKey string `json:"plain_key"` // only returned once
}

func (h *APIKeyHandler) Create(c echo.Context) error {
	var req CreateAPIKeyRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	userID := c.Get("user_id").(string)
	uid, _ := uuid.Parse(userID)

	// Generate random key
	bytes := make([]byte, 32)
	rand.Read(bytes)
	plainKey := hex.EncodeToString(bytes)

	keyHash, err := bcrypt.GenerateFromPassword([]byte(plainKey), 12)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to hash key")
	}

	apiKey := &domain.APIKey{
		ID:        uuid.New(),
		UserID:    uid,
		Name:      req.Name,
		KeyHash:   string(keyHash),
		KeyPrefix: plainKey[:8],
		Scopes:    req.Scopes,
		Revoked:   false,
		CreatedAt: time.Now(),
	}

	if err := h.apiKeyRepo.Create(c.Request().Context(), apiKey); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create API key")
	}

	go h.auditRepo.Create(c.Request().Context(), &domain.AuditLog{
		ActorID:   uid,
		Action:    "apikey.create",
		Resource:  "api_keys:" + apiKey.ID.String(),
		IPAddress: c.RealIP(),
		Metadata:  map[string]any{"name": apiKey.Name},
		CreatedAt: time.Now(),
	})

	return c.JSON(http.StatusCreated, CreateAPIKeyResponse{
		APIKey:   apiKey,
		PlainKey: plainKey,
	})
}

func (h *APIKeyHandler) List(c echo.Context) error {
	userID := c.Get("user_id").(string)
	uid, _ := uuid.Parse(userID)

	keys, err := h.apiKeyRepo.List(c.Request().Context(), uid)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list API keys")
	}

	return c.JSON(http.StatusOK, keys)
}

func (h *APIKeyHandler) Revoke(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid key id")
	}

	if err := h.apiKeyRepo.Revoke(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to revoke key")
	}

	go h.auditRepo.Create(c.Request().Context(), &domain.AuditLog{
		ActorID:   uuid.MustParse(c.Get("user_id").(string)),
		Action:    "apikey.revoke",
		Resource:  "api_keys:" + id.String(),
		IPAddress: c.RealIP(),
		CreatedAt: time.Now(),
	})

	return c.JSON(http.StatusOK, map[string]string{"status": "revoked"})
}
