package phytoanalysis

import (
	"context"
	"testing"

	"github.com/ESG-Project/suassu-api/internal/app/types"
	"github.com/ESG-Project/suassu-api/internal/apperr"
	"github.com/stretchr/testify/require"
)

func TestEnsureUniqueTitleForProject(t *testing.T) {
	t.Parallel()

	repo := &noopRepo{
		phytos: []*types.PhytoAnalysisWithProject{
			{Title: "Inventário Florestal"},
		},
	}

	err := ensureUniqueTitleForProject(
		context.Background(),
		repo,
		"project-1",
		"  INVENTÁRIO FLORESTAL  ",
	)

	require.Error(t, err)
	require.Equal(t, apperr.CodeConflict, apperr.CodeOf(err))
}
