package auditlog

import (
	"context"
	"log"
	"time"

	domainauditlog "github.com/ESG-Project/suassu-api/internal/domain/auditlog"
	"github.com/google/uuid"
)

type Service struct {
	repo Repo
	now  func() time.Time
}

var _ Recorder = (*Service)(nil)

func NewService(r Repo) *Service {
	return &Service{repo: r, now: time.Now}
}

// Record grava um evento de auditoria. Ver Recorder para o motivo de a falha
// ser apenas logada.
func (s *Service) Record(ctx context.Context, actorID *string, enterpriseID, description string) {
	entry := domainauditlog.NewLog(uuid.NewString(), actorID, enterpriseID, description, s.now())
	if err := entry.Validate(); err != nil {
		log.Printf("auditlog: entrada inválida (%v): %s", err, description)
		return
	}
	if err := s.repo.Create(ctx, entry); err != nil {
		log.Printf("auditlog: falha ao gravar (%v): %s", err, description)
	}
}
