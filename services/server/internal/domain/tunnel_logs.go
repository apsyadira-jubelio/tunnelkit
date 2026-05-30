package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// TunnelLog represents a request log entry for a tunnel
type TunnelLog struct {
	ID        int64     `json:"id"`
	TunnelID  uuid.UUID `json:"tunnel_id"`
	Method    string    `json:"method"`
	Path      string    `json:"path"`
	Query     string    `json:"query,omitempty"`
	StatusCode int      `json:"status_code"`
	Duration  int       `json:"duration_ms"` // milliseconds
	RequestBody  []byte `json:"request_body,omitempty"`
	ResponseBody []byte `json:"response_body,omitempty"`
	RequestHeaders  map[string]string `json:"request_headers,omitempty"`
	ResponseHeaders map[string]string `json:"response_headers,omitempty"`
	ClientIP   string   `json:"client_ip"`
	UserAgent  string   `json:"user_agent,omitempty"`
	Error      string   `json:"error,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

// TunnelLogRepository interface
type TunnelLogRepository interface {
	Create(ctx context.Context, log *TunnelLog) error
	List(ctx context.Context, tunnelID uuid.UUID, offset, limit int) ([]*TunnelLog, error)
	Stream(ctx context.Context, tunnelID uuid.UUID, since time.Time) <-chan *TunnelLog
}

// TunnelMetrics represents tunnel statistics
type TunnelMetrics struct {
	TunnelID      uuid.UUID `json:"tunnel_id"`
	TotalRequests int64     `json:"total_requests"`
	TotalBytes    int64     `json:"total_bytes"`
	AvgLatency    float64   `json:"avg_latency_ms"`
	ErrorCount    int64     `json:"error_count"`
	LastActivity  time.Time `json:"last_activity"`
	RequestsPerMin float64  `json:"requests_per_min"`
}
