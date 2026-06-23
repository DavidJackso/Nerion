package service

import (
	"context"
	"errors"
	"log/slog"

	"nerion/internal/domain"
	"nerion/internal/entity"
	"nerion/pkg/apierrors"
)

type spaceMemberService struct {
	memberRepo domain.SpaceMemberRepository
	userRepo   domain.UserRepository
	emailSend  domain.EmailSender
	logger     *slog.Logger
}

func NewSpaceMemberService(
	memberRepo domain.SpaceMemberRepository,
	userRepo domain.UserRepository,
	emailSend domain.EmailSender,
	logger *slog.Logger,
) domain.SpaceMemberService {
	return &spaceMemberService{
		memberRepo: memberRepo,
		userRepo:   userRepo,
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
		// User doesn't exist — send invitation email (stub)
		err = s.emailSend.Send(email, "Приглашение в Nerion",
			"Вас пригласили в пространство. Зарегистрируйтесь на https://nerionapp.ru")
		if err != nil {
			s.logger.Error("Failed to send invitation email", "email", email, "error", err)
			return errors.New("не удалось отправить приглашение")
		}

		return nil
	}

	return s.memberRepo.Add(ctx, &entity.SpaceMember{
		SpaceID: spaceID,
		UserID:  user.ID,
		Role:    entity.SpaceMemberRoleMember,
	})
}

func (s *spaceMemberService) ChangeRole(ctx context.Context, spaceID, requesterID, targetUserID int64, role entity.SpaceMemberRole) error {
	reqRole, err := s.memberRepo.GetRole(ctx, spaceID, requesterID)
	if err != nil || reqRole != entity.SpaceMemberRoleAdmin {
		return apierrors.ErrForbidden
	}
	// Protect last admin
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
