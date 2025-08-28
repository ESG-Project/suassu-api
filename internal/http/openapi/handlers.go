package openapi

import (
	"embed"
	"net/http"
	"strings"

	"github.com/ESG-Project/suassu-api/internal/apperr"
	"github.com/ESG-Project/suassu-api/internal/http/httperr"
)

//go:embed openapi.yaml docs.html
var files embed.FS

// ServeOpenAPISpec serve o arquivo openapi.yaml
func ServeOpenAPISpec(w http.ResponseWriter, r *http.Request) {
	content, err := files.ReadFile("openapi.yaml")
	if err != nil {
		httperr.Handle(w, r, apperr.New(apperr.CodeInternal, "failed to read openapi spec"))
		return
	}

	w.Header().Set("Content-Type", "application/yaml")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusOK)
	w.Write(content)
}

// ServeDocs serve a documentação Swagger UI
func ServeDocs(w http.ResponseWriter, r *http.Request) {
	content, err := files.ReadFile("docs.html")
	if err != nil {
		httperr.Handle(w, r, apperr.New(apperr.CodeInternal, "failed to read docs"))
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusOK)
	w.Write(content)
}

// ServeStatic serve arquivos estáticos do OpenAPI (para casos específicos)
func ServeStatic(w http.ResponseWriter, r *http.Request) {
	// Remove o prefixo /api/v1/openapi/static/ do path
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/openapi/static/")
	if path == "" {
		httperr.Handle(w, r, apperr.New(apperr.CodeNotFound, "path not specified"))
		return
	}

	// Tenta ler o arquivo do embed
	content, err := files.ReadFile(path)
	if err != nil {
		httperr.Handle(w, r, apperr.New(apperr.CodeNotFound, "file not found"))
		return
	}

	// Define o Content-Type baseado na extensão do arquivo
	contentType := "application/octet-stream"
	switch {
	case strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml"):
		contentType = "application/yaml"
	case strings.HasSuffix(path, ".json"):
		contentType = "application/json"
	case strings.HasSuffix(path, ".html"):
		contentType = "text/html; charset=utf-8"
	case strings.HasSuffix(path, ".css"):
		contentType = "text/css"
	case strings.HasSuffix(path, ".js"):
		contentType = "application/javascript"
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "public, max-age=3600") // Cache por 1 hora para arquivos estáticos
	w.WriteHeader(http.StatusOK)
	w.Write(content)
}
