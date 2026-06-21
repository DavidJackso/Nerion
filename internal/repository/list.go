package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"nerion/internal/domain"
	"nerion/internal/entity"
	"nerion/pkg/apierrors"
)

type listRepository struct {
	pool *pgxpool.Pool
}

func NewListRepository(pool *pgxpool.Pool) domain.ListRepository {
	return &listRepository{pool: pool}
}

const listSelectCols = `
	l.id, l.space_id, l.slug, l.source_table_id,
	tm.slug AS table_slug,
	l.field_config, l.filter_config, l.sort_config, l.row_limit,
	l.published_at, l.unpublished_at, l.created_at`

func scanList(row pgx.Row) (*entity.List, error) {
	l := &entity.List{}
	var fieldConfigJSON, filterConfigJSON, sortConfigJSON []byte
	err := row.Scan(
		&l.ID, &l.SpaceID, &l.Slug, &l.SourceTableID, &l.TableSlug,
		&fieldConfigJSON, &filterConfigJSON, &sortConfigJSON, &l.RowLimit,
		&l.PublishedAt, &l.UnpublishedAt, &l.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	if len(fieldConfigJSON) > 0 {
		_ = json.Unmarshal(fieldConfigJSON, &l.FieldConfig)
	}
	if len(filterConfigJSON) > 0 {
		_ = json.Unmarshal(filterConfigJSON, &l.FilterConfig)
	}
	if len(sortConfigJSON) > 0 {
		_ = json.Unmarshal(sortConfigJSON, &l.SortConfig)
	}
	if l.FieldConfig == nil {
		l.FieldConfig = []entity.ListFieldConfig{}
	}
	if l.FilterConfig == nil {
		l.FilterConfig = map[string]any{}
	}
	if l.SortConfig == nil {
		l.SortConfig = []entity.ListSortConfig{}
	}
	return l, nil
}

func (r *listRepository) Create(ctx context.Context, l *entity.List) error {
	fieldJSON, _ := json.Marshal(l.FieldConfig)
	filterJSON, _ := json.Marshal(l.FilterConfig)
	sortJSON, _ := json.Marshal(l.SortConfig)

	limit := l.RowLimit
	if limit <= 0 {
		limit = 100
	}

	var publishedAt any
	if l.IsPublished() {
		publishedAt = "now()"
	}
	_ = publishedAt

	return r.pool.QueryRow(ctx, `
		INSERT INTO lists (space_id, slug, source_table_id, field_config, filter_config, sort_config, row_limit, published_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at`,
		l.SpaceID, l.Slug, l.SourceTableID,
		fieldJSON, filterJSON, sortJSON, limit,
		publishedAtPtr(l),
	).Scan(&l.ID, &l.CreatedAt)
}

func publishedAtPtr(l *entity.List) any {
	if l.PublishedAt != nil {
		return l.PublishedAt
	}
	return nil
}

func (r *listRepository) GetBySlug(ctx context.Context, spaceID int64, slug string) (*entity.List, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT `+listSelectCols+`
		FROM lists l JOIN table_meta tm ON l.source_table_id = tm.id
		WHERE l.space_id = $1 AND l.slug = $2`, spaceID, slug)
	l, err := scanList(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apierrors.ErrNotFound
	}
	return l, err
}

func (r *listRepository) ListBySpace(ctx context.Context, spaceID int64) ([]*entity.List, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT `+listSelectCols+`
		FROM lists l JOIN table_meta tm ON l.source_table_id = tm.id
		WHERE l.space_id = $1 ORDER BY l.created_at DESC`, spaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*entity.List
	for rows.Next() {
		l, err := scanList(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, l)
	}
	return list, rows.Err()
}

func (r *listRepository) Update(ctx context.Context, l *entity.List) error {
	fieldJSON, _ := json.Marshal(l.FieldConfig)
	filterJSON, _ := json.Marshal(l.FilterConfig)
	sortJSON, _ := json.Marshal(l.SortConfig)

	limit := l.RowLimit
	if limit <= 0 {
		limit = 100
	}

	_, err := r.pool.Exec(ctx, `
		UPDATE lists SET
			field_config   = $1,
			filter_config  = $2,
			sort_config    = $3,
			row_limit      = $4,
			published_at   = $5,
			unpublished_at = $6
		WHERE id = $7`,
		fieldJSON, filterJSON, sortJSON, limit,
		l.PublishedAt, l.UnpublishedAt, l.ID,
	)
	return err
}

func (r *listRepository) GetPublic(ctx context.Context, spaceSlug, listSlug string) (*entity.List, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT `+listSelectCols+`
		FROM lists l
		JOIN table_meta tm ON l.source_table_id = tm.id
		JOIN spaces s ON l.space_id = s.id
		WHERE s.slug = $1 AND l.slug = $2
		  AND l.published_at IS NOT NULL AND l.unpublished_at IS NULL`,
		spaceSlug, listSlug)
	l, err := scanList(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apierrors.ErrNotFound
	}
	return l, err
}

func (r *listRepository) QueryPublic(ctx context.Context, spaceSlug string, l *entity.List, fields []*entity.FieldMeta) ([]map[string]any, error) {
	fieldBySlug := make(map[string]*entity.FieldMeta, len(fields))
	for _, f := range fields {
		fieldBySlug[f.Slug] = f
	}

	var selectExprs []string
	for _, fc := range l.FieldConfig {
		f, ok := fieldBySlug[fc.FieldSlug]
		if !ok {
			continue
		}
		publicName := fc.PublicName
		if publicName == "" {
			publicName = fc.FieldSlug
		}
		if f.Type == entity.FieldTypeNumber {
			selectExprs = append(selectExprs, fmt.Sprintf(`%q::float8 AS %q`, fc.FieldSlug, publicName))
		} else {
			selectExprs = append(selectExprs, fmt.Sprintf(`%q AS %q`, fc.FieldSlug, publicName))
		}
	}
	if len(selectExprs) == 0 {
		return []map[string]any{}, nil
	}
	// Always include id
	selectExprs = append([]string{"id"}, selectExprs...)

	tbl := recQualifiedTable(spaceSlug, l.TableSlug)

	var whereParts []string
	var args []any
	argN := 1
	whereParts = append(whereParts, "deleted_at IS NULL")

	for slug, val := range l.FilterConfig {
		if _, ok := fieldBySlug[slug]; ok {
			whereParts = append(whereParts, fmt.Sprintf(`%q = $%d`, slug, argN))
			args = append(args, val)
			argN++
		}
	}

	var orderParts []string
	for _, sc := range l.SortConfig {
		if _, ok := fieldBySlug[sc.FieldSlug]; !ok {
			continue
		}
		dir := "ASC"
		if strings.ToUpper(sc.Direction) == "DESC" {
			dir = "DESC"
		}
		orderParts = append(orderParts, fmt.Sprintf(`%q %s`, sc.FieldSlug, dir))
	}
	if len(orderParts) == 0 {
		orderParts = []string{"id ASC"}
	}

	limit := l.RowLimit
	if limit <= 0 {
		limit = 100
	}
	args = append(args, limit)

	query := fmt.Sprintf(
		"SELECT %s FROM %s WHERE %s ORDER BY %s LIMIT $%d",
		strings.Join(selectExprs, ", "),
		tbl,
		strings.Join(whereParts, " AND "),
		strings.Join(orderParts, ", "),
		argN,
	)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	fds := rows.FieldDescriptions()
	var records []map[string]any
	for rows.Next() {
		vals, err := rows.Values()
		if err != nil {
			return nil, err
		}
		rec := make(map[string]any, len(vals))
		for i, fd := range fds {
			rec[fd.Name] = vals[i]
		}
		records = append(records, rec)
	}
	if records == nil {
		records = []map[string]any{}
	}
	return records, rows.Err()
}
