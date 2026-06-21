package thttp

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"nerion/internal/entity"
	authmw "nerion/internal/middleware"
)

// Protected list management routes (JWT required, admin only for write).
func (s *Server) listRoutes() chi.Router {
	r := chi.NewRouter()
	r.Use(authmw.Auth(s.jwtManager))
	r.Use(authmw.LoadSpace(s.spaceRepo))

	r.Get("/", s.listLists)
	r.Post("/", s.createList)
	r.Put("/{listSlug}", s.updateList)

	return r
}

// Public list route — no auth, CORS allowed.
func (s *Server) publicListRoutes() chi.Router {
	r := chi.NewRouter()
	r.Use(corsAllowAll)

	r.Options("/*", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	r.Get("/{space}/{listSlug}", s.getPublicList)

	return r
}

func corsAllowAll(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		next.ServeHTTP(w, r)
	})
}

// --- Response types ---

type listResponse struct {
	ID            int64                    `json:"id"`
	Slug          string                   `json:"slug"`
	TableSlug     string                   `json:"table_slug"`
	FieldConfig   []entity.ListFieldConfig `json:"field_config"`
	FilterConfig  map[string]any           `json:"filter_config"`
	SortConfig    []entity.ListSortConfig  `json:"sort_config"`
	RowLimit      int                      `json:"row_limit"`
	Published     bool                     `json:"published"`
	PublishedAt   *time.Time               `json:"published_at"`
	UnpublishedAt *time.Time               `json:"unpublished_at"`
	CreatedAt     time.Time                `json:"created_at"`
}

func toListResponse(l *entity.List) listResponse {
	return listResponse{
		ID:            l.ID,
		Slug:          l.Slug,
		TableSlug:     l.TableSlug,
		FieldConfig:   l.FieldConfig,
		FilterConfig:  l.FilterConfig,
		SortConfig:    l.SortConfig,
		RowLimit:      l.RowLimit,
		Published:     l.IsPublished(),
		PublishedAt:   l.PublishedAt,
		UnpublishedAt: l.UnpublishedAt,
		CreatedAt:     l.CreatedAt,
	}
}

// --- Handlers ---

func (s *Server) listLists(w http.ResponseWriter, r *http.Request) {
	claims, _ := authmw.ClaimsFrom(r.Context())
	space, _ := authmw.SpaceFrom(r.Context())

	lists, err := s.listService.List(r.Context(), space.Slug, claims.UserID)
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	resp := make([]listResponse, 0, len(lists))
	for _, l := range lists {
		resp = append(resp, toListResponse(l))
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) createList(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req createListRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errBody("bad_request", "Некорректное тело запроса"))
		return
	}
	claims, _ := authmw.ClaimsFrom(r.Context())
	space, _ := authmw.SpaceFrom(r.Context())

	cfg := entity.ListConfig{
		FieldConfig:  req.FieldConfig,
		FilterConfig: req.FilterConfig,
		SortConfig:   req.SortConfig,
		RowLimit:     req.RowLimit,
	}
	l, err := s.listService.Create(r.Context(), space.Slug, req.TableSlug, req.Slug, claims.UserID, cfg, req.Published)
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, toListResponse(l))
}

func (s *Server) updateList(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req updateListRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errBody("bad_request", "Некорректное тело запроса"))
		return
	}
	claims, _ := authmw.ClaimsFrom(r.Context())
	space, _ := authmw.SpaceFrom(r.Context())
	listSlug := chi.URLParam(r, "listSlug")

	var cfg *entity.ListConfig
	if req.FieldConfig != nil || req.FilterConfig != nil || req.SortConfig != nil || req.RowLimit != nil {
		cfg = &entity.ListConfig{
			FieldConfig:  req.FieldConfig,
			FilterConfig: req.FilterConfig,
			SortConfig:   req.SortConfig,
		}
		if req.RowLimit != nil {
			cfg.RowLimit = *req.RowLimit
		}
	}

	l, err := s.listService.Update(r.Context(), space.Slug, listSlug, claims.UserID, cfg, req.Published)
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toListResponse(l))
}

func (s *Server) getPublicList(w http.ResponseWriter, r *http.Request) {
	spaceSlug := chi.URLParam(r, "space")
	listSlug := chi.URLParam(r, "listSlug")

	records, err := s.listService.GetPublicData(r.Context(), spaceSlug, listSlug)
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": records})
}
