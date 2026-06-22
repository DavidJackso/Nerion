package thttp

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"nerion/internal/entity"
	authmw "nerion/internal/middleware"
	tmpl "nerion/internal/template"
)

func (s *Server) schemaRoutes() chi.Router {
	r := chi.NewRouter()
	r.Use(authmw.Auth(s.jwtManager))

	r.Get("/spaces/{slug}/tables", s.listTables)
	r.Post("/spaces/{slug}/tables", s.createTable)
	r.Get("/spaces/{slug}/tables/{table}", s.getTable)
	r.Put("/spaces/{slug}/tables/{table}/fields", s.updateFields)
	r.Delete("/spaces/{slug}/tables/{table}", s.deleteTable)

	return r
}

func (s *Server) listTables(w http.ResponseWriter, r *http.Request) {
	claims, _ := authmw.ClaimsFrom(r.Context())
	slug := chi.URLParam(r, "slug")
	tables, err := s.schemaService.ListTables(r.Context(), slug, claims.UserID)
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, tables)
}

func (s *Server) createTable(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req createTableRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errBody("bad_request", "Некорректное тело запроса"))
		return
	}
	claims, _ := authmw.ClaimsFrom(r.Context())
	slug := chi.URLParam(r, "slug")

	// Resolve fields from template if requested
	var templateFields []*entity.FieldMeta
	if req.TemplateID != "" {
		t := tmpl.FindByID(req.TemplateID)
		if t == nil {
			writeJSON(w, http.StatusBadRequest, errBody("not_found", "Шаблон не найден"))
			return
		}
		templateFields = t.Fields
	}

	table, err := s.schemaService.CreateTable(r.Context(), slug, req.Name, req.Slug, claims.UserID)
	if err != nil {
		s.writeError(w, r, err)
		return
	}

	if len(templateFields) > 0 {
		if err := s.schemaService.UpdateFields(r.Context(), slug, req.Slug, claims.UserID, templateFields); err != nil {
			s.writeError(w, r, err)
			return
		}
		// Reload table with fields
		if full, err := s.schemaService.GetTable(r.Context(), slug, req.Slug, claims.UserID); err == nil {
			table = full
		}
	}

	if space, _ := authmw.SpaceFrom(r.Context()); space != nil {
		s.auditForSpace(space.ID, claims.UserID, "schema.table_create", "table", table.Slug, map[string]any{"name": table.Name})
	}
	writeJSON(w, http.StatusCreated, table)
}

func (s *Server) listTemplates(w http.ResponseWriter, r *http.Request) {
	type templateItem struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	items := make([]templateItem, len(tmpl.Catalog))
	for i, t := range tmpl.Catalog {
		items[i] = templateItem{ID: t.ID, Name: t.Name, Description: t.Description}
	}
	writeJSON(w, http.StatusOK, items)
}

func (s *Server) getTable(w http.ResponseWriter, r *http.Request) {
	claims, _ := authmw.ClaimsFrom(r.Context())
	slug := chi.URLParam(r, "slug")
	tableSlug := chi.URLParam(r, "table")
	t, err := s.schemaService.GetTable(r.Context(), slug, tableSlug, claims.UserID)
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, t)
}

func (s *Server) updateFields(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req updateFieldsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errBody("bad_request", "Некорректное тело запроса"))
		return
	}
	claims, _ := authmw.ClaimsFrom(r.Context())
	slug := chi.URLParam(r, "slug")
	tableSlug := chi.URLParam(r, "table")

	fields := make([]*entity.FieldMeta, len(req.Fields))
	for i, f := range req.Fields {
		fields[i] = &entity.FieldMeta{
			Name:                f.Name,
			Slug:                f.Slug,
			Type:                entity.FieldType(f.Type),
			Required:            f.Required,
			DefaultValue:        f.DefaultValue,
			Unique:              f.Unique,
			EnumValues:          f.EnumValues,
			RelationTableID:     f.RelationTableID,
			RelationCardinality: f.RelationCardinality,
		}
	}

	if err := s.schemaService.UpdateFields(r.Context(), slug, tableSlug, claims.UserID, fields); err != nil {
		s.writeError(w, r, err)
		return
	}
	space, _ := authmw.SpaceFrom(r.Context())
	s.auditForSpace(space.ID, claims.UserID, "schema.fields_update", "table", tableSlug, map[string]any{"field_count": len(fields)})
	writeJSON(w, http.StatusOK, map[string]string{"message": "Поля обновлены"})
}

func (s *Server) deleteTable(w http.ResponseWriter, r *http.Request) {
	claims, _ := authmw.ClaimsFrom(r.Context())
	slug := chi.URLParam(r, "slug")
	tableSlug := chi.URLParam(r, "table")
	if err := s.schemaService.DeleteTable(r.Context(), slug, tableSlug, claims.UserID); err != nil {
		s.writeError(w, r, err)
		return
	}
	space, _ := authmw.SpaceFrom(r.Context())
	s.auditForSpace(space.ID, claims.UserID, "schema.table_delete", "table", tableSlug, nil)
	w.WriteHeader(http.StatusNoContent)
}
