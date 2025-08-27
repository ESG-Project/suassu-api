package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/ESG-Project/suassu-api/internal/apperr"
	"github.com/ESG-Project/suassu-api/internal/http/httperr"
	chimw "github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

// RecoveryWithLogger cria um middleware de panic recovery que:
// 1. Captura panics e os converte para erros HTTP padronizados
// 2. Loga o panic com stack trace e request context
// 3. Retorna resposta JSON consistente via httperr
func RecoveryWithLogger(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rvr := recover(); rvr != nil {
					// Log estruturado com contexto completo
					logger.Error("panic_recovered",
						zap.Any("panic_value", rvr),
						zap.String("request_id", chimw.GetReqID(r.Context())),
						zap.String("method", r.Method),
						zap.String("path", r.URL.Path),
						zap.String("remote_addr", r.RemoteAddr),
						zap.String("user_agent", r.UserAgent()),
						zap.String("stack_trace", string(debug.Stack())),
					)

					// Converte panic para erro da aplicação
					var err error
					switch v := rvr.(type) {
					case error:
						err = v
					case string:
						err = fmt.Errorf("panic: %s", v)
					default:
						err = fmt.Errorf("panic: %v", v)
					}

					// Cria erro interno com contexto do panic
					appErr := apperr.WithFields(
						apperr.Wrap(err, apperr.CodeInternal, "internal server error"),
						map[string]any{
							"panic_value": rvr,
							"recovered":   true,
						},
					)

					// Usa httperr para resposta consistente
					httperr.Handle(w, r, appErr)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
