// Package bank contém o cliente do catálogo público de bancos consumido por
// GET /all-banks.
package bank

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ESG-Project/suassu-api/internal/apperr"
)

// HTTPCatalog busca o catálogo de bancos numa API externa (por padrão a da
// Cora) e repassa a resposta sem transformação.
type HTTPCatalog struct {
	url    string
	client *http.Client
}

func NewHTTPCatalog(url string) *HTTPCatalog {
	return &HTTPCatalog{url: url, client: &http.Client{Timeout: 10 * time.Second}}
}

func (c *HTTPCatalog) List(ctx context.Context) (json.RawMessage, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url, nil)
	if err != nil {
		return nil, apperr.Wrap(err, apperr.CodeInternal, "Erro ao buscar os bancos.")
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, apperr.Wrap(err, apperr.CodeInternal, "Erro ao buscar os bancos.")
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, apperr.New(apperr.CodeInternal,
			fmt.Sprintf("Erro ao buscar os bancos: catálogo respondeu %d.", resp.StatusCode))
	}

	var body json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, apperr.Wrap(err, apperr.CodeInternal, "Erro ao buscar os bancos: resposta inválida.")
	}
	return body, nil
}
