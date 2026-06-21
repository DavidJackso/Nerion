package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"nerion/internal/domain"
	"nerion/pkg/apierrors"
)

// --- Email Verification ---

type emailVerificationRepository struct {
	pool *pgxpool.Pool
}

func NewEmailVerificationRepository(pool *pgxpool.Pool) domain.EmailVerificationRepository {
	return &emailVerificationRepository{pool: pool}
}

func (r *emailVerificationRepository) Create(ctx context.Context, userID int64, tokenHash string, expiresAt time.Time) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO email_verifications (user_id, token_hash, expires_at) VALUES ($1, $2, $3)`,
		userID, tokenHash, expiresAt,
	)
	return err
}

func (r *emailVerificationRepository) GetByTokenHash(ctx context.Context, hash string) (int64, time.Time, *time.Time, error) {
	var userID int64
	var expiresAt time.Time
	var usedAt *time.Time
	err := r.pool.QueryRow(ctx,
		`SELECT user_id, expires_at, used_at FROM email_verifications WHERE token_hash = $1`,
		hash,
	).Scan(&userID, &expiresAt, &usedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, time.Time{}, nil, apierrors.ErrNotFound
	}
	if err != nil {
		return 0, time.Time{}, nil, err
	}
	return userID, expiresAt, usedAt, nil
}

func (r *emailVerificationRepository) MarkUsed(ctx context.Context, tokenHash string) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE email_verifications SET used_at = $1 WHERE token_hash = $2 AND used_at IS NULL`,
		time.Now(), tokenHash,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return apierrors.ErrNotFound
	}
	return nil
}

// --- Password Reset ---

type passwordResetRepository struct {
	pool *pgxpool.Pool
}

func NewPasswordResetRepository(pool *pgxpool.Pool) domain.PasswordResetRepository {
	return &passwordResetRepository{pool: pool}
}

func (r *passwordResetRepository) Create(ctx context.Context, userID int64, tokenHash string, expiresAt time.Time) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO password_resets (user_id, token_hash, expires_at) VALUES ($1, $2, $3)`,
		userID, tokenHash, expiresAt,
	)
	return err
}

func (r *passwordResetRepository) GetByTokenHash(ctx context.Context, hash string) (int64, time.Time, *time.Time, error) {
	var userID int64
	var expiresAt time.Time
	var usedAt *time.Time
	err := r.pool.QueryRow(ctx,
		`SELECT user_id, expires_at, used_at FROM password_resets WHERE token_hash = $1`,
		hash,
	).Scan(&userID, &expiresAt, &usedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, time.Time{}, nil, apierrors.ErrNotFound
	}
	if err != nil {
		return 0, time.Time{}, nil, err
	}
	return userID, expiresAt, usedAt, nil
}

func (r *passwordResetRepository) MarkUsed(ctx context.Context, tokenHash string) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE password_resets SET used_at = $1 WHERE token_hash = $2 AND used_at IS NULL`,
		time.Now(), tokenHash,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return apierrors.ErrNotFound
	}
	return nil
}
