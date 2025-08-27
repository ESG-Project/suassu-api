package main

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"

	appuser "github.com/ESG-Project/suassu-api/internal/app/user"
	"github.com/ESG-Project/suassu-api/internal/config"
	userhttp "github.com/ESG-Project/suassu-api/internal/http/v1/user"
	"github.com/ESG-Project/suassu-api/internal/infra/db/postgres"

	appauth "github.com/ESG-Project/suassu-api/internal/app/auth"
	httperr "github.com/ESG-Project/suassu-api/internal/http/httperr"
	httpmw "github.com/ESG-Project/suassu-api/internal/http/middleware"
	authhttp "github.com/ESG-Project/suassu-api/internal/http/v1/auth"
	infraauth "github.com/ESG-Project/suassu-api/internal/infra/auth"
)

func main() {
	// 1) Config
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}
	logger, _ := config.BuildLogger(cfg)
	defer logger.Sync()
	httperr.SetLogger(logger)

	// 2) DB
	ctx := context.Background()
	db, err := config.OpenPostgres(ctx, cfg)
	if err != nil {
		logger.Fatal("db open/ping", zap.Error(err))
	}
	defer func(db *sql.DB) { _ = db.Close() }(db)

	// 3) Dependencies
	userRepo := postgres.NewUserRepo(db)
	hasher := infraauth.NewBCrypt()
	userSvc := appuser.NewService(userRepo, hasher)

	// JWT e Auth
	jwtIssuer := infraauth.NewJWT(cfg)
	authSvc := appauth.NewService(userRepo, hasher, jwtIssuer)
	authH := authhttp.NewHandler(authSvc)

	// 4) HTTP router
	r := chi.NewRouter()
	r.Use(
		middleware.RequestID,
		middleware.RealIP,
		middleware.Recoverer,
		middleware.Timeout(30*time.Second),
	)
	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	r.Route("/api/v1", func(v1 chi.Router) {
		v1.Route("/auth", func(auth chi.Router) {
			// público
			auth.Group(func(pub chi.Router) {
				authH.RegisterPublic(pub)
			})
			// privado
			auth.Group(func(priv chi.Router) {
				priv.Use(httpmw.AuthJWT(jwtIssuer))
				authH.RegisterPrivate(priv)
			})
		})

		// demais rotas protegidas
		v1.Group(func(priv chi.Router) {
			priv.Use(httpmw.AuthJWT(jwtIssuer))
			priv.Use(httpmw.RequireEnterprise)
			priv.Mount("/users", userhttp.Routes(userSvc))
		})
	})

	// 5) Server
	srv := &http.Server{
		Addr:         ":" + cfg.HTTPPort,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	logger.Info("server starting", zap.String("port", cfg.HTTPPort))
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatal("server error", zap.Error(err))
	}
}
