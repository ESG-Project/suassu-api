package species

import (
	"context"

	"github.com/ESG-Project/suassu-api/internal/app/types"
	"github.com/ESG-Project/suassu-api/internal/apperr"
	domainspecies "github.com/ESG-Project/suassu-api/internal/domain/species"
	"github.com/google/uuid"
)

type ServiceInterface interface {
	Create(ctx context.Context, in CreateInput) (string, error)
	RequestChange(ctx context.Context, in ChangeInput) (string, error)
	Approve(ctx context.Context, id string) error
	Refuse(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*types.SpeciesWithLegislation, error)
	GetByScientificName(ctx context.Context, scientificName string) (*types.SpeciesWithLegislation, error)
	GetOrCreate(ctx context.Context, in CreateInput) (*types.SpeciesWithLegislation, error)
	List(ctx context.Context, limit, offset int32) ([]*types.SpeciesWithLegislation, error)
	ListVisible(ctx context.Context, enterpriseID string, limit, offset int32) ([]*types.SpeciesWithLegislation, error)
	ListVisiblePaged(ctx context.Context, f types.SpeciesListFilter) ([]*types.SpeciesWithLegislation, int64, error)
	ListPending(ctx context.Context, limit, offset int32) ([]*types.SpeciesWithLegislation, error)
}

type Service struct {
	repo Repo
}

func NewService(r Repo) *Service {
	return &Service{repo: r}
}

type CreateInput struct {
	ScientificName string
	Family         string
	PopularName    *string
	Habit          *string
	// Autoria/tenant (opcionais: nil = catálogo legado/global)
	EnterpriseID *string
	CreatedBy    *string
	// Legislação (opcional): só é criada se LawScope != ""
	LawScope            string
	LawID               *string
	IsLawActive         bool
	SpeciesFormFactor   float64
	IsSpeciesProtected  bool
	SpeciesThreatStatus string
	SpeciesOrigin       string
	SuccessionalEcology string
}

// ChangeInput representa a solicitação de alteração (nova versão) de uma espécie.
type ChangeInput struct {
	BaseSpeciesID string // versão-base que está sendo editada
	// Nome científico da nova versão (pode diferir da base). Vazio = mantém o da base.
	ScientificName string
	Family         string
	PopularName    *string
	Habit          *string
	EnterpriseID   *string
	CreatedBy      *string
	// Legislação (opcional): só é criada se LawScope != ""
	LawScope            string
	LawID               *string
	IsLawActive         bool
	SpeciesFormFactor   float64
	IsSpeciesProtected  bool
	SpeciesThreatStatus string
	SpeciesOrigin       string
	SuccessionalEcology string
}

// createVersion persiste uma nova versão de espécie (raiz ou derivada) e sua
// legislação opcional, computando a versão de forma incremental pelo nome científico.
func (s *Service) createVersion(ctx context.Context, in CreateInput, parentID *string) (string, error) {
	speciesID := uuid.NewString()
	sp := domainspecies.NewSpecies(speciesID, in.ScientificName, in.Family)

	if in.PopularName != nil {
		sp.SetPopularName(in.PopularName)
	}
	if in.Habit != nil {
		sp.SetHabit(in.Habit)
	}
	sp.SetOwnership(in.CreatedBy, in.EnterpriseID)

	// Versão = nº de versões já existentes na linhagem da espécie-base + 1.
	// Conta a árvore inteira (raiz + todos os descendentes ligados por parent_id),
	// então o incremento é global na linhagem e independe do nome científico.
	// Raiz (criação sem parent) começa em 1.
	var nextVersion int32 = 1
	if parentID != nil {
		count, err := s.repo.CountLineage(ctx, *parentID)
		if err != nil {
			return "", apperr.Wrap(err, apperr.CodeInternal, "failed to compute species version")
		}
		nextVersion = count + 1
	}
	sp.SetVersioning(nextVersion, parentID)

	if err := sp.Validate(); err != nil {
		return "", apperr.Wrap(err, apperr.CodeInvalid, "invalid species data")
	}

	if err := s.repo.CreateSpecies(ctx, sp); err != nil {
		return "", apperr.Wrap(err, apperr.CodeInternal, "failed to create species")
	}

	// Legislação é opcional; só cria quando o escopo é informado.
	if in.LawScope != "" {
		legislationID := uuid.NewString()
		legislation := domainspecies.NewSpeciesLegislation(
			legislationID,
			in.LawScope,
			in.LawID,
			in.IsLawActive,
			in.SpeciesFormFactor,
			in.IsSpeciesProtected,
			in.SpeciesThreatStatus,
			in.SpeciesOrigin,
			in.SuccessionalEcology,
			&speciesID,
		)
		if err := legislation.Validate(); err != nil {
			return "", apperr.Wrap(err, apperr.CodeInvalid, "invalid legislation data")
		}
		if err := s.repo.CreateLegislation(ctx, legislation); err != nil {
			return "", apperr.Wrap(err, apperr.CodeInternal, "failed to create species legislation")
		}
	}

	return speciesID, nil
}

// Create cria uma nova espécie (versão raiz, status PENDING).
func (s *Service) Create(ctx context.Context, in CreateInput) (string, error) {
	if in.ScientificName == "" || in.Family == "" {
		return "", apperr.New(apperr.CodeInvalid, "missing required fields")
	}
	return s.createVersion(ctx, in, nil)
}

// RequestChange cria uma nova versão a partir de uma espécie existente,
// preservando o registro-base (histórico). Status inicial PENDING.
func (s *Service) RequestChange(ctx context.Context, in ChangeInput) (string, error) {
	if in.BaseSpeciesID == "" || in.Family == "" {
		return "", apperr.New(apperr.CodeInvalid, "missing required fields")
	}

	base, err := s.repo.GetByID(ctx, in.BaseSpeciesID)
	if err != nil {
		return "", apperr.Wrap(err, apperr.CodeNotFound, "base species not found")
	}

	// O nome científico pode ser alterado na edição; se vier vazio, mantém o da base.
	// A nova versão preserva o vínculo com a base via parent_id (histórico intacto).
	scientificName := in.ScientificName
	if scientificName == "" {
		scientificName = base.ScientificName
	}

	create := CreateInput{
		ScientificName: scientificName,
		Family:         in.Family,
		PopularName:    in.PopularName,
		Habit:          in.Habit,
		EnterpriseID:   in.EnterpriseID,
		CreatedBy:      in.CreatedBy,
		// Legislação da nova versão (opcional).
		LawScope:            in.LawScope,
		LawID:               in.LawID,
		IsLawActive:         in.IsLawActive,
		SpeciesFormFactor:   in.SpeciesFormFactor,
		IsSpeciesProtected:  in.IsSpeciesProtected,
		SpeciesThreatStatus: in.SpeciesThreatStatus,
		SpeciesOrigin:       in.SpeciesOrigin,
		SuccessionalEcology: in.SuccessionalEcology,
	}
	return s.createVersion(ctx, create, &in.BaseSpeciesID)
}

// Approve marca uma espécie pendente como aprovada (torna-se oficial global).
func (s *Service) Approve(ctx context.Context, id string) error {
	return s.setStatus(ctx, id, domainspecies.StatusApproved)
}

// Refuse marca uma espécie pendente como recusada.
func (s *Service) Refuse(ctx context.Context, id string) error {
	return s.setStatus(ctx, id, domainspecies.StatusRefused)
}

func (s *Service) setStatus(ctx context.Context, id, status string) error {
	current, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return apperr.Wrap(err, apperr.CodeNotFound, "species not found")
	}
	if current.Status != domainspecies.StatusPending {
		return apperr.New(apperr.CodeInvalid, "only pending species can be evaluated")
	}
	if err := s.repo.UpdateStatus(ctx, id, status); err != nil {
		return apperr.Wrap(err, apperr.CodeInternal, "failed to update species status")
	}
	return nil
}

func (s *Service) GetByID(ctx context.Context, id string) (*types.SpeciesWithLegislation, error) {
	species, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, apperr.Wrap(err, apperr.CodeNotFound, "species not found")
	}
	return species, nil
}

func (s *Service) GetByScientificName(ctx context.Context, scientificName string) (*types.SpeciesWithLegislation, error) {
	species, err := s.repo.GetByScientificName(ctx, scientificName)
	if err != nil {
		return nil, apperr.Wrap(err, apperr.CodeNotFound, "species not found")
	}
	return species, nil
}

// GetOrCreate busca uma espécie pelo nome científico ou cria se não existir
func (s *Service) GetOrCreate(ctx context.Context, in CreateInput) (*types.SpeciesWithLegislation, error) {
	// Tentar buscar primeiro
	species, err := s.repo.GetByScientificName(ctx, in.ScientificName)
	if err == nil {
		return species, nil
	}

	// Verificar se o erro é "not found" - se for outro erro, retornar
	if apperr.CodeOf(err) != apperr.CodeNotFound {
		return nil, err
	}

	// Se não encontrou, criar
	id, err := s.Create(ctx, in)
	if err != nil {
		return nil, err
	}

	// Buscar a espécie recém-criada
	return s.repo.GetByID(ctx, id)
}

func (s *Service) List(ctx context.Context, limit, offset int32) ([]*types.SpeciesWithLegislation, error) {
	// Sem limite ou limite muito alto significa retornar todas
	if limit <= 0 {
		limit = 999999
	}

	return s.repo.List(ctx, limit, offset)
}

// ListVisible retorna as espécies visíveis para uma empresa (aprovadas + próprias).
func (s *Service) ListVisible(ctx context.Context, enterpriseID string, limit, offset int32) ([]*types.SpeciesWithLegislation, error) {
	if limit <= 0 {
		limit = 999999
	}
	return s.repo.ListVisible(ctx, enterpriseID, limit, offset)
}

// ListVisiblePaged retorna as espécies visíveis para a empresa aplicando
// filtros, ordenação, paginação e contagem total no banco (server-side).
func (s *Service) ListVisiblePaged(ctx context.Context, f types.SpeciesListFilter) ([]*types.SpeciesWithLegislation, int64, error) {
	if f.Limit <= 0 {
		f.Limit = 50
	}
	if f.Offset < 0 {
		f.Offset = 0
	}
	return s.repo.ListVisiblePaged(ctx, f)
}

// ListPending retorna todas as espécies pendentes (fila do super-admin).
func (s *Service) ListPending(ctx context.Context, limit, offset int32) ([]*types.SpeciesWithLegislation, error) {
	if limit <= 0 {
		limit = 999999
	}
	return s.repo.ListPending(ctx, limit, offset)
}
