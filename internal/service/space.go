package service

import (
	"context"
	"errors"
	"log/slog"
	"regexp"

	"nerion/internal/domain"
	"nerion/internal/entity"
	"nerion/pkg/apierrors"
)

var slugRe = regexp.MustCompile(`^[a-z0-9][a-z0-9\-]{1,62}[a-z0-9]$`)

type spaceService struct {
	spaceRepo  domain.SpaceRepository
	memberRepo domain.SpaceMemberRepository
	logger     *slog.Logger
}

func NewSpaceService(spaceRepo domain.SpaceRepository, memberRepo domain.SpaceMemberRepository, logger *slog.Logger) domain.SpaceService {
	return &spaceService{spaceRepo: spaceRepo, memberRepo: memberRepo, logger: logger}
}

func (s *spaceService) Create(ctx context.Context, userID int64, name, slug string) (*entity.Space, error) {
	if name == "" || slug == "" {
		return nil, apierrors.ErrBadRequest
	}
	if !slugRe.MatchString(slug) {
		return nil, apierrors.NewValidationError(map[string]string{
			"slug": "Только латинские буквы, цифры и дефис (3–64 символа)",
		})
	}

	space := &entity.Space{Name: name, Slug: slug, OwnerID: userID}
	if err := s.spaceRepo.Create(ctx, space); err != nil {
		return nil, err
	}

	if err := s.spaceRepo.CreateSchema(ctx, slug); err != nil {
		return nil, err
	}

	member := &entity.SpaceMember{
		SpaceID: space.ID,
		UserID:  userID,
		Role:    entity.SpaceMemberRoleAdmin,
	}
	if err := s.memberRepo.Add(ctx, member); err != nil {
		return nil, err
	}
	s.logger.Info("space created", "space", slug, "user_id", userID)
	return space, nil
}

func (s *spaceService) Get(ctx context.Context, userID int64, slug string) (*entity.Space, error) {
	space, err := s.spaceRepo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	if _, err := s.memberRepo.GetRole(ctx, space.ID, userID); err != nil {
		if errors.Is(err, apierrors.ErrForbidden) {
			return nil, apierrors.ErrForbidden
		}
		return nil, err
	}
	return space, nil
}

func (s *spaceService) List(ctx context.Context, userID int64) ([]*entity.Space, error) {
	return s.spaceRepo.ListByUserID(ctx, userID)
}

func (s *spaceService) Rename(ctx context.Context, userID int64, slug, newName string) error {
	if newName == "" {
		return apierrors.ErrBadRequest
	}
	space, err := s.spaceRepo.GetBySlug(ctx, slug)
	if err != nil {
		return err
	}
	role, err := s.memberRepo.GetRole(ctx, space.ID, userID)
	if err != nil {
		return apierrors.ErrForbidden
	}
	if role != entity.SpaceMemberRoleAdmin {
		return apierrors.ErrForbidden
	}
	if err := s.spaceRepo.UpdateName(ctx, space.ID, newName); err != nil {
		return err
	}
	s.logger.Info("space renamed", "space", slug, "new_name", newName, "user_id", userID)
	return nil
}

func (s *spaceService) Delete(ctx context.Context, userID int64, slug, confirmName string) error {
	space, err := s.spaceRepo.GetBySlug(ctx, slug)
	if err != nil {
		return err
	}
	role, err := s.memberRepo.GetRole(ctx, space.ID, userID)
	if err != nil || role != entity.SpaceMemberRoleAdmin {
		return apierrors.ErrForbidden
	}
	if confirmName != space.Name {
		return apierrors.NewValidationError(map[string]string{
			"confirm_name": "Название пространства не совпадает",
		})
	}
	if err := s.spaceRepo.DropSchema(ctx, slug); err != nil {
		return err
	}
	if err := s.spaceRepo.Delete(ctx, space.ID); err != nil {
		return err
	}
	s.logger.Info("space deleted", "space", slug, "user_id", userID)
	return nil
}
