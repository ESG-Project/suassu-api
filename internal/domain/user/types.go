package user

import (
	"github.com/google/uuid"
)

// UserCursorKey representa a chave de cursor para paginação
type UserCursorKey struct {
	Email string    `json:"email"`
	ID    uuid.UUID `json:"id"`
}

// PageInfo representa informações de paginação
type PageInfo struct {
	Next    *UserCursorKey
	HasMore bool
}
