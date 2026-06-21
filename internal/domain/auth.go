package domain

import (
	"context"
	"time"

	"nerion/internal/entity"
)

type SessionRepository interface {
	Create(ctx context.Context, s *entity.Session) error
	GetByTokenHash(ctx context.Context, hash string) (*entity.Session, error)
	RevokeByTokenHash(ctx context.Context, hash string) error
	RevokeAllByUserID(ctx context.Context, userID int64) error
}

type EmailVerificationRepository interface {
	Create(ctx context.Context, userID int64, tokenHash string, expiresAt time.Time) error
	GetByTokenHash(ctx context.Context, hash string) (userID int64, expiresAt time.Time, usedAt *time.Time, err error)
	MarkUsed(ctx context.Context, tokenHash string) error
}

type PasswordResetRepository interface {
	Create(ctx context.Context, userID int64, tokenHash string, expiresAt time.Time) error
	GetByTokenHash(ctx context.Context, hash string) (userID int64, expiresAt time.Time, usedAt *time.Time, err error)
	MarkUsed(ctx context.Context, tokenHash string) error
}

type AuthService interface {
	Register(ctx context.Context, name, email, password string) (*entity.User, error)
	Login(ctx context.Context, email, password string) (accessToken, refreshToken string, err error)
	Refresh(ctx context.Context, refreshToken string) (newAccessToken, newRefreshToken string, err error)
	Logout(ctx context.Context, refreshToken string) error
	RequestPasswordReset(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, token, newPassword string) error
	VerifyEmail(ctx context.Context, token string) error
	GetMe(ctx context.Context, userID int64) (*entity.User, error)
	ChangePassword(ctx context.Context, userID int64, currentPassword, newPassword string) error
}
