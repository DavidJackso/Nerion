package entity

import "time"

type FieldType string

const (
	FieldTypeText     FieldType = "text"
	FieldTypeLongtext FieldType = "longtext"
	FieldTypeNumber   FieldType = "number"
	FieldTypeDate     FieldType = "date"
	FieldTypeDatetime FieldType = "datetime"
	FieldTypeBoolean  FieldType = "boolean"
	FieldTypeEnum     FieldType = "enum"
	FieldTypeEmail    FieldType = "email"
	FieldTypePhone    FieldType = "phone"
	FieldTypeURL      FieldType = "url"
	FieldTypeFile     FieldType = "file"
	FieldTypeRelation FieldType = "relation"
)

type TableMeta struct {
	ID        int64
	SpaceID   int64
	Name      string
	Slug      string
	CreatedAt time.Time
	Fields    []*FieldMeta
}

type FieldMeta struct {
	ID                  int64
	TableID             int64
	Name                string
	Slug                string
	Type                FieldType
	Required            bool
	DefaultValue        *string
	Unique              bool
	EnumValues          []string
	RelationTableID     *int64
	RelationCardinality *string
	Position            int
}
