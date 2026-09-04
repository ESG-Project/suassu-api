package auditlog

import (
	"errors"
	"strings"
	"time"
)

// Log é uma entrada da trilha de auditoria (tabela "Log" do schema legado).
//
// ActorID mapeia a coluna "tag" — nome herdado do Prisma que, apesar do nome,
// é a FK para "User". É opcional: eventos gerados pelo próprio sistema (sem
// usuário) gravam NULL, e a leitura em /logs os exibe como "Gerado pelo
// Sistema".
type Log struct {
	ID           string
	ActorID      *string
	EnterpriseID string
	Description  string
	CreatedAt    time.Time
}

func NewLog(id string, actorID *string, enterpriseID, description string, createdAt time.Time) *Log {
	return &Log{
		ID:           id,
		ActorID:      actorID,
		EnterpriseID: enterpriseID,
		Description:  description,
		CreatedAt:    createdAt,
	}
}

func (l *Log) Validate() error {
	if strings.TrimSpace(l.ID) == "" {
		return errors.New("id is required")
	}
	if strings.TrimSpace(l.EnterpriseID) == "" {
		return errors.New("enterpriseId is required")
	}
	if strings.TrimSpace(l.Description) == "" {
		return errors.New("description is required")
	}
	return nil
}
