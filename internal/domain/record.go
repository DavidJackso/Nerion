package domain

import (
	"context"

	"nerion/internal/entity"
)

type RecordRepository interface {
	List(ctx context.Context, spaceSlug, tableSlug string, fields []*entity.FieldMeta, params entity.ListParams) ([]map[string]any, int64, error)
	GetByID(ctx context.Context, spaceSlug, tableSlug string, fields []*entity.FieldMeta, id int64) (map[string]any, error)
	Create(ctx context.Context, spaceSlug, tableSlug string, fields []*entity.FieldMeta, data map[string]any) (map[string]any, error)
	Update(ctx context.Context, spaceSlug, tableSlug string, fields []*entity.FieldMeta, id int64, data map[string]any) (map[string]any, error)
	Delete(ctx context.Context, spaceSlug, tableSlug string, id int64) error
	CheckUnique(ctx context.Context, spaceSlug, tableSlug, colSlug string, val any, excludeID *int64) (bool, error)
}

type RecordService interface {
	List(ctx context.Context, spaceSlug, tableSlug string, userID int64, params entity.ListParams) ([]map[string]any, int64, error)
	GetByID(ctx context.Context, spaceSlug, tableSlug string, userID, id int64) (map[string]any, error)
	Create(ctx context.Context, spaceSlug, tableSlug string, userID int64, data map[string]any) (map[string]any, error)
	Update(ctx context.Context, spaceSlug, tableSlug string, userID, id int64, data map[string]any) (map[string]any, error)
	Delete(ctx context.Context, spaceSlug, tableSlug string, userID, id int64) error
}
