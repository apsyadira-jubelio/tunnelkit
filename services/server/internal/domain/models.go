package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// User represents a system user
type User struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	Password  string    `json:"-"` // bcrypt hash, never expose
	Role      string    `json:"role"` // owner | admin | member | viewer
	CreatedAt time.Time `json:"created_at"`
}

// APIKey represents an API key
type APIKey struct {
	ID        uuid.UUID  `json:"id"`
	UserID    uuid.UUID  `json:"user_id"`
	Name      string     `json:"name"`
	KeyHash   string     `json:"-"` // bcrypt hash
	KeyPrefix string     `json:"key_prefix"` // first 8 chars for display
	Scopes    []string   `json:"scopes"`
	LastUsed  *time.Time `json:"last_used"`
	ExpiresAt *time.Time `json:"expires_at"`
	Revoked   bool       `json:"revoked"`
	CreatedAt time.Time  `json:"created_at"`
}

// Tunnel represents a tunnel configuration
type Tunnel struct {
	ID           uuid.UUID  `json:"id"`
	UserID       uuid.UUID  `json:"user_id"`
	Name         string     `json:"name"`
	Protocol     string     `json:"protocol"` // http | https | tcp
	Subdomain    *string    `json:"subdomain"`
	RemotePort   *int       `json:"remote_port"`
	AuthType     string     `json:"auth_type"` // none | basic | token
	AuthConfig   map[string]any `json:"auth_config,omitempty"`
	IPAllowlist  []string   `json:"ip_allowlist,omitempty"`
	Status       string     `json:"status"` // active | inactive | error
	CreatedAt    time.Time  `json:"created_at"`
}

// AuditLog represents an audit entry
type AuditLog struct {
	ID        int64     `json:"id"`
	ActorID   uuid.UUID `json:"actor_id"`
	Action    string    `json:"action"`
	Resource  string    `json:"resource"`
	IPAddress string    `json:"ip_address"`
	Metadata  map[string]any `json:"metadata,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// Repository interfaces
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	List(ctx context.Context, offset, limit int) ([]*User, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type TunnelRepository interface {
	Create(ctx context.Context, tunnel *Tunnel) error
	GetByID(ctx context.Context, id uuid.UUID) (*Tunnel, error)
	GetBySubdomain(ctx context.Context, subdomain string) (*Tunnel, error)
	List(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*Tunnel, error)
	Update(ctx context.Context, tunnel *Tunnel) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type APIKeyRepository interface {
	Create(ctx context.Context, key *APIKey) error
	GetByKeyHash(ctx context.Context, hash string) (*APIKey, error)
	List(ctx context.Context, userID uuid.UUID) ([]*APIKey, error)
	Revoke(ctx context.Context, id uuid.UUID) error
	UpdateLastUsed(ctx context.Context, id uuid.UUID) error
}

type AuditLogRepository interface {
	Create(ctx context.Context, log *AuditLog) error
	List(ctx context.Context, filter AuditLogFilter) ([]*AuditLog, int, error)
}

type AuditLogFilter struct {
	ActorID   *uuid.UUID
	Action    string
	StartTime *time.Time
	EndTime   *time.Time
	Offset    int
	Limit     int
}
