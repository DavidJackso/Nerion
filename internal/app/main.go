package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"nerion/internal/adapter/brevo"
	"nerion/internal/adapter/storage"
	"nerion/internal/config"
	"nerion/internal/domain"
	"nerion/internal/jwtauth"
	"nerion/internal/repository"
	"nerion/internal/service"
	thttp "nerion/internal/transport/http"
)

type App struct {
	cfg    *config.Config
	pool   *pgxpool.Pool
	server *thttp.Server
	logger *slog.Logger
}

func New(ctx context.Context, cfg *config.Config, logger *slog.Logger) (*App, error) {
	pool, err := pgxpool.New(ctx, cfg.DB.DSN)
	if err != nil {
		return nil, fmt.Errorf("db: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("db ping: %w", err)
	}

	ttl, err := time.ParseDuration(cfg.JWT.TTL)
	if err != nil {
		return nil, fmt.Errorf("jwt.ttl: %w", err)
	}
	jm := jwtauth.New(cfg.JWT.Secret, ttl)

	// Repositories
	userRepo := repository.NewUserRepository(pool)
	sessionRepo := repository.NewSessionRepository(pool)
	emailVerifRepo := repository.NewEmailVerificationRepository(pool)
	pwdResetRepo := repository.NewPasswordResetRepository(pool)
	spaceRepo := repository.NewSpaceRepository(pool)
	memberRepo := repository.NewSpaceMemberRepository(pool)
	tableRepo := repository.NewTableRepository(pool)
	fieldRepo := repository.NewFieldRepository(pool)
	recordRepo := repository.NewRecordRepository(pool)
	apiKeyRepo := repository.NewAPIKeyRepository(pool)
	listRepo := repository.NewListRepository(pool)
	pdfRepo := repository.NewPDFRepository(pool)
	auditRepo := repository.NewAuditRepository(pool)
	ddl := repository.NewDDLExecutor(pool)

	// Adapters
	emailSender := brevo.New(cfg.Brevo.APIKey, cfg.Brevo.From, cfg.Brevo.Name, cfg.Brevo.Host)

	presignTTL, err := time.ParseDuration(cfg.Storage.PresignTTL)
	if err != nil {
		presignTTL = time.Hour
	}
	var storageAdapter domain.StorageAdapter
	if cfg.Storage.S3Bucket != "" {
		storageAdapter, err = storage.NewS3Adapter(cfg.Storage)
		if err != nil {
			return nil, fmt.Errorf("storage: %w", err)
		}
	} else {
		storageAdapter = storage.NewLocalAdapter(cfg.Storage.UploadDir)
	}

	// Services
	userSvc := service.NewUserService(userRepo)
	authSvc := service.NewAuthService(userRepo, sessionRepo, emailVerifRepo, pwdResetRepo, jm, emailSender, logger)
	spaceSvc := service.NewSpaceService(spaceRepo, memberRepo)
	memberSvc := service.NewSpaceMemberService(memberRepo, userRepo, emailSender, logger)
	schemaSvc := service.NewSchemaService(spaceRepo, memberRepo, tableRepo, fieldRepo, ddl)
	recordSvc := service.NewRecordService(spaceRepo, memberRepo, tableRepo, fieldRepo, recordRepo)
	apiKeySvc := service.NewAPIKeyService(spaceRepo, memberRepo, apiKeyRepo)
	listSvc := service.NewListService(spaceRepo, memberRepo, tableRepo, fieldRepo, listRepo)
	pdfSvc := service.NewPDFService(spaceRepo, memberRepo, tableRepo, fieldRepo, recordRepo, pdfRepo, storageAdapter, cfg.Storage.UploadDir)

	server := thttp.NewServer(thttp.ServerConfig{
		UserService:   userSvc,
		AuthService:   authSvc,
		SpaceService:  spaceSvc,
		MemberService: memberSvc,
		SchemaService: schemaSvc,
		RecordService: recordSvc,
		APIKeyService: apiKeySvc,
		ListService:   listSvc,
		PDFService:    pdfSvc,
		Storage:       storageAdapter,
		PresignTTL:    presignTTL,
		SpaceRepo:     spaceRepo,
		MemberRepo:    memberRepo,
		TableRepo:     tableRepo,
		FieldRepo:     fieldRepo,
		RecordRepo:    recordRepo,
		APIKeyRepo:    apiKeyRepo,
		AuditRepo:     auditRepo,
		JWTManager:    jm,
		Logger:        logger,
	})

	return &App{cfg: cfg, pool: pool, server: server, logger: logger}, nil
}

func (a *App) Run(ctx context.Context) error {
	srv := &http.Server{
		Addr:         a.cfg.HTTP.Addr,
		Handler:      a.server,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		a.logger.Info("shutting down")
		return srv.Shutdown(shutCtx)
	case err := <-errCh:
		return err
	}
}

func (a *App) Close() {
	a.pool.Close()
}
