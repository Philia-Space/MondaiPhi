package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/philiaspace/mondaiphi/config"
	"github.com/philiaspace/mondaiphi/handlers"
	"github.com/philiaspace/mondaiphi/repositories/postgres"
	"github.com/philiaspace/phi-core/observability"
	"github.com/philiaspace/phi-middleware"
	storage "github.com/philiaspace/phi-storage/s3"
)

func main() {
	logger := observability.NewLogger(os.Getenv("LOG_LEVEL"))
	cfg := config.Load()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Connect to database
	db, err := postgres.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error(ctx, "failed to connect to database", "err", err)
		os.Exit(1)
	}
	defer db.Close()
	logger.Info(ctx, "database connected")

	// Initialize repositories
	repo := postgres.NewQuestionRepository(db)

	// Initialize S3 storage
	var s3Client *storage.S3Client
	if cfg.StorageAccessKey != "" && cfg.StorageSecretKey != "" {
		var err error
		s3Client, err = storage.NewS3Client(storage.S3Config{
			Endpoint:   cfg.StorageEndpoint,
			Region:     cfg.StorageRegion,
			Bucket:     cfg.StorageBucket,
			AccessKey:  cfg.StorageAccessKey,
			SecretKey:  cfg.StorageSecretKey,
			UseSSL:     cfg.UseSSL,
			PresignTTL: cfg.StoragePresignTTL,
		})
		if err != nil {
			logger.Error(ctx, "failed to initialize S3 storage", "err", err)
		} else {
			logger.Info(ctx, "S3 storage initialized", "bucket", cfg.StorageBucket)
		}
	}

	// Initialize handlers
	questionHandler := handlers.NewQuestionHandler(repo, s3Client)
	adminHandler := handlers.NewAdminHandler(repo)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Register routes
	questionHandler.RegisterRoutes(mux)
	adminHandler.RegisterRoutes(mux)

	// Apply middleware chain
	handler := middleware.Chain(mux,
		middleware.Recovery(logger),
		middleware.Logger(logger),
		middleware.CORS(),
		middleware.RateLimit(100),
		middleware.AuthJWKS(middleware.JWKSAuthConfig{
			IssuerURL:      cfg.AuthJWKSURL,
			JWKSEndpoint:   "/.well-known/jwks.json",
			ExpectedIssuer: cfg.AuthJWKSURL,
			Audience:       "philia-space",
			CacheTTL:       5 * time.Minute,
			SkipPaths:      []string{"/health", "/.well-known", "/questions", "/passages", "/templates", "/assets"},
		}),
	)

	addr := ":" + cfg.ServerPort
	server := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		logger.Info(ctx, "MondaiPhi starting", "addr", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error(ctx, "server error", "err", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	logger.Info(ctx, "shutting down MondaiPhi")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error(ctx, "shutdown error", "err", err)
	}
}
