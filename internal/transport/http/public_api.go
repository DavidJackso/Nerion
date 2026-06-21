package thttp

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"nerion/internal/entity"
	authmw "nerion/internal/middleware"
)

func (s *Server) publicAPIRoutes() chi.Router {
	r := chi.NewRouter()
	r.Use(authmw.APIKeyAuth(s.apiKeyRepo, s.auditRepo))

	r.Get("/{space}/openapi.json", s.publicOpenAPI)

	r.Get("/{space}/{table}", s.publicListRecords)
	r.Post("/{space}/{table}", s.publicCreateRecord)
	r.Get("/{space}/{table}/{id}", s.publicGetRecord)
	r.Put("/{space}/{table}/{id}", s.publicUpdateRecord)
	r.Delete("/{space}/{table}/{id}", s.publicDeleteRecord)

	return r
}

// resolvePublicTable resolves space + verifies key belongs to space + resolves table + fields.
func (s *Server) resolvePublicTable(w http.ResponseWriter, r *http.Request) (*entity.Space, *entity.TableMeta, []*entity.FieldMeta, bool) {
	key, _ := authmw.APIKeyFrom(r.Context())
	spaceSlug := chi.URLParam(r, "space")
	tableSlug := chi.URLParam(r, "table")

	space, err := s.spaceRepo.GetBySlug(r.Context(), spaceSlug)
	if err != nil {
		writeJSON(w, http.StatusNotFound, errBody("not_found", "Пространство не найдено"))
		return nil, nil, nil, false
	}
	if space.ID != key.SpaceID {
		writeJSON(w, http.StatusForbidden, errBody("forbidden", "API ключ не принадлежит данному пространству"))
		return nil, nil, nil, false
	}
	t, err := s.tableRepo.GetBySlug(r.Context(), space.ID, tableSlug)
	if err != nil {
		writeJSON(w, http.StatusNotFound, errBody("not_found", "Таблица не найдена"))
		return nil, nil, nil, false
	}
	fields, err := s.fieldRepo.ListByTable(r.Context(), t.ID)
	if err != nil {
		s.writeError(w, r, err)
		return nil, nil, nil, false
	}
	return space, t, fields, true
}

func (s *Server) requireWriteScope(w http.ResponseWriter, r *http.Request) bool {
	key, _ := authmw.APIKeyFrom(r.Context())
	if key.Scope == "read" {
		writeJSON(w, http.StatusForbidden, errBody("forbidden", "API ключ имеет права только на чтение"))
		return false
	}
	return true
}

func (s *Server) publicListRecords(w http.ResponseWriter, r *http.Request) {
	space, _, fields, ok := s.resolvePublicTable(w, r)
	if !ok {
		return
	}
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))

	params := entity.ListParams{
		Limit:   limit,
		Offset:  offset,
		SortBy:  q.Get("sort_by"),
		SortDir: q.Get("sort_dir"),
		Search:  q.Get("search"),
	}
	tableSlug := chi.URLParam(r, "table")
	records, total, err := s.recordRepo.List(r.Context(), space.Slug, tableSlug, fields, params)
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	if records == nil {
		records = []map[string]any{}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"data": records,
		"meta": map[string]any{"total": total, "limit": params.Limit, "offset": params.Offset},
	})
}

func (s *Server) publicGetRecord(w http.ResponseWriter, r *http.Request) {
	space, _, fields, ok := s.resolvePublicTable(w, r)
	if !ok {
		return
	}
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errBody("bad_request", "Некорректный ID"))
		return
	}
	tableSlug := chi.URLParam(r, "table")
	rec, err := s.recordRepo.GetByID(r.Context(), space.Slug, tableSlug, fields, id)
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, rec)
}

func (s *Server) publicCreateRecord(w http.ResponseWriter, r *http.Request) {
	if !s.requireWriteScope(w, r) {
		return
	}
	space, _, fields, ok := s.resolvePublicTable(w, r)
	if !ok {
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 4<<20)
	var data map[string]any
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		writeJSON(w, http.StatusBadRequest, errBody("bad_request", "Некорректное тело запроса"))
		return
	}
	tableSlug := chi.URLParam(r, "table")
	rec, err := s.recordRepo.Create(r.Context(), space.Slug, tableSlug, fields, data)
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, rec)
}

func (s *Server) publicUpdateRecord(w http.ResponseWriter, r *http.Request) {
	if !s.requireWriteScope(w, r) {
		return
	}
	space, _, fields, ok := s.resolvePublicTable(w, r)
	if !ok {
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 4<<20)
	var data map[string]any
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		writeJSON(w, http.StatusBadRequest, errBody("bad_request", "Некорректное тело запроса"))
		return
	}
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errBody("bad_request", "Некорректный ID"))
		return
	}
	tableSlug := chi.URLParam(r, "table")
	rec, err := s.recordRepo.Update(r.Context(), space.Slug, tableSlug, fields, id, data)
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, rec)
}

func (s *Server) publicDeleteRecord(w http.ResponseWriter, r *http.Request) {
	if !s.requireWriteScope(w, r) {
		return
	}
	space, _, _, ok := s.resolvePublicTable(w, r)
	if !ok {
		return
	}
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errBody("bad_request", "Некорректный ID"))
		return
	}
	tableSlug := chi.URLParam(r, "table")
	if err := s.recordRepo.Delete(r.Context(), space.Slug, tableSlug, id); err != nil {
		s.writeError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- OpenAPI 3.0 generation ---

func (s *Server) publicOpenAPI(w http.ResponseWriter, r *http.Request) {
	key, _ := authmw.APIKeyFrom(r.Context())
	spaceSlug := chi.URLParam(r, "space")

	space, err := s.spaceRepo.GetBySlug(r.Context(), spaceSlug)
	if err != nil {
		writeJSON(w, http.StatusNotFound, errBody("not_found", "Пространство не найдено"))
		return
	}
	if space.ID != key.SpaceID {
		writeJSON(w, http.StatusForbidden, errBody("forbidden", "Доступ запрещён"))
		return
	}

	tables, err := s.tableRepo.ListBySpace(r.Context(), space.ID)
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	for _, t := range tables {
		t.Fields, _ = s.fieldRepo.ListByTable(r.Context(), t.ID)
	}

	spec := buildOpenAPISpec(space, tables)
	writeJSON(w, http.StatusOK, spec)
}

func fieldTypeToOAPI(ft entity.FieldType) map[string]any {
	switch ft {
	case entity.FieldTypeNumber:
		return map[string]any{"type": "number"}
	case entity.FieldTypeBoolean:
		return map[string]any{"type": "boolean"}
	case entity.FieldTypeDate:
		return map[string]any{"type": "string", "format": "date"}
	case entity.FieldTypeDatetime:
		return map[string]any{"type": "string", "format": "date-time"}
	case entity.FieldTypeRelation:
		return map[string]any{"type": "integer"}
	case entity.FieldTypeEnum:
		return map[string]any{"type": "string"} // enum values added separately
	default:
		return map[string]any{"type": "string"}
	}
}

func buildOpenAPISpec(space *entity.Space, tables []*entity.TableMeta) map[string]any {
	paths := map[string]any{}
	schemas := map[string]any{}
	baseURL := fmt.Sprintf("/api/%s", space.Slug)

	paginationParams := []map[string]any{
		{"name": "limit", "in": "query", "schema": map[string]any{"type": "integer", "default": 50}},
		{"name": "offset", "in": "query", "schema": map[string]any{"type": "integer", "default": 0}},
		{"name": "sort_by", "in": "query", "schema": map[string]any{"type": "string"}},
		{"name": "sort_dir", "in": "query", "schema": map[string]any{"type": "string", "enum": []string{"asc", "desc"}}},
		{"name": "search", "in": "query", "schema": map[string]any{"type": "string"}},
	}

	for _, t := range tables {
		props := map[string]any{}
		inputProps := map[string]any{}
		required := []string{}

		for _, f := range t.Fields {
			schema := fieldTypeToOAPI(f.Type)
			if f.Type == entity.FieldTypeEnum && len(f.EnumValues) > 0 {
				schema["enum"] = f.EnumValues
			}
			props[f.Slug] = schema
			inputProps[f.Slug] = schema
			if f.Required {
				required = append(required, f.Slug)
			}
		}

		// Output schema includes system fields
		outSchema := map[string]any{
			"type": "object",
			"properties": mergeMaps(map[string]any{
				"id":         map[string]any{"type": "integer"},
				"created_at": map[string]any{"type": "string", "format": "date-time"},
				"updated_at": map[string]any{"type": "string", "format": "date-time"},
			}, props),
		}
		inSchema := map[string]any{"type": "object", "properties": inputProps}
		if len(required) > 0 {
			inSchema["required"] = required
		}

		schemas[t.Slug] = outSchema
		schemas[t.Slug+"Input"] = inSchema

		ref := fmt.Sprintf("#/components/schemas/%s", t.Slug)
		refIn := fmt.Sprintf("#/components/schemas/%sInput", t.Slug)

		listResp := map[string]any{
			"200": map[string]any{
				"description": "OK",
				"content": map[string]any{"application/json": map[string]any{
					"schema": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"data": map[string]any{"type": "array", "items": map[string]any{"$ref": ref}},
							"meta": map[string]any{"type": "object", "properties": map[string]any{
								"total":  map[string]any{"type": "integer"},
								"limit":  map[string]any{"type": "integer"},
								"offset": map[string]any{"type": "integer"},
							}},
						},
					},
				}},
			},
		}

		paths["/"+t.Slug] = map[string]any{
			"get": map[string]any{
				"summary":    fmt.Sprintf("List %s", t.Name),
				"parameters": paginationParams,
				"responses":  listResp,
			},
			"post": map[string]any{
				"summary": fmt.Sprintf("Create %s", t.Name),
				"requestBody": map[string]any{
					"required": true,
					"content": map[string]any{"application/json": map[string]any{
						"schema": map[string]any{"$ref": refIn},
					}},
				},
				"responses": map[string]any{
					"201": map[string]any{"description": "Created", "content": map[string]any{"application/json": map[string]any{"schema": map[string]any{"$ref": ref}}}},
				},
			},
		}

		idParam := map[string]any{"name": "id", "in": "path", "required": true, "schema": map[string]any{"type": "integer"}}
		paths["/"+t.Slug+"/{id}"] = map[string]any{
			"get": map[string]any{
				"summary":    fmt.Sprintf("Get %s", t.Name),
				"parameters": []map[string]any{idParam},
				"responses": map[string]any{
					"200": map[string]any{"description": "OK", "content": map[string]any{"application/json": map[string]any{"schema": map[string]any{"$ref": ref}}}},
				},
			},
			"put": map[string]any{
				"summary":    fmt.Sprintf("Update %s", t.Name),
				"parameters": []map[string]any{idParam},
				"requestBody": map[string]any{
					"required": true,
					"content": map[string]any{"application/json": map[string]any{"schema": map[string]any{"$ref": refIn}}},
				},
				"responses": map[string]any{
					"200": map[string]any{"description": "OK", "content": map[string]any{"application/json": map[string]any{"schema": map[string]any{"$ref": ref}}}},
				},
			},
			"delete": map[string]any{
				"summary":    fmt.Sprintf("Delete %s", t.Name),
				"parameters": []map[string]any{idParam},
				"responses":  map[string]any{"204": map[string]any{"description": "No Content"}},
			},
		}
	}

	return map[string]any{
		"openapi": "3.0.3",
		"info":    map[string]any{"title": fmt.Sprintf("Nerion API — %s", space.Name), "version": "1.0.0"},
		"servers": []map[string]any{{"url": baseURL}},
		"paths":   paths,
		"components": map[string]any{
			"schemas": schemas,
			"securitySchemes": map[string]any{
				"ApiKeyAuth": map[string]any{"type": "apiKey", "in": "header", "name": "X-Api-Key"},
			},
		},
		"security": []map[string]any{{"ApiKeyAuth": []string{}}},
	}
}

func mergeMaps(a, b map[string]any) map[string]any {
	out := make(map[string]any, len(a)+len(b))
	for k, v := range a {
		out[k] = v
	}
	for k, v := range b {
		out[k] = v
	}
	return out
}

// --- API status endpoint (7.4) — requires JWT auth, lives here for locality ---

func (s *Server) apiStatus(w http.ResponseWriter, r *http.Request) {
	claims, _ := authmw.ClaimsFrom(r.Context())
	spaceSlug := chi.URLParam(r, "slug")
	space, err := s.spaceRepo.GetBySlug(r.Context(), spaceSlug)
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	if _, err := s.memberRepo.GetRole(r.Context(), space.ID, claims.UserID); err != nil {
		writeJSON(w, http.StatusForbidden, errBody("forbidden", "Доступ запрещён"))
		return
	}

	tables, err := s.tableRepo.ListBySpace(r.Context(), space.ID)
	if err != nil {
		s.writeError(w, r, err)
		return
	}

	keys, err := s.apiKeyRepo.ListBySpace(r.Context(), space.ID)
	if err != nil {
		s.writeError(w, r, err)
		return
	}

	activeKeys := 0
	for _, k := range keys {
		if k.RevokedAt == nil {
			activeKeys++
		}
	}

	tableSlugs := make([]string, 0, len(tables))
	for _, t := range tables {
		tableSlugs = append(tableSlugs, t.Slug)
	}

	status := "online"
	if activeKeys == 0 {
		status = "offline"
	} else if len(tables) == 0 {
		status = "setup"
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"status":      status,
		"base_url":    fmt.Sprintf("/api/%s", spaceSlug),
		"active_keys": activeKeys,
		"tables":      tableSlugs,
	})
}
