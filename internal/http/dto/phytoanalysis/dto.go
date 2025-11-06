package phytoanalysisdto

import (
	"time"

	"github.com/ESG-Project/suassu-api/internal/app/types"
)

// CreatePhytoAnalysisRequest representa a requisição para criar uma análise fitossociológica
type CreatePhytoAnalysisRequest struct {
	Title           string          `json:"title"`
	InitialDate     time.Time       `json:"initialDate"`
	PortionQuantity int             `json:"portionQuantity"`
	PortionArea     float64         `json:"portionArea"`
	TotalArea       float64         `json:"totalArea"`
	SampledArea     float64         `json:"sampledArea"`
	Description     *string         `json:"description,omitempty"`
	ProjectID       string          `json:"projectId"`
	Specimens       []SpecimenInput `json:"specimens,omitempty"`
}

type SpecimenInput struct {
	Portion      string    `json:"portion"`
	Height       float64   `json:"height"`
	Cap1         float64   `json:"cap1"`
	Cap2         *float64  `json:"cap2,omitempty"`
	Cap3         *float64  `json:"cap3,omitempty"`
	Cap4         *float64  `json:"cap4,omitempty"`
	Cap5         *float64  `json:"cap5,omitempty"`
	Cap6         *float64  `json:"cap6,omitempty"`
	AverageDap   float64   `json:"averageDap"`
	BasalArea    float64   `json:"basalArea"`
	Volume       float64   `json:"volume"`
	RegisterDate time.Time `json:"registerDate"`
	// Dados da espécie
	SpecieID       *string `json:"specieId,omitempty"`
	ScientificName *string `json:"scientificName,omitempty"`
	Family         *string `json:"family,omitempty"`
	PopularName    *string `json:"popularName,omitempty"`
	// Dados da legislação (opcional, com valores padrão)
	LawScope            *string  `json:"lawScope,omitempty"`
	LawID               *string  `json:"lawId,omitempty"`
	IsLawActive         *bool    `json:"isLawActive,omitempty"`
	SpeciesFormFactor   *float64 `json:"speciesFormFactor,omitempty"`
	IsSpeciesProtected  *bool    `json:"isSpeciesProtected,omitempty"`
	SpeciesThreatStatus *string  `json:"speciesThreatStatus,omitempty"`
	SpeciesOrigin       *string  `json:"speciesOrigin,omitempty"`
}

// UpdatePhytoAnalysisRequest representa a requisição para atualizar uma análise fitossociológica
type UpdatePhytoAnalysisRequest struct {
	Title           string    `json:"title"`
	InitialDate     time.Time `json:"initialDate"`
	PortionQuantity int       `json:"portionQuantity"`
	PortionArea     float64   `json:"portionArea"`
	TotalArea       float64   `json:"totalArea"`
	SampledArea     float64   `json:"sampledArea"`
	Description     *string   `json:"description,omitempty"`
}

// PhytoAnalysisResponse representa a resposta de uma análise fitossociológica
type PhytoAnalysisResponse struct {
	ID              string             `json:"id"`
	Title           string             `json:"title"`
	InitialDate     time.Time          `json:"initialDate"`
	PortionQuantity int                `json:"portionQuantity"`
	PortionArea     float64            `json:"portionArea"`
	TotalArea       float64            `json:"totalArea"`
	SampledArea     float64            `json:"sampledArea"`
	Description     *string            `json:"description,omitempty"`
	ProjectID       string             `json:"projectId"`
	CreatedAt       time.Time          `json:"createdAt"`
	UpdatedAt       time.Time          `json:"updatedAt"`
	Project         *ProjectInfo       `json:"project,omitempty"`
	Specimens       []SpecimenResponse `json:"specimens,omitempty"`
}

type ProjectInfo struct {
	ID       string  `json:"id"`
	Title    string  `json:"title"`
	CNPJ     *string `json:"cnpj,omitempty"`
	Activity string  `json:"activity"`
	ClientID string  `json:"clientId"`
}

type SpecimenResponse struct {
	ID             string    `json:"id"`
	Portion        string    `json:"portion"`
	Height         float64   `json:"height"`
	Cap1           float64   `json:"cap1"`
	Cap2           *float64  `json:"cap2,omitempty"`
	Cap3           *float64  `json:"cap3,omitempty"`
	Cap4           *float64  `json:"cap4,omitempty"`
	Cap5           *float64  `json:"cap5,omitempty"`
	Cap6           *float64  `json:"cap6,omitempty"`
	AverageDap     float64   `json:"averageDap"`
	BasalArea      float64   `json:"basalArea"`
	Volume         float64   `json:"volume"`
	RegisterDate   time.Time `json:"registerDate"`
	SpecieID       string    `json:"specieId"`
	ScientificName string    `json:"scientificName"`
	Family         string    `json:"family"`
	PopularName    *string   `json:"popularName,omitempty"`
}

// ToPhytoAnalysisResponse converte tipos internos para resposta HTTP
func ToPhytoAnalysisResponse(p *types.PhytoAnalysisWithProject) *PhytoAnalysisResponse {
	return &PhytoAnalysisResponse{
		ID:              p.ID,
		Title:           p.Title,
		InitialDate:     p.InitialDate,
		PortionQuantity: p.PortionQuantity,
		PortionArea:     p.PortionArea,
		TotalArea:       p.TotalArea,
		SampledArea:     p.SampledArea,
		Description:     p.Description,
		ProjectID:       p.ProjectID,
		CreatedAt:       p.CreatedAt,
		UpdatedAt:       p.UpdatedAt,
		Project: &ProjectInfo{
			ID:       p.ProjectID,
			Title:    p.ProjectTitle,
			CNPJ:     p.ProjectCNPJ,
			Activity: p.ProjectActivity,
			ClientID: p.ProjectClientID,
		},
	}
}

// ToPhytoAnalysisCompleteResponse converte análise completa para resposta HTTP
func ToPhytoAnalysisCompleteResponse(p *types.PhytoAnalysisComplete) *PhytoAnalysisResponse {
	specimens := make([]SpecimenResponse, 0, len(p.Specimens))
	for _, s := range p.Specimens {
		specimens = append(specimens, SpecimenResponse{
			ID:             s.ID,
			Portion:        s.Portion,
			Height:         s.Height,
			Cap1:           s.Cap1,
			Cap2:           s.Cap2,
			Cap3:           s.Cap3,
			Cap4:           s.Cap4,
			Cap5:           s.Cap5,
			Cap6:           s.Cap6,
			AverageDap:     s.AverageDap,
			BasalArea:      s.BasalArea,
			Volume:         s.Volume,
			RegisterDate:   s.RegisterDate,
			SpecieID:       s.SpecieID,
			ScientificName: s.ScientificName,
			Family:         s.Family,
			PopularName:    s.PopularName,
		})
	}

	return &PhytoAnalysisResponse{
		ID:              p.ID,
		Title:           p.Title,
		InitialDate:     p.InitialDate,
		PortionQuantity: p.PortionQuantity,
		PortionArea:     p.PortionArea,
		TotalArea:       p.TotalArea,
		SampledArea:     p.SampledArea,
		Description:     p.Description,
		ProjectID:       p.ProjectID,
		CreatedAt:       p.CreatedAt,
		UpdatedAt:       p.UpdatedAt,
		Project: &ProjectInfo{
			ID:       p.ProjectID,
			Title:    p.ProjectTitle,
			CNPJ:     p.ProjectCNPJ,
			Activity: p.ProjectActivity,
			ClientID: p.ProjectClientID,
		},
		Specimens: specimens,
	}
}
