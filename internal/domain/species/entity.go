package species

import (
	"errors"
	"strings"
	"time"
)

// Species representa a entidade de espécie no domínio
type Species struct {
	ID             string
	ScientificName string
	Family         string
	PopularName    *string
	Habit          *string // ARB, ANF, ARV, EME FIX, FLU FIX, FLU LIV, HERB, PAL, TREP
	CreatedAt      time.Time
	UpdatedAt      time.Time
	Legislations   []*SpeciesLegislation
}

// SpeciesLegislation representa a legislação da espécie
type SpeciesLegislation struct {
	ID                  string
	LawScope            string // FEDERAL, STATE, MUNICIPAL
	LawID               *string
	IsLawActive         bool
	SpeciesFormFactor   float64
	IsSpeciesProtected  bool
	SpeciesThreatStatus string // LC, CR, NT, EN, VU
	SpeciesOrigin       string // EX, EXI, N
	SuccessionalEcology string // P, IS, S, C, LS, MS, AS
	SpeciesID           *string
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// NewSpecies cria uma nova instância de Species
func NewSpecies(
	id, scientificName, family string,
) *Species {
	now := time.Now()
	return &Species{
		ID:             id,
		ScientificName: scientificName,
		Family:         family,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

// NewSpeciesLegislation cria uma nova instância de SpeciesLegislation
func NewSpeciesLegislation(
	id, lawScope string,
	lawID *string,
	isLawActive bool,
	speciesFormFactor float64,
	isSpeciesProtected bool,
	speciesThreatStatus, speciesOrigin, successionalEcology string,
	speciesID *string,
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
		SuccessionalEcology: successionalEcology,
		SpeciesID:           speciesID,
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

	// Validar habit se fornecido
	if s.Habit != nil {
		validHabits := map[string]bool{
			"ARB": true, "ANF": true, "ARV": true, "EME FIX": true,
			"FLU FIX": true, "FLU LIV": true, "HERB": true, "PAL": true, "TREP": true,
		}
		if !validHabits[*s.Habit] {
			return errors.New("invalid habit")
		}
	}

	return nil
}

// Validate valida se a legislação está em um estado válido
func (sl *SpeciesLegislation) Validate() error {
	validLawScopes := map[string]bool{"FEDERAL": true, "STATE": true, "MUNICIPAL": true}
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

	validSuccessionalEcologies := map[string]bool{"P": true, "IS": true, "S": true, "C": true, "LS": true, "MS": true, "AS": true}
	if !validSuccessionalEcologies[sl.SuccessionalEcology] {
		return errors.New("invalid successional ecology")
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

// SetHabit define o hábito da espécie
func (s *Species) SetHabit(habit *string) {
	s.Habit = habit
}

// AddLegislation adiciona uma legislação à espécie
func (s *Species) AddLegislation(legislation *SpeciesLegislation) {
	s.Legislations = append(s.Legislations, legislation)
}
