package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"go-template/internal/domain"
	"go-template/internal/entity"
	"go-template/pkg/apierrors"
)

type userRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) domain.UserRepository {
	return &userRepository{pool: pool}
}

func (r *userRepository) GetByID(ctx context.Context, id int64) (*entity.User, error) {
	u := &entity.User{}
	err := r.pool.QueryRow(ctx,
		"SELECT id, name, email, role, password_hash FROM users WHERE id = $1", id,
	).Scan(&u.ID, &u.Name, &u.Email, &u.Role, &u.PasswordHash)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apierrors.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	u := &entity.User{}
	err := r.pool.QueryRow(ctx,
		"SELECT id, name, email, role, password_hash FROM users WHERE email = $1", email,
	).Scan(&u.ID, &u.Name, &u.Email, &u.Role, &u.PasswordHash)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apierrors.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *userRepository) Create(ctx context.Context, user *entity.User) error {
	err := r.pool.QueryRow(ctx,
		"INSERT INTO users (name, email, role, password_hash) VALUES ($1, $2, $3, $4) RETURNING id",
		user.Name, user.Email, user.Role, user.PasswordHash,
	).Scan(&user.ID)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return apierrors.ErrConflict
	}
	return err
}

func (r *userRepository) List(ctx context.Context, limit, offset int) ([]*entity.User, error) {
	rows, err := r.pool.Query(ctx,
		"SELECT id, name, email, role FROM users ORDER BY id LIMIT $1 OFFSET $2",
		limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]*entity.User, 0, limit)
	for rows.Next() {
		u := &entity.User{}
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Role); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}
