package entity

import "time"

type AuditEntry struct {
	ID         int64
	SpaceID    *int64
	UserID     *int64
	Action     string
	EntityType string
	EntityID   string
	Meta       map[string]any
	CreatedAt  time.Time
}
