package service

import (
	"context"
	"log/slog"
	"time"

	"nerion/internal/domain"
	"nerion/internal/entity"
	"nerion/pkg/apierrors"
)

type listService struct {
	spaceRepo  domain.SpaceRepository
	memberRepo domain.SpaceMemberRepository
	tableRepo  domain.TableRepository
	fieldRepo  domain.FieldRepository
	listRepo   domain.ListRepository
	logger     *slog.Logger
}

func NewListService(
	spaceRepo domain.SpaceRepository,
	memberRepo domain.SpaceMemberRepository,
	tableRepo domain.TableRepository,
	fieldRepo domain.FieldRepository,
	listRepo domain.ListRepository,
	logger *slog.Logger,
) domain.ListService {
	return &listService{
		spaceRepo:  spaceRepo,
		memberRepo: memberRepo,
		tableRepo:  tableRepo,
		fieldRepo:  fieldRepo,
		listRepo:   listRepo,
		logger:     logger,
	}
}

func (s *listService) requireAdmin(ctx context.Context, spaceSlug string, userID int64) (*entity.Space, error) {
	space, err := s.spaceRepo.GetBySlug(ctx, spaceSlug)
	if err != nil {
		return nil, err
	}
	role, err := s.memberRepo.GetRole(ctx, space.ID, userID)
	if err != nil {
		return nil, apierrors.ErrForbidden
	}
	if role != entity.SpaceMemberRoleAdmin {
		return nil, apierrors.ErrForbidden
	}
	return space, nil
}

func (s *listService) requireMember(ctx context.Context, spaceSlug string, userID int64) (*entity.Space, error) {
	space, err := s.spaceRepo.GetBySlug(ctx, spaceSlug)
	if err != nil {
		return nil, err
	}
	if _, err := s.memberRepo.GetRole(ctx, space.ID, userID); err != nil {
		return nil, apierrors.ErrForbidden
	}
	return space, nil
}

func (s *listService) Create(ctx context.Context, spaceSlug, tableSlug, listSlug string, userID int64, cfg entity.ListConfig, publish bool) (*entity.List, error) {
	if listSlug == "" {
		return nil, apierrors.NewValidationError(map[string]string{"slug": "Обязательное поле"})
	}
	if !slugRe.MatchString(listSlug) {
		return nil, apierrors.NewValidationError(map[string]string{"slug": "Только латинские буквы, цифры и дефис (3–64 символа)"})
	}

	space, err := s.requireAdmin(ctx, spaceSlug, userID)
	if err != nil {
		return nil, err
	}

	t, err := s.tableRepo.GetBySlug(ctx, space.ID, tableSlug)
	if err != nil {
		return nil, err
	}

	if cfg.RowLimit <= 0 {
		cfg.RowLimit = 100
	}
	if cfg.FieldConfig == nil {
		cfg.FieldConfig = []entity.ListFieldConfig{}
	}
	if cfg.FilterConfig == nil {
		cfg.FilterConfig = map[string]any{}
	}
	if cfg.SortConfig == nil {
		cfg.SortConfig = []entity.ListSortConfig{}
	}

	now := time.Now()
	l := &entity.List{
		SpaceID:       space.ID,
		Slug:          listSlug,
		SourceTableID: t.ID,
		TableSlug:     t.Slug,
		ListConfig:    cfg,
	}
	if publish {
		l.PublishedAt = &now
	}

	if err := s.listRepo.Create(ctx, l); err != nil {
		return nil, err
	}
	s.logger.Info("list created", "space", spaceSlug, "table", tableSlug, "list", listSlug, "user_id", userID)
	return l, nil
}

func (s *listService) List(ctx context.Context, spaceSlug string, userID int64) ([]*entity.List, error) {
	space, err := s.requireMember(ctx, spaceSlug, userID)
	if err != nil {
		return nil, err
	}
	return s.listRepo.ListBySpace(ctx, space.ID)
}

func (s *listService) Update(ctx context.Context, spaceSlug, listSlug string, userID int64, cfg *entity.ListConfig, publish *bool) (*entity.List, error) {
	space, err := s.requireAdmin(ctx, spaceSlug, userID)
	if err != nil {
		return nil, err
	}

	l, err := s.listRepo.GetBySlug(ctx, space.ID, listSlug)
	if err != nil {
		return nil, err
	}

	if cfg != nil {
		if cfg.FieldConfig != nil {
			l.FieldConfig = cfg.FieldConfig
		}
		if cfg.FilterConfig != nil {
			l.FilterConfig = cfg.FilterConfig
		}
		if cfg.SortConfig != nil {
			l.SortConfig = cfg.SortConfig
		}
		if cfg.RowLimit > 0 {
			l.RowLimit = cfg.RowLimit
		}
	}

	if publish != nil {
		now := time.Now()
		if *publish {
			l.PublishedAt = &now
			l.UnpublishedAt = nil
		} else {
			l.UnpublishedAt = &now
		}
	}

	if err := s.listRepo.Update(ctx, l); err != nil {
		return nil, err
	}
	return l, nil
}

func (s *listService) GetPublicData(ctx context.Context, spaceSlug, listSlug string) ([]map[string]any, error) {
	l, err := s.listRepo.GetPublic(ctx, spaceSlug, listSlug)
	if err != nil {
		return nil, err
	}

	t, err := s.tableRepo.GetBySlug(ctx, l.SpaceID, l.TableSlug)
	if err != nil {
		return nil, err
	}

	fields, err := s.fieldRepo.ListByTable(ctx, t.ID)
	if err != nil {
		return nil, err
	}

	return s.listRepo.QueryPublic(ctx, spaceSlug, l, fields)
}
