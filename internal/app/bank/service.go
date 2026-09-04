package bank

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	appauditlog "github.com/ESG-Project/suassu-api/internal/app/auditlog"
	"github.com/ESG-Project/suassu-api/internal/app/types"
	"github.com/ESG-Project/suassu-api/internal/apperr"
	domainbank "github.com/ESG-Project/suassu-api/internal/domain/bank"
	"github.com/google/uuid"
)

type Service struct {
	repo    Repo
	catalog Catalog
	audit   appauditlog.Recorder
}

func NewService(r Repo, c Catalog, audit appauditlog.Recorder) *Service {
	return &Service{repo: r, catalog: c, audit: audit}
}

// ListCatalog devolve o catálogo público de bancos, repassado da fonte
// externa sem transformação.
func (s *Service) ListCatalog(ctx context.Context) (json.RawMessage, error) {
	return s.catalog.List(ctx)
}

// ListByEnterprise lista os bancos vinculados à empresa.
func (s *Service) ListByEnterprise(ctx context.Context, enterpriseID string) ([]types.EnterpriseBankRow, error) {
	return s.repo.ListByEnterprise(ctx, enterpriseID)
}

// Link vincula um banco à empresa, cadastrando-o no catálogo global se o
// código ainda não existir lá.
func (s *Service) Link(ctx context.Context, actorID, enterpriseID, code, name string) (*domainbank.EnterpriseBank, error) {
	code = strings.TrimSpace(code)
	name = strings.TrimSpace(name)
	if code == "" {
		return nil, apperr.New(apperr.CodeInvalid, "Código do banco é obrigatório.")
	}
	if name == "" {
		return nil, apperr.New(apperr.CodeInvalid, "Nome do banco é obrigatório.")
	}

	b, err := s.repo.GetByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	if b == nil {
		b = domainbank.NewBank(uuid.NewString(), code, name)
		if err := s.repo.Create(ctx, b); err != nil {
			return nil, err
		}
	}

	existing, err := s.repo.GetEnterpriseBank(ctx, b.ID, enterpriseID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, apperr.New(apperr.CodeConflict, "Este banco já está vinculado a esta empresa.")
	}

	eb := domainbank.NewEnterpriseBank(uuid.NewString(), enterpriseID, b.ID)
	if err := s.repo.CreateEnterpriseBank(ctx, eb); err != nil {
		return nil, err
	}

	s.audit.Record(ctx, &actorID, enterpriseID, fmt.Sprintf(
		"Criação da associação banco-empresa (Banco: %s, Código: %s). Empresa: %s", b.Name, b.Code, enterpriseID))

	return eb, nil
}

// Unlink remove o vínculo empresa-banco. O banco em si continua no catálogo
// global, que é compartilhado entre as empresas.
func (s *Service) Unlink(ctx context.Context, actorID, enterpriseBankID, enterpriseID string) error {
	eb, err := s.repo.GetEnterpriseBankByID(ctx, enterpriseBankID)
	if err != nil {
		return err
	}
	if eb == nil {
		return apperr.New(apperr.CodeNotFound, "Associação banco-empresa não encontrada.")
	}
	if eb.EnterpriseID != enterpriseID {
		return apperr.New(apperr.CodeForbidden,
			"Acesso negado: Você não tem permissão para excluir esta associação banco-empresa. A associação pertence a outra empresa.")
	}

	if err := s.repo.DeleteEnterpriseBank(ctx, eb.ID); err != nil {
		return err
	}

	s.audit.Record(ctx, &actorID, enterpriseID, fmt.Sprintf(
		"Exclusão da associação banco-empresa (ID: %s). Empresa relacionada: %s", eb.ID, eb.EnterpriseName))

	return nil
}
