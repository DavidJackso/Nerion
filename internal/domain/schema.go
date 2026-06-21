package domain

import (
	"context"

	"nerion/internal/entity"
)

type TableRepository interface {
	Create(ctx context.Context, t *entity.TableMeta) error
	GetBySlug(ctx context.Context, spaceID int64, slug string) (*entity.TableMeta, error)
	ListBySpace(ctx context.Context, spaceID int64) ([]*entity.TableMeta, error)
	Delete(ctx context.Context, id int64) error
}

type FieldRepository interface {
	ListByTable(ctx context.Context, tableID int64) ([]*entity.FieldMeta, error)
	Upsert(ctx context.Context, f *entity.FieldMeta) error
	Delete(ctx context.Context, id int64) error
}

type SchemaService interface {
	ListTables(ctx context.Context, spaceSlug string, userID int64) ([]*entity.TableMeta, error)
	GetTable(ctx context.Context, spaceSlug, tableSlug string, userID int64) (*entity.TableMeta, error)
	CreateTable(ctx context.Context, spaceSlug, name, slug string, userID int64) (*entity.TableMeta, error)
	UpdateFields(ctx context.Context, spaceSlug, tableSlug string, userID int64, fields []*entity.FieldMeta) error
	DeleteTable(ctx context.Context, spaceSlug, tableSlug string, userID int64) error
}
