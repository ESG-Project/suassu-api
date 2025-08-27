package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ESG-Project/suassu-api/internal/apperr"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestRecoveryWithLogger(t *testing.T) {
	t.Run("recovers from panic and returns 500", func(t *testing.T) {
		logger := zaptest.NewLogger(t)
		recovery := RecoveryWithLogger(logger)

		// Handler que sempre faz panic
		panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			panic("test panic")
		})

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		// Executa o middleware
		recovery(panicHandler).ServeHTTP(w, req)

		// Verifica status 500
		require.Equal(t, http.StatusInternalServerError, w.Code)
		require.Equal(t, "application/json", w.Header().Get("Content-Type"))

		// Verifica resposta JSON
		var response map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		require.Contains(t, response, "error")
		require.Contains(t, response, "requestId")

		errorObj := response["error"].(map[string]any)
		require.Equal(t, "internal", errorObj["code"])
		require.Equal(t, "internal server error", errorObj["message"])
	})

	t.Run("recovers from panic with error type", func(t *testing.T) {
		logger := zaptest.NewLogger(t)
		recovery := RecoveryWithLogger(logger)

		// Handler que faz panic com error
		panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			panic(apperr.New(apperr.CodeNotFound, "not found"))
		})

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		recovery(panicHandler).ServeHTTP(w, req)

		// Deve retornar 500 mesmo com erro de domínio (panic é sempre interno)
		require.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("recovers from panic with non-error type", func(t *testing.T) {
		logger := zaptest.NewLogger(t)
		recovery := RecoveryWithLogger(logger)

		// Handler que faz panic com int
		panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			panic(42)
		})

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		recovery(panicHandler).ServeHTTP(w, req)

		require.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("normal request passes through without panic", func(t *testing.T) {
		logger := zaptest.NewLogger(t)
		recovery := RecoveryWithLogger(logger)

		// Handler normal
		normalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("success"))
		})

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		recovery(normalHandler).ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		require.Equal(t, "success", w.Body.String())
	})

	t.Run("includes panic context in error fields", func(t *testing.T) {
		logger := zaptest.NewLogger(t)
		recovery := RecoveryWithLogger(logger)

		panicValue := "test panic value"
		panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			panic(panicValue)
		})

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		recovery(panicHandler).ServeHTTP(w, req)

		// Verifica que o panic foi convertido para erro interno
		require.Equal(t, http.StatusInternalServerError, w.Code)

		// O middleware deve ter logado o panic com contexto completo
		// (teste de integração seria necessário para verificar logs)
	})
}
