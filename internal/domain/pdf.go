package domain

import (
	"context"

	"nerion/internal/entity"
)

type PDFRepository interface {
	CreateTemplate(ctx context.Context, t *entity.PDFTemplate) error
	ListTemplates(ctx context.Context, spaceID int64) ([]*entity.PDFTemplate, error)
	GetTemplate(ctx context.Context, id, spaceID int64) (*entity.PDFTemplate, error)
	UpdateTemplateStatus(ctx context.Context, id int64, status string) error

	SaveMappings(ctx context.Context, templateID int64, mappings []*entity.PDFMapping) error
	GetMappings(ctx context.Context, templateID int64) ([]*entity.PDFMapping, error)

	CreateJob(ctx context.Context, j *entity.PDFJob) error
	GetJob(ctx context.Context, jobID string, spaceID int64) (*entity.PDFJob, error)
	UpdateJob(ctx context.Context, j *entity.PDFJob) error
	ListJobs(ctx context.Context, spaceID int64) ([]*entity.PDFJob, error)
}

type PDFService interface {
	UploadTemplate(ctx context.Context, spaceSlug, name string, userID int64, data []byte, filename string) (*entity.PDFTemplate, error)
	ListTemplates(ctx context.Context, spaceSlug string, userID int64) ([]*entity.PDFTemplate, error)
	SaveMapping(ctx context.Context, spaceSlug string, templateID, userID int64, mappings []*entity.PDFMapping) error
	Preview(ctx context.Context, spaceSlug string, templateID, recordID int64, tableSlug string, userID int64) (map[string]any, error)
	Generate(ctx context.Context, spaceSlug string, templateID, userID int64, tableSlug string, recordIDs []int64) (*entity.PDFJob, error)
	GetJob(ctx context.Context, spaceSlug, jobID string, userID int64) (*entity.PDFJob, error)
	ListArchive(ctx context.Context, spaceSlug string, userID int64) ([]*entity.PDFJob, error)
}
