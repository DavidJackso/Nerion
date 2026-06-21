package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"nerion/internal/domain"
	"nerion/internal/entity"
	"nerion/pkg/apierrors"
)

type recordRepository struct {
	pool *pgxpool.Pool
}

func NewRecordRepository(pool *pgxpool.Pool) domain.RecordRepository {
	return &recordRepository{pool: pool}
}

func recQualifiedTable(spaceSlug, tableSlug string) string {
	return fmt.Sprintf("%q.%q", "space_"+spaceSlug, tableSlug)
}

// buildSelectExprs returns the SELECT expression list for a records query.
// NUMERIC columns are cast to float8 so pgx returns float64 (JSON-safe).
func buildSelectExprs(fields []*entity.FieldMeta) string {
	exprs := []string{"id", "created_at", "updated_at"}
	for _, f := range fields {
		if f.Type == entity.FieldTypeNumber {
			exprs = append(exprs, fmt.Sprintf(`%q::float8 AS %q`, f.Slug, f.Slug))
		} else {
			exprs = append(exprs, fmt.Sprintf(`%q`, f.Slug))
		}
	}
	return strings.Join(exprs, ", ")
}

func (r *recordRepository) List(ctx context.Context, spaceSlug, tableSlug string, fields []*entity.FieldMeta, params entity.ListParams) ([]map[string]any, int64, error) {
	tbl := recQualifiedTable(spaceSlug, tableSlug)
	selectExprs := buildSelectExprs(fields)

	var whereParts []string
	var args []any
	argN := 1

	whereParts = append(whereParts, "deleted_at IS NULL")

	if params.Search != "" {
		var searchConds []string
		for _, f := range fields {
			switch f.Type {
			case entity.FieldTypeText, entity.FieldTypeLongtext, entity.FieldTypeEmail,
				entity.FieldTypePhone, entity.FieldTypeURL:
				searchConds = append(searchConds, fmt.Sprintf(`%q ILIKE $%d`, f.Slug, argN))
			}
		}
		if len(searchConds) > 0 {
			args = append(args, "%"+params.Search+"%")
			argN++
			whereParts = append(whereParts, "("+strings.Join(searchConds, " OR ")+")")
		}
	}

	where := "WHERE " + strings.Join(whereParts, " AND ")

	var total int64
	if err := r.pool.QueryRow(ctx,
		fmt.Sprintf("SELECT COUNT(*) FROM %s %s", tbl, where),
		args...,
	).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Validate sort column against known field slugs to prevent injection.
	sortBy := "id"
	for _, f := range fields {
		if f.Slug == params.SortBy {
			sortBy = params.SortBy
			break
		}
	}
	sortDir := "ASC"
	if strings.ToUpper(params.SortDir) == "DESC" {
		sortDir = "DESC"
	}

	limit := params.Limit
	if limit <= 0 || limit > 500 {
		limit = 50
	}
	offset := params.Offset
	if offset < 0 {
		offset = 0
	}

	listArgs := make([]any, len(args), len(args)+2)
	copy(listArgs, args)
	listArgs = append(listArgs, limit, offset)

	query := fmt.Sprintf(
		`SELECT %s FROM %s %s ORDER BY %q %s LIMIT $%d OFFSET $%d`,
		selectExprs, tbl, where, sortBy, sortDir, argN, argN+1,
	)

	rows, err := r.pool.Query(ctx, query, listArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	fds := rows.FieldDescriptions()
	var records []map[string]any
	for rows.Next() {
		vals, err := rows.Values()
		if err != nil {
			return nil, 0, err
		}
		rec := make(map[string]any, len(vals))
		for i, fd := range fds {
			rec[fd.Name] = vals[i]
		}
		records = append(records, rec)
	}
	return records, total, rows.Err()
}

func (r *recordRepository) GetByID(ctx context.Context, spaceSlug, tableSlug string, fields []*entity.FieldMeta, id int64) (map[string]any, error) {
	tbl := recQualifiedTable(spaceSlug, tableSlug)
	query := fmt.Sprintf(`SELECT %s FROM %s WHERE id = $1 AND deleted_at IS NULL`,
		buildSelectExprs(fields), tbl)

	rows, err := r.pool.Query(ctx, query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, apierrors.ErrNotFound
	}
	vals, err := rows.Values()
	if err != nil {
		return nil, err
	}
	fds := rows.FieldDescriptions()
	rec := make(map[string]any, len(vals))
	for i, fd := range fds {
		rec[fd.Name] = vals[i]
	}
	return rec, rows.Err()
}

func (r *recordRepository) Create(ctx context.Context, spaceSlug, tableSlug string, fields []*entity.FieldMeta, data map[string]any) (map[string]any, error) {
	tbl := recQualifiedTable(spaceSlug, tableSlug)
	selectExprs := buildSelectExprs(fields)

	var cols []string
	var placeholders []string
	var args []any
	argN := 1

	for _, f := range fields {
		val, ok := data[f.Slug]
		if !ok || val == nil {
			continue
		}
		cols = append(cols, fmt.Sprintf("%q", f.Slug))
		placeholders = append(placeholders, fmt.Sprintf("$%d", argN))
		args = append(args, val)
		argN++
	}

	var query string
	if len(cols) == 0 {
		query = fmt.Sprintf(`INSERT INTO %s DEFAULT VALUES RETURNING %s`, tbl, selectExprs)
	} else {
		query = fmt.Sprintf(
			`INSERT INTO %s (%s) VALUES (%s) RETURNING %s`,
			tbl,
			strings.Join(cols, ", "),
			strings.Join(placeholders, ", "),
			selectExprs,
		)
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, fmt.Errorf("insert returned no rows")
	}
	vals, err := rows.Values()
	if err != nil {
		return nil, err
	}
	fds := rows.FieldDescriptions()
	rec := make(map[string]any, len(vals))
	for i, fd := range fds {
		rec[fd.Name] = vals[i]
	}
	return rec, rows.Err()
}

func (r *recordRepository) Update(ctx context.Context, spaceSlug, tableSlug string, fields []*entity.FieldMeta, id int64, data map[string]any) (map[string]any, error) {
	tbl := recQualifiedTable(spaceSlug, tableSlug)
	selectExprs := buildSelectExprs(fields)

	fieldBySlug := make(map[string]struct{}, len(fields))
	for _, f := range fields {
		fieldBySlug[f.Slug] = struct{}{}
	}

	var setClauses []string
	var args []any
	argN := 1

	for slug, val := range data {
		if _, ok := fieldBySlug[slug]; !ok {
			continue
		}
		setClauses = append(setClauses, fmt.Sprintf("%q = $%d", slug, argN))
		args = append(args, val)
		argN++
	}

	if len(setClauses) == 0 {
		return r.GetByID(ctx, spaceSlug, tableSlug, fields, id)
	}

	setClauses = append(setClauses, "updated_at = now()")
	args = append(args, id)

	query := fmt.Sprintf(
		`UPDATE %s SET %s WHERE id = $%d AND deleted_at IS NULL RETURNING %s`,
		tbl, strings.Join(setClauses, ", "), argN, selectExprs,
	)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, apierrors.ErrNotFound
	}
	vals, err := rows.Values()
	if err != nil {
		return nil, err
	}
	fds := rows.FieldDescriptions()
	rec := make(map[string]any, len(vals))
	for i, fd := range fds {
		rec[fd.Name] = vals[i]
	}
	return rec, rows.Err()
}

func (r *recordRepository) Delete(ctx context.Context, spaceSlug, tableSlug string, id int64) error {
	tbl := recQualifiedTable(spaceSlug, tableSlug)
	tag, err := r.pool.Exec(ctx,
		fmt.Sprintf(`UPDATE %s SET deleted_at = now() WHERE id = $1 AND deleted_at IS NULL`, tbl),
		id,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return apierrors.ErrNotFound
	}
	return nil
}

func (r *recordRepository) CheckUnique(ctx context.Context, spaceSlug, tableSlug, colSlug string, val any, excludeID *int64) (bool, error) {
	tbl := recQualifiedTable(spaceSlug, tableSlug)
	args := []any{val}
	query := fmt.Sprintf(`SELECT EXISTS(SELECT 1 FROM %s WHERE %q = $1 AND deleted_at IS NULL`, tbl, colSlug)
	if excludeID != nil {
		query += " AND id != $2"
		args = append(args, *excludeID)
	}
	query += ")"
	var exists bool
	return exists, r.pool.QueryRow(ctx, query, args...).Scan(&exists)
}
