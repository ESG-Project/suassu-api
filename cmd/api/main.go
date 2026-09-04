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
	"github.com/joho/godotenv"
	"go.uber.org/zap"

	appaddress "github.com/ESG-Project/suassu-api/internal/app/address"
	appadminuser "github.com/ESG-Project/suassu-api/internal/app/adminuser"
	appenterprise "github.com/ESG-Project/suassu-api/internal/app/enterprise"
	appfeatures "github.com/ESG-Project/suassu-api/internal/app/feature"
	appperm "github.com/ESG-Project/suassu-api/internal/app/permission"
	appphyto "github.com/ESG-Project/suassu-api/internal/app/phytoanalysis"
	approle "github.com/ESG-Project/suassu-api/internal/app/role"
	appspecies "github.com/ESG-Project/suassu-api/internal/app/species"
	appspecimen "github.com/ESG-Project/suassu-api/internal/app/specimen"
	appuser "github.com/ESG-Project/suassu-api/internal/app/user"
	"github.com/ESG-Project/suassu-api/internal/config"
	enterprisehttp "github.com/ESG-Project/suassu-api/internal/http/v1/enterprise"
	featurehttp "github.com/ESG-Project/suassu-api/internal/http/v1/feature"
	legacyuserhttp "github.com/ESG-Project/suassu-api/internal/http/v1/legacyuser"
	permissionhttp "github.com/ESG-Project/suassu-api/internal/http/v1/permission"
	phytohttp "github.com/ESG-Project/suassu-api/internal/http/v1/phytoanalysis"
	rolehttp "github.com/ESG-Project/suassu-api/internal/http/v1/role"
	specieshttp "github.com/ESG-Project/suassu-api/internal/http/v1/species"
	specimenhttp "github.com/ESG-Project/suassu-api/internal/http/v1/specimen"
	userhttp "github.com/ESG-Project/suassu-api/internal/http/v1/user"
	"github.com/ESG-Project/suassu-api/internal/infra/db/postgres"

	appauth "github.com/ESG-Project/suassu-api/internal/app/auth"
	"github.com/ESG-Project/suassu-api/internal/http/cookie"
	"github.com/ESG-Project/suassu-api/internal/http/httperr"
	httpmw "github.com/ESG-Project/suassu-api/internal/http/middleware"
	"github.com/ESG-Project/suassu-api/internal/http/openapi"
	authhttp "github.com/ESG-Project/suassu-api/internal/http/v1/auth"
	infraauth "github.com/ESG-Project/suassu-api/internal/infra/auth"
)

func main() {
	// 0) Carrega .env (best-effort): variáveis já definidas no ambiente têm precedência.
	// Em produção o .env normalmente não existe e as vars vêm do ambiente — por isso o erro é ignorado.
	_ = godotenv.Load()

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
	enterpriseSvc := appenterprise.NewServiceWithTx(enterpriseRepo, addressSvc, hasher, txm)

	// Role / Permission (Fase 0 da migração user-crud -> suassu-api)
	roleRepo := postgres.NewRoleRepo(db)
	permissionRepo := postgres.NewPermissionRepo(db)
	roleSvc := approle.NewService(roleRepo, permissionRepo, userSvc)
	permissionSvc := appperm.NewService(permissionRepo, roleRepo, featureRepo, userSvc)

	// Admin de usuários (CRUD administrativo legado em /user, distinto do
	// self-service /auth/me) — completa a Fase 0 da migração.
	clientRepo := postgres.NewClientRepo(db)
	technicianRepo := postgres.NewTechnicianRepo(db)
	adminUserSvc := appadminuser.NewService(userRepo, roleRepo, permissionRepo, clientRepo, technicianRepo, userSvc, hasher, txm)

	// PhytoAnalysis
	phytoRepo := postgres.NewPhytoAnalysisRepo(db)
	phytoSvc := appphyto.NewService(phytoRepo, txm)

	// Species
	speciesRepo := postgres.NewSpeciesRepo(db)
	speciesSvc := appspecies.NewService(speciesRepo)

	// Specimen
	specimenRepo := postgres.NewSpecimenRepo(db)
	specimenSvc := appspecimen.NewService(specimenRepo)

	// Refresh Tokens
	refreshTokenRepo := postgres.NewRefreshTokenRepo(db)

	// Cookie Manager
	cookieMgr := cookie.NewManager(cookie.Config{
		Domain: cfg.CookieDomain,
		Secure: cfg.CookieSecure,
	})

	// JWT e Auth
	jwtIssuer := infraauth.NewJWT(cfg)
	authSvc := appauth.NewServiceWithRefresh(userRepo, userSvc, hasher, jwtIssuer, refreshTokenRepo)
	authH := authhttp.NewHandler(authSvc, cookieMgr)

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
		// Rotas públicas (sem autenticação)
		v1.Group(func(pub chi.Router) {
			// POST /enterprises é público (criação); GET/{id} e PUT exigem auth,
			// aplicada por rota dentro do próprio router de enterprise.
			pub.Mount("/enterprises", enterprisehttp.Routes(
				enterpriseSvc,
				httpmw.AuthJWT(jwtIssuer),
				httpmw.RequireEnterprise,
			))
		})

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

		// Rotas privadas (requerem autenticação)
		v1.Group(func(priv chi.Router) {
			priv.Use(httpmw.AuthJWT(jwtIssuer))
			priv.Use(httpmw.RequireEnterprise)
			priv.Mount("/users", userhttp.Routes(userSvc))
			priv.Mount("/phyto-analyses", phytohttp.Routes(phytoSvc))
			priv.Mount("/specimens", specimenhttp.Routes(specimenSvc))
			priv.Mount("/species", specieshttp.Routes(
				speciesSvc,
				httpmw.RequirePermission(userSvc, "Species", "create"),
				httpmw.RequirePermission(userSvc, "Species", "update"),
				httpmw.RequirePermission(userSvc, "ManageSpecies", "update"),
			))
			priv.Mount("/role", rolehttp.Routes(
				roleSvc,
				httpmw.RequirePermission(userSvc, "Role", "read"),
				httpmw.RequirePermissionAny(userSvc, "read", "Client", "Technician", "User", "Financial"),
				httpmw.RequirePermission(userSvc, "Role", "create"),
				httpmw.RequirePermission(userSvc, "Role", "delete"),
			))
			priv.Mount("/permission", permissionhttp.Routes(
				permissionSvc,
				httpmw.RequirePermission(userSvc, "Permission", "read"),
				httpmw.RequirePermission(userSvc, "Permission", "create"),
				httpmw.RequirePermission(userSvc, "Permission", "update"),
				httpmw.RequirePermission(userSvc, "Permission", "delete"),
			))
			// Sem permissionVerification no user-crud (gap de RBAC preexistente);
			// preservado aqui para manter o contrato idêntico.
			priv.Mount("/feature", featurehttp.Routes(featureSvc))

			// /user, /user-enterprise (legado, singular) — CRUD administrativo
			// completo, distinto do self-service em /auth/me e da API nova em
			// /users (plural). Ver plano de migração, Fase 0.
			legacyuserhttp.RegisterRoutes(priv, adminUserSvc,
				httpmw.RequirePermission(userSvc, "Client", "read"),
				httpmw.RequirePermission(userSvc, "User", "read"),
				httpmw.RequirePermissionAny(userSvc, "create", "Client", "Technician", "User", "Financial"),
				httpmw.RequirePermissionAny(userSvc, "update", "Client", "Technician", "User", "Financial"),
				httpmw.RequirePermissionAny(userSvc, "delete", "Client", "Technician", "User", "Financial"),
			)
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
