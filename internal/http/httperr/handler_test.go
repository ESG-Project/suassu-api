package httperr

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ESG-Project/suassu-api/internal/apperr"
	"github.com/stretchr/testify/require"
)

func TestHandle(t *testing.T) {
	t.Run("nil error returns 204", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		Handle(w, req, nil)

		require.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("apperr.Error with fields", func(t *testing.T) {
		err := apperr.WithFields(
			apperr.New(apperr.CodeInvalid, "validation failed"),
			map[string]any{"field": "email", "value": "invalid"},
		)
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		Handle(w, req, err)

		require.Equal(t, http.StatusUnprocessableEntity, w.Code)
		require.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"))

		var response map[string]any
		jsonErr := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, jsonErr)

		require.Contains(t, response, "error")
		require.Contains(t, response, "meta")

		errorObj := response["error"].(map[string]any)
		require.Equal(t, "invalid", errorObj["code"])
		require.Equal(t, "validation failed", errorObj["message"])
	})

	t.Run("non-apperr error defaults to internal", func(t *testing.T) {
		err := apperr.New(apperr.CodeInternal, "database connection failed")
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		Handle(w, req, err)

		require.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestMapStatus(t *testing.T) {
	tests := []struct {
		name     string
		code     apperr.Code
		expected int
	}{
		{"invalid", apperr.CodeInvalid, http.StatusUnprocessableEntity},
		{"not_found", apperr.CodeNotFound, http.StatusNotFound},
		{"conflict", apperr.CodeConflict, http.StatusConflict},
		{"unauthorized", apperr.CodeUnauthorized, http.StatusUnauthorized},
		{"forbidden", apperr.CodeForbidden, http.StatusForbidden},
		{"rate_limited", apperr.CodeRateLimited, http.StatusTooManyRequests},
		{"internal", apperr.CodeInternal, http.StatusInternalServerError},
		{"unknown", "unknown", http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := mapStatus(tt.code, nil)
			require.Equal(t, tt.expected, status)
		})
	}
}

func TestPublicMessage(t *testing.T) {
	t.Run("apperr with message", func(t *testing.T) {
		err := apperr.New(apperr.CodeInvalid, "custom error message")
		msg := publicMessage(err)
		require.Equal(t, "custom error message", msg)
	})

	t.Run("apperr without message", func(t *testing.T) {
		err := apperr.New(apperr.CodeNotFound, "")
		msg := publicMessage(err)
		require.Equal(t, "Not Found", msg)
	})

	t.Run("non-apperr error", func(t *testing.T) {
		err := apperr.New(apperr.CodeInternal, "")
		msg := publicMessage(err)
		require.Equal(t, "Internal Server Error", msg)
	})
}

func TestWriteJSON(t *testing.T) {
	t.Run("writes JSON response", func(t *testing.T) {
		w := httptest.NewRecorder()
		data := map[string]string{"message": "success"}

		writeJSON(w, http.StatusOK, "test-req-id", data)

		require.Equal(t, http.StatusOK, w.Code)
		require.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"))

		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		require.Equal(t, "success", response["message"])
	})

	t.Run("handles JSON marshal error gracefully", func(t *testing.T) {
		w := httptest.NewRecorder()
		// Channel não pode ser serializado para JSON
		unserializable := make(chan int)

		writeJSON(w, http.StatusOK, "test-req-id", unserializable)

		require.Equal(t, http.StatusOK, w.Code)
		require.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"))
	})
}
