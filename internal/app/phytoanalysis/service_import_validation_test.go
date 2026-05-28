package phytoanalysis

import (
	"context"
	"testing"
	"time"

	"github.com/ESG-Project/suassu-api/internal/app/types"
	"github.com/ESG-Project/suassu-api/internal/apperr"
	domainphyto "github.com/ESG-Project/suassu-api/internal/domain/phytoanalysis"
	postgres "github.com/ESG-Project/suassu-api/internal/infra/db/postgres"
	"github.com/stretchr/testify/require"
)

type noopRepo struct{}

func (n *noopRepo) Create(ctx context.Context, p *domainphyto.PhytoAnalysis) error {
	return nil
}

func (n *noopRepo) GetByID(ctx context.Context, id string) (*types.PhytoAnalysisWithProject, error) {
	return nil, nil
}

func (n *noopRepo) ListByProject(ctx context.Context, projectID string) ([]*types.PhytoAnalysisWithProject, error) {
	return nil, nil
}

func (n *noopRepo) ListByEnterprise(ctx context.Context, enterpriseID string) ([]*types.PhytoAnalysisWithProject, error) {
	return nil, nil
}

func (n *noopRepo) ListAll(ctx context.Context, limit, offset int32) ([]*types.PhytoAnalysisWithProject, error) {
	return nil, nil
}

func (n *noopRepo) Update(ctx context.Context, p *domainphyto.PhytoAnalysis) error {
	return nil
}

func (n *noopRepo) Delete(ctx context.Context, id string) error {
	return nil
}

func (n *noopRepo) GetWithSpecimens(ctx context.Context, id string) (*types.PhytoAnalysisComplete, error) {
	return nil, nil
}

type mockTxManager struct {
	runInTxFunc func(ctx context.Context, fn func(postgres.Repos) error) error
}

func (m *mockTxManager) RunInTx(ctx context.Context, fn func(postgres.Repos) error) error {
	if m.runInTxFunc != nil {
		return m.runInTxFunc(ctx, fn)
	}
	return fn(postgres.Repos{})
}

func TestNormalizeAndValidateSpecimens_IgnoresBlankRows(t *testing.T) {
	t.Parallel()
	registerDate := time.Date(2026, time.January, 10, 0, 0, 0, 0, time.UTC)

	rows, invalidRows := normalizeAndValidateSpecimens([]SpecimenInput{
		{},
		{
			Portion:        " A1 ",
			Height:         12,
			Cap1:           22,
			RegisterDate:   registerDate,
			ScientificName: "  Copaifera langsdorffii  ",
		},
		{},
	})

	require.Empty(t, invalidRows)
	require.Len(t, rows, 1)
	require.Equal(t, 2, rows[0].RowNumber)
	require.Equal(t, "A1", rows[0].Specimen.Portion)
	require.Equal(t, "Copaifera langsdorffii", rows[0].Specimen.ScientificName)
}

func TestNormalizeAndValidateSpecimens_ReportsPartialRows(t *testing.T) {
	t.Parallel()

	rows, invalidRows := normalizeAndValidateSpecimens([]SpecimenInput{{Portion: "A1"}})

	require.Empty(t, rows)
	require.Len(t, invalidRows, 1)
	require.Equal(t, 1, invalidRows[0].RowNumber)
	require.Contains(t, invalidRows[0].Errors, "height must be positive")
	require.Contains(t, invalidRows[0].Errors, "cap1 must be positive")
	require.Contains(t, invalidRows[0].Errors, "register date is required")
	require.Contains(t, invalidRows[0].Errors, "scientific name is required")
}

func TestCreate_ReturnsInvalidRowsDetails(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	svc := NewService(&noopRepo{}, &mockTxManager{
		runInTxFunc: func(ctx context.Context, fn func(postgres.Repos) error) error {
			return fn(postgres.Repos{})
		},
	})

	_, err := svc.Create(ctx, CreateInput{
		Title:           "Analise",
		InitialDate:     time.Now(),
		PortionQuantity: 1,
		PortionArea:     100,
		TotalArea:       100,
		ProjectID:       "proj-1",
		Specimens: []SpecimenInput{
			{},
			{Portion: "A1"},
		},
	})

	require.Error(t, err)
	require.Equal(t, apperr.CodeInvalid, apperr.CodeOf(err))

	appErr, ok := err.(*apperr.Error)
	require.True(t, ok)
	require.Equal(t, "invalid specimen rows", appErr.Msg)

	invalidRowsRaw, exists := appErr.Fields["invalidRows"]
	require.True(t, exists)

	invalidRows, ok := invalidRowsRaw.([]invalidSpecimenRow)
	require.True(t, ok)
	require.Len(t, invalidRows, 1)
	require.Equal(t, 2, invalidRows[0].RowNumber)
	require.Contains(t, invalidRows[0].Errors, "height must be positive")
}
