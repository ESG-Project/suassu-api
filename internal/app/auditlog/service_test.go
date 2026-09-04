package auditlog_test

import (
	"context"
	"errors"
	"testing"

	"github.com/ESG-Project/suassu-api/internal/app/auditlog"
	domainauditlog "github.com/ESG-Project/suassu-api/internal/domain/auditlog"
	"github.com/stretchr/testify/require"
)

type fakeRepo struct {
	saved []domainauditlog.Log
	err   error
}

func (f *fakeRepo) Create(ctx context.Context, l *domainauditlog.Log) error {
	if f.err != nil {
		return f.err
	}
	f.saved = append(f.saved, *l)
	return nil
}

func TestService_Record(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("grava a entrada com ator, empresa e descrição", func(t *testing.T) {
		repo := &fakeRepo{}
		actor := "user-1"

		auditlog.NewService(repo).Record(ctx, &actor, "ent-1", "Registro de produto X, ")

		require.Len(t, repo.saved, 1)
		entry := repo.saved[0]
		require.NotEmpty(t, entry.ID)
		require.Equal(t, "user-1", *entry.ActorID)
		require.Equal(t, "ent-1", entry.EnterpriseID)
		require.Equal(t, "Registro de produto X, ", entry.Description)
		require.False(t, entry.CreatedAt.IsZero())
	})

	t.Run("aceita evento sem ator (gerado pelo sistema)", func(t *testing.T) {
		repo := &fakeRepo{}

		auditlog.NewService(repo).Record(ctx, nil, "ent-1", "Evento automático")

		require.Len(t, repo.saved, 1)
		require.Nil(t, repo.saved[0].ActorID)
	})

	t.Run("falha do repositório não interrompe quem chamou", func(t *testing.T) {
		repo := &fakeRepo{err: errors.New("db down")}

		require.NotPanics(t, func() {
			auditlog.NewService(repo).Record(ctx, nil, "ent-1", "Evento")
		})
		require.Empty(t, repo.saved)
	})

	t.Run("entrada inválida é descartada em vez de ir ao banco", func(t *testing.T) {
		repo := &fakeRepo{}

		auditlog.NewService(repo).Record(ctx, nil, "", "Evento sem empresa")

		require.Empty(t, repo.saved)
	})
}
