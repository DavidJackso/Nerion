package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"nerion/internal/domain"
	"nerion/internal/entity"
	"nerion/pkg/apierrors"
)

type sessionRepository struct {
	pool *pgxpool.Pool
}

func NewSessionRepository(pool *pgxpool.Pool) domain.SessionRepository {
	return &sessionRepository{pool: pool}
}

func (r *sessionRepository) Create(ctx context.Context, s *entity.Session) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO sessions (id, user_id, token_hash, created_at, expires_at)
		 VALUES ($1, $2, $3, $4, $5)`,
		s.ID, s.UserID, s.TokenHash, s.CreatedAt, s.ExpiresAt,
	)
	return err
}

func (r *sessionRepository) GetByTokenHash(ctx context.Context, hash string) (*entity.Session, error) {
	s := &entity.Session{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, token_hash, created_at, expires_at, revoked_at
		 FROM sessions WHERE token_hash = $1`,
		hash,
	).Scan(&s.ID, &s.UserID, &s.TokenHash, &s.CreatedAt, &s.ExpiresAt, &s.RevokedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apierrors.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (r *sessionRepository) RevokeByTokenHash(ctx context.Context, hash string) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE sessions SET revoked_at = $1 WHERE token_hash = $2 AND revoked_at IS NULL`,
		time.Now(), hash,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return apierrors.ErrNotFound
	}
	return nil
}

func (r *sessionRepository) RevokeAllByUserID(ctx context.Context, userID int64) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE sessions SET revoked_at = $1 WHERE user_id = $2 AND revoked_at IS NULL`,
		time.Now(), userID,
	)
	return err
}
