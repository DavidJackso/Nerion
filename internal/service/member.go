package service

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"nerion/internal/domain"
	"nerion/internal/entity"
	"nerion/pkg/apierrors"
)

type spaceMemberService struct {
	memberRepo domain.SpaceMemberRepository
	userRepo   domain.UserRepository
	inviteRepo domain.SpaceInvitationRepository
	emailSend  domain.EmailSender
	logger     *slog.Logger
}

func NewSpaceMemberService(
	memberRepo domain.SpaceMemberRepository,
	userRepo domain.UserRepository,
	inviteRepo domain.SpaceInvitationRepository,
	emailSend domain.EmailSender,
	logger *slog.Logger,
) domain.SpaceMemberService {
	return &spaceMemberService{
		memberRepo: memberRepo,
		userRepo:   userRepo,
		inviteRepo: inviteRepo,
		emailSend:  emailSend,
		logger:     logger,
	}
}

func (s *spaceMemberService) GetRole(ctx context.Context, spaceID, userID int64) (entity.SpaceMemberRole, error) {
	return s.memberRepo.GetRole(ctx, spaceID, userID)
}

func (s *spaceMemberService) ListMembers(ctx context.Context, spaceID int64) ([]*entity.SpaceMember, error) {
	return s.memberRepo.List(ctx, spaceID)
}

func (s *spaceMemberService) Invite(ctx context.Context, spaceID, inviterID int64, email string) error {
	role, err := s.memberRepo.GetRole(ctx, spaceID, inviterID)
	if err != nil || role != entity.SpaceMemberRoleAdmin {
		return apierrors.ErrForbidden
	}

	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if !errors.Is(err, apierrors.ErrNotFound) {
			return err
		}
		rawHex, raw, genErr := generateRawToken()
		if genErr != nil {
			s.logger.Error("generate invite token", "err", genErr)
			return fmt.Errorf("не удалось создать приглашение")
		}
		inv := &entity.SpaceInvitation{
			SpaceID:   spaceID,
			Email:     email,
			InvitedBy: inviterID,
			TokenHash: hashToken(raw),
			ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		}
		if storeErr := s.inviteRepo.Create(ctx, inv); storeErr != nil {
			s.logger.Error("store invite", "err", storeErr)
			return fmt.Errorf("не удалось создать приглашение")
		}
		if sendErr := s.emailSend.Send(email, "Приглашение в Nerion",
			fmt.Sprintf("Вас пригласили в пространство Nerion. Перейдите по ссылке, чтобы принять приглашение: https://nerionapp.ru/invite?token=%s", rawHex),
		); sendErr != nil {
			s.logger.Error("send invite email", "email", email, "err", sendErr)
			return fmt.Errorf("не удалось отправить приглашение")
		}
		return nil
	}

	return s.memberRepo.Add(ctx, &entity.SpaceMember{
		SpaceID: spaceID,
		UserID:  user.ID,
		Role:    entity.SpaceMemberRoleMember,
	})
}

func (s *spaceMemberService) GetInviteInfo(ctx context.Context, token string) (*entity.SpaceInvitation, error) {
	raw, err := hex.DecodeString(token)
	if err != nil {
		return nil, apierrors.ErrNotFound
	}
	inv, err := s.inviteRepo.GetByTokenHash(ctx, hashToken(raw))
	if err != nil {
		return nil, err
	}
	if time.Now().After(inv.ExpiresAt) {
		return nil, apierrors.NewValidationError(map[string]string{"token": "Приглашение истекло"})
	}
	if inv.UsedAt != nil {
		return nil, apierrors.NewValidationError(map[string]string{"token": "Приглашение уже использовано"})
	}
	return inv, nil
}

func (s *spaceMemberService) AcceptInvite(ctx context.Context, token string, userID int64) error {
	inv, err := s.GetInviteInfo(ctx, token)
	if err != nil {
		return err
	}
	if err := s.memberRepo.Add(ctx, &entity.SpaceMember{
		SpaceID: inv.SpaceID,
		UserID:  userID,
		Role:    entity.SpaceMemberRoleMember,
	}); err != nil {
		return err
	}
	return s.inviteRepo.MarkUsed(ctx, inv.ID)
}

func (s *spaceMemberService) ChangeRole(ctx context.Context, spaceID, requesterID, targetUserID int64, role entity.SpaceMemberRole) error {
	reqRole, err := s.memberRepo.GetRole(ctx, spaceID, requesterID)
	if err != nil || reqRole != entity.SpaceMemberRoleAdmin {
		return apierrors.ErrForbidden
	}
	if role != entity.SpaceMemberRoleAdmin {
		currentRole, _ := s.memberRepo.GetRole(ctx, spaceID, targetUserID)
		if currentRole == entity.SpaceMemberRoleAdmin {
			count, err := s.memberRepo.AdminCount(ctx, spaceID)
			if err != nil {
				return err
			}
			if count <= 1 {
				return apierrors.NewValidationError(map[string]string{
					"role": "Нельзя понизить единственного администратора",
				})
			}
		}
	}
	return s.memberRepo.UpdateRole(ctx, spaceID, targetUserID, role)
}

func (s *spaceMemberService) RemoveMember(ctx context.Context, spaceID, requesterID, targetUserID int64) error {
	reqRole, err := s.memberRepo.GetRole(ctx, spaceID, requesterID)
	if err != nil || reqRole != entity.SpaceMemberRoleAdmin {
		return apierrors.ErrForbidden
	}
	targetRole, _ := s.memberRepo.GetRole(ctx, spaceID, targetUserID)
	if targetRole == entity.SpaceMemberRoleAdmin {
		count, err := s.memberRepo.AdminCount(ctx, spaceID)
		if err != nil {
			return err
		}
		if count <= 1 {
			return apierrors.NewValidationError(map[string]string{
				"user_id": "Нельзя удалить единственного администратора",
			})
		}
	}
	return s.memberRepo.Remove(ctx, spaceID, targetUserID)
}
