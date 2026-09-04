package product

import (
	"context"
	"fmt"
	"strings"

	appauditlog "github.com/ESG-Project/suassu-api/internal/app/auditlog"
	"github.com/ESG-Project/suassu-api/internal/app/types"
	"github.com/ESG-Project/suassu-api/internal/apperr"
	domainproduct "github.com/ESG-Project/suassu-api/internal/domain/product"
	"github.com/google/uuid"
)

type Service struct {
	repo  Repo
	audit appauditlog.Recorder
}

func NewService(r Repo, audit appauditlog.Recorder) *Service {
	return &Service{repo: r, audit: audit}
}

func (s *Service) List(ctx context.Context, enterpriseID string) ([]types.ProductDetailRow, error) {
	return s.repo.ListDetailedByEnterprise(ctx, enterpriseID)
}

// CreateInput carrega os campos de POST /product.
type CreateInput struct {
	Name           string
	SuggestedValue *string
	TypeProductID  *string
	Deliverable    bool
	IsDefault      *bool
}

func (s *Service) Create(ctx context.Context, actorID, enterpriseID string, in CreateInput) (*domainproduct.Product, error) {
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return nil, apperr.New(apperr.CodeInvalid, "Nome do produto é obrigatório.")
	}

	suggested, err := NormalizeSuggestedValue(in.SuggestedValue)
	if err != nil {
		return nil, err
	}

	p := domainproduct.NewProduct(uuid.NewString(), name, enterpriseID, in.Deliverable)
	p.SetSuggestedValue(suggested)
	p.SetTypeProductID(emptyToNil(in.TypeProductID))
	if in.IsDefault != nil {
		p.SetIsDefault(*in.IsDefault)
	}

	if err := s.repo.Create(ctx, p); err != nil {
		return nil, err
	}

	// A vírgula final vem do user-crud; mantida para que as entradas antigas e
	// as novas da trilha de auditoria continuem idênticas.
	s.audit.Record(ctx, &actorID, enterpriseID, fmt.Sprintf("Registro de produto %s, ", p.Name))
	return p, nil
}

// UpdateInput carrega os campos de PUT /product. Os ponteiros reproduzem a
// semântica de `undefined` do Prisma: campo ausente no corpo permanece como
// está no banco.
type UpdateInput struct {
	ID             string
	Name           string
	SuggestedValue *string
	TypeProductID  *string
	Deliverable    *bool
	IsDefault      *bool
}

func (s *Service) Update(ctx context.Context, actorID, enterpriseID string, in UpdateInput) (*domainproduct.Product, error) {
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return nil, apperr.New(apperr.CodeInvalid, "Nome do produto é obrigatório.")
	}

	suggested, err := NormalizeSuggestedValue(in.SuggestedValue)
	if err != nil {
		return nil, err
	}

	existing, err := s.own(ctx, in.ID, enterpriseID, "editar")
	if err != nil {
		return nil, err
	}

	existing.Name = name
	existing.SetSuggestedValue(suggested)
	if in.TypeProductID != nil {
		existing.SetTypeProductID(emptyToNil(in.TypeProductID))
	}
	if in.Deliverable != nil {
		existing.Deliverable = *in.Deliverable
	}
	if in.IsDefault != nil {
		existing.SetIsDefault(*in.IsDefault)
	}

	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, err
	}

	s.audit.Record(ctx, &actorID, enterpriseID, fmt.Sprintf("Edição do produto %s", existing.Name))
	return existing, nil
}

func (s *Service) Delete(ctx context.Context, actorID, enterpriseID, id string) error {
	existing, err := s.own(ctx, id, enterpriseID, "excluir")
	if err != nil {
		return err
	}

	if err := s.repo.Delete(ctx, id, enterpriseID); err != nil {
		return err
	}

	s.audit.Record(ctx, &actorID, enterpriseID, fmt.Sprintf("Exclusão de produto %s", existing.Name))
	return nil
}

// own carrega o produto garantindo que ele pertence à empresa do ator. verb
// entra na mensagem de 403 ("...permissão para <verb> este produto").
func (s *Service) own(ctx context.Context, id, enterpriseID, verb string) (*domainproduct.Product, error) {
	existing, err := s.repo.GetByIDAnyEnterprise(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, apperr.New(apperr.CodeNotFound, "Produto não encontrado.")
	}
	if existing.EnterpriseID != enterpriseID {
		return nil, apperr.New(apperr.CodeForbidden, fmt.Sprintf(
			"Acesso negado: Você não tem permissão para %s este produto. O produto pertence a outra empresa.", verb))
	}
	return existing, nil
}

// emptyToNil trata "" como "sem tipo": o select do formulário manda string
// vazia quando o operador limpa o campo, e gravá-la violaria a FK.
func emptyToNil(s *string) *string {
	if s == nil || strings.TrimSpace(*s) == "" {
		return nil
	}
	return s
}
