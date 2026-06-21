package entity

import "time"

type ListFieldConfig struct {
	FieldSlug  string `json:"field_slug"`
	PublicName string `json:"public_name"`
}

type ListSortConfig struct {
	FieldSlug string `json:"field_slug"`
	Direction string `json:"direction"` // "asc" | "desc"
}

type ListConfig struct {
	FieldConfig  []ListFieldConfig
	FilterConfig map[string]any
	SortConfig   []ListSortConfig
	RowLimit     int
}

type List struct {
	ID            int64
	SpaceID       int64
	Slug          string
	SourceTableID int64
	TableSlug     string // populated via join
	ListConfig
	PublishedAt   *time.Time
	UnpublishedAt *time.Time
	CreatedAt     time.Time
}

func (l *List) IsPublished() bool {
	return l.PublishedAt != nil && l.UnpublishedAt == nil
}
