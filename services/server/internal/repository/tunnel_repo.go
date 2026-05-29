package repository

import (
	"context"
	"github.com/tunnelkit/services/server/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/google/uuid"
)

type PostgresTunnelRepo struct {
	pool *pgxpool.Pool
}

func NewPostgresTunnelRepo(pool *pgxpool.Pool) *PostgresTunnelRepo {
	return &PostgresTunnelRepo{pool: pool}
}

func (r *PostgresTunnelRepo) Create(ctx context.Context, t *domain.Tunnel) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO tunnels (id, user_id, name, protocol, subdomain, remote_port, auth_type, auth_config, ip_allowlist, status, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		t.ID, t.UserID, t.Name, t.Protocol, t.Subdomain, t.RemotePort,
		t.AuthType, t.AuthConfig, t.IPAllowlist, t.Status, t.CreatedAt,
	)
	return err
}

func (r *PostgresTunnelRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Tunnel, error) {
	var t domain.Tunnel
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, name, protocol, subdomain, remote_port, auth_type, auth_config, ip_allowlist, status, created_at
		 FROM tunnels WHERE id = $1`, id,
	).Scan(&t.ID, &t.UserID, &t.Name, &t.Protocol, &t.Subdomain, &t.RemotePort,
		&t.AuthType, &t.AuthConfig, &t.IPAllowlist, &t.Status, &t.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *PostgresTunnelRepo) GetBySubdomain(ctx context.Context, subdomain string) (*domain.Tunnel, error) {
	var t domain.Tunnel
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, name, protocol, subdomain, remote_port, auth_type, auth_config, ip_allowlist, status, created_at
		 FROM tunnels WHERE subdomain = $1`, subdomain,
	).Scan(&t.ID, &t.UserID, &t.Name, &t.Protocol, &t.Subdomain, &t.RemotePort,
		&t.AuthType, &t.AuthConfig, &t.IPAllowlist, &t.Status, &t.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *PostgresTunnelRepo) List(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*domain.Tunnel, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, name, protocol, subdomain, remote_port, auth_type, auth_config, ip_allowlist, status, created_at
		 FROM tunnels WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tunnels []*domain.Tunnel
	for rows.Next() {
		var t domain.Tunnel
		if err := rows.Scan(&t.ID, &t.UserID, &t.Name, &t.Protocol, &t.Subdomain, &t.RemotePort,
			&t.AuthType, &t.AuthConfig, &t.IPAllowlist, &t.Status, &t.CreatedAt); err != nil {
			return nil, err
		}
		tunnels = append(tunnels, &t)
	}
	return tunnels, nil
}

func (r *PostgresTunnelRepo) Update(ctx context.Context, t *domain.Tunnel) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE tunnels SET name = $2, subdomain = $3, auth_type = $4, auth_config = $5, ip_allowlist = $6, status = $7 WHERE id = $1`,
		t.ID, t.Name, t.Subdomain, t.AuthType, t.AuthConfig, t.IPAllowlist, t.Status,
	)
	return err
}

func (r *PostgresTunnelRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM tunnels WHERE id = $1`, id)
	return err
}
