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
	FieldTypeFiles    FieldType = "files"
	FieldTypeRelation FieldType = "relation"
)

type TableMeta struct {
	ID        int64       `json:"id"`
	SpaceID   int64       `json:"space_id"`
	Name      string      `json:"name"`
	Slug      string      `json:"slug"`
	CreatedAt time.Time   `json:"created_at"`
	Fields    []*FieldMeta `json:"fields,omitempty"`
}

type FieldMeta struct {
	ID                  int64      `json:"id"`
	TableID             int64      `json:"table_id"`
	Name                string     `json:"name"`
	Slug                string     `json:"slug"`
	Type                FieldType  `json:"type"`
	Required            bool       `json:"required"`
	DefaultValue        *string    `json:"default_value,omitempty"`
	Unique              bool       `json:"unique"`
	EnumValues          []string   `json:"enum_values,omitempty"`
	RelationTableID     *int64     `json:"relation_table_id,omitempty"`
	RelationCardinality *string    `json:"relation_cardinality,omitempty"`
	RelationTarget      *string    `json:"relation_target,omitempty"`
	Position            int        `json:"position"`
}
