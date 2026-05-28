package specimendto

import (
	"math"
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
	VolumeM3        float64   `json:"volumeM3"`
	DbhCm           float64   `json:"dbhCm"`
	BasalAreaM2     float64   `json:"basalAreaM2"`
}

// ToSpecimenResponse converte tipos internos para resposta HTTP
func ToSpecimenResponse(s *types.SpecimenWithSpecies) *SpecimenResponse {
	volumeM3 := calcVolumeM3(s)
	abi := calcABI(s) // cm²
	dbhCm, basalM2 := calcDbhAndBasalFromABI(abi)

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
		VolumeM3:        volumeM3,
		DbhCm:           dbhCm,
		BasalAreaM2:     basalM2,
	}
}

// calcula ABI em cm²
func calcABI(s *types.SpecimenWithSpecies) float64 {
	caps := []float64{s.Cap1}

	if s.Cap2 != nil {
		caps = append(caps, *s.Cap2)
	}
	if s.Cap3 != nil {
		caps = append(caps, *s.Cap3)
	}
	if s.Cap4 != nil {
		caps = append(caps, *s.Cap4)
	}
	if s.Cap5 != nil {
		caps = append(caps, *s.Cap5)
	}
	if s.Cap6 != nil {
		caps = append(caps, *s.Cap6)
	}

	abi := 0.0
	for _, cap := range caps {
		abi += (cap * cap) / (4 * math.Pi)
	}
	return abi
}

// calcula volume em m³
func calcVolumeM3(s *types.SpecimenWithSpecies) float64 {
	abi := calcABI(s) // cm²

	// CR11.4: G(m²) = ABI / 10.000
	g := abi / 10000.0

	// CR11.5: Volume = G(m²) × Height (m)
	return g * s.Height
}

// DAP (cm) e área basal (m²) a partir da ABI em cm²
func calcDbhAndBasalFromABI(abiCm2 float64) (dbhCm, basalM2 float64) {
	if abiCm2 <= 0 {
		return 0, 0
	}

	// CR11.2 – CAP_mean = √( ABI × 4π )
	capMean := math.Sqrt(abiCm2 * 4 * math.Pi) // cm

	// CR11.3 – DBH(cm) = CAP_mean / π
	dbhCm = capMean / math.Pi

	// CR11.4 – G(m²) = ABI / 10.000
	basalM2 = abiCm2 / 10000.0

	return dbhCm, basalM2
}
