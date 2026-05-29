package api

import (
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
	"github.com/tunnelkit/services/server/internal/domain"
	"github.com/tunnelkit/services/server/internal/config"
)

type AuthHandler struct {
	userRepo  domain.UserRepository
	auditRepo domain.AuditLogRepository
	config    *config.Config
}

func NewAuthHandler(userRepo domain.UserRepository, auditRepo domain.AuditLogRepository, cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		userRepo:  userRepo,
		auditRepo: auditRepo,
		config:    cfg,
	}
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

type LoginResponse struct {
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expires_at"`
	User      *domain.User `json:"user"`
}

func (h *AuthHandler) Login(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	user, err := h.userRepo.GetByEmail(c.Request().Context(), req.Email)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid credentials")
	}

	// Generate JWT
	expiresAt := time.Now().Add(15 * time.Minute)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID.String(),
		"email":   user.Email,
		"role":    user.Role,
		"exp":     expiresAt.Unix(),
	})

	tokenString, err := token.SignedString([]byte(h.config.JWTSecret))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to sign token")
	}

	// Audit log
	go h.auditRepo.Create(c.Request().Context(), &domain.AuditLog{
		ActorID:   user.ID,
		Action:    "user.login",
		Resource:  "users:" + user.ID.String(),
		IPAddress: c.RealIP(),
		CreatedAt: time.Now(),
	})

	return c.JSON(http.StatusOK, LoginResponse{
		Token:     tokenString,
		ExpiresAt: expiresAt.Unix(),
		User:      user,
	})
}

type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

func (h *AuthHandler) Register(c echo.Context) error {
	var req RegisterRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	existing, _ := h.userRepo.GetByEmail(c.Request().Context(), req.Email)
	if existing != nil {
		return echo.NewHTTPError(http.StatusConflict, "email already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to hash password")
	}

	user := &domain.User{
		ID:        uuid.New(),
		Email:     req.Email,
		Password:  string(hashedPassword),
		Role:      "member",
		CreatedAt: time.Now(),
	}

	if err := h.userRepo.Create(c.Request().Context(), user); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create user")
	}

	return c.JSON(http.StatusCreated, user)
}

func (h *AuthHandler) GetMe(c echo.Context) error {
	userID := c.Get("user_id").(string)
	id, err := uuid.Parse(userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid user")
	}

	user, err := h.userRepo.GetByID(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "user not found")
	}

	return c.JSON(http.StatusOK, user)
}
