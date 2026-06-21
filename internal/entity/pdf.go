package entity

import "time"

type PDFTemplate struct {
	ID           int64
	SpaceID      int64
	Name         string
	StoragePath  string
	Placeholders []string
	Status       string // needs_mapping | ready | error
	CreatedAt    time.Time
}

type PDFMapping struct {
	ID            int64
	TemplateID    int64
	Placeholder   string
	SourceFieldID *int64
	Expression    *string
}

type PDFJob struct {
	ID           string // UUID
	SpaceID      int64
	TemplateID   int64
	Status       string // pending | processing | done | error
	TotalRecords *int
	Processed    int
	StoragePath  *string
	CreatedBy    int64
	CreatedAt    time.Time
	CompletedAt  *time.Time
}
