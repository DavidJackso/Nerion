package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"golang.org/x/crypto/bcrypt"

	"nerion/internal/domain"
	"nerion/internal/entity"
	"nerion/internal/jwtauth"
	"nerion/pkg/apierrors"
)

type authService struct {
	userRepo       domain.UserRepository
	sessionRepo    domain.SessionRepository
	emailVerifRepo domain.EmailVerificationRepository
	pwdResetRepo   domain.PasswordResetRepository
	jm             jwtauth.Tokenizer
	emailSender    domain.EmailSender
	logger         *slog.Logger
}

func NewAuthService(
	userRepo domain.UserRepository,
	sessionRepo domain.SessionRepository,
	emailVerifRepo domain.EmailVerificationRepository,
	pwdResetRepo domain.PasswordResetRepository,
	jm jwtauth.Tokenizer,
	emailSender domain.EmailSender,
	logger *slog.Logger,
) domain.AuthService {
	return &authService{
		userRepo:       userRepo,
		sessionRepo:    sessionRepo,
		emailVerifRepo: emailVerifRepo,
		pwdResetRepo:   pwdResetRepo,
		jm:             jm,
		emailSender:    emailSender,
		logger:         logger,
	}
}

func hashToken(raw []byte) string {
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

func generateRawToken() (rawHex string, raw []byte, err error) {
	raw = make([]byte, 32)
	if _, err = rand.Read(raw); err != nil {
		return "", nil, fmt.Errorf("rand: %w", err)
	}
	return hex.EncodeToString(raw), raw, nil
}

// newUUID generates a random UUID v4 string without external deps.
func newUUID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func (s *authService) Register(ctx context.Context, name, email, password string) (*entity.User, error) {
	if name == "" || email == "" || len(password) < 8 {
		return nil, apierrors.ErrBadRequest
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, apierrors.ErrInternal
	}
	user := &entity.User{
		Name:          name,
		Email:         email,
		Role:          entity.RoleUser,
		PasswordHash:  string(hash),
		EmailVerified: false,
	}
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	rawHex, raw, err := generateRawToken()
	if err != nil {
		s.logger.Error("generate email verif token", "err", err)
		return user, nil
	}
	tokenHash := hashToken(raw)
	if err := s.emailVerifRepo.Create(ctx, user.ID, tokenHash, time.Now().Add(24*time.Hour)); err != nil {
		s.logger.Error("store email verif token", "err", err)
		return user, nil
	}
	_ = s.emailSender.Send(
		email,
		"Подтвердите email",
		fmt.Sprintf("Ссылка: https://app.nerion.ru/auth/verify?token=%s", rawHex),
	)
	return user, nil
}

func (s *authService) Login(ctx context.Context, email, password string) (string, string, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, apierrors.ErrNotFound) {
			return "", "", apierrors.ErrUnauthorized
		}
		return "", "", err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", "", apierrors.ErrUnauthorized
	}

	accessToken, err := s.jm.Generate(user.ID, string(user.Role))
	if err != nil {
		return "", "", fmt.Errorf("generate access token: %w", err)
	}

	rawHex, raw, err := generateRawToken()
	if err != nil {
		return "", "", fmt.Errorf("generate refresh token: %w", err)
	}
	session := &entity.Session{
		ID:        newUUID(),
		UserID:    user.ID,
		TokenHash: hashToken(raw),
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
	}
	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return "", "", fmt.Errorf("create session: %w", err)
	}
	return accessToken, rawHex, nil
}

func (s *authService) Refresh(ctx context.Context, refreshToken string) (string, string, error) {
	raw, err := hex.DecodeString(refreshToken)
	if err != nil {
		return "", "", apierrors.ErrUnauthorized
	}
	hash := hashToken(raw)

	session, err := s.sessionRepo.GetByTokenHash(ctx, hash)
	if err != nil {
		if errors.Is(err, apierrors.ErrNotFound) {
			return "", "", apierrors.ErrUnauthorized
		}
		return "", "", err
	}
	if session.RevokedAt != nil || time.Now().After(session.ExpiresAt) {
		return "", "", apierrors.ErrUnauthorized
	}

	user, err := s.userRepo.GetByID(ctx, session.UserID)
	if err != nil {
		return "", "", err
	}

	if err := s.sessionRepo.RevokeByTokenHash(ctx, hash); err != nil {
		return "", "", fmt.Errorf("revoke old session: %w", err)
	}

	accessToken, err := s.jm.Generate(user.ID, string(user.Role))
	if err != nil {
		return "", "", fmt.Errorf("generate access token: %w", err)
	}

	newRawHex, newRaw, err := generateRawToken()
	if err != nil {
		return "", "", fmt.Errorf("generate refresh token: %w", err)
	}
	newSession := &entity.Session{
		ID:        newUUID(),
		UserID:    user.ID,
		TokenHash: hashToken(newRaw),
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
	}
	if err := s.sessionRepo.Create(ctx, newSession); err != nil {
		return "", "", fmt.Errorf("create new session: %w", err)
	}
	return accessToken, newRawHex, nil
}

func (s *authService) Logout(ctx context.Context, refreshToken string) error {
	raw, err := hex.DecodeString(refreshToken)
	if err != nil {
		return apierrors.ErrUnauthorized
	}
	err = s.sessionRepo.RevokeByTokenHash(ctx, hashToken(raw))
	if errors.Is(err, apierrors.ErrNotFound) {
		return apierrors.ErrUnauthorized
	}
	return err
}

func (s *authService) RequestPasswordReset(ctx context.Context, email string) error {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, apierrors.ErrNotFound) {
			return nil // no enumeration
		}
		return err
	}
	rawHex, raw, err := generateRawToken()
	if err != nil {
		s.logger.Error("generate pwd reset token", "err", err)
		return nil
	}
	tokenHash := hashToken(raw)
	if err := s.pwdResetRepo.Create(ctx, user.ID, tokenHash, time.Now().Add(60*time.Minute)); err != nil {
		s.logger.Error("store pwd reset token", "err", err)
		return nil
	}
	_ = s.emailSender.Send(
		email,
		"Сброс пароля",
		fmt.Sprintf("Ссылка: https://app.nerion.ru/auth/password/reset?token=%s", rawHex),
	)
	return nil
}

func (s *authService) ResetPassword(ctx context.Context, token, newPassword string) error {
	if len(newPassword) < 8 {
		return apierrors.ErrBadRequest
	}
	raw, err := hex.DecodeString(token)
	if err != nil {
		return apierrors.ErrBadRequest
	}
	hash := hashToken(raw)

	userID, expiresAt, usedAt, err := s.pwdResetRepo.GetByTokenHash(ctx, hash)
	if err != nil {
		if errors.Is(err, apierrors.ErrNotFound) {
			return apierrors.ErrBadRequest
		}
		return err
	}
	if usedAt != nil || time.Now().After(expiresAt) {
		return apierrors.ErrBadRequest
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return apierrors.ErrInternal
	}
	if err := s.userRepo.UpdatePassword(ctx, userID, string(passwordHash)); err != nil {
		return err
	}
	if err := s.pwdResetRepo.MarkUsed(ctx, hash); err != nil {
		s.logger.Error("mark pwd reset used", "err", err)
	}
	if err := s.sessionRepo.RevokeAllByUserID(ctx, userID); err != nil {
		s.logger.Error("revoke all sessions after pwd reset", "err", err)
	}
	return nil
}

func (s *authService) VerifyEmail(ctx context.Context, token string) error {
	raw, err := hex.DecodeString(token)
	if err != nil {
		return apierrors.ErrBadRequest
	}
	hash := hashToken(raw)

	userID, expiresAt, usedAt, err := s.emailVerifRepo.GetByTokenHash(ctx, hash)
	if err != nil {
		if errors.Is(err, apierrors.ErrNotFound) {
			return apierrors.ErrBadRequest
		}
		return err
	}
	if usedAt != nil || time.Now().After(expiresAt) {
		return apierrors.ErrBadRequest
	}

	if err := s.userRepo.SetEmailVerified(ctx, userID); err != nil {
		return err
	}
	if err := s.emailVerifRepo.MarkUsed(ctx, hash); err != nil {
		s.logger.Error("mark email verif used", "err", err)
	}
	return nil
}

func (s *authService) GetMe(ctx context.Context, userID int64) (*entity.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}

func (s *authService) ChangePassword(ctx context.Context, userID int64, currentPassword, newPassword string) error {
	if len(newPassword) < 8 {
		return apierrors.NewError(400, "validation_error", "Новый пароль должен быть не менее 8 символов")
	}
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(currentPassword)); err != nil {
		return apierrors.NewError(400, "wrong_password", "Текущий пароль неверен")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return apierrors.ErrInternal
	}
	if err := s.userRepo.UpdatePassword(ctx, userID, string(hash)); err != nil {
		return err
	}
	// Revoke all sessions so other devices are logged out.
	return s.sessionRepo.RevokeAllByUserID(ctx, userID)
}
