package typeproduct

import (
	"context"
	"fmt"
	"strings"

	appauditlog "github.com/ESG-Project/suassu-api/internal/app/auditlog"
	"github.com/ESG-Project/suassu-api/internal/apperr"
	domaintypeproduct "github.com/ESG-Project/suassu-api/internal/domain/typeproduct"
	"github.com/google/uuid"
)

type Service struct {
	repo  Repo
	audit appauditlog.Recorder
}

func NewService(r Repo, audit appauditlog.Recorder) *Service {
	return &Service{repo: r, audit: audit}
}

func (s *Service) List(ctx context.Context, enterpriseID string) ([]domaintypeproduct.TypeProduct, error) {
	return s.repo.List(ctx, enterpriseID)
}

func (s *Service) Create(ctx context.Context, actorID, enterpriseID, typ string) (*domaintypeproduct.TypeProduct, error) {
	typ = strings.TrimSpace(typ)
	if typ == "" {
		return nil, apperr.New(apperr.CodeInvalid, "Tipo de produto é obrigatório.")
	}

	t := domaintypeproduct.NewTypeProduct(uuid.NewString(), typ, enterpriseID)
	if err := s.repo.Create(ctx, t); err != nil {
		return nil, err
	}

	s.audit.Record(ctx, &actorID, enterpriseID, fmt.Sprintf("Registro de um tipo de produto %s", t.Type))
	return t, nil
}

func (s *Service) Update(ctx context.Context, actorID, enterpriseID, id, typ string) (*domaintypeproduct.TypeProduct, error) {
	typ = strings.TrimSpace(typ)
	if typ == "" {
		return nil, apperr.New(apperr.CodeInvalid, "Tipo de produto é obrigatório.")
	}

	existing, err := s.own(ctx, id, enterpriseID, "alterar")
	if err != nil {
		return nil, err
	}

	existing.SetType(typ)
	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, err
	}

	s.audit.Record(ctx, &actorID, enterpriseID, fmt.Sprintf("Edição de um tipo de produto %s", existing.Type))
	return existing, nil
}

func (s *Service) Delete(ctx context.Context, actorID, enterpriseID, id string) error {
	existing, err := s.own(ctx, id, enterpriseID, "deletar")
	if err != nil {
		return err
	}

	if err := s.repo.Delete(ctx, id, enterpriseID); err != nil {
		return err
	}

	s.audit.Record(ctx, &actorID, enterpriseID, fmt.Sprintf("Exclusão de um tipo de produto %s", existing.Type))
	return nil
}

// own carrega o tipo de produto garantindo que ele pertence à empresa do
// ator. verb entra na mensagem de 403 ("...permissão para <verb> este...").
func (s *Service) own(ctx context.Context, id, enterpriseID, verb string) (*domaintypeproduct.TypeProduct, error) {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, apperr.New(apperr.CodeNotFound, "Tipo de produto não encontrado.")
	}
	if existing.EnterpriseID != enterpriseID {
		return nil, apperr.New(apperr.CodeForbidden, fmt.Sprintf(
			"Acesso negado: Você não tem permissão para %s este tipo de produto. O tipo de produto pertence a outra empresa.", verb))
	}
	return existing, nil
}
