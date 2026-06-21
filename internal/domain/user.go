package domain

import (
	"context"

	"nerion/internal/entity"
)

type UserRepository interface {
	GetByID(ctx context.Context, id int64) (*entity.User, error)
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	Create(ctx context.Context, user *entity.User) error
	List(ctx context.Context, limit, offset int) ([]*entity.User, error)
	UpdatePassword(ctx context.Context, userID int64, passwordHash string) error
	SetEmailVerified(ctx context.Context, userID int64) error
	UpdateProfile(ctx context.Context, userID int64, name, email string) error
	Delete(ctx context.Context, userID int64) error
	IsLastAdminAnywhere(ctx context.Context, userID int64) (bool, error)
}

type UserService interface {
	GetUser(ctx context.Context, id int64) (*entity.User, error)
	CreateUser(ctx context.Context, name, email, password string) (*entity.User, error)
	ListUsers(ctx context.Context, limit, offset int) ([]*entity.User, error)
	Login(ctx context.Context, email, password string) (*entity.User, error)
	UpdateProfile(ctx context.Context, userID int64, name, email string) error
	DeleteAccount(ctx context.Context, userID int64) error
}
