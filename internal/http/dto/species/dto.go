package speciesdto

import (
	"time"

	"github.com/ESG-Project/suassu-api/internal/app/types"
)

// SpeciesResponse representa a resposta de uma espécie
type SpeciesResponse struct {
	ID             string                `json:"id"`
	ScientificName string                `json:"scientificName"`
	Family         string                `json:"family"`
	PopularName    *string               `json:"popularName,omitempty"`
	Habit          *string               `json:"habit,omitempty"`
	CreatedAt      time.Time             `json:"createdAt"`
	UpdatedAt      time.Time             `json:"updatedAt"`
	Legislations   []LegislationResponse `json:"legislations,omitempty"`
}

// LegislationResponse representa a resposta de uma legislação de espécie
type LegislationResponse struct {
	ID                  string    `json:"id"`
	LawScope            string    `json:"lawScope"`
	LawID               *string   `json:"lawId,omitempty"`
	IsLawActive         bool      `json:"isLawActive"`
	SpeciesFormFactor   float64   `json:"speciesFormFactor"`
	IsSpeciesProtected  bool      `json:"isSpeciesProtected"`
	SpeciesThreatStatus string    `json:"speciesThreatStatus"`
	SpeciesOrigin       string    `json:"speciesOrigin"`
	SuccessionalEcology string    `json:"successionalEcology"`
	CreatedAt           time.Time `json:"createdAt"`
	UpdatedAt           time.Time `json:"updatedAt"`
}

// ToSpeciesResponse converte tipos internos para resposta HTTP
func ToSpeciesResponse(s *types.SpeciesWithLegislation) *SpeciesResponse {
	legislations := make([]LegislationResponse, 0, len(s.Legislations))
	for _, l := range s.Legislations {
		legislations = append(legislations, LegislationResponse{
			ID:                  l.ID,
			LawScope:            l.LawScope,
			LawID:               l.LawID,
			IsLawActive:         l.IsLawActive,
			SpeciesFormFactor:   l.SpeciesFormFactor,
			IsSpeciesProtected:  l.IsSpeciesProtected,
			SpeciesThreatStatus: l.SpeciesThreatStatus,
			SpeciesOrigin:       l.SpeciesOrigin,
			SuccessionalEcology: l.SuccessionalEcology,
			CreatedAt:           l.CreatedAt,
			UpdatedAt:           l.UpdatedAt,
		})
	}

	return &SpeciesResponse{
		ID:             s.ID,
		ScientificName: s.ScientificName,
		Family:         s.Family,
		PopularName:    s.PopularName,
		Habit:          s.Habit,
		CreatedAt:      s.CreatedAt,
		UpdatedAt:      s.UpdatedAt,
		Legislations:   legislations,
	}
}
