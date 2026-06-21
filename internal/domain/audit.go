package domain

import (
	"context"

	"nerion/internal/entity"
)

type AuditRepository interface {
	Log(ctx context.Context, e *entity.AuditEntry) error
	List(ctx context.Context, spaceID int64, limit, offset int) ([]*entity.AuditEntry, error)
	CountByAction(ctx context.Context, spaceID int64, action string) (map[string]int64, error)
}
