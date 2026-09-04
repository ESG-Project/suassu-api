package parameter

import (
	"context"
	"fmt"
	"strings"

	appauditlog "github.com/ESG-Project/suassu-api/internal/app/auditlog"
	"github.com/ESG-Project/suassu-api/internal/apperr"
	domainparameter "github.com/ESG-Project/suassu-api/internal/domain/parameter"
)

type Service struct {
	repo  Repo
	audit appauditlog.Recorder
}

func NewService(r Repo, audit appauditlog.Recorder) *Service {
	return &Service{repo: r, audit: audit}
}

func (s *Service) List(ctx context.Context, enterpriseID string) ([]domainparameter.Parameter, error) {
	return s.repo.List(ctx, enterpriseID)
}

// UpdateInput carrega os campos de PUT /parameter. Value e IsDefault são
// ponteiros para reproduzir a semântica de `undefined` do Prisma: campo
// ausente no corpo permanece como está no banco. O front (models/parameter.ts)
// não envia isDefault ao editar, e sobrescrevê-lo com false apagaria a marca
// dos parâmetros criados no onboarding da empresa.
type UpdateInput struct {
	ID        string
	Title     string
	Value     *string
	IsDefault *bool
}

func (s *Service) Update(ctx context.Context, actorID, enterpriseID string, in UpdateInput) (*domainparameter.Parameter, error) {
	if strings.TrimSpace(in.Title) == "" {
		return nil, apperr.New(apperr.CodeInvalid, "Título do parâmetro é obrigatório.")
	}

	existing, err := s.repo.GetByIDAnyEnterprise(ctx, in.ID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, apperr.New(apperr.CodeNotFound, "Parâmetro não encontrado.")
	}
	if existing.EnterpriseID != enterpriseID {
		return nil, apperr.New(apperr.CodeForbidden,
			"Acesso negado: Você não tem permissão para editar este parâmetro. O parâmetro pertence a outra empresa.")
	}

	existing.Title = in.Title
	if in.Value != nil {
		existing.SetValue(in.Value)
	}
	if in.IsDefault != nil {
		existing.SetIsDefault(*in.IsDefault)
	}

	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, err
	}

	s.audit.Record(ctx, &actorID, enterpriseID, fmt.Sprintf("Edição do parâmetro %s", existing.Title))
	return existing, nil
}
