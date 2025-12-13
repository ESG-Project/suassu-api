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
	VolumeTotalM3    float64            `json:"volumeTotalM3"`    // Volume total (m³)
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
	VolumeM3       float64   `json:"volumeM3"`    // volume individual (m³)
	DbhCm          float64   `json:"dbhCm"`       // DAP individual (cm)
	BasalAreaM2    float64   `json:"basalAreaM2"` // área basal individual (m²)
	StdDevDbhCm    float64   `json:"stdDevDbhCm"` // desvio padrão do DAP da espécie (cm)
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

// calcula ABI em cm² a partir dos CAPs do SpecimenWithSpecies
func calcABIFromSpecimen(s *types.SpecimenWithSpecies) float64 {
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
		// CR11.1: ABI = Σ (CAPi² / 4π)
		abi += (cap * cap) / (4 * math.Pi)
	}
	return abi
}

// volume em m³ a partir de ABI (cm²) e altura (m)
func calcVolumeFromABI(abiCm2, heightM float64) float64 {
	if abiCm2 <= 0 || heightM <= 0 {
		return 0
	}
	// CR11.4: G(m²) = ABI / 10.000
	g := abiCm2 / 10000.0
	// CR11.5: Volume = G(m²) × Height(m)
	return g * heightM
}

// DAP (cm) e área basal (m²) a partir da ABI em cm²
func calcDbhAndBasalFromABI(abiCm2 float64) (dbhCm, basalM2 float64) {
	if abiCm2 <= 0 {
		return 0, 0
	}
	// CR11.2 – CAP_mean = √( ABI × 4π )
	capMean := math.Sqrt(abiCm2 * 4 * math.Pi) // cm
	// CR11.3 – DBH(cm) = CAP_mean / π
	dbhCm = capMean / math.Pi // cm
	// CR11.4 – G(m²) = ABI / 10.000
	basalM2 = abiCm2 / 10000.0 // m²
	return dbhCm, basalM2
}

// ToPhytoAnalysisCompleteResponse converte análise completa para resposta HTTP
func ToPhytoAnalysisCompleteResponse(p *types.PhytoAnalysisComplete) *PhytoAnalysisResponse {
	specimens := make([]SpecimenResponse, 0, len(p.Specimens))
	uniqueSpecies := make(map[string]bool) // Para contar espécies únicas

	var (
		sumDbhCm  float64 // soma dos DAPs (cm)
		sumHeight float64 // soma das alturas (m)
		sumBasal  float64 // soma das áreas basais (m²)
		sumVolume float64 // soma dos volumes individuais (m³)
	)

	// Estrutura auxiliar para guardar métricas por espécime
	type specMetrics struct {
		s        *types.SpecimenWithSpecies
		abiCm2   float64
		dbhCm    float64
		basalM2  float64
		volumeM3 float64
	}

	metrics := make([]specMetrics, 0, len(p.Specimens))
	dapBySpecies := make(map[string][]float64)

	// Primeira passada: calcula métricas individuais e acumula para agregados
	for _, s := range p.Specimens {
		abi := calcABIFromSpecimen(s) // cm²

		var (
			dbhCm    float64
			basalM2  float64
			volumeM3 float64
		)

		if abi > 0 {
			dbhCm, basalM2 = calcDbhAndBasalFromABI(abi)
			volumeM3 = calcVolumeFromABI(abi, s.Height)

			sumDbhCm += dbhCm
			sumBasal += basalM2
			sumVolume += volumeM3
		}

		metrics = append(metrics, specMetrics{
			s:        s,
			abiCm2:   abi,
			dbhCm:    dbhCm,
			basalM2:  basalM2,
			volumeM3: volumeM3,
		})

		// chave para agrupar por espécie (preferência por nome científico)
		key := s.ScientificName
		if key == "" {
			key = s.SpecieID
		}
		if key != "" && dbhCm > 0 {
			dapBySpecies[key] = append(dapBySpecies[key], dbhCm)
		}

		// mapa de espécies únicas
		if s.ScientificName != "" {
			uniqueSpecies[s.ScientificName] = true
		}

		// alturas para média
		if s.Height > 0 {
			sumHeight += s.Height
		}
	}

	// Segunda passada: desvio padrão do DAP por espécie
	stdDevBySpecies := make(map[string]float64)
	for key, daps := range dapBySpecies {
		if len(daps) <= 1 {
			stdDevBySpecies[key] = 0
			continue
		}

		var sum float64
		for _, v := range daps {
			sum += v
		}
		mean := sum / float64(len(daps))

		var sq float64
		for _, v := range daps {
			diff := v - mean
			sq += diff * diff
		}

		// desvio padrão amostral: / (n-1)
		stdDevBySpecies[key] = math.Sqrt(sq / float64(len(daps)-1))
	}

	// Monta SpecimenResponse com DAP, área basal, volume e desvio padrão
	for _, m := range metrics {
		s := m.s

		key := s.ScientificName
		if key == "" {
			key = s.SpecieID
		}
		stdDev := stdDevBySpecies[key]

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
			VolumeM3:       m.volumeM3,
			DbhCm:          m.dbhCm,
			BasalAreaM2:    m.basalM2,
			StdDevDbhCm:    stdDev,
		})
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
		IndividualsCount: len(p.Specimens),
		SpeciesCount:     len(uniqueSpecies),
		MeanDBHCm:        meanDbhCm,
		MeanHeightM:      meanHeightM,
		DensityIndHa:     densityIndHa,
		VolumeTotalM3:    sumVolume,
		VolumePerHa:      volumePerHa,
		BasalAreaPerHa:   basalPerHa,
	}
}
