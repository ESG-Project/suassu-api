package main

import (
	"context"
	"database/sql"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"

	appuser "github.com/ESG-Project/suassu-api/internal/app/user"
	userhttp "github.com/ESG-Project/suassu-api/internal/http/v1/user"
	"github.com/ESG-Project/suassu-api/internal/infra/auth"
	"github.com/ESG-Project/suassu-api/internal/infra/db/postgres"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	port := getenv("HTTP_PORT", "8080")
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = os.Getenv("DATABASE_URL")
	}
	if dsn == "" {
		logger.Fatal("missing DB_DSN or DATABASE_URL")
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		logger.Fatal("db open", zap.Error(err))
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		logger.Fatal("db ping", zap.Error(err))
	}

	userRepo := postgres.NewUserRepo(db)
	hasher := auth.NewBCrypt()
	userSvc := appuser.NewService(userRepo, hasher)

	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.RealIP, middleware.Recoverer, middleware.Timeout(30*time.Second))
	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })

	r.Route("/api/v1", func(v1 chi.Router) {
		v1.Mount("/users", userhttp.Routes(userSvc)) // rotas no plural
	})

	srv := &http.Server{Addr: ":" + port, Handler: r}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("listen", zap.Error(err))
		}
	}()
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	_ = srv.Shutdown(context.Background())
}

func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
