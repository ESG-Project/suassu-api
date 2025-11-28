package phytoanalysisdto

import (
	"math"
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
	RegisterDate time.Time `json:"registerDate"`
	// Nome científico da espécie (obrigatório)
	ScientificName string `json:"scientificName"`
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
	ID               string             `json:"id"`
	Title            string             `json:"title"`
	InitialDate      time.Time          `json:"initialDate"`
	PortionQuantity  int                `json:"portionQuantity"`
	PortionArea      float64            `json:"portionArea"`
	TotalArea        float64            `json:"totalArea"`
	SampledArea      float64            `json:"sampledArea"`
	Description      *string            `json:"description,omitempty"`
	ProjectID        string             `json:"projectId"`
	CreatedAt        time.Time          `json:"createdAt"`
	UpdatedAt        time.Time          `json:"updatedAt"`
	Project          *ProjectInfo       `json:"project,omitempty"`
	Specimens        []SpecimenResponse `json:"specimens,omitempty"`
	IndividualsCount int                `json:"individualsCount"` // Número de indivíduos (total de specimens)
	SpeciesCount     int                `json:"speciesCount"`     // Número de espécies (scientific names únicos)
	MeanDBHCm        float64            `json:"meanDbhCm"`        // DAP médio (cm)
	MeanHeightM      float64            `json:"meanHeightM"`      // Altura média (m)
	DensityIndHa     float64            `json:"densityIndHa"`     // Densidade (ind/ha)
	VolumeTotalM3    float64            `json:"volumeTotalM3"`    // Volume total (m³ ou mst)
	VolumePerHa      float64            `json:"volumePerHa"`      // Volume (m³/ha)
	BasalAreaPerHa   float64            `json:"basalAreaPerHa"`   // Área basal (m²/ha)
}

type ProjectInfo struct {
	ID       string          `json:"id"`
	Title    string          `json:"title"`
	CNPJ     *string         `json:"cnpj,omitempty"`
	Activity string          `json:"activity"`
	ClientID string          `json:"clientId"`
	Address  *ProjectAddress `json:"address,omitempty"`
}

type ProjectAddress struct {
	ZipCode      *string `json:"zipCode,omitempty"`
	State        *string `json:"state,omitempty"`
	City         *string `json:"city,omitempty"`
	Neighborhood *string `json:"neighborhood,omitempty"`
	Street       *string `json:"street,omitempty"`
	Num          *string `json:"num,omitempty"`
	Latitude     *string `json:"latitude,omitempty"`
	Longitude    *string `json:"longitude,omitempty"`
	AddInfo      *string `json:"addInfo,omitempty"`
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
	uniqueSpecies := make(map[string]bool) // Para contar espécies únicas

	var (
		sumDbhCm  float64 // soma dos DAPs (cm)
		sumHeight float64 // soma das alturas (m)
		sumBasal  float64 // soma das áreas basais (m²)
		sumVolume float64 // soma dos volumes individuais (m³) – por enquanto 0 se não houver campo
	)

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
			RegisterDate:   s.RegisterDate,
			SpecieID:       s.SpecieID,
			ScientificName: s.ScientificName,
			Family:         s.Family,
			PopularName:    s.PopularName,
		})

		// Adicionar nome científico ao mapa para contar espécies únicas
		if s.ScientificName != "" {
			uniqueSpecies[s.ScientificName] = true
		}

		if s.Cap1 > 0 {
			dapCm := s.Cap1 / math.Pi
			sumDbhCm += dapCm

			// Área basal individual (m²): π * (DAP_m / 2)²
			dapM := dapCm / 100.0
			basal := math.Pi * math.Pow(dapM/2, 2)
			sumBasal += basal
		}

		// 🔹 Altura
		if s.Height > 0 {
			sumHeight += s.Height
		}
	}

	n := len(p.Specimens)
	sampledAreaHa := p.SampledArea

	var (
		meanDbhCm    float64
		meanHeightM  float64
		densityIndHa float64
		volumePerHa  float64
		basalPerHa   float64
	)

	if n > 0 {
		meanDbhCm = sumDbhCm / float64(n)
		meanHeightM = sumHeight / float64(n)
	}

	if sampledAreaHa > 0 {
		densityIndHa = float64(n) / sampledAreaHa
		volumePerHa = sumVolume / sampledAreaHa
		basalPerHa = sumBasal / sampledAreaHa
	}

	// Montar endereço do projeto se houver dados
	var projectAddress *ProjectAddress
	if p.ProjectZipCode != nil || p.ProjectState != nil || p.ProjectCity != nil {
		projectAddress = &ProjectAddress{
			ZipCode:      p.ProjectZipCode,
			State:        p.ProjectState,
			City:         p.ProjectCity,
			Neighborhood: p.ProjectNeighborhood,
			Street:       p.ProjectStreet,
			Num:          p.ProjectNum,
			Latitude:     p.ProjectLatitude,
			Longitude:    p.ProjectLongitude,
			AddInfo:      p.ProjectAddInfo,
		}
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
			Address:  projectAddress,
		},
		Specimens:        specimens,
		IndividualsCount: len(p.Specimens),   // Total de specimens
		SpeciesCount:     len(uniqueSpecies), // Total de espécies únicas
		MeanDBHCm:        meanDbhCm,
		MeanHeightM:      meanHeightM,
		DensityIndHa:     densityIndHa,
		VolumeTotalM3:    sumVolume,
		VolumePerHa:      volumePerHa,
		BasalAreaPerHa:   basalPerHa,
	}
}
