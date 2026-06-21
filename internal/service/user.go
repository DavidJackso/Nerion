package service

import (
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"

	"go-template/internal/domain"
	"go-template/internal/entity"
	"go-template/pkg/apierrors"
)

type userService struct {
	repo domain.UserRepository
}

func NewUserService(repo domain.UserRepository) domain.UserService {
	return &userService{repo: repo}
}

func (s *userService) GetUser(ctx context.Context, id int64) (*entity.User, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *userService) CreateUser(ctx context.Context, name, email, password string) (*entity.User, error) {
	if name == "" || email == "" || len(password) < 8 {
		return nil, apierrors.ErrBadRequest
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, apierrors.ErrInternal
	}
	user := &entity.User{
		Name:         name,
		Email:        email,
		Role:         entity.RoleUser,
		PasswordHash: string(hash),
	}
	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *userService) ListUsers(ctx context.Context, limit, offset int) ([]*entity.User, error) {
	return s.repo.List(ctx, limit, offset)
}

func (s *userService) Login(ctx context.Context, email, password string) (*entity.User, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, apierrors.ErrNotFound) {
			return nil, apierrors.ErrUnauthorized
		}
		return nil, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, apierrors.ErrUnauthorized
	}
	user.PasswordHash = ""
	return user, nil
}
