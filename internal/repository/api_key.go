package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"nerion/internal/domain"
	"nerion/internal/entity"
	"nerion/pkg/apierrors"
)

type apiKeyRepository struct {
	pool *pgxpool.Pool
}

func NewAPIKeyRepository(pool *pgxpool.Pool) domain.APIKeyRepository {
	return &apiKeyRepository{pool: pool}
}

func (r *apiKeyRepository) Create(ctx context.Context, key *entity.APIKey, keyHash string) error {
	return r.pool.QueryRow(ctx, `
		INSERT INTO api_keys (space_id, name, key_hash, key_prefix, scope)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at`,
		key.SpaceID, key.Name, keyHash, key.KeyPrefix, key.Scope,
	).Scan(&key.ID, &key.CreatedAt)
}

func (r *apiKeyRepository) ListBySpace(ctx context.Context, spaceID int64) ([]*entity.APIKey, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, space_id, name, key_prefix, scope, created_at, last_used_at, revoked_at
		FROM api_keys WHERE space_id = $1 AND revoked_at IS NULL ORDER BY created_at DESC`, spaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*entity.APIKey
	for rows.Next() {
		k := &entity.APIKey{}
		if err := rows.Scan(&k.ID, &k.SpaceID, &k.Name, &k.KeyPrefix, &k.Scope,
			&k.CreatedAt, &k.LastUsedAt, &k.RevokedAt); err != nil {
			return nil, err
		}
		list = append(list, k)
	}
	return list, rows.Err()
}

func (r *apiKeyRepository) Revoke(ctx context.Context, id, spaceID int64) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE api_keys SET revoked_at = now() WHERE id = $1 AND space_id = $2 AND revoked_at IS NULL`,
		id, spaceID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return apierrors.ErrNotFound
	}
	return nil
}

func (r *apiKeyRepository) FindByHash(ctx context.Context, hash string) (*entity.APIKey, error) {
	k := &entity.APIKey{}
	err := r.pool.QueryRow(ctx, `
		SELECT id, space_id, name, key_prefix, scope, created_at, last_used_at, revoked_at
		FROM api_keys WHERE key_hash = $1 AND revoked_at IS NULL`, hash,
	).Scan(&k.ID, &k.SpaceID, &k.Name, &k.KeyPrefix, &k.Scope,
		&k.CreatedAt, &k.LastUsedAt, &k.RevokedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apierrors.ErrUnauthorized
	}
	return k, err
}

func (r *apiKeyRepository) UpdateLastUsed(ctx context.Context, id int64) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE api_keys SET last_used_at = now() WHERE id = $1`, id)
	return err
}
