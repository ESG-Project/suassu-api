package userhttp_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	appuser "github.com/ESG-Project/suassu-api/internal/app/user"
	userhttp "github.com/ESG-Project/suassu-api/internal/http/v1/user"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
)

type fakeSvc struct {
	createdID string
}

func (f *fakeSvc) Create(ctx context.Context, in appuser.CreateInput) (string, error) {
	return f.createdID, nil
}
func (f *fakeSvc) GetByEmail(context.Context, string) (appuser.Entity, error) {
	return appuser.Entity{}, nil
}
func (f *fakeSvc) List(context.Context, int32, int32) ([]appuser.Entity, error) { return nil, nil }

func TestPOST_Users_Create(t *testing.T) {
	svc := &fakeSvc{createdID: "uuid-1"}

	r := chi.NewRouter()
	r.Mount("/users", userhttp.Routes(svc))

	body := map[string]any{
		"name": "Ana", "email": "ana@ex.com", "password": "123", "document": "999", "enterpriseId": "e1",
	}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/users/", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code) // Mudar de 200 para 201
	require.Contains(t, w.Body.String(), `"id":"uuid-1"`)
}
