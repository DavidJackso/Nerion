package domain

import (
	"context"

	"nerion/internal/entity"
)

type ListRepository interface {
	Create(ctx context.Context, l *entity.List) error
	GetBySlug(ctx context.Context, spaceID int64, slug string) (*entity.List, error)
	ListBySpace(ctx context.Context, spaceID int64) ([]*entity.List, error)
	Update(ctx context.Context, l *entity.List) error
	GetPublic(ctx context.Context, spaceSlug, listSlug string) (*entity.List, error)
	QueryPublic(ctx context.Context, spaceSlug string, l *entity.List, fields []*entity.FieldMeta) ([]map[string]any, error)
}

type ListService interface {
	Create(ctx context.Context, spaceSlug, tableSlug, listSlug string, userID int64, cfg entity.ListConfig, publish bool) (*entity.List, error)
	List(ctx context.Context, spaceSlug string, userID int64) ([]*entity.List, error)
	Update(ctx context.Context, spaceSlug, listSlug string, userID int64, cfg *entity.ListConfig, publish *bool) (*entity.List, error)
	GetPublicData(ctx context.Context, spaceSlug, listSlug string) ([]map[string]any, error)
}
