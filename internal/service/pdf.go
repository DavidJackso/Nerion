package service

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"nerion/internal/domain"
	"nerion/internal/entity"
	"nerion/pkg/apierrors"
)

const pdfPresignTTL = 24 * time.Hour

var placeholderRe = regexp.MustCompile(`\{\{(\w+)\}\}`)

type pdfService struct {
	spaceRepo  domain.SpaceRepository
	memberRepo domain.SpaceMemberRepository
	tableRepo  domain.TableRepository
	fieldRepo  domain.FieldRepository
	recordRepo domain.RecordRepository
	pdfRepo    domain.PDFRepository
	storage    domain.StorageAdapter
	uploadDir  string
}

func NewPDFService(
	spaceRepo domain.SpaceRepository,
	memberRepo domain.SpaceMemberRepository,
	tableRepo domain.TableRepository,
	fieldRepo domain.FieldRepository,
	recordRepo domain.RecordRepository,
	pdfRepo domain.PDFRepository,
	storage domain.StorageAdapter,
	uploadDir string,
) domain.PDFService {
	return &pdfService{
		spaceRepo:  spaceRepo,
		memberRepo: memberRepo,
		tableRepo:  tableRepo,
		fieldRepo:  fieldRepo,
		recordRepo: recordRepo,
		pdfRepo:    pdfRepo,
		storage:    storage,
		uploadDir:  uploadDir,
	}
}

func (s *pdfService) requireAdmin(ctx context.Context, spaceSlug string, userID int64) (*entity.Space, error) {
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

func (s *pdfService) requireMember(ctx context.Context, spaceSlug string, userID int64) (*entity.Space, error) {
	space, err := s.spaceRepo.GetBySlug(ctx, spaceSlug)
	if err != nil {
		return nil, err
	}
	if _, err := s.memberRepo.GetRole(ctx, space.ID, userID); err != nil {
		return nil, apierrors.ErrForbidden
	}
	return space, nil
}

// extractPlaceholders reads DOCX (ZIP+XML) and finds all {{name}} patterns.
func extractPlaceholders(data []byte) ([]string, error) {
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("не удалось открыть DOCX: %w", err)
	}
	for _, f := range r.File {
		if f.Name != "word/document.xml" {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return nil, err
		}
		content, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			return nil, err
		}
		matches := placeholderRe.FindAllSubmatch(content, -1)
		seen := map[string]bool{}
		var result []string
		for _, m := range matches {
			name := string(m[1])
			if !seen[name] {
				seen[name] = true
				result = append(result, name)
			}
		}
		return result, nil
	}
	return []string{}, nil
}

// uploadToStorage saves data via StorageAdapter and returns the storage key.
func (s *pdfService) uploadToStorage(ctx context.Context, key string, data []byte, contentType string) (string, error) {
	return key, s.storage.Upload(ctx, key, bytes.NewReader(data), int64(len(data)), contentType)
}

func (s *pdfService) UploadTemplate(ctx context.Context, spaceSlug, name string, userID int64, data []byte, filename string) (*entity.PDFTemplate, error) {
	space, err := s.requireAdmin(ctx, spaceSlug, userID)
	if err != nil {
		return nil, err
	}

	placeholders, err := extractPlaceholders(data)
	if err != nil {
		return nil, apierrors.NewError(400, "invalid_file", err.Error())
	}

	key := fmt.Sprintf("pdf-templates/%s/%d_%s", spaceSlug, time.Now().UnixMilli(), filepath.Base(filename))
	storagePath, err := s.uploadToStorage(ctx, key, data, "application/vnd.openxmlformats-officedocument.wordprocessingml.document")
	if err != nil {
		return nil, err
	}

	t := &entity.PDFTemplate{
		SpaceID:      space.ID,
		Name:         name,
		StoragePath:  storagePath,
		Placeholders: placeholders,
		Status:       "needs_mapping",
	}
	if err := s.pdfRepo.CreateTemplate(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}

func (s *pdfService) ListTemplates(ctx context.Context, spaceSlug string, userID int64) ([]*entity.PDFTemplate, error) {
	space, err := s.requireMember(ctx, spaceSlug, userID)
	if err != nil {
		return nil, err
	}
	return s.pdfRepo.ListTemplates(ctx, space.ID)
}

func (s *pdfService) SaveMapping(ctx context.Context, spaceSlug string, templateID, userID int64, mappings []*entity.PDFMapping) error {
	space, err := s.requireAdmin(ctx, spaceSlug, userID)
	if err != nil {
		return err
	}
	t, err := s.pdfRepo.GetTemplate(ctx, templateID, space.ID)
	if err != nil {
		return err
	}

	for _, m := range mappings {
		m.TemplateID = templateID
	}
	if err := s.pdfRepo.SaveMappings(ctx, templateID, mappings); err != nil {
		return err
	}

	// Update status: ready if all placeholders mapped
	mappedSet := map[string]bool{}
	for _, m := range mappings {
		mappedSet[m.Placeholder] = true
	}
	status := "ready"
	for _, p := range t.Placeholders {
		if !mappedSet[p] {
			status = "needs_mapping"
			break
		}
	}
	return s.pdfRepo.UpdateTemplateStatus(ctx, templateID, status)
}

// applyMappings substitutes placeholders in src using record data + mappings.
func applyMappings(src string, mappings []*entity.PDFMapping, record map[string]any, fields []*entity.FieldMeta) string {
	fieldByID := map[int64]*entity.FieldMeta{}
	for _, f := range fields {
		fieldByID[f.ID] = f
	}
	for _, m := range mappings {
		var val string
		if m.SourceFieldID != nil {
			if f, ok := fieldByID[*m.SourceFieldID]; ok {
				if v := record[f.Slug]; v != nil {
					val = fmt.Sprintf("%v", v)
				}
			}
		} else if m.Expression != nil {
			val = *m.Expression
		}
		src = strings.ReplaceAll(src, "{{"+m.Placeholder+"}}", val)
	}
	return src
}

func (s *pdfService) Preview(ctx context.Context, spaceSlug string, templateID, recordID int64, tableSlug string, userID int64) (map[string]any, error) {
	space, err := s.requireMember(ctx, spaceSlug, userID)
	if err != nil {
		return nil, err
	}
	t, err := s.pdfRepo.GetTemplate(ctx, templateID, space.ID)
	if err != nil {
		return nil, err
	}
	mappings, err := s.pdfRepo.GetMappings(ctx, templateID)
	if err != nil {
		return nil, err
	}

	table, err := s.tableRepo.GetBySlug(ctx, space.ID, tableSlug)
	if err != nil {
		return nil, err
	}
	fields, err := s.fieldRepo.ListByTable(ctx, table.ID)
	if err != nil {
		return nil, err
	}
	record, err := s.recordRepo.GetByID(ctx, spaceSlug, tableSlug, fields, recordID)
	if err != nil {
		return nil, err
	}

	// Build substitution preview
	subs := map[string]string{}
	fieldByID := map[int64]*entity.FieldMeta{}
	for _, f := range fields {
		fieldByID[f.ID] = f
	}
	for _, m := range mappings {
		var val string
		if m.SourceFieldID != nil {
			if f, ok := fieldByID[*m.SourceFieldID]; ok {
				if v := record[f.Slug]; v != nil {
					val = fmt.Sprintf("%v", v)
				}
			}
		} else if m.Expression != nil {
			val = *m.Expression
		}
		subs[m.Placeholder] = val
	}

	return map[string]any{
		"template_id":   t.ID,
		"template_name": t.Name,
		"record_id":     recordID,
		"substitutions": subs,
	}, nil
}

func (s *pdfService) Generate(ctx context.Context, spaceSlug string, templateID, userID int64, tableSlug string, recordIDs []int64) (*entity.PDFJob, error) {
	space, err := s.requireMember(ctx, spaceSlug, userID)
	if err != nil {
		return nil, err
	}
	t, err := s.pdfRepo.GetTemplate(ctx, templateID, space.ID)
	if err != nil {
		return nil, err
	}
	if t.Status != "ready" {
		return nil, apierrors.NewError(400, "template_not_ready", "Шаблон ещё не настроен: заполните маппинг полей")
	}

	total := len(recordIDs)
	job := &entity.PDFJob{
		SpaceID:      space.ID,
		TemplateID:   templateID,
		Status:       "processing",
		TotalRecords: &total,
		Processed:    0,
		CreatedBy:    userID,
	}
	if err := s.pdfRepo.CreateJob(ctx, job); err != nil {
		return nil, err
	}

	// Stub: mark done immediately.
	// Real impl: call Python/WeasyPrint microservice, stream result here, upload to S3.
	now := time.Now()
	job.Status = "done"
	job.Processed = total
	job.CompletedAt = &now

	// Upload stub output to storage so the archive has a real key.
	stubContent := []byte(fmt.Sprintf("PDF stub for job %s (%d records)", job.ID, total))
	outputKey := fmt.Sprintf("pdf-jobs/%s/%s/output.pdf", spaceSlug, job.ID)
	if err := s.storage.Upload(ctx, outputKey, bytes.NewReader(stubContent), int64(len(stubContent)), "application/pdf"); err == nil {
		job.StoragePath = &outputKey
	}

	_ = s.pdfRepo.UpdateJob(ctx, job)
	return job, nil
}

func (s *pdfService) GetJob(ctx context.Context, spaceSlug, jobID string, userID int64) (*entity.PDFJob, error) {
	space, err := s.requireMember(ctx, spaceSlug, userID)
	if err != nil {
		return nil, err
	}
	return s.pdfRepo.GetJob(ctx, jobID, space.ID)
}

func (s *pdfService) ListArchive(ctx context.Context, spaceSlug string, userID int64) ([]*entity.PDFJob, error) {
	space, err := s.requireMember(ctx, spaceSlug, userID)
	if err != nil {
		return nil, err
	}
	return s.pdfRepo.ListJobs(ctx, space.ID)
}
