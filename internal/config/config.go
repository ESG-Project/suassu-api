package config

import (
	"errors"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	AppName  string
	AppEnv   string // dev | staging | prod
	HTTPPort string

	// Banco
	DBDSN           string // se vazio, cai para DATABASE_URL
	DBMaxOpenConns  int
	DBMaxIdleConns  int
	DBConnMaxIdleMS int // ms
	DBConnMaxLifeMS int // ms

	// Log
	LogLevel string // debug|info|warn|error

	// JWT
	JWTSecret         string // HMAC secret
	JWTIssuer         string
	JWTAudience       string
	JWTAccessTTLMin   int // minutos (ex: 15)
	JWTRefreshTTLDays int // dias (ex: 7)

	// CORS
	CORSAllowedOrigins []string

	// Cookie
	CookieDomain string // domínio do cookie (ex: ".suassu.com")
	CookieSecure bool   // true em produção (HTTPS)
}

// Load lê variáveis de ambiente e aplica defaults sensatos.
// Observação: o carregamento de .env deve ser feito pelo shell (ex.: `source .env`).
func Load() (*Config, error) {
	corsOrigins := getenv("CORS_ALLOWED_ORIGINS", "http://localhost:3000")

	appEnv := getenv("APP_ENV", "dev")

	cfg := &Config{
		AppName:         getenv("APP_NAME", "suassu-api"),
		AppEnv:          appEnv,
		HTTPPort:        getenv("HTTP_PORT", "8080"),
		DBDSN:           getenv("DB_DSN", ""), // fallback p/ DATABASE_URL abaixo
		DBMaxOpenConns:  getint("DB_MAX_OPEN_CONNS", 20),
		DBMaxIdleConns:  getint("DB_MAX_IDLE_CONNS", 10),
		DBConnMaxIdleMS: getint("DB_CONN_MAX_IDLE_MS", 60000),  // 1min
		DBConnMaxLifeMS: getint("DB_CONN_MAX_LIFE_MS", 300000), // 5min
		LogLevel:        strings.ToLower(getenv("LOG_LEVEL", "info")),

		JWTSecret:         getenv("JWT_SECRET", "dev-secret-change-me"),
		JWTIssuer:         getenv("JWT_ISSUER", "suassu-api"),
		JWTAudience:       getenv("JWT_AUDIENCE", "suassu-clients"),
		JWTAccessTTLMin:   getint("JWT_ACCESS_TTL_MIN", 15),
		JWTRefreshTTLDays: getint("JWT_REFRESH_TTL_DAYS", 7),

		CORSAllowedOrigins: strings.Split(corsOrigins, ","),

		CookieDomain: getenv("COOKIE_DOMAIN", ""), // vazio = mesmo domínio
		CookieSecure: appEnv == "prod",            // true apenas em produção
	}

	if cfg.DBDSN == "" {
		cfg.DBDSN = getenv("DATABASE_URL", "")
	}

	// Validações básicas
	if cfg.AppEnv == "prod" {
		if cfg.DBDSN == "" {
			return nil, errors.New("missing DB_DSN/DATABASE_URL in prod")
		}
		if cfg.JWTSecret == "" || cfg.JWTSecret == "dev-secret-change-me" {
			return nil, errors.New("missing strong JWT_SECRET in prod")
		}
	}
	if cfg.HTTPPort == "" {
		return nil, errors.New("missing HTTP_PORT")
	}
	return cfg, nil
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getint(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return def
}
