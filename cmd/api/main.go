package main

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"

	appaddress "github.com/ESG-Project/suassu-api/internal/app/address"
	appenterprise "github.com/ESG-Project/suassu-api/internal/app/enterprise"
	appfeatures "github.com/ESG-Project/suassu-api/internal/app/feature"
	appphyto "github.com/ESG-Project/suassu-api/internal/app/phytoanalysis"
	appuser "github.com/ESG-Project/suassu-api/internal/app/user"
	"github.com/ESG-Project/suassu-api/internal/config"
	enterprisehttp "github.com/ESG-Project/suassu-api/internal/http/v1/enterprise"
	phytohttp "github.com/ESG-Project/suassu-api/internal/http/v1/phytoanalysis"
	userhttp "github.com/ESG-Project/suassu-api/internal/http/v1/user"
	"github.com/ESG-Project/suassu-api/internal/infra/db/postgres"

	appauth "github.com/ESG-Project/suassu-api/internal/app/auth"
	"github.com/ESG-Project/suassu-api/internal/http/httperr"
	httpmw "github.com/ESG-Project/suassu-api/internal/http/middleware"
	openapi "github.com/ESG-Project/suassu-api/internal/http/openapi"
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

	// 3) Dependencies & Seeding
	hasher := infraauth.NewBCrypt()
	featureRepo := postgres.NewFeatureRepo(db)
	featureSvc := appfeatures.NewService(featureRepo, hasher)
	featureSvc.SeedFeatures(ctx)

	txm := &postgres.TxManager{DB: db}

	userRepo := postgres.NewUserRepoWithTx(db, txm)
	addressRepo := postgres.NewAddressRepo(db)
	enterpriseRepo := postgres.NewEnterpriseRepo(db)

	addressSvc := appaddress.NewService(addressRepo, hasher)
	userSvc := appuser.NewServiceWithTx(userRepo, addressSvc, hasher, txm)
	enterpriseSvc := appenterprise.NewService(enterpriseRepo, addressSvc, hasher)

	// PhytoAnalysis
	phytoRepo := postgres.NewPhytoAnalysisRepo(db)
	phytoSvc := appphyto.NewService(phytoRepo, txm)

	// JWT e Auth
	jwtIssuer := infraauth.NewJWT(cfg)
	authSvc := appauth.NewService(userRepo, userSvc, hasher, jwtIssuer)
	authH := authhttp.NewHandler(authSvc)

	// 4) HTTP router
	r := chi.NewRouter()
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.CORSAllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // 5 minutos
	}))
	r.Use(
		middleware.RequestID,
		middleware.RealIP,
		httpmw.RecoveryWithLogger(logger),
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
				priv.Use(httpmw.RequireEnterprise)
				authH.RegisterPrivate(priv)
			})
		})

		v1.Group(func(priv chi.Router) {
			priv.Use(httpmw.AuthJWT(jwtIssuer))
			priv.Use(httpmw.RequireEnterprise)
			priv.Mount("/users", userhttp.Routes(userSvc))
			priv.Mount("/enterprises", enterprisehttp.Routes(enterpriseSvc))
			priv.Mount("/phyto-analyses", phytohttp.Routes(phytoSvc))
		})

		v1.Mount("/", openapi.Routes())
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
