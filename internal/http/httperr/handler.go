package httperr

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/ESG-Project/suassu-api/internal/apperr"
	chimw "github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

var logger *zap.Logger // setado no main

// SetLogger permite logar erros dentro do handler.
func SetLogger(l *zap.Logger) { logger = l }

// Handle converte um error em resposta JSON padronizada.
func Handle(w http.ResponseWriter, r *http.Request, err error) {
	if err == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	status := mapStatus(apperr.CodeOf(err), err)
	payload := map[string]any{
		"error": map[string]any{
			"code":    string(apperr.CodeOf(err)),
			"message": publicMessage(err),
		},
		"requestId": chimw.GetReqID(r.Context()),
	}

	// log estruturado (sem vazar sensíveis)
	if logger != nil {
		fields := []zap.Field{
			zap.String("code", string(apperr.CodeOf(err))),
			zap.Int("status", status),
			zap.String("request_id", chimw.GetReqID(r.Context())),
		}
		var ae *apperr.Error
		if errors.As(err, &ae) && ae.Fields != nil {
			fields = append(fields, zap.Any("fields", ae.Fields))
		}
		logger.Error("http_error", append(fields, zap.Error(err))...)
	}

	writeJSON(w, status, payload)
}

func mapStatus(code apperr.Code, err error) int {
	switch code {
	case apperr.CodeInvalid:
		return http.StatusUnprocessableEntity // 422
	case apperr.CodeNotFound:
		return http.StatusNotFound
	case apperr.CodeConflict:
		return http.StatusConflict
	case apperr.CodeUnauthorized:
		return http.StatusUnauthorized
	case apperr.CodeForbidden:
		return http.StatusForbidden
	case apperr.CodeRateLimited:
		return http.StatusTooManyRequests
	default:
		return http.StatusInternalServerError
	}
}

func publicMessage(err error) string {
	// Mensagens amigáveis/curtas; evite vazar detalhes internos
	var ae *apperr.Error
	if errors.As(err, &ae) && ae.Msg != "" {
		return ae.Msg
	}
	return http.StatusText(mapStatus(apperr.CodeOf(err), err))
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
