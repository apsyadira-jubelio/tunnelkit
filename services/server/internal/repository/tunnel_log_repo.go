package repository

import (
	"context"
	"time"
	"github.com/tunnelkit/services/server/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/google/uuid"
)

type PostgresTunnelLogRepo struct {
	pool *pgxpool.Pool
}

func NewPostgresTunnelLogRepo(pool *pgxpool.Pool) *PostgresTunnelLogRepo {
	return &PostgresTunnelLogRepo{pool: pool}
}

func (r *PostgresTunnelLogRepo) Create(ctx context.Context, log *domain.TunnelLog) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO tunnel_logs (tunnel_id, method, path, query, status_code, duration_ms, 
		 request_body, response_body, request_headers, response_headers, client_ip, user_agent, error, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`,
		log.TunnelID, log.Method, log.Path, log.Query, log.StatusCode, log.Duration,
		log.RequestBody, log.ResponseBody, log.RequestHeaders, log.ResponseHeaders,
		log.ClientIP, log.UserAgent, log.Error, log.CreatedAt,
	)
	return err
}

func (r *PostgresTunnelLogRepo) List(ctx context.Context, tunnelID uuid.UUID, offset, limit int) ([]*domain.TunnelLog, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, tunnel_id, method, path, query, status_code, duration_ms, 
		 request_body, response_body, request_headers, response_headers, client_ip, user_agent, error, created_at
		 FROM tunnel_logs WHERE tunnel_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		tunnelID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*domain.TunnelLog
	for rows.Next() {
		var l domain.TunnelLog
		if err := rows.Scan(&l.ID, &l.TunnelID, &l.Method, &l.Path, &l.Query, &l.StatusCode, &l.Duration,
			&l.RequestBody, &l.ResponseBody, &l.RequestHeaders, &l.ResponseHeaders,
			&l.ClientIP, &l.UserAgent, &l.Error, &l.CreatedAt); err != nil {
			return nil, err
		}
		logs = append(logs, &l)
	}
	return logs, nil
}

func (r *PostgresTunnelLogRepo) Stream(ctx context.Context, tunnelID uuid.UUID, since time.Time) <-chan *domain.TunnelLog {
	ch := make(chan *domain.TunnelLog, 100)
	go func() {
		defer close(ch)
		// Poll for new logs - simplified implementation
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		lastID := int64(0)
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				rows, err := r.pool.Query(ctx,
					`SELECT id, tunnel_id, method, path, query, status_code, duration_ms, 
					 request_body, response_body, request_headers, response_headers, client_ip, user_agent, error, created_at
					 FROM tunnel_logs WHERE tunnel_id = $1 AND id > $2 ORDER BY id ASC LIMIT 50`,
					tunnelID, lastID,
				)
				if err != nil {
					continue
				}
				for rows.Next() {
					var l domain.TunnelLog
					if err := rows.Scan(&l.ID, &l.TunnelID, &l.Method, &l.Path, &l.Query, &l.StatusCode, &l.Duration,
						&l.RequestBody, &l.ResponseBody, &l.RequestHeaders, &l.ResponseHeaders,
						&l.ClientIP, &l.UserAgent, &l.Error, &l.CreatedAt); err != nil {
						continue
					}
					lastID = l.ID
					ch <- &l
				}
				rows.Close()
			}
		}
	}()
	return ch
}
