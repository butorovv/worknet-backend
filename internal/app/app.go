package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	authdelivery "github.com/butorovv/worknet-backend/internal/modules/auth/delivery"
	authhttp "github.com/butorovv/worknet-backend/internal/modules/auth/delivery/http"
	authmw "github.com/butorovv/worknet-backend/internal/modules/auth/delivery/middlewares"
	authrepo "github.com/butorovv/worknet-backend/internal/modules/auth/repository"
	authusecase "github.com/butorovv/worknet-backend/internal/modules/auth/usecase"
	"github.com/butorovv/worknet-backend/internal/platform/config"
	"github.com/butorovv/worknet-backend/internal/platform/httpx"
	"github.com/butorovv/worknet-backend/internal/platform/password"
	"github.com/butorovv/worknet-backend/internal/platform/token"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type App struct {
	cfg    config.Config
	server *http.Server
	db     *pgxpool.Pool
	redis  *redis.Client
}

func New(ctx context.Context) (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	db, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("create postgres pool: %w", err)
	}
	if err := db.Ping(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})
	if err := redisClient.Ping(ctx).Err(); err != nil {
		db.Close()
		_ = redisClient.Close()
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	authRepository := authrepo.NewAuthRepoPostgres(db)
	sessionStore := authrepo.NewSessionRepoRedis(redisClient, "auth:session:")

	authUsecase, err := authusecase.NewAuthUseCase(
		authRepository,
		sessionStore,
		password.NewBcryptHasher(cfg.BcryptCost),
		token.NewRandomGenerator(32),
		cfg.SessionTTL,
	)
	if err != nil {
		db.Close()
		_ = redisClient.Close()
		return nil, err
	}

	cookieConfig := authdelivery.CookieConfig{
		Name:   cfg.SessionCookieName,
		TTL:    cfg.SessionTTL,
		Secure: cfg.SessionSecure,
	}

	authHandler := authhttp.NewAuthHandler(authUsecase, cookieConfig)
	authMiddleware := authmw.NewMiddleware(authUsecase, cookieConfig)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", health)
	registerAuthRoutes(mux, authHandler, authMiddleware)

	return &App{
		cfg: cfg,
		server: &http.Server{
			Addr:              cfg.Addr(),
			Handler:           mux,
			ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		},
		db:    db,
		redis: redisClient,
	}, nil
}

func (a *App) Start() error {
	slog.Info("starting HTTP server", "addr", a.server.Addr)
	if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (a *App) Shutdown(ctx context.Context) error {
	var shutdownErr error
	if err := a.server.Shutdown(ctx); err != nil {
		shutdownErr = err
	}
	a.db.Close()
	if err := a.redis.Close(); err != nil && shutdownErr == nil {
		shutdownErr = err
	}
	return shutdownErr
}

func (a *App) ShutdownTimeout() time.Duration {
	return a.cfg.ShutdownTimeout
}

func registerAuthRoutes(mux *http.ServeMux, handler *authhttp.AuthHandler, middleware *authmw.Middleware) {
	mux.HandleFunc("/api/v1/auth/register", handler.Register)
	mux.HandleFunc("/api/v1/auth/login", handler.Login)
	mux.Handle("/api/v1/auth/logout", middleware.RequireAuth(http.HandlerFunc(handler.Logout)))
	mux.Handle("/api/v1/auth/me", middleware.RequireAuth(http.HandlerFunc(handler.Me)))
}

func health(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpx.MethodNotAllowed(w)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
