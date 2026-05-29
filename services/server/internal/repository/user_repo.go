package repository

import (
	"context"
	"github.com/tunnelkit/services/server/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/google/uuid"
)

type PostgresUserRepo struct {
	pool *pgxpool.Pool
}

func NewPostgresUserRepo(pool *pgxpool.Pool) *PostgresUserRepo {
	return &PostgresUserRepo{pool: pool}
}

func (r *PostgresUserRepo) Create(ctx context.Context, user *domain.User) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO users (id, email, password, role, created_at) VALUES ($1, $2, $3, $4, $5)`,
		user.ID, user.Email, user.Password, user.Role, user.CreatedAt,
	)
	return err
}

func (r *PostgresUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	var user domain.User
	err := r.pool.QueryRow(ctx,
		`SELECT id, email, password, role, created_at FROM users WHERE id = $1`, id,
	).Scan(&user.ID, &user.Email, &user.Password, &user.Role, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *PostgresUserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	err := r.pool.QueryRow(ctx,
		`SELECT id, email, password, role, created_at FROM users WHERE email = $1`, email,
	).Scan(&user.ID, &user.Email, &user.Password, &user.Role, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *PostgresUserRepo) List(ctx context.Context, offset, limit int) ([]*domain.User, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, email, password, role, created_at FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2`,
		limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(&u.ID, &u.Email, &u.Password, &u.Role, &u.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, &u)
	}
	return users, nil
}

func (r *PostgresUserRepo) Update(ctx context.Context, user *domain.User) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE users SET email = $2, role = $3 WHERE id = $1`,
		user.ID, user.Email, user.Role,
	)
	return err
}

func (r *PostgresUserRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, id)
	return err
}
