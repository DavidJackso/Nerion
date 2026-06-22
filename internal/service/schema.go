package service

import (
	"context"
	"log/slog"

	"nerion/internal/domain"
	"nerion/internal/entity"
	"nerion/internal/repository"
	"nerion/pkg/apierrors"
)

type schemaService struct {
	spaceRepo  domain.SpaceRepository
	memberRepo domain.SpaceMemberRepository
	tableRepo  domain.TableRepository
	fieldRepo  domain.FieldRepository
	ddl        *repository.DDLExecutor
	logger     *slog.Logger
}

func NewSchemaService(
	spaceRepo domain.SpaceRepository,
	memberRepo domain.SpaceMemberRepository,
	tableRepo domain.TableRepository,
	fieldRepo domain.FieldRepository,
	ddl *repository.DDLExecutor,
	logger *slog.Logger,
) domain.SchemaService {
	return &schemaService{
		spaceRepo:  spaceRepo,
		memberRepo: memberRepo,
		tableRepo:  tableRepo,
		fieldRepo:  fieldRepo,
		ddl:        ddl,
		logger:     logger,
	}
}

func (s *schemaService) requireMember(ctx context.Context, spaceSlug string, userID int64) (*entity.Space, error) {
	space, err := s.spaceRepo.GetBySlug(ctx, spaceSlug)
	if err != nil {
		return nil, err
	}
	if _, err := s.memberRepo.GetRole(ctx, space.ID, userID); err != nil {
		return nil, apierrors.ErrForbidden
	}
	return space, nil
}

func (s *schemaService) ListTables(ctx context.Context, spaceSlug string, userID int64) ([]*entity.TableMeta, error) {
	space, err := s.requireMember(ctx, spaceSlug, userID)
	if err != nil {
		return nil, err
	}
	return s.tableRepo.ListBySpace(ctx, space.ID)
}

func (s *schemaService) GetTable(ctx context.Context, spaceSlug, tableSlug string, userID int64) (*entity.TableMeta, error) {
	space, err := s.requireMember(ctx, spaceSlug, userID)
	if err != nil {
		return nil, err
	}
	t, err := s.tableRepo.GetBySlug(ctx, space.ID, tableSlug)
	if err != nil {
		return nil, err
	}
	fields, err := s.fieldRepo.ListByTable(ctx, t.ID)
	if err != nil {
		return nil, err
	}
	t.Fields = fields
	return t, nil
}

func (s *schemaService) CreateTable(ctx context.Context, spaceSlug, name, slug string, userID int64) (*entity.TableMeta, error) {
	if name == "" || slug == "" {
		return nil, apierrors.ErrBadRequest
	}
	if !slugRe.MatchString(slug) {
		return nil, apierrors.NewValidationError(map[string]string{
			"slug": "Только латинские буквы, цифры и дефис (3–64 символа)",
		})
	}

	space, err := s.requireMember(ctx, spaceSlug, userID)
	if err != nil {
		return nil, err
	}

	t := &entity.TableMeta{SpaceID: space.ID, Name: name, Slug: slug}
	if err := s.tableRepo.Create(ctx, t); err != nil {
		return nil, err
	}
	if err := s.ddl.CreateTable(ctx, spaceSlug, slug); err != nil {
		return nil, err
	}
	s.logger.Info("table created", "space", spaceSlug, "table", slug, "user_id", userID)
	return t, nil
}

func (s *schemaService) UpdateFields(ctx context.Context, spaceSlug, tableSlug string, userID int64, fields []*entity.FieldMeta) error {
	space, err := s.requireMember(ctx, spaceSlug, userID)
	if err != nil {
		return err
	}
	t, err := s.tableRepo.GetBySlug(ctx, space.ID, tableSlug)
	if err != nil {
		return err
	}

	existing, err := s.fieldRepo.ListByTable(ctx, t.ID)
	if err != nil {
		return err
	}

	// Map existing by slug
	existingMap := make(map[string]*entity.FieldMeta, len(existing))
	for _, f := range existing {
		existingMap[f.Slug] = f
	}
	newSlugs := make(map[string]bool, len(fields))
	for _, f := range fields {
		newSlugs[f.Slug] = true
	}

	var dropped, added int

	// Drop columns removed from the list
	for _, ef := range existing {
		if !newSlugs[ef.Slug] {
			if err := s.ddl.DropColumn(ctx, spaceSlug, tableSlug, ef.Slug); err != nil {
				return err
			}
			if err := s.fieldRepo.Delete(ctx, ef.ID); err != nil {
				return err
			}
			dropped++
		}
	}

	// Upsert new/changed fields
	for i, f := range fields {
		f.TableID = t.ID
		f.Position = i
		if _, exists := existingMap[f.Slug]; !exists {
			if err := s.ddl.AddColumn(ctx, spaceSlug, tableSlug, f); err != nil {
				return err
			}
			added++
		}
		if err := s.fieldRepo.Upsert(ctx, f); err != nil {
			return err
		}
	}
	s.logger.Info("fields updated", "space", spaceSlug, "table", tableSlug, "added", added, "dropped", dropped, "user_id", userID)
	return nil
}

func (s *schemaService) DeleteTable(ctx context.Context, spaceSlug, tableSlug string, userID int64) error {
	space, err := s.requireMember(ctx, spaceSlug, userID)
	if err != nil {
		return err
	}
	t, err := s.tableRepo.GetBySlug(ctx, space.ID, tableSlug)
	if err != nil {
		return err
	}
	if err := s.ddl.DropTable(ctx, spaceSlug, tableSlug); err != nil {
		return err
	}
	if err := s.tableRepo.Delete(ctx, t.ID); err != nil {
		return err
	}
	s.logger.Info("table deleted", "space", spaceSlug, "table", tableSlug, "user_id", userID)
	return nil
}
