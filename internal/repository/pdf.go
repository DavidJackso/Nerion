package repository

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"nerion/internal/domain"
	"nerion/internal/entity"
	"nerion/pkg/apierrors"
)

type pdfRepository struct {
	pool *pgxpool.Pool
}

func NewPDFRepository(pool *pgxpool.Pool) domain.PDFRepository {
	return &pdfRepository{pool: pool}
}

func scanTemplate(row pgx.Row) (*entity.PDFTemplate, error) {
	t := &entity.PDFTemplate{}
	var placeholdersJSON []byte
	err := row.Scan(&t.ID, &t.SpaceID, &t.Name, &t.StoragePath, &placeholdersJSON, &t.Status, &t.CreatedAt)
	if err != nil {
		return nil, err
	}
	if len(placeholdersJSON) > 0 {
		_ = json.Unmarshal(placeholdersJSON, &t.Placeholders)
	}
	if t.Placeholders == nil {
		t.Placeholders = []string{}
	}
	return t, nil
}

func (r *pdfRepository) CreateTemplate(ctx context.Context, t *entity.PDFTemplate) error {
	placeholdersJSON, _ := json.Marshal(t.Placeholders)
	return r.pool.QueryRow(ctx, `
		INSERT INTO pdf_templates (space_id, name, storage_path, placeholders, status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at`,
		t.SpaceID, t.Name, t.StoragePath, placeholdersJSON, t.Status,
	).Scan(&t.ID, &t.CreatedAt)
}

func (r *pdfRepository) ListTemplates(ctx context.Context, spaceID int64) ([]*entity.PDFTemplate, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, space_id, name, storage_path, placeholders, status, created_at
		FROM pdf_templates WHERE space_id = $1 ORDER BY created_at DESC`, spaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*entity.PDFTemplate
	for rows.Next() {
		t, err := scanTemplate(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, t)
	}
	return list, rows.Err()
}

func (r *pdfRepository) GetTemplate(ctx context.Context, id, spaceID int64) (*entity.PDFTemplate, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, space_id, name, storage_path, placeholders, status, created_at
		FROM pdf_templates WHERE id = $1 AND space_id = $2`, id, spaceID)
	t, err := scanTemplate(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apierrors.ErrNotFound
	}
	return t, err
}

func (r *pdfRepository) UpdateTemplateStatus(ctx context.Context, id int64, status string) error {
	_, err := r.pool.Exec(ctx, `UPDATE pdf_templates SET status = $1 WHERE id = $2`, status, id)
	return err
}

func (r *pdfRepository) SaveMappings(ctx context.Context, templateID int64, mappings []*entity.PDFMapping) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `DELETE FROM pdf_mappings WHERE template_id = $1`, templateID); err != nil {
		return err
	}
	for _, m := range mappings {
		if err := tx.QueryRow(ctx, `
			INSERT INTO pdf_mappings (template_id, placeholder, source_field_id, expression)
			VALUES ($1, $2, $3, $4)
			RETURNING id`,
			templateID, m.Placeholder, m.SourceFieldID, m.Expression,
		).Scan(&m.ID); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *pdfRepository) GetMappings(ctx context.Context, templateID int64) ([]*entity.PDFMapping, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, template_id, placeholder, source_field_id, expression
		FROM pdf_mappings WHERE template_id = $1`, templateID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*entity.PDFMapping
	for rows.Next() {
		m := &entity.PDFMapping{}
		if err := rows.Scan(&m.ID, &m.TemplateID, &m.Placeholder, &m.SourceFieldID, &m.Expression); err != nil {
			return nil, err
		}
		list = append(list, m)
	}
	return list, rows.Err()
}

func (r *pdfRepository) CreateJob(ctx context.Context, j *entity.PDFJob) error {
	return r.pool.QueryRow(ctx, `
		INSERT INTO pdf_jobs (space_id, template_id, status, total_records, processed, created_by)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at`,
		j.SpaceID, j.TemplateID, j.Status, j.TotalRecords, j.Processed, j.CreatedBy,
	).Scan(&j.ID, &j.CreatedAt)
}

func (r *pdfRepository) GetJob(ctx context.Context, jobID string, spaceID int64) (*entity.PDFJob, error) {
	j := &entity.PDFJob{}
	err := r.pool.QueryRow(ctx, `
		SELECT id, space_id, template_id, status, total_records, processed,
		       storage_path, created_by, created_at, completed_at
		FROM pdf_jobs WHERE id = $1 AND space_id = $2`, jobID, spaceID,
	).Scan(&j.ID, &j.SpaceID, &j.TemplateID, &j.Status, &j.TotalRecords,
		&j.Processed, &j.StoragePath, &j.CreatedBy, &j.CreatedAt, &j.CompletedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apierrors.ErrNotFound
	}
	return j, err
}

func (r *pdfRepository) UpdateJob(ctx context.Context, j *entity.PDFJob) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE pdf_jobs SET status=$1, processed=$2, storage_path=$3, completed_at=$4
		WHERE id=$5`,
		j.Status, j.Processed, j.StoragePath, j.CompletedAt, j.ID)
	return err
}

func (r *pdfRepository) ListJobs(ctx context.Context, spaceID int64) ([]*entity.PDFJob, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, space_id, template_id, status, total_records, processed,
		       storage_path, created_by, created_at, completed_at
		FROM pdf_jobs WHERE space_id = $1 ORDER BY created_at DESC`, spaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*entity.PDFJob
	for rows.Next() {
		j := &entity.PDFJob{}
		if err := rows.Scan(&j.ID, &j.SpaceID, &j.TemplateID, &j.Status, &j.TotalRecords,
			&j.Processed, &j.StoragePath, &j.CreatedBy, &j.CreatedAt, &j.CompletedAt); err != nil {
			return nil, err
		}
		list = append(list, j)
	}
	return list, rows.Err()
}
