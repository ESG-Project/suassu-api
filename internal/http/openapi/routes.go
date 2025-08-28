package openapi

import (
	"github.com/go-chi/chi/v5"
)

func Routes() chi.Router {
	r := chi.NewRouter()

	// GET /api/v1/openapi.yaml - Especificação OpenAPI
	r.Get("/openapi.yaml", ServeOpenAPISpec)

	// GET /api/v1/docs - Documentação Swagger UI
	r.Get("/docs", ServeDocs)

	// GET /api/v1/openapi/static/* - Arquivos estáticos (se necessário)
	r.Get("/openapi/static/*", ServeStatic)

	return r
}
