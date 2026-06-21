package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"nerion/internal/domain"
	"nerion/internal/entity"
	"nerion/pkg/apierrors"
)

type spaceRepository struct {
	pool *pgxpool.Pool
}

func NewSpaceRepository(pool *pgxpool.Pool) domain.SpaceRepository {
	return &spaceRepository{pool: pool}
}

func (r *spaceRepository) Create(ctx context.Context, s *entity.Space) error {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO spaces (name, slug, owner_id) VALUES ($1, $2, $3) RETURNING id, created_at`,
		s.Name, s.Slug, s.OwnerID,
	).Scan(&s.ID, &s.CreatedAt)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return apierrors.ErrConflict
	}
	return err
}

func (r *spaceRepository) GetBySlug(ctx context.Context, slug string) (*entity.Space, error) {
	s := &entity.Space{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, slug, owner_id, created_at FROM spaces WHERE slug = $1`, slug,
	).Scan(&s.ID, &s.Name, &s.Slug, &s.OwnerID, &s.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apierrors.ErrNotFound
	}
	return s, err
}

func (r *spaceRepository) ListByUserID(ctx context.Context, userID int64) ([]*entity.Space, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT s.id, s.name, s.slug, s.owner_id, s.created_at
		FROM spaces s
		JOIN space_members sm ON sm.space_id = s.id
		WHERE sm.user_id = $1
		ORDER BY s.created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*entity.Space
	for rows.Next() {
		s := &entity.Space{}
		if err := rows.Scan(&s.ID, &s.Name, &s.Slug, &s.OwnerID, &s.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, s)
	}
	return list, rows.Err()
}

func (r *spaceRepository) UpdateName(ctx context.Context, id int64, name string) error {
	tag, err := r.pool.Exec(ctx, `UPDATE spaces SET name = $1 WHERE id = $2`, name, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return apierrors.ErrNotFound
	}
	return nil
}

func (r *spaceRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM spaces WHERE id = $1`, id)
	return err
}

func (r *spaceRepository) TableCount(ctx context.Context, spaceID int64) (int, error) {
	var n int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM table_meta WHERE space_id = $1`, spaceID,
	).Scan(&n)
	return n, err
}

func (r *spaceRepository) CreateSchema(ctx context.Context, slug string) error {
	schemaName := fmt.Sprintf("space_%s", slug)
	_, err := r.pool.Exec(ctx, fmt.Sprintf(`CREATE SCHEMA IF NOT EXISTS %q`, schemaName))
	return err
}

func (r *spaceRepository) DropSchema(ctx context.Context, slug string) error {
	schemaName := fmt.Sprintf("space_%s", slug)
	_, err := r.pool.Exec(ctx, fmt.Sprintf(`DROP SCHEMA IF EXISTS %q CASCADE`, schemaName))
	return err
}
