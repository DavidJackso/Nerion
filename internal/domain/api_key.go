package domain

import (
	"context"

	"nerion/internal/entity"
)

type APIKeyRepository interface {
	Create(ctx context.Context, key *entity.APIKey, keyHash string) error
	ListBySpace(ctx context.Context, spaceID int64) ([]*entity.APIKey, error)
	Revoke(ctx context.Context, id, spaceID int64) error
	// FindByHash and UpdateLastUsed are used by the auto-REST API key middleware (section 7).
	FindByHash(ctx context.Context, hash string) (*entity.APIKey, error)
	UpdateLastUsed(ctx context.Context, id int64) error
}

type APIKeyService interface {
	Create(ctx context.Context, spaceSlug, name, scope string, userID int64) (*entity.APIKey, string, error)
	List(ctx context.Context, spaceSlug string, userID int64) ([]*entity.APIKey, error)
	Revoke(ctx context.Context, spaceSlug string, keyID, userID int64) error
}
