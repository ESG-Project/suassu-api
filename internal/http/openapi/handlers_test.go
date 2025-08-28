package openapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestServeOpenAPISpec(t *testing.T) {
	req, err := http.NewRequest("GET", "/openapi.yaml", nil)
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(ServeOpenAPISpec)

	handler.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	require.Equal(t, "application/yaml", rr.Header().Get("Content-Type"))

	body := rr.Body.String()
	require.Contains(t, body, "openapi: \"3.1.0\"")
}

func TestServeDocs(t *testing.T) {
	req, err := http.NewRequest("GET", "/docs", nil)
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(ServeDocs)

	handler.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	require.Equal(t, "text/html; charset=utf-8", rr.Header().Get("Content-Type"))

	body := rr.Body.String()
	require.Contains(t, body, "SwaggerUIBundle")
}

func TestServeStatic(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		expectedStatus int
		expectedType   string
		expectError    bool
	}{
		{
			name:           "yaml file",
			path:           "/api/v1/openapi/static/openapi.yaml",
			expectedStatus: http.StatusOK,
			expectedType:   "application/yaml",
			expectError:    false,
		},
		{
			name:           "html file",
			path:           "/api/v1/openapi/static/docs.html",
			expectedStatus: http.StatusOK,
			expectedType:   "text/html; charset=utf-8",
			expectError:    false,
		},
		{
			name:           "path vazio",
			path:           "/api/v1/openapi/static/",
			expectedStatus: http.StatusNotFound,
			expectedType:   "application/json; charset=utf-8",
			expectError:    true,
		},
		{
			name:           "non-existent file",
			path:           "/api/v1/openapi/static/notfound.txt",
			expectedStatus: http.StatusNotFound,
			expectedType:   "application/json; charset=utf-8",
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", tt.path, nil)
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(ServeStatic)

			handler.ServeHTTP(rr, req)

			require.Equal(t, tt.expectedStatus, rr.Code)
			require.Equal(t, tt.expectedType, rr.Header().Get("Content-Type"))

			if tt.expectError {
				// Verifica se a resposta é JSON de erro padronizado
				var response map[string]any
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				require.NoError(t, err)

				require.Contains(t, response, "error")
				require.Contains(t, response, "meta")

				errorObj := response["error"].(map[string]any)
				require.Contains(t, errorObj, "code")
				require.Contains(t, errorObj, "message")

				// Verifica se o código de erro é correto
				if tt.name == "path vazio" {
					require.Equal(t, "not_found", errorObj["code"])
					require.Equal(t, "path not specified", errorObj["message"])
				} else if tt.name == "non-existent file" {
					require.Equal(t, "not_found", errorObj["code"])
					require.Equal(t, "file not found", errorObj["message"])
				}
			}
		})
	}
}

// Teste adicional para verificar se os erros internos retornam JSON padronizado
func TestServeOpenAPISpec_Error(t *testing.T) {
	// Este teste seria mais complexo pois precisaria mockar o filesystem
	// Por enquanto, apenas verifica se a função existe e não quebra
	req, err := http.NewRequest("GET", "/openapi.yaml", nil)
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(ServeOpenAPISpec)

	handler.ServeHTTP(rr, req)

	// Se não houver erro, deve retornar 200
	if rr.Code == http.StatusOK {
		require.Equal(t, "application/yaml", rr.Header().Get("Content-Type"))
	} else {
		// Se houver erro, deve retornar JSON padronizado
		require.Equal(t, "application/json; charset=utf-8", rr.Header().Get("Content-Type"))
		require.Equal(t, http.StatusInternalServerError, rr.Code)
	}
}
