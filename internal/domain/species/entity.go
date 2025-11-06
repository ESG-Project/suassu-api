package species

import (
	"errors"
	"strings"
	"time"
)

// Species representa a entidade de espécie no domínio
type Species struct {
	ID               string
	ScientificName   string
	Family           string
	PopularName      *string
	SpeciesDetailID  string
	CreatedAt        time.Time
	UpdatedAt        time.Time
	SpeciesDetail    *SpeciesLegislation
}

// SpeciesLegislation representa a legislação da espécie
type SpeciesLegislation struct {
	ID                   string
	LawScope             string // Federal, State, Municipal
	LawID                string
	IsLawActive          bool
	SpeciesFormFactor    float64
	IsSpeciesProtected   bool
	SpeciesThreatStatus  string // LC, CR, NT, EN, VU
	SpeciesOrigin        string // EX, EXI, N
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

// NewSpecies cria uma nova instância de Species
func NewSpecies(
	id, scientificName, family, speciesDetailID string,
) *Species {
	now := time.Now()
	return &Species{
		ID:              id,
		ScientificName:  scientificName,
		Family:          family,
		SpeciesDetailID: speciesDetailID,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}

// NewSpeciesLegislation cria uma nova instância de SpeciesLegislation
func NewSpeciesLegislation(
	id, lawScope, lawID string,
	isLawActive bool,
	speciesFormFactor float64,
	isSpeciesProtected bool,
	speciesThreatStatus, speciesOrigin string,
) *SpeciesLegislation {
	now := time.Now()
	return &SpeciesLegislation{
		ID:                  id,
		LawScope:            lawScope,
		LawID:               lawID,
		IsLawActive:         isLawActive,
		SpeciesFormFactor:   speciesFormFactor,
		IsSpeciesProtected:  isSpeciesProtected,
		SpeciesThreatStatus: speciesThreatStatus,
		SpeciesOrigin:       speciesOrigin,
		CreatedAt:           now,
		UpdatedAt:           now,
	}
}

// Validate valida se a espécie está em um estado válido
func (s *Species) Validate() error {
	if strings.TrimSpace(s.ScientificName) == "" {
		return errors.New("scientific name is required")
	}
	if strings.TrimSpace(s.Family) == "" {
		return errors.New("family is required")
	}
	if strings.TrimSpace(s.SpeciesDetailID) == "" {
		return errors.New("species detail ID is required")
	}
	return nil
}

// Validate valida se a legislação está em um estado válido
func (sl *SpeciesLegislation) Validate() error {
	validLawScopes := map[string]bool{"Federal": true, "State": true, "Municipal": true}
	if !validLawScopes[sl.LawScope] {
		return errors.New("invalid law scope")
	}

	validThreatStatuses := map[string]bool{"LC": true, "CR": true, "NT": true, "EN": true, "VU": true}
	if !validThreatStatuses[sl.SpeciesThreatStatus] {
		return errors.New("invalid threat status")
	}

	validOrigins := map[string]bool{"EX": true, "EXI": true, "N": true}
	if !validOrigins[sl.SpeciesOrigin] {
		return errors.New("invalid species origin")
	}

	if strings.TrimSpace(sl.LawID) == "" {
		return errors.New("law ID is required")
	}

	if sl.SpeciesFormFactor <= 0 {
		return errors.New("species form factor must be positive")
	}

	return nil
}

// SetPopularName define o nome popular da espécie
func (s *Species) SetPopularName(popularName *string) {
	s.PopularName = popularName
}

// SetSpeciesDetail define os detalhes da legislação
func (s *Species) SetSpeciesDetail(detail *SpeciesLegislation) {
	s.SpeciesDetail = detail
}

