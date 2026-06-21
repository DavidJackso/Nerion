package entity

import "time"

type APIKey struct {
	ID         int64
	SpaceID    int64
	Name       string
	KeyPrefix  string
	Scope      string
	CreatedAt  time.Time
	LastUsedAt *time.Time
	RevokedAt  *time.Time
}

func (k *APIKey) Status() string {
	if k.RevokedAt != nil {
		return "revoked"
	}
	return "active"
}
