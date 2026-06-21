package template

import "nerion/internal/entity"

type Template struct {
	ID          string
	Name        string
	Description string
	Fields      []*entity.FieldMeta
}

func ptr(s string) *string { return &s }

var Catalog = []Template{
	{
		ID:          "contacts",
		Name:        "Контакты",
		Description: "Список контактов: имя, email, телефон, компания",
		Fields: []*entity.FieldMeta{
			{Name: "Имя", Slug: "name", Type: entity.FieldTypeText, Required: true},
			{Name: "Email", Slug: "email", Type: entity.FieldTypeEmail},
			{Name: "Телефон", Slug: "phone", Type: entity.FieldTypePhone},
			{Name: "Компания", Slug: "company", Type: entity.FieldTypeText},
			{Name: "Заметки", Slug: "notes", Type: entity.FieldTypeLongtext},
		},
	},
	{
		ID:          "tasks",
		Name:        "Задачи",
		Description: "Трекер задач: название, статус, срок, приоритет",
		Fields: []*entity.FieldMeta{
			{Name: "Название", Slug: "title", Type: entity.FieldTypeText, Required: true},
			{Name: "Описание", Slug: "description", Type: entity.FieldTypeLongtext},
			{
				Name: "Статус", Slug: "status", Type: entity.FieldTypeEnum,
				EnumValues: []string{"new", "in_progress", "done", "cancelled"},
				DefaultValue: ptr("new"),
			},
			{
				Name: "Приоритет", Slug: "priority", Type: entity.FieldTypeEnum,
				EnumValues: []string{"low", "medium", "high"},
				DefaultValue: ptr("medium"),
			},
			{Name: "Срок", Slug: "due_date", Type: entity.FieldTypeDate},
		},
	},
	{
		ID:          "inventory",
		Name:        "Склад",
		Description: "Учёт товаров: название, SKU, количество, цена",
		Fields: []*entity.FieldMeta{
			{Name: "Название", Slug: "name", Type: entity.FieldTypeText, Required: true},
			{Name: "SKU", Slug: "sku", Type: entity.FieldTypeText, Unique: true},
			{Name: "Количество", Slug: "quantity", Type: entity.FieldTypeNumber},
			{Name: "Цена", Slug: "price", Type: entity.FieldTypeNumber},
			{Name: "Описание", Slug: "description", Type: entity.FieldTypeLongtext},
		},
	},
	{
		ID:          "events",
		Name:        "События",
		Description: "Мероприятия: название, дата, место, описание",
		Fields: []*entity.FieldMeta{
			{Name: "Название", Slug: "title", Type: entity.FieldTypeText, Required: true},
			{Name: "Дата начала", Slug: "start_at", Type: entity.FieldTypeDatetime, Required: true},
			{Name: "Дата окончания", Slug: "end_at", Type: entity.FieldTypeDatetime},
			{Name: "Место", Slug: "location", Type: entity.FieldTypeText},
			{Name: "Описание", Slug: "description", Type: entity.FieldTypeLongtext},
		},
	},
	{
		ID:          "clients",
		Name:        "Клиенты",
		Description: "CRM: клиенты с контактами и статусом",
		Fields: []*entity.FieldMeta{
			{Name: "Имя", Slug: "name", Type: entity.FieldTypeText, Required: true},
			{Name: "Email", Slug: "email", Type: entity.FieldTypeEmail, Unique: true},
			{Name: "Телефон", Slug: "phone", Type: entity.FieldTypePhone},
			{
				Name: "Статус", Slug: "status", Type: entity.FieldTypeEnum,
				EnumValues:   []string{"lead", "active", "inactive"},
				DefaultValue: ptr("lead"),
			},
			{Name: "Сайт", Slug: "website", Type: entity.FieldTypeURL},
			{Name: "Заметки", Slug: "notes", Type: entity.FieldTypeLongtext},
		},
	},
}

func FindByID(id string) *Template {
	for i := range Catalog {
		if Catalog[i].ID == id {
			return &Catalog[i]
		}
	}
	return nil
}
