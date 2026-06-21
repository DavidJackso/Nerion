package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"nerion/internal/domain"
	"nerion/internal/entity"
	"nerion/pkg/apierrors"
)

type spaceMemberRepository struct {
	pool *pgxpool.Pool
}

func NewSpaceMemberRepository(pool *pgxpool.Pool) domain.SpaceMemberRepository {
	return &spaceMemberRepository{pool: pool}
}

func (r *spaceMemberRepository) Add(ctx context.Context, m *entity.SpaceMember) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO space_members (space_id, user_id, role) VALUES ($1, $2, $3)
		 ON CONFLICT (space_id, user_id) DO NOTHING`,
		m.SpaceID, m.UserID, string(m.Role),
	)
	return err
}

func (r *spaceMemberRepository) GetRole(ctx context.Context, spaceID, userID int64) (entity.SpaceMemberRole, error) {
	var role string
	err := r.pool.QueryRow(ctx,
		`SELECT role FROM space_members WHERE space_id = $1 AND user_id = $2`,
		spaceID, userID,
	).Scan(&role)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", apierrors.ErrForbidden
	}
	if err != nil {
		return "", err
	}
	return entity.SpaceMemberRole(role), nil
}

func (r *spaceMemberRepository) List(ctx context.Context, spaceID int64) ([]*entity.SpaceMember, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT sm.space_id, sm.user_id, u.name, u.email, sm.role, sm.joined_at
		FROM space_members sm
		JOIN users u ON u.id = sm.user_id
		WHERE sm.space_id = $1
		ORDER BY sm.joined_at`, spaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*entity.SpaceMember
	for rows.Next() {
		m := &entity.SpaceMember{}
		var role string
		if err := rows.Scan(&m.SpaceID, &m.UserID, &m.UserName, &m.UserEmail, &role, &m.JoinedAt); err != nil {
			return nil, err
		}
		m.Role = entity.SpaceMemberRole(role)
		list = append(list, m)
	}
	return list, rows.Err()
}

func (r *spaceMemberRepository) UpdateRole(ctx context.Context, spaceID, userID int64, role entity.SpaceMemberRole) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE space_members SET role = $1 WHERE space_id = $2 AND user_id = $3`,
		string(role), spaceID, userID,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return apierrors.ErrNotFound
	}
	return nil
}

func (r *spaceMemberRepository) Remove(ctx context.Context, spaceID, userID int64) error {
	tag, err := r.pool.Exec(ctx,
		`DELETE FROM space_members WHERE space_id = $1 AND user_id = $2`,
		spaceID, userID,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return apierrors.ErrNotFound
	}
	return nil
}

func (r *spaceMemberRepository) AdminCount(ctx context.Context, spaceID int64) (int, error) {
	var n int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM space_members WHERE space_id = $1 AND role = 'admin'`, spaceID,
	).Scan(&n)
	return n, err
}
