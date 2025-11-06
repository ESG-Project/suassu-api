package types

import "time"

// PhytoAnalysisWithProject representa uma análise fitossociológica com dados do projeto
type PhytoAnalysisWithProject struct {
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
	// Dados do projeto
	ProjectTitle    string
	ProjectCNPJ     *string
	ProjectActivity string
	ProjectClientID string
}

// PhytoAnalysisComplete representa uma análise fitossociológica completa com projeto e espécimes
type PhytoAnalysisComplete struct {
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
	// Dados do projeto
	ProjectTitle    string
	ProjectCNPJ     *string
	ProjectActivity string
	ProjectClientID string
	// Lista de espécimes
	Specimens []*SpecimenWithSpecies
}

// SpecimenWithSpecies representa um espécime com dados da espécie
type SpecimenWithSpecies struct {
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
	// Dados da espécie
	ScientificName string
	Family         string
	PopularName    *string
}

// SpeciesWithLegislation representa uma espécie com dados da legislação
type SpeciesWithLegislation struct {
	ID              string
	ScientificName  string
	Family          string
	PopularName     *string
	SpeciesDetailID string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	// Dados da legislação
	LawScope            string
	LawID               string
	IsLawActive         bool
	SpeciesFormFactor   float64
	IsSpeciesProtected  bool
	SpeciesThreatStatus string
	SpeciesOrigin       string
}

