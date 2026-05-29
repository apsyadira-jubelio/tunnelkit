package repository

import (
	"context"
	"github.com/tunnelkit/services/server/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/google/uuid"
)

type PostgresAPIKeyRepo struct {
	pool *pgxpool.Pool
}

func NewPostgresAPIKeyRepo(pool *pgxpool.Pool) *PostgresAPIKeyRepo {
	return &PostgresAPIKeyRepo{pool: pool}
}

func (r *PostgresAPIKeyRepo) Create(ctx context.Context, key *domain.APIKey) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO api_keys (id, user_id, name, key_hash, key_prefix, scopes, last_used, expires_at, revoked, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		key.ID, key.UserID, key.Name, key.KeyHash, key.KeyPrefix, key.Scopes,
		key.LastUsed, key.ExpiresAt, key.Revoked, key.CreatedAt,
	)
	return err
}

func (r *PostgresAPIKeyRepo) GetByKeyHash(ctx context.Context, hash string) (*domain.APIKey, error) {
	var k domain.APIKey
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, name, key_hash, key_prefix, scopes, last_used, expires_at, revoked, created_at
		 FROM api_keys WHERE key_hash = $1 AND revoked = false`, hash,
	).Scan(&k.ID, &k.UserID, &k.Name, &k.KeyHash, &k.KeyPrefix, &k.Scopes,
		&k.LastUsed, &k.ExpiresAt, &k.Revoked, &k.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &k, nil
}

func (r *PostgresAPIKeyRepo) List(ctx context.Context, userID uuid.UUID) ([]*domain.APIKey, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, name, key_hash, key_prefix, scopes, last_used, expires_at, revoked, created_at
		 FROM api_keys WHERE user_id = $1 ORDER BY created_at DESC`, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []*domain.APIKey
	for rows.Next() {
		var k domain.APIKey
		if err := rows.Scan(&k.ID, &k.UserID, &k.Name, &k.KeyHash, &k.KeyPrefix, &k.Scopes,
			&k.LastUsed, &k.ExpiresAt, &k.Revoked, &k.CreatedAt); err != nil {
			return nil, err
		}
		keys = append(keys, &k)
	}
	return keys, nil
}

func (r *PostgresAPIKeyRepo) Revoke(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE api_keys SET revoked = true WHERE id = $1`, id)
	return err
}

func (r *PostgresAPIKeyRepo) UpdateLastUsed(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE api_keys SET last_used = NOW() WHERE id = $1`, id)
	return err
}
