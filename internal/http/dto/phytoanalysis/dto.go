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
	ID               string                       `json:"id"`
	Title            string                       `json:"title"`
	InitialDate      time.Time                    `json:"initialDate"`
	PortionQuantity  int                          `json:"portionQuantity"`
	PortionArea      float64                      `json:"portionArea"`
	TotalArea        float64                      `json:"totalArea"`
	SampledArea      float64                      `json:"sampledArea"`
	Description      *string                      `json:"description,omitempty"`
	ProjectID        string                       `json:"projectId"`
	CreatedAt        time.Time                    `json:"createdAt"`
	UpdatedAt        time.Time                    `json:"updatedAt"`
	Project          *ProjectInfo                 `json:"project,omitempty"`
	Specimens        []SpecimenResponse           `json:"specimens,omitempty"`
	IndividualsCount int                          `json:"individualsCount"` // Número de indivíduos (total de specimens)
	SpeciesCount     int                          `json:"speciesCount"`     // Número de espécies (scientific names únicos)
	Indicators       *PhytosociologicalIndicators `json:"indicators"`       // Indicadores fitossociológicos calculados
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

// SpeciesPhytosociologicalData representa dados fitossociológicos por espécie
type SpeciesPhytosociologicalData struct {
	ScientificName string  `json:"scientificName"`
	DA             float64 `json:"da"` // Densidade Absoluta (ind/ha)
	DR             float64 `json:"dr"` // Densidade Relativa (%)
	FA             float64 `json:"fa"` // Frequência Absoluta (%)
}

// CollectorCurvePoint representa um ponto da curva coletor
type CollectorCurvePoint struct {
	CumulativeArea  float64 `json:"cumulativeArea"`  // Área acumulada (m²)
	ObservedSpecies int     `json:"observedSpecies"` // Número acumulado de espécies observadas
	TrendSpecies    float64 `json:"trendSpecies"`    // Número acumulado de espécies (tendência)
}

// CollectorCurveData representa os dados da curva coletor
type CollectorCurveData struct {
	Points []CollectorCurvePoint `json:"points"` // Pontos da curva
}

// PhytosociologicalIndicators representa os indicadores calculados da análise
type PhytosociologicalIndicators struct {
	// Campos básicos (já existentes)
	IndividualsCount int     `json:"individualsCount"` // Número de indivíduos
	SpeciesCount     int     `json:"speciesCount"`     // Número de espécies
	PlotsCount       int     `json:"plotsCount"`       // Número de parcelas
	PlotsArea        float64 `json:"plotsArea"`        // Área de parcelas (m²)

	// Campos calculados
	Density             *float64 `json:"density,omitempty"`             // Densidade Geral (ind/ha)
	BasalArea           *float64 `json:"basalArea,omitempty"`           // Área basal (m²/ha)
	Volume              *float64 `json:"volume,omitempty"`              // Volume (m³/ha)
	ReplacementVolume   *float64 `json:"replacementVolume,omitempty"`   // Volume de reposição (m³)
	SampledAreaHa       *float64 `json:"sampledAreaHa,omitempty"`       // Área amostrada (ha)
	ShannonIndex        *float64 `json:"shannonIndex,omitempty"`        // Índice de Shannon (H')
	SimpsonIndex        *float64 `json:"simpsonIndex,omitempty"`        // Índice de Simpson (D)
	PielouEvennessIndex *float64 `json:"pielouEvennessIndex,omitempty"` // Índice de Equabilidade de Pielou (J')

	// Dados para gráficos
	SpeciesData    []SpeciesPhytosociologicalData `json:"speciesData,omitempty"`    // DA, DR, FA por espécie
	CollectorCurve *CollectorCurveData            `json:"collectorCurve,omitempty"` // Dados da curva coletor
}

const (
	pi = 3.14159265358979323846
)

// calculateABI calcula a Área Basal Individual (cm²)
// ABI = (CAP1²)/(4π) + (CAP2²)/(4π) + ... + (CAP6²)/(4π)
func calculateABI(cap1 float64, cap2, cap3, cap4, cap5, cap6 *float64) float64 {
	abi := (cap1 * cap1) / (4 * pi)

	if cap2 != nil {
		abi += (*cap2 * *cap2) / (4 * pi)
	}
	if cap3 != nil {
		abi += (*cap3 * *cap3) / (4 * pi)
	}
	if cap4 != nil {
		abi += (*cap4 * *cap4) / (4 * pi)
	}
	if cap5 != nil {
		abi += (*cap5 * *cap5) / (4 * pi)
	}
	if cap6 != nil {
		abi += (*cap6 * *cap6) / (4 * pi)
	}

	return abi
}

// calculateBasalArea calcula a Área Basal (G) em m²
// G(m²) = ABI / 10,000
func calculateBasalArea(abi float64) float64 {
	return abi / 10000.0
}

// calculateVolume calcula o Volume em m³
// Vol = G(m²) × Height
func calculateVolume(basalArea, height float64) float64 {
	return basalArea * height
}

// calculateCollectorCurve calcula os dados da curva coletor
// Retorna pontos com área acumulada e número acumulado de espécies
// Os specimens já vêm ordenados por parcela e data do banco de dados
func calculateCollectorCurve(specimens []*types.SpecimenWithSpecies, portionArea float64) *CollectorCurveData {
	if len(specimens) == 0 {
		return nil
	}

	// Processar specimens na ordem que vêm (já ordenados por parcela e data)
	// Agrupar por parcela mantendo a ordem
	type plotData struct {
		plot    string
		species map[string]bool
	}

	plotList := make([]plotData, 0)
	plotMap := make(map[string]int) // parcela -> índice na lista

	for _, s := range specimens {
		if s.ScientificName == "" {
			continue
		}

		idx, exists := plotMap[s.Portion]
		if !exists {
			// Nova parcela
			idx = len(plotList)
			plotList = append(plotList, plotData{
				plot:    s.Portion,
				species: make(map[string]bool),
			})
			plotMap[s.Portion] = idx
		}

		// Adicionar espécie à parcela
		plotList[idx].species[s.ScientificName] = true
	}

	// Calcular pontos da curva
	points := make([]CollectorCurvePoint, 0, len(plotList)+1)
	cumulativeArea := 0.0
	observedSpeciesSet := make(map[string]bool)

	// Adicionar ponto inicial (0, 0)
	points = append(points, CollectorCurvePoint{
		CumulativeArea:  0,
		ObservedSpecies: 0,
		TrendSpecies:    0,
	})

	// Para cada parcela na ordem
	for _, plotInfo := range plotList {
		// Adicionar área da parcela
		cumulativeArea += portionArea

		// Adicionar espécies encontradas nesta parcela ao conjunto total
		for species := range plotInfo.species {
			observedSpeciesSet[species] = true
		}

		// Adicionar ponto após processar esta parcela
		points = append(points, CollectorCurvePoint{
			CumulativeArea:  cumulativeArea,
			ObservedSpecies: len(observedSpeciesSet),
			TrendSpecies:    0, // Será calculado depois
		})
	}

	// Calcular curva de tendência usando regressão linear simples
	// y = a + b*x, onde y é número de espécies e x é área acumulada
	if len(points) > 1 {
		// Calcular coeficientes da regressão linear
		n := float64(len(points))
		var sumX, sumY, sumXY, sumX2 float64

		for _, p := range points {
			x := p.CumulativeArea
			y := float64(p.ObservedSpecies)
			sumX += x
			sumY += y
			sumXY += x * y
			sumX2 += x * x
		}

		// Calcular coeficientes: b = (n*ΣXY - ΣX*ΣY) / (n*ΣX² - (ΣX)²)
		denominator := n*sumX2 - sumX*sumX
		var b, a float64

		if denominator != 0 {
			b = (n*sumXY - sumX*sumY) / denominator
			a = (sumY - b*sumX) / n
		}

		// Aplicar tendência aos pontos
		for i := range points {
			points[i].TrendSpecies = a + b*points[i].CumulativeArea
			// Garantir que a tendência não seja negativa
			if points[i].TrendSpecies < 0 {
				points[i].TrendSpecies = 0
			}
		}
	}

	return &CollectorCurveData{
		Points: points,
	}
}

// calculatePhytosociologicalIndicators calcula todos os indicadores fitossociológicos
func calculatePhytosociologicalIndicators(p *types.PhytoAnalysisComplete) *PhytosociologicalIndicators {
	if len(p.Specimens) == 0 {
		return &PhytosociologicalIndicators{
			IndividualsCount: 0,
			SpeciesCount:     0,
			PlotsCount:       p.PortionQuantity,
			PlotsArea:        p.PortionArea * float64(p.PortionQuantity),
		}
	}

	// Constantes
	N := len(p.Specimens) // Número total de indivíduos

	// Contar espécies únicas e indivíduos por espécie
	uniqueSpecies := make(map[string]bool)
	speciesCount := make(map[string]int)             // n_i: número de indivíduos por espécie
	speciesPlots := make(map[string]map[string]bool) // parcelas onde cada espécie ocorre

	for _, s := range p.Specimens {
		if s.ScientificName != "" {
			uniqueSpecies[s.ScientificName] = true
			speciesCount[s.ScientificName]++

			// Registrar parcela onde a espécie ocorre
			if speciesPlots[s.ScientificName] == nil {
				speciesPlots[s.ScientificName] = make(map[string]bool)
			}
			speciesPlots[s.ScientificName][s.Portion] = true
		}
	}

	S := len(uniqueSpecies) // Número de espécies
	P := p.PortionQuantity  // Número de parcelas

	// Calcular área de parcelas (m²)
	plotsArea := p.PortionArea * float64(P)

	// Calcular área amostrada em hectares
	sampledAreaHa := plotsArea / 10000.0

	// Calcular densidade (ind/ha)
	var density *float64
	if sampledAreaHa > 0 {
		d := float64(N) / sampledAreaHa
		density = &d
	}

	// Calcular área basal total e volume total
	var totalBasalArea float64 // em m²
	var totalVolume float64    // em m³

	for _, s := range p.Specimens {
		abi := calculateABI(s.Cap1, s.Cap2, s.Cap3, s.Cap4, s.Cap5, s.Cap6)
		g := calculateBasalArea(abi)
		vol := calculateVolume(g, s.Height)

		totalBasalArea += g
		totalVolume += vol
	}

	// Calcular área basal por hectare (m²/ha)
	var basalArea *float64
	if sampledAreaHa > 0 {
		ba := totalBasalArea / sampledAreaHa
		basalArea = &ba
	}

	// Calcular volume por hectare (m³/ha)
	var volume *float64
	if sampledAreaHa > 0 {
		v := totalVolume / sampledAreaHa
		volume = &v
	}

	// Volume de reposição (m³) - valor agregado, não dividido por ha
	replacementVolume := totalVolume

	// Calcular índices de diversidade
	var shannonIndex *float64
	var simpsonIndex *float64
	var pielouIndex *float64

	if S > 0 && N > 0 {
		// Índice de Shannon: H' = -Σ(p_i × ln(p_i))
		var shannon float64
		// Índice de Simpson: D = Σ(p_i²)
		var simpson float64

		for _, n_i := range speciesCount {
			p_i := float64(n_i) / float64(N)
			if p_i > 0 {
				shannon -= p_i * math.Log(p_i)
				simpson += p_i * p_i
			}
		}

		shannonIndex = &shannon
		simpsonIndex = &simpson

		// Índice de Pielou: J' = H' / ln(S)
		if S > 1 {
			lnS := math.Log(float64(S))
			if lnS > 0 {
				pielou := shannon / lnS
				pielouIndex = &pielou
			}
		}
	}

	// Calcular dados para gráficos (DA, DR, FA por espécie)
	speciesData := make([]SpeciesPhytosociologicalData, 0, S)

	for speciesName := range uniqueSpecies {
		n_i := speciesCount[speciesName]
		p_i := len(speciesPlots[speciesName]) // número de parcelas onde a espécie ocorre

		// DA_i = n_i / Área_amostrada_ha
		var da float64
		if sampledAreaHa > 0 {
			da = float64(n_i) / sampledAreaHa
		}

		// DR_i = (n_i / N) × 100
		var dr float64
		if N > 0 {
			dr = (float64(n_i) / float64(N)) * 100.0
		}

		// FA_i = (p_i / P) × 100
		var fa float64
		if P > 0 {
			fa = (float64(p_i) / float64(P)) * 100.0
		}

		speciesData = append(speciesData, SpeciesPhytosociologicalData{
			ScientificName: speciesName,
			DA:             da,
			DR:             dr,
			FA:             fa,
		})
	}

	// Calcular curva coletor
	collectorCurve := calculateCollectorCurve(p.Specimens, p.PortionArea)

	return &PhytosociologicalIndicators{
		IndividualsCount:    N,
		SpeciesCount:        S,
		PlotsCount:          P,
		PlotsArea:           plotsArea,
		Density:             density,
		BasalArea:           basalArea,
		Volume:              volume,
		ReplacementVolume:   &replacementVolume,
		SampledAreaHa:       &sampledAreaHa,
		ShannonIndex:        shannonIndex,
		SimpsonIndex:        simpsonIndex,
		PielouEvennessIndex: pielouIndex,
		SpeciesData:         speciesData,
		CollectorCurve:      collectorCurve,
	}
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

	// Calcular indicadores fitossociológicos
	indicators := calculatePhytosociologicalIndicators(p)

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
		Indicators:       indicators,
	}
}
