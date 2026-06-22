package thttp

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	"nerion/internal/domain"
	"nerion/internal/jwtauth"
	authmw "nerion/internal/middleware"
	"nerion/pkg/apierrors"
)

type Server struct {
	userService   domain.UserService
	authService   domain.AuthService
	spaceService  domain.SpaceService
	memberService domain.SpaceMemberService
	schemaService domain.SchemaService
	recordService domain.RecordService
	apiKeyService domain.APIKeyService
	listService   domain.ListService
	pdfService    domain.PDFService
	storage       domain.StorageAdapter
	presignTTL    time.Duration
	spaceRepo     domain.SpaceRepository
	memberRepo    domain.SpaceMemberRepository
	tableRepo     domain.TableRepository
	fieldRepo     domain.FieldRepository
	recordRepo    domain.RecordRepository
	apiKeyRepo    domain.APIKeyRepository
	auditRepo     domain.AuditRepository
	jwtManager    jwtauth.Tokenizer
	logger        *slog.Logger
	router        chi.Router
}

type ServerConfig struct {
	UserService   domain.UserService
	AuthService   domain.AuthService
	SpaceService  domain.SpaceService
	MemberService domain.SpaceMemberService
	SchemaService domain.SchemaService
	RecordService domain.RecordService
	APIKeyService domain.APIKeyService
	ListService   domain.ListService
	PDFService    domain.PDFService
	Storage       domain.StorageAdapter
	PresignTTL    time.Duration
	SpaceRepo     domain.SpaceRepository
	MemberRepo    domain.SpaceMemberRepository
	TableRepo     domain.TableRepository
	FieldRepo     domain.FieldRepository
	RecordRepo    domain.RecordRepository
	APIKeyRepo    domain.APIKeyRepository
	AuditRepo     domain.AuditRepository
	JWTManager    jwtauth.Tokenizer
	Logger        *slog.Logger
	CORSOrigins   []string
}

func NewServer(cfg ServerConfig) *Server {
	s := &Server{
		userService:   cfg.UserService,
		authService:   cfg.AuthService,
		spaceService:  cfg.SpaceService,
		memberService: cfg.MemberService,
		schemaService: cfg.SchemaService,
		recordService: cfg.RecordService,
		apiKeyService: cfg.APIKeyService,
		listService:   cfg.ListService,
		pdfService:    cfg.PDFService,
		storage:       cfg.Storage,
		presignTTL:    cfg.PresignTTL,
		spaceRepo:     cfg.SpaceRepo,
		memberRepo:    cfg.MemberRepo,
		tableRepo:     cfg.TableRepo,
		fieldRepo:     cfg.FieldRepo,
		recordRepo:    cfg.RecordRepo,
		apiKeyRepo:    cfg.APIKeyRepo,
		auditRepo:     cfg.AuditRepo,
		jwtManager:    cfg.JWTManager,
		logger:        cfg.Logger,
		router:        chi.NewRouter(),
	}
	s.router.Use(
		chimw.RequestID,
		chimw.RealIP,
		authmw.RequestLogger(cfg.Logger),
		authmw.SecurityHeaders,
		authmw.CORS(cfg.CORSOrigins),
		chimw.Recoverer,
		chimw.CleanPath,
		chimw.Timeout(8*time.Second),
	)
	s.registerRoutes()
	return s
}


func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *Server) registerRoutes() {
	s.router.Mount("/auth", s.authRoutes())
	s.router.Mount("/users", s.userRoutes())
	s.router.Mount("/me", s.meRoutes())
	s.router.Mount("/spaces", s.spaceRoutes())
	s.router.Mount("/spaces/{slug}/members", s.memberRoutes())
	s.router.Mount("/spaces/{slug}/api-keys", s.apiKeyRoutes())
	s.router.Mount("/spaces/{slug}/lists", s.listRoutes())
	s.router.Mount("/spaces/{slug}/pdf", s.pdfRoutes())
	s.router.Mount("/spaces/{slug}/files", s.fileRoutes())
	s.router.Mount("/spaces/{slug}/audit", s.auditRoutes())
	s.router.Mount("/lists", s.publicListRoutes())
	s.router.Mount("/api", s.publicAPIRoutes())

	s.router.Group(func(r chi.Router) {
		r.Use(authmw.Auth(s.jwtManager))
		r.Get("/spaces/{slug}/tables", s.listTables)
		r.Post("/spaces/{slug}/tables", s.createTable)
		r.Get("/spaces/{slug}/tables/{table}", s.getTable)
		r.Put("/spaces/{slug}/tables/{table}/fields", s.updateFields)
		r.Delete("/spaces/{slug}/tables/{table}", s.deleteTable)
		r.Get("/spaces/{slug}/tables/{table}/records", s.listRecords)
		r.Post("/spaces/{slug}/tables/{table}/records", s.createRecord)
		r.Get("/spaces/{slug}/tables/{table}/records/{id}", s.getRecord)
		r.Put("/spaces/{slug}/tables/{table}/records/{id}", s.updateRecord)
		r.Delete("/spaces/{slug}/tables/{table}/records/{id}", s.deleteRecord)
	})

	s.router.Group(func(r chi.Router) {
		r.Use(authmw.Auth(s.jwtManager))
		r.Get("/me", s.getMe)
		r.Get("/templates", s.listTemplates)
		r.Get("/spaces/{slug}/api/status", s.apiStatus)
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

type errorBody struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

type errorResponse struct {
	Error errorBody `json:"error"`
}

func (s *Server) writeError(w http.ResponseWriter, r *http.Request, err error) {
	var apiErr *apierrors.APIError
	if errors.As(err, &apiErr) {
		writeJSON(w, apiErr.Code, errorResponse{Error: errorBody{
			Code:    apiErr.ErrorCode,
			Message: apiErr.Message,
			Fields:  apiErr.Fields,
		}})
		return
	}
	s.logger.Error("unexpected error",
		"request_id", chimw.GetReqID(r.Context()),
		"err", err,
	)
	writeJSON(w, http.StatusInternalServerError, errorResponse{Error: errorBody{
		Code:    "internal_error",
		Message: "Внутренняя ошибка сервера",
	}})
}
