package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"nerion/internal/domain"
	"nerion/internal/entity"
	"nerion/pkg/apierrors"
)

type spaceInvitationRepository struct {
	pool *pgxpool.Pool
}

func NewSpaceInvitationRepository(pool *pgxpool.Pool) domain.SpaceInvitationRepository {
	return &spaceInvitationRepository{pool: pool}
}

func (r *spaceInvitationRepository) Create(ctx context.Context, inv *entity.SpaceInvitation) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO space_invitations (space_id, email, invited_by, token_hash, expires_at)
		 VALUES ($1, $2, $3, $4, $5)`,
		inv.SpaceID, inv.Email, inv.InvitedBy, inv.TokenHash, inv.ExpiresAt,
	)
	return err
}

func (r *spaceInvitationRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*entity.SpaceInvitation, error) {
	inv := &entity.SpaceInvitation{}
	err := r.pool.QueryRow(ctx,
		`SELECT i.id, i.space_id, s.name, i.email, i.invited_by, i.token_hash, i.created_at, i.expires_at, i.used_at
		 FROM space_invitations i
		 JOIN spaces s ON s.id = i.space_id
		 WHERE i.token_hash = $1`,
		tokenHash,
	).Scan(&inv.ID, &inv.SpaceID, &inv.SpaceName, &inv.Email, &inv.InvitedBy, &inv.TokenHash, &inv.CreatedAt, &inv.ExpiresAt, &inv.UsedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apierrors.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return inv, nil
}

func (r *spaceInvitationRepository) MarkUsed(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE space_invitations SET used_at = $1 WHERE id = $2`,
		time.Now(), id,
	)
	return err
}
