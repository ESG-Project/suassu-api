package specimendto

import (
	"time"

	"github.com/ESG-Project/suassu-api/internal/app/types"
)

// CreateSpecimenRequest representa a requisição para criar um specimen
type CreateSpecimenRequest struct {
	Portion         string    `json:"portion"`
	Height          float64   `json:"height"`
	Cap1            float64   `json:"cap1"`
	Cap2            *float64  `json:"cap2,omitempty"`
	Cap3            *float64  `json:"cap3,omitempty"`
	Cap4            *float64  `json:"cap4,omitempty"`
	Cap5            *float64  `json:"cap5,omitempty"`
	Cap6            *float64  `json:"cap6,omitempty"`
	RegisterDate    time.Time `json:"registerDate"`
	PhytoAnalysisID string    `json:"phytoAnalysisId"`
	SpecieID        string    `json:"specieId"`
}

// UpdateSpecimenRequest representa a requisição para atualizar um specimen
type UpdateSpecimenRequest struct {
	Portion      string    `json:"portion"`
	Height       float64   `json:"height"`
	Cap1         float64   `json:"cap1"`
	Cap2         *float64  `json:"cap2,omitempty"`
	Cap3         *float64  `json:"cap3,omitempty"`
	Cap4         *float64  `json:"cap4,omitempty"`
	Cap5         *float64  `json:"cap5,omitempty"`
	Cap6         *float64  `json:"cap6,omitempty"`
	RegisterDate time.Time `json:"registerDate"`
	SpecieID     string    `json:"specieId"`
}

// SpecimenResponse representa a resposta de um specimen
type SpecimenResponse struct {
	ID              string    `json:"id"`
	Portion         string    `json:"portion"`
	Height          float64   `json:"height"`
	Cap1            float64   `json:"cap1"`
	Cap2            *float64  `json:"cap2,omitempty"`
	Cap3            *float64  `json:"cap3,omitempty"`
	Cap4            *float64  `json:"cap4,omitempty"`
	Cap5            *float64  `json:"cap5,omitempty"`
	Cap6            *float64  `json:"cap6,omitempty"`
	RegisterDate    time.Time `json:"registerDate"`
	PhytoAnalysisID string    `json:"phytoAnalysisId"`
	SpecieID        string    `json:"specieId"`
	ScientificName  string    `json:"scientificName"`
	Family          string    `json:"family"`
	PopularName     *string   `json:"popularName,omitempty"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

// ToSpecimenResponse converte tipos internos para resposta HTTP
func ToSpecimenResponse(s *types.SpecimenWithSpecies) *SpecimenResponse {
	return &SpecimenResponse{
		ID:              s.ID,
		Portion:         s.Portion,
		Height:          s.Height,
		Cap1:            s.Cap1,
		Cap2:            s.Cap2,
		Cap3:            s.Cap3,
		Cap4:            s.Cap4,
		Cap5:            s.Cap5,
		Cap6:            s.Cap6,
		RegisterDate:    s.RegisterDate,
		PhytoAnalysisID: s.PhytoAnalysisID,
		SpecieID:        s.SpecieID,
		ScientificName:  s.ScientificName,
		Family:          s.Family,
		PopularName:     s.PopularName,
		CreatedAt:       s.CreatedAt,
		UpdatedAt:       s.UpdatedAt,
	}
}

