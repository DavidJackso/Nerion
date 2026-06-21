package service

import (
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"

	"nerion/internal/domain"
	"nerion/internal/entity"
	"nerion/pkg/apierrors"
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

func (s *userService) UpdateProfile(ctx context.Context, userID int64, name, email string) error {
	if name == "" {
		return apierrors.NewError(400, "validation_error", "Имя не может быть пустым")
	}
	return s.repo.UpdateProfile(ctx, userID, name, email)
}

func (s *userService) DeleteAccount(ctx context.Context, userID int64) error {
	last, err := s.repo.IsLastAdminAnywhere(ctx, userID)
	if err != nil {
		return err
	}
	if last {
		return apierrors.NewError(409, "last_admin", "Нельзя удалить аккаунт: вы единственный администратор одного из пространств")
	}
	return s.repo.Delete(ctx, userID)
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
