package apperr

import (
	"errors"
	"fmt"
)

// Code classifica o tipo do erro sem acoplar à camada HTTP.
type Code string

const (
	CodeInvalid      Code = "invalid" // validação / dados inválidos
	CodeNotFound     Code = "not_found"
	CodeConflict     Code = "conflict"
	CodeUnauthorized Code = "unauthorized"
	CodeForbidden    Code = "forbidden"
	CodeRateLimited  Code = "rate_limited"
	CodeInternal     Code = "internal"
)

// Error é um erro da aplicação com código e wrap opcional.
type Error struct {
	Code   Code
	Msg    string
	Cause  error
	Fields map[string]any // opcional: detalhes para troubleshooting
}

func (e *Error) Error() string {
	if e.Msg != "" {
		if e.Cause != nil {
			return fmt.Sprintf("%s: %v", e.Msg, e.Cause)
		}
		return e.Msg
	}
	if e.Cause != nil {
		return e.Cause.Error()
	}
	return string(e.Code)
}
func (e *Error) Unwrap() error { return e.Cause }

// New cria um erro com código e mensagem.
func New(code Code, msg string) *Error {
	return &Error{Code: code, Msg: msg}
}

// Wrap envolve um erro pré-existente atribuindo um código e mensagem.
func Wrap(err error, code Code, msg string) *Error {
	if err == nil {
		return New(code, msg)
	}
	return &Error{Code: code, Msg: msg, Cause: err}
}

// WithFields adiciona detalhes (sem expor dados sensíveis!).
func WithFields(err error, fields map[string]any) *Error {
	var e *Error
	if errors.As(err, &e) {
		// copia rala para não mutar o original
		n := *e
		if n.Fields == nil {
			n.Fields = map[string]any{}
		}
		for k, v := range fields {
			n.Fields[k] = v
		}
		return &n
	}
	return &Error{Code: CodeInternal, Msg: fmt.Sprintf("wrapped non-app error: %v", err), Cause: err, Fields: fields}
}

// CodeOf busca o Code no chain de errors; default = internal.
func CodeOf(err error) Code {
	var e *Error
	if errors.As(err, &e) {
		return e.Code
	}
	return CodeInternal
}

// escalationField marca, em Fields, um Forbidden como escalação de privilégio
// (tentativa de conceder/atribuir mais do que o próprio ator possui).
const escalationField = "escalation"

// NewPrivilegeEscalation cria um erro CodeForbidden marcado como escalação de
// privilégio. Replica PrivilegeEscalationError do user-crud: a camada HTTP
// deve devolver um `code` de nível superior no corpo (não só em `error.code`)
// para que o front exiba a mensagem em vez de redirecionar para
// /unauthorized, como faz com os demais 403 — a tela em que o usuário está
// continua legítima, só a ação foi recusada.
func NewPrivilegeEscalation(msg string) *Error {
	return &Error{Code: CodeForbidden, Msg: msg, Fields: map[string]any{escalationField: true}}
}

// IsPrivilegeEscalation reporta se err foi criado com NewPrivilegeEscalation.
func IsPrivilegeEscalation(err error) bool {
	var e *Error
	if errors.As(err, &e) && e.Fields != nil {
		v, _ := e.Fields[escalationField].(bool)
		return v
	}
	return false
}
