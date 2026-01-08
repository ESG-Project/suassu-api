package phytoanalysis

import (
	"errors"
	"strings"
	"time"
)

// PhytoAnalysis representa a entidade de análise fitossociológica no domínio
type PhytoAnalysis struct {
	ID              string
	Title           string
	InitialDate     time.Time
	PortionQuantity int
	PortionArea     float64
	TotalArea       float64
	SampledArea     float64
	Description     *string
	ProjectID       string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// NewPhytoAnalysis cria uma nova instância de PhytoAnalysis
func NewPhytoAnalysis(
	id string,
	title string,
	initialDate time.Time,
	portionQuantity int,
	portionArea, totalArea, sampledArea float64,
	projectID string,
) *PhytoAnalysis {
	now := time.Now()
	return &PhytoAnalysis{
		ID:              id,
		Title:           title,
		InitialDate:     initialDate,
		PortionQuantity: portionQuantity,
		PortionArea:     portionArea,
		TotalArea:       totalArea,
		SampledArea:     sampledArea,
		ProjectID:       projectID,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}

// Validate valida se a análise fitossociológica está em um estado válido
func (p *PhytoAnalysis) Validate() error {
	if strings.TrimSpace(p.Title) == "" {
		return errors.New("title is required")
	}
	if strings.TrimSpace(p.ProjectID) == "" {
		return errors.New("project ID is required")
	}
	if p.PortionQuantity <= 0 {
		return errors.New("portion quantity must be positive")
	}
	if p.PortionArea <= 0 {
		return errors.New("portion area must be positive")
	}
	if p.TotalArea <= 0 {
		return errors.New("total area must be positive")
	}
	if p.SampledArea <= 0 {
		return errors.New("sampled area must be positive")
	}
	if p.InitialDate.IsZero() {
		return errors.New("initial date is required")
	}
	return nil
}

// SetDescription define a descrição da análise
func (p *PhytoAnalysis) SetDescription(description *string) {
	p.Description = description
}

// Update atualiza os dados da análise
func (p *PhytoAnalysis) Update(
	title string,
	initialDate time.Time,
	portionQuantity int,
	portionArea, totalArea, sampledArea float64,
	description *string,
) {
	p.Title = title
	p.InitialDate = initialDate
	p.PortionQuantity = portionQuantity
	p.PortionArea = portionArea
	p.TotalArea = totalArea
	p.SampledArea = sampledArea
	p.Description = description
	p.UpdatedAt = time.Now()
}
