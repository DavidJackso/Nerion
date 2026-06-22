package service

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"time"

	"nerion/internal/domain"
	"nerion/internal/entity"
	"nerion/pkg/apierrors"
)

var (
	emailRe   = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)
	urlRe     = regexp.MustCompile(`^https?://`)
	phoneRe   = regexp.MustCompile(`^[+\d][\d\s\-().]{4,}$`)
	datetimeLayouts = []string{
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02T15:04",
		"2006-01-02 15:04:05",
	}
)

type recordService struct {
	spaceRepo  domain.SpaceRepository
	memberRepo domain.SpaceMemberRepository
	tableRepo  domain.TableRepository
	fieldRepo  domain.FieldRepository
	recordRepo domain.RecordRepository
	logger     *slog.Logger
}

func NewRecordService(
	spaceRepo domain.SpaceRepository,
	memberRepo domain.SpaceMemberRepository,
	tableRepo domain.TableRepository,
	fieldRepo domain.FieldRepository,
	recordRepo domain.RecordRepository,
	logger *slog.Logger,
) domain.RecordService {
	return &recordService{
		spaceRepo:  spaceRepo,
		memberRepo: memberRepo,
		tableRepo:  tableRepo,
		fieldRepo:  fieldRepo,
		recordRepo: recordRepo,
		logger:     logger,
	}
}

func (s *recordService) resolveTable(ctx context.Context, spaceSlug, tableSlug string, userID int64) (*entity.Space, []*entity.FieldMeta, error) {
	space, err := s.spaceRepo.GetBySlug(ctx, spaceSlug)
	if err != nil {
		return nil, nil, err
	}
	if _, err := s.memberRepo.GetRole(ctx, space.ID, userID); err != nil {
		return nil, nil, apierrors.ErrForbidden
	}
	t, err := s.tableRepo.GetBySlug(ctx, space.ID, tableSlug)
	if err != nil {
		return nil, nil, err
	}
	fields, err := s.fieldRepo.ListByTable(ctx, t.ID)
	if err != nil {
		return nil, nil, err
	}
	return space, fields, nil
}

func (s *recordService) validateData(ctx context.Context, spaceSlug, tableSlug string, fields []*entity.FieldMeta, data map[string]any, excludeID *int64) error {
	fieldErrors := map[string]string{}
	for _, f := range fields {
		val, exists := data[f.Slug]
		empty := !exists || val == nil || val == ""
		if f.Required && empty {
			fieldErrors[f.Slug] = "Поле обязательно для заполнения"
			continue
		}
		if empty {
			continue
		}
		switch f.Type {
		case entity.FieldTypeNumber, entity.FieldTypeRelation:
			switch val.(type) {
			case float64, int, int64, float32:
				// ok
			default:
				fieldErrors[f.Slug] = "Ожидается числовое значение"
				continue
			}
		case entity.FieldTypeBoolean:
			if _, ok := val.(bool); !ok {
				fieldErrors[f.Slug] = "Ожидается булево значение"
				continue
			}
		case entity.FieldTypeEnum:
			strVal, ok := val.(string)
			if !ok {
				fieldErrors[f.Slug] = "Ожидается строковое значение"
				continue
			}
			valid := false
			for _, ev := range f.EnumValues {
				if ev == strVal {
					valid = true
					break
				}
			}
			if !valid {
				fieldErrors[f.Slug] = fmt.Sprintf("Допустимые значения: %s", strings.Join(f.EnumValues, ", "))
			}
		case entity.FieldTypeDate:
			strVal, ok := val.(string)
			if !ok {
				fieldErrors[f.Slug] = "Ожидается строка с датой"
				continue
			}
			if _, err := time.Parse("2006-01-02", strVal); err != nil {
				fieldErrors[f.Slug] = "Неверный формат даты (ожидается ГГГГ-ММ-ДД)"
			}
		case entity.FieldTypeDatetime:
			strVal, ok := val.(string)
			if !ok {
				fieldErrors[f.Slug] = "Ожидается строка с датой и временем"
				continue
			}
			valid := false
			for _, layout := range datetimeLayouts {
				if _, err := time.Parse(layout, strVal); err == nil {
					valid = true
					break
				}
			}
			if !valid {
				fieldErrors[f.Slug] = "Неверный формат даты и времени"
			}
		case entity.FieldTypeEmail:
			strVal, ok := val.(string)
			if !ok || !emailRe.MatchString(strVal) {
				fieldErrors[f.Slug] = "Неверный формат email"
			}
		case entity.FieldTypeURL:
			strVal, ok := val.(string)
			if !ok || !urlRe.MatchString(strVal) {
				fieldErrors[f.Slug] = "URL должен начинаться с http:// или https://"
			}
		case entity.FieldTypePhone:
			strVal, ok := val.(string)
			if !ok || !phoneRe.MatchString(strVal) {
				fieldErrors[f.Slug] = "Неверный формат номера телефона"
			}
		case entity.FieldTypeFile:
			if _, ok := val.(string); !ok {
				fieldErrors[f.Slug] = "Ожидается строка с ключом файла"
			}
		case entity.FieldTypeText, entity.FieldTypeLongtext:
			if _, ok := val.(string); !ok {
				fieldErrors[f.Slug] = "Ожидается строковое значение"
			}
		}
		if f.Unique {
			dup, err := s.recordRepo.CheckUnique(ctx, spaceSlug, tableSlug, f.Slug, val, excludeID)
			if err != nil {
				return err
			}
			if dup {
				fieldErrors[f.Slug] = "Значение должно быть уникальным"
			}
		}
	}
	if len(fieldErrors) > 0 {
		return apierrors.NewValidationError(fieldErrors)
	}
	return nil
}

func (s *recordService) List(ctx context.Context, spaceSlug, tableSlug string, userID int64, params entity.ListParams) ([]map[string]any, int64, error) {
	_, fields, err := s.resolveTable(ctx, spaceSlug, tableSlug, userID)
	if err != nil {
		return nil, 0, err
	}
	return s.recordRepo.List(ctx, spaceSlug, tableSlug, fields, params)
}

func (s *recordService) GetByID(ctx context.Context, spaceSlug, tableSlug string, userID, id int64) (map[string]any, error) {
	_, fields, err := s.resolveTable(ctx, spaceSlug, tableSlug, userID)
	if err != nil {
		return nil, err
	}
	return s.recordRepo.GetByID(ctx, spaceSlug, tableSlug, fields, id)
}

func (s *recordService) Create(ctx context.Context, spaceSlug, tableSlug string, userID int64, data map[string]any) (map[string]any, error) {
	_, fields, err := s.resolveTable(ctx, spaceSlug, tableSlug, userID)
	if err != nil {
		return nil, err
	}
	if err := s.validateData(ctx, spaceSlug, tableSlug, fields, data, nil); err != nil {
		return nil, err
	}
	rec, err := s.recordRepo.Create(ctx, spaceSlug, tableSlug, fields, data)
	if err != nil {
		return nil, err
	}
	s.logger.Info("record created", "space", spaceSlug, "table", tableSlug, "user_id", userID)
	return rec, nil
}

func (s *recordService) Update(ctx context.Context, spaceSlug, tableSlug string, userID, id int64, data map[string]any) (map[string]any, error) {
	_, fields, err := s.resolveTable(ctx, spaceSlug, tableSlug, userID)
	if err != nil {
		return nil, err
	}
	if err := s.validateData(ctx, spaceSlug, tableSlug, fields, data, &id); err != nil {
		return nil, err
	}
	return s.recordRepo.Update(ctx, spaceSlug, tableSlug, fields, id, data)
}

func (s *recordService) Delete(ctx context.Context, spaceSlug, tableSlug string, userID, id int64) error {
	_, _, err := s.resolveTable(ctx, spaceSlug, tableSlug, userID)
	if err != nil {
		return err
	}
	if err := s.recordRepo.Delete(ctx, spaceSlug, tableSlug, id); err != nil {
		return err
	}
	s.logger.Info("record deleted", "space", spaceSlug, "table", tableSlug, "id", id, "user_id", userID)
	return nil
}
