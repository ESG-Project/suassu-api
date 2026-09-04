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

	reqID := chimw.GetReqID(r.Context())
	status := mapStatus(apperr.CodeOf(err), err)

	// Monta payload padronizado
	payload := map[string]any{
		"error": map[string]any{
			"code":    string(apperr.CodeOf(err)),
			"message": publicMessage(err),
		},
		"meta": map[string]any{
			"requestId": reqID,
		},
	}

	// Anexa details quando existir (ex.: validação campo a campo)
	var ae *apperr.Error
	if errors.As(err, &ae) && ae.Fields != nil {
		payload["error"].(map[string]any)["details"] = ae.Fields
	}

	// log estruturado (sem vazar sensíveis)
	if logger != nil {
		fields := []zap.Field{
			zap.String("code", string(apperr.CodeOf(err))),
			zap.Int("status", status),
			zap.String("request_id", reqID),
		}
		if ae != nil && ae.Fields != nil {
			fields = append(fields, zap.Any("fields", ae.Fields))
		}
		logger.Error("http_error", append(fields, zap.Error(err))...)
	}

	writeJSON(w, status, reqID, payload)
}

// HandlePrivilegeAware é como Handle, mas dá tratamento especial a erros
// marcados com apperr.NewPrivilegeEscalation: escreve `code` no nível
// superior do corpo (não só em `error.code`), replicando sendControllerError
// do user-crud. O interceptor global do front (services/api.ts) só evita
// redirecionar um 403 para /unauthorized quando enxerga esse `code` no topo
// — sem isto, a tela de cadastro de usuário/permissão expulsaria o operador
// no meio do preenchimento por causa de uma tentativa de escalação recusada.
func HandlePrivilegeAware(w http.ResponseWriter, r *http.Request, err error) {
	if err == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if !apperr.IsPrivilegeEscalation(err) {
		Handle(w, r, err)
		return
	}

	var ae *apperr.Error
	errors.As(err, &ae)
	msg := publicMessage(err)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusForbidden)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"success": false,
		"code":    "PRIVILEGE_ESCALATION",
		"message": msg,
		"error":   msg,
	})
}

func mapStatus(code apperr.Code, _ error) int {
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
	// Mensagens amigáveis e curtas
	var ae *apperr.Error
	if errors.As(err, &ae) && ae.Msg != "" {
		return ae.Msg
	}
	return http.StatusText(mapStatus(apperr.CodeOf(err), err))
}

func writeJSON(w http.ResponseWriter, status int, reqID string, v any) {
	// Opcional: propagar X-Request-ID no header para o gateway/clients
	if reqID != "" {
		w.Header().Set("X-Request-ID", reqID)
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// HandleLegacy escreve o erro no formato herdado do user-crud —
// `{"error": "<mensagem>"}`, com a mensagem como string no nível superior.
//
// É o contrato que alguns models do front leem diretamente
// (`e.response?.data?.error`, ex.: models/bank.ts, cashFlow.ts,
// transaction.ts); com o envelope padrão de Handle esse campo viraria um
// objeto e o toast exibiria "[object Object]". O status continua sendo o
// mapeado de apperr (404/403/409/...), e não o 400 uniforme do user-crud: o
// front trata qualquer 4xx da mesma forma nesses fluxos.
func HandleLegacy(w http.ResponseWriter, r *http.Request, err error) {
	if err == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	reqID := chimw.GetReqID(r.Context())
	status := mapStatus(apperr.CodeOf(err), err)

	if logger != nil {
		logger.Error("http_error",
			zap.String("code", string(apperr.CodeOf(err))),
			zap.Int("status", status),
			zap.String("request_id", reqID),
			zap.Error(err),
		)
	}

	writeJSON(w, status, reqID, map[string]any{"error": publicMessage(err)})
}
