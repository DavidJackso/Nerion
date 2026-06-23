package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"log/slog"

	"nerion/internal/domain"
	"nerion/internal/entity"
	"nerion/pkg/apierrors"
)

type apiKeyService struct {
	spaceRepo  domain.SpaceRepository
	memberRepo domain.SpaceMemberRepository
	keyRepo    domain.APIKeyRepository
	logger     *slog.Logger
}

func NewAPIKeyService(
	spaceRepo domain.SpaceRepository,
	memberRepo domain.SpaceMemberRepository,
	keyRepo domain.APIKeyRepository,
	logger *slog.Logger,
) domain.APIKeyService {
	return &apiKeyService{
		spaceRepo:  spaceRepo,
		memberRepo: memberRepo,
		keyRepo:    keyRepo,
		logger:     logger,
	}
}

func generateKey() (fullKey, prefix, hash string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return
	}
	rawHex := hex.EncodeToString(b) // 64 hex chars
	fullKey = "nrn_" + rawHex
	prefix = fullKey[:12] // "nrn_" + first 8 hex chars
	h := sha256.Sum256([]byte(fullKey))
	hash = hex.EncodeToString(h[:])
	return
}

func (s *apiKeyService) requireAdmin(ctx context.Context, spaceSlug string, userID int64) (*entity.Space, error) {
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

func (s *apiKeyService) Create(ctx context.Context, spaceSlug, name, scope string, userID int64) (*entity.APIKey, string, error) {
	if name == "" {
		return nil, "", apierrors.NewValidationError(map[string]string{"name": "Обязательное поле"})
	}
	if scope != "read" && scope != "write" {
		scope = "write"
	}

	space, err := s.requireAdmin(ctx, spaceSlug, userID)
	if err != nil {
		return nil, "", err
	}

	fullKey, prefix, hash, err := generateKey()
	if err != nil {
		return nil, "", err
	}

	key := &entity.APIKey{
		SpaceID:   space.ID,
		Name:      name,
		KeyPrefix: prefix,
		Scope:     scope,
	}
	if err := s.keyRepo.Create(ctx, key, hash); err != nil {
		return nil, "", err
	}
	s.logger.Info("api key created", "space", spaceSlug, "name", name, "scope", scope, "user_id", userID)
	return key, fullKey, nil
}

func (s *apiKeyService) List(ctx context.Context, spaceSlug string, userID int64) ([]*entity.APIKey, error) {
	space, err := s.requireAdmin(ctx, spaceSlug, userID)
	if err != nil {
		return nil, err
	}
	return s.keyRepo.ListBySpace(ctx, space.ID)
}

func (s *apiKeyService) Revoke(ctx context.Context, spaceSlug string, keyID, userID int64) error {
	space, err := s.requireAdmin(ctx, spaceSlug, userID)
	if err != nil {
		return err
	}
	if err := s.keyRepo.Revoke(ctx, keyID, space.ID); err != nil {
		return err
	}
	s.logger.Info("api key revoked", "space", spaceSlug, "key_id", keyID, "user_id", userID)
	return nil
}
