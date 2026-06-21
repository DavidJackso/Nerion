package entity

import "time"

type Space struct {
	ID        int64
	Name      string
	Slug      string
	OwnerID   int64
	CreatedAt time.Time
}

type SpaceMemberRole string

const (
	SpaceMemberRoleAdmin  SpaceMemberRole = "admin"
	SpaceMemberRoleMember SpaceMemberRole = "member"
)

type SpaceMember struct {
	SpaceID  int64
	UserID   int64
	UserName string
	UserEmail string
	Role     SpaceMemberRole
	JoinedAt time.Time
}
