package specimen

import (
	"errors"
	"strings"
	"time"
)

// Specimen representa a entidade de espécime no domínio
type Specimen struct {
	ID              string
	Portion         string
	Height          float64
	Cap1            float64
	Cap2            *float64
	Cap3            *float64
	Cap4            *float64
	Cap5            *float64
	Cap6            *float64
	AverageDap      float64
	BasalArea       float64
	Volume          float64
	RegisterDate    time.Time
	PhytoAnalysisID string
	SpecieID        string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// NewSpecimen cria uma nova instância de Specimen
func NewSpecimen(
	id string,
	portion string,
	height, cap1, averageDap, basalArea, volume float64,
	registerDate time.Time,
	phytoAnalysisID, specieID string,
) *Specimen {
	now := time.Now()
	return &Specimen{
		ID:              id,
		Portion:         portion,
		Height:          height,
		Cap1:            cap1,
		AverageDap:      averageDap,
		BasalArea:       basalArea,
		Volume:          volume,
		RegisterDate:    registerDate,
		PhytoAnalysisID: phytoAnalysisID,
		SpecieID:        specieID,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}

// Validate valida se o espécime está em um estado válido
func (s *Specimen) Validate() error {
	if strings.TrimSpace(s.Portion) == "" {
		return errors.New("portion is required")
	}
	if s.Height <= 0 {
		return errors.New("height must be positive")
	}
	if s.Cap1 <= 0 {
		return errors.New("cap1 must be positive")
	}
	if s.AverageDap <= 0 {
		return errors.New("average DAP must be positive")
	}
	if s.BasalArea <= 0 {
		return errors.New("basal area must be positive")
	}
	if s.Volume <= 0 {
		return errors.New("volume must be positive")
	}
	if strings.TrimSpace(s.PhytoAnalysisID) == "" {
		return errors.New("phyto analysis ID is required")
	}
	if strings.TrimSpace(s.SpecieID) == "" {
		return errors.New("specie ID is required")
	}
	if s.RegisterDate.IsZero() {
		return errors.New("register date is required")
	}
	return nil
}

// SetOptionalCaps define as circunferências opcionais
func (s *Specimen) SetOptionalCaps(cap2, cap3, cap4, cap5, cap6 *float64) {
	s.Cap2 = cap2
	s.Cap3 = cap3
	s.Cap4 = cap4
	s.Cap5 = cap5
	s.Cap6 = cap6
}

