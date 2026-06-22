package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"nerion/internal/domain"
	"nerion/internal/entity"
	"nerion/pkg/apierrors"
)

// --- TableRepository ---

type tableRepository struct {
	pool *pgxpool.Pool
}

func NewTableRepository(pool *pgxpool.Pool) domain.TableRepository {
	return &tableRepository{pool: pool}
}

func (r *tableRepository) Create(ctx context.Context, t *entity.TableMeta) error {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO table_meta (space_id, name, slug) VALUES ($1, $2, $3) RETURNING id, created_at`,
		t.SpaceID, t.Name, t.Slug,
	).Scan(&t.ID, &t.CreatedAt)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return apierrors.NewError(409, "conflict", "Таблица с таким slug уже существует")
	}
	return err
}

func (r *tableRepository) GetBySlug(ctx context.Context, spaceID int64, slug string) (*entity.TableMeta, error) {
	t := &entity.TableMeta{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, space_id, name, slug, created_at FROM table_meta WHERE space_id = $1 AND slug = $2`,
		spaceID, slug,
	).Scan(&t.ID, &t.SpaceID, &t.Name, &t.Slug, &t.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apierrors.ErrNotFound
	}
	return t, err
}

func (r *tableRepository) ListBySpace(ctx context.Context, spaceID int64) ([]*entity.TableMeta, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, space_id, name, slug, created_at FROM table_meta WHERE space_id = $1 ORDER BY created_at`,
		spaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]*entity.TableMeta, 0)
	for rows.Next() {
		t := &entity.TableMeta{}
		if err := rows.Scan(&t.ID, &t.SpaceID, &t.Name, &t.Slug, &t.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, t)
	}
	return list, rows.Err()
}

func (r *tableRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM table_meta WHERE id = $1`, id)
	return err
}

// --- FieldRepository ---

type fieldRepository struct {
	pool *pgxpool.Pool
}

func NewFieldRepository(pool *pgxpool.Pool) domain.FieldRepository {
	return &fieldRepository{pool: pool}
}

func (r *fieldRepository) ListByTable(ctx context.Context, tableID int64) ([]*entity.FieldMeta, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, table_id, name, slug, type, required, default_value, "unique",
		       enum_values, relation_table_id, relation_cardinality, position
		FROM field_meta WHERE table_id = $1 ORDER BY position, id`, tableID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]*entity.FieldMeta, 0)
	for rows.Next() {
		f := &entity.FieldMeta{}
		var ftype string
		if err := rows.Scan(
			&f.ID, &f.TableID, &f.Name, &f.Slug, &ftype, &f.Required,
			&f.DefaultValue, &f.Unique, &f.EnumValues,
			&f.RelationTableID, &f.RelationCardinality, &f.Position,
		); err != nil {
			return nil, err
		}
		f.Type = entity.FieldType(ftype)
		list = append(list, f)
	}
	return list, rows.Err()
}

func (r *fieldRepository) Upsert(ctx context.Context, f *entity.FieldMeta) error {
	return r.pool.QueryRow(ctx, `
		INSERT INTO field_meta (table_id, name, slug, type, required, default_value, "unique",
		                        enum_values, relation_table_id, relation_cardinality, position)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		ON CONFLICT (table_id, slug) DO UPDATE SET
		  name=EXCLUDED.name, type=EXCLUDED.type, required=EXCLUDED.required,
		  default_value=EXCLUDED.default_value, "unique"=EXCLUDED.unique,
		  enum_values=EXCLUDED.enum_values, relation_table_id=EXCLUDED.relation_table_id,
		  relation_cardinality=EXCLUDED.relation_cardinality, position=EXCLUDED.position
		RETURNING id`,
		f.TableID, f.Name, f.Slug, string(f.Type), f.Required,
		f.DefaultValue, f.Unique, f.EnumValues,
		f.RelationTableID, f.RelationCardinality, f.Position,
	).Scan(&f.ID)
}

func (r *fieldRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM field_meta WHERE id = $1`, id)
	return err
}

// DDLExecutor executes raw DDL against a space schema.
type DDLExecutor struct {
	pool *pgxpool.Pool
}

func NewDDLExecutor(pool *pgxpool.Pool) *DDLExecutor {
	return &DDLExecutor{pool: pool}
}

func (e *DDLExecutor) CreateTable(ctx context.Context, spaceSlug, tableSlug string) error {
	schema := fmt.Sprintf("space_%s", spaceSlug)
	ddl := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %q.%q (
		id         BIGSERIAL    PRIMARY KEY,
		created_at TIMESTAMPTZ  NOT NULL DEFAULT now(),
		updated_at TIMESTAMPTZ  NOT NULL DEFAULT now(),
		deleted_at TIMESTAMPTZ
	)`, schema, tableSlug)
	_, err := e.pool.Exec(ctx, ddl)
	return err
}

func (e *DDLExecutor) DropTable(ctx context.Context, spaceSlug, tableSlug string) error {
	schema := fmt.Sprintf("space_%s", spaceSlug)
	_, err := e.pool.Exec(ctx, fmt.Sprintf(`DROP TABLE IF EXISTS %q.%q CASCADE`, schema, tableSlug))
	return err
}

func (e *DDLExecutor) fieldTypeToPG(ft entity.FieldType) string {
	switch ft {
	case entity.FieldTypeNumber:
		return "NUMERIC"
	case entity.FieldTypeBoolean:
		return "BOOLEAN"
	case entity.FieldTypeDate:
		return "DATE"
	case entity.FieldTypeDatetime:
		return "TIMESTAMPTZ"
	case entity.FieldTypeRelation:
		return "BIGINT"
	case entity.FieldTypeLongtext:
		return "TEXT"
	default:
		return "TEXT"
	}
}

func (e *DDLExecutor) AddColumn(ctx context.Context, spaceSlug, tableSlug string, f *entity.FieldMeta) error {
	schema := fmt.Sprintf("space_%s", spaceSlug)
	pgType := e.fieldTypeToPG(f.Type)
	notNull := ""
	if f.Required {
		notNull = " NOT NULL"
	}
	defVal := ""
	if f.DefaultValue != nil {
		defVal = fmt.Sprintf(" DEFAULT '%s'", strings.ReplaceAll(*f.DefaultValue, "'", "''"))
	}
	ddl := fmt.Sprintf(`ALTER TABLE %q.%q ADD COLUMN IF NOT EXISTS %q %s%s%s`,
		schema, tableSlug, f.Slug, pgType, notNull, defVal)
	_, err := e.pool.Exec(ctx, ddl)
	return err
}

func (e *DDLExecutor) DropColumn(ctx context.Context, spaceSlug, tableSlug, colSlug string) error {
	schema := fmt.Sprintf("space_%s", spaceSlug)
	_, err := e.pool.Exec(ctx, fmt.Sprintf(`ALTER TABLE %q.%q DROP COLUMN IF EXISTS %q`,
		schema, tableSlug, colSlug))
	return err
}
