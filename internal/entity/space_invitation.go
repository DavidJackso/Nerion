package entity

import "time"

type SpaceInvitation struct {
	ID        string
	SpaceID   int64
	SpaceName string
	Email     string
	InvitedBy int64
	TokenHash string
	CreatedAt time.Time
	ExpiresAt time.Time
	UsedAt    *time.Time
}
