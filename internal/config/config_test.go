package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name    string
		env     map[string]string
		wantErr bool
		check   func(*Config) bool
	}{
		{
			name: "defaults",
			env:  map[string]string{},
			check: func(cfg *Config) bool {
				return cfg.AppName == "suassu-api" &&
					cfg.AppEnv == "dev" &&
					cfg.HTTPPort == "8080" &&
					cfg.LogLevel == "info" &&
					cfg.DBMaxOpenConns == 20 &&
					cfg.DBMaxIdleConns == 10
			},
		},
		{
			name: "custom values",
			env: map[string]string{
				"APP_NAME":          "test-api",
				"APP_ENV":           "staging",
				"HTTP_PORT":         "9090",
				"LOG_LEVEL":         "debug",
				"DB_MAX_OPEN_CONNS": "50",
				"DB_MAX_IDLE_CONNS": "25",
			},
			check: func(cfg *Config) bool {
				return cfg.AppName == "test-api" &&
					cfg.AppEnv == "staging" &&
					cfg.HTTPPort == "9090" &&
					cfg.LogLevel == "debug" &&
					cfg.DBMaxOpenConns == 50 &&
					cfg.DBMaxIdleConns == 25
			},
		},
		{
			name: "DB_DSN fallback to DATABASE_URL",
			env: map[string]string{
				"DATABASE_URL": "postgres://user:pass@localhost:5432/db",
			},
			check: func(cfg *Config) bool {
				return cfg.DBDSN == "postgres://user:pass@localhost:5432/db"
			},
		},
		{
			name: "DB_DSN takes precedence over DATABASE_URL",
			env: map[string]string{
				"DB_DSN":       "postgres://user:pass@localhost:5432/db1",
				"DATABASE_URL": "postgres://user:pass@localhost:5432/db2",
			},
			check: func(cfg *Config) bool {
				return cfg.DBDSN == "postgres://user:pass@localhost:5432/db1"
			},
		},
		{
			name: "prod validation - missing DB_DSN",
			env: map[string]string{
				"APP_ENV": "prod",
			},
			wantErr: true,
		},
		{
			name: "prod validation - with DB_DSN",
			env: map[string]string{
				"APP_ENV": "prod",
				"DB_DSN":  "postgres://user:pass@localhost:5432/db",
			},
			check: func(cfg *Config) bool {
				return cfg.AppEnv == "prod" && cfg.DBDSN != ""
			},
		},
		{
			name: "invalid integer values fallback to defaults",
			env: map[string]string{
				"DB_MAX_OPEN_CONNS": "invalid",
				"DB_MAX_IDLE_CONNS": "not-a-number",
			},
			check: func(cfg *Config) bool {
				return cfg.DBMaxOpenConns == 20 && cfg.DBMaxIdleConns == 10
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup environment
			for k, v := range tt.env {
				os.Setenv(k, v)
			}
			defer func() {
				for k := range tt.env {
					os.Unsetenv(k)
				}
			}()

			// Test
			cfg, err := Load()

			// Assertions
			if tt.wantErr {
				if err == nil {
					t.Errorf("Load() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Load() unexpected error: %v", err)
				return
			}

			if tt.check != nil && !tt.check(cfg) {
				t.Errorf("Load() config validation failed")
			}
		})
	}
}

func TestGetEnvHelpers(t *testing.T) {
	// Test getenv
	os.Setenv("TEST_KEY", "test_value")
	defer os.Unsetenv("TEST_KEY")

	if got := getenv("TEST_KEY", "default"); got != "test_value" {
		t.Errorf("getenv() = %v, want %v", got, "test_value")
	}

	if got := getenv("NONEXISTENT_KEY", "default"); got != "default" {
		t.Errorf("getenv() = %v, want %v", got, "default")
	}

	// Test getint
	os.Setenv("TEST_INT", "42")
	defer os.Unsetenv("TEST_INT")

	if got := getint("TEST_INT", 0); got != 42 {
		t.Errorf("getint() = %v, want %v", got, 42)
	}

	if got := getint("NONEXISTENT_INT", 100); got != 100 {
		t.Errorf("getint() = %v, want %v", got, 100)
	}

	if got := getint("INVALID_INT", 200); got != 200 {
		t.Errorf("getint() = %v, want %v", got, 200)
	}
}
