package domain

import (
	"context"

	"go-template/internal/entity"
)

type UserRepository interface {
	GetByID(ctx context.Context, id int64) (*entity.User, error)
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	Create(ctx context.Context, user *entity.User) error
	List(ctx context.Context, limit, offset int) ([]*entity.User, error)
}

type UserService interface {
	GetUser(ctx context.Context, id int64) (*entity.User, error)
	CreateUser(ctx context.Context, name, email, password string) (*entity.User, error)
	ListUsers(ctx context.Context, limit, offset int) ([]*entity.User, error)
	Login(ctx context.Context, email, password string) (*entity.User, error)
}
