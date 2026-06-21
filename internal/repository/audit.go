package repository

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"nerion/internal/domain"
	"nerion/internal/entity"
	"nerion/pkg/apierrors"
)

type auditRepository struct {
	pool *pgxpool.Pool
}

func NewAuditRepository(pool *pgxpool.Pool) domain.AuditRepository {
	return &auditRepository{pool: pool}
}

func (r *auditRepository) Log(ctx context.Context, e *entity.AuditEntry) error {
	var metaJSON []byte
	if e.Meta != nil {
		var err error
		metaJSON, err = json.Marshal(e.Meta)
		if err != nil {
			return err
		}
	}
	_, err := r.pool.Exec(ctx, `
		INSERT INTO audit_log (space_id, user_id, action, entity_type, entity_id, meta)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, e.SpaceID, e.UserID, e.Action, e.EntityType, e.EntityID, metaJSON)
	return err
}

func (r *auditRepository) List(ctx context.Context, spaceID int64, limit, offset int) ([]*entity.AuditEntry, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, space_id, user_id, action, entity_type, entity_id, meta, created_at
		FROM audit_log
		WHERE space_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, spaceID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []*entity.AuditEntry
	for rows.Next() {
		e := &entity.AuditEntry{}
		var metaJSON []byte
		if err := rows.Scan(&e.ID, &e.SpaceID, &e.UserID, &e.Action, &e.EntityType, &e.EntityID, &metaJSON, &e.CreatedAt); err != nil {
			return nil, err
		}
		if metaJSON != nil {
			_ = json.Unmarshal(metaJSON, &e.Meta)
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

// CountByAction returns per-entity_id counts for a given action within a space.
func (r *auditRepository) CountByAction(ctx context.Context, spaceID int64, action string) (map[string]int64, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT entity_id, COUNT(*) FROM audit_log
		WHERE space_id = $1 AND action = $2 AND entity_id IS NOT NULL
		GROUP BY entity_id
	`, spaceID, action)
	if errors.Is(err, pgx.ErrNoRows) {
		return map[string]int64{}, nil
	}
	if err != nil {
		return nil, apierrors.ErrInternal
	}
	defer rows.Close()

	result := map[string]int64{}
	for rows.Next() {
		var key string
		var count int64
		if err := rows.Scan(&key, &count); err != nil {
			return nil, err
		}
		result[key] = count
	}
	return result, rows.Err()
}
