package domain

import (
	"context"

	"nerion/internal/entity"
)

type SpaceRepository interface {
	Create(ctx context.Context, space *entity.Space) error
	GetBySlug(ctx context.Context, slug string) (*entity.Space, error)
	ListByUserID(ctx context.Context, userID int64) ([]*entity.Space, error)
	UpdateName(ctx context.Context, id int64, name string) error
	Delete(ctx context.Context, id int64) error
	TableCount(ctx context.Context, spaceID int64) (int, error)
	CreateSchema(ctx context.Context, slug string) error
	DropSchema(ctx context.Context, slug string) error
}

type SpaceMemberRepository interface {
	Add(ctx context.Context, m *entity.SpaceMember) error
	GetRole(ctx context.Context, spaceID, userID int64) (entity.SpaceMemberRole, error)
	List(ctx context.Context, spaceID int64) ([]*entity.SpaceMember, error)
	UpdateRole(ctx context.Context, spaceID, userID int64, role entity.SpaceMemberRole) error
	Remove(ctx context.Context, spaceID, userID int64) error
	AdminCount(ctx context.Context, spaceID int64) (int, error)
}

type SpaceService interface {
	Create(ctx context.Context, userID int64, name, slug string) (*entity.Space, error)
	Get(ctx context.Context, userID int64, slug string) (*entity.Space, error)
	List(ctx context.Context, userID int64) ([]*entity.Space, error)
	Rename(ctx context.Context, userID int64, slug, newName string) error
	Delete(ctx context.Context, userID int64, slug, confirmName string) error
}

type SpaceMemberService interface {
	GetRole(ctx context.Context, spaceID, userID int64) (entity.SpaceMemberRole, error)
	ListMembers(ctx context.Context, spaceID int64) ([]*entity.SpaceMember, error)
	Invite(ctx context.Context, spaceID, inviterID int64, email string) error
	ChangeRole(ctx context.Context, spaceID, requesterID, targetUserID int64, role entity.SpaceMemberRole) error
	RemoveMember(ctx context.Context, spaceID, requesterID, targetUserID int64) error
}
