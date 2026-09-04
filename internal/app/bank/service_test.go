package bank_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	appbank "github.com/ESG-Project/suassu-api/internal/app/bank"
	"github.com/ESG-Project/suassu-api/internal/app/types"
	"github.com/ESG-Project/suassu-api/internal/apperr"
	domainbank "github.com/ESG-Project/suassu-api/internal/domain/bank"
	"github.com/stretchr/testify/require"
)

type fakeRepo struct {
	byCode      map[string]*domainbank.Bank
	links       map[string]*domainbank.EnterpriseBank // chave: bankID+"/"+enterpriseID
	byID        map[string]*types.EnterpriseBankDetail
	rows        []types.EnterpriseBankRow
	createdBank *domainbank.Bank
	createdLink *domainbank.EnterpriseBank
	deletedID   string
	err         error
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{
		byCode: map[string]*domainbank.Bank{},
		links:  map[string]*domainbank.EnterpriseBank{},
		byID:   map[string]*types.EnterpriseBankDetail{},
	}
}

func (f *fakeRepo) GetByCode(ctx context.Context, code string) (*domainbank.Bank, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.byCode[code], nil
}

func (f *fakeRepo) Create(ctx context.Context, b *domainbank.Bank) error {
	f.createdBank = b
	f.byCode[b.Code] = b
	return nil
}

func (f *fakeRepo) GetEnterpriseBank(ctx context.Context, bankID, enterpriseID string) (*domainbank.EnterpriseBank, error) {
	return f.links[bankID+"/"+enterpriseID], nil
}

func (f *fakeRepo) GetEnterpriseBankByID(ctx context.Context, id string) (*types.EnterpriseBankDetail, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.byID[id], nil
}

func (f *fakeRepo) CreateEnterpriseBank(ctx context.Context, eb *domainbank.EnterpriseBank) error {
	f.createdLink = eb
	return nil
}

func (f *fakeRepo) DeleteEnterpriseBank(ctx context.Context, id string) error {
	f.deletedID = id
	return nil
}

func (f *fakeRepo) ListByEnterprise(ctx context.Context, enterpriseID string) ([]types.EnterpriseBankRow, error) {
	return f.rows, nil
}

type fakeCatalog struct {
	body json.RawMessage
	err  error
}

func (f fakeCatalog) List(ctx context.Context) (json.RawMessage, error) { return f.body, f.err }

type fakeAudit struct{ entries []string }

func (f *fakeAudit) Record(ctx context.Context, actorID *string, enterpriseID, description string) {
	f.entries = append(f.entries, description)
}

func TestService_Link(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("cadastra o banco no catálogo quando o código é novo", func(t *testing.T) {
		repo, audit := newFakeRepo(), &fakeAudit{}
		svc := appbank.NewService(repo, fakeCatalog{}, audit)

		eb, err := svc.Link(ctx, "user-1", "ent-1", "341", "Itaú")
		require.NoError(t, err)
		require.NotNil(t, repo.createdBank)
		require.Equal(t, "341", repo.createdBank.Code)
		require.Equal(t, repo.createdBank.ID, eb.BankID)
		require.Equal(t, "ent-1", eb.EnterpriseID)
		require.Len(t, audit.entries, 1)
		require.Contains(t, audit.entries[0], "Criação da associação banco-empresa")
	})

	t.Run("reaproveita o banco já existente no catálogo", func(t *testing.T) {
		repo, audit := newFakeRepo(), &fakeAudit{}
		repo.byCode["341"] = &domainbank.Bank{ID: "bank-1", Code: "341", Name: "Itaú"}
		svc := appbank.NewService(repo, fakeCatalog{}, audit)

		eb, err := svc.Link(ctx, "user-1", "ent-1", "341", "Itaú Unibanco")
		require.NoError(t, err)
		require.Nil(t, repo.createdBank, "o catálogo é global; não deve duplicar o banco")
		require.Equal(t, "bank-1", eb.BankID)
	})

	t.Run("recusa vínculo duplicado na mesma empresa", func(t *testing.T) {
		repo, audit := newFakeRepo(), &fakeAudit{}
		repo.byCode["341"] = &domainbank.Bank{ID: "bank-1", Code: "341", Name: "Itaú"}
		repo.links["bank-1/ent-1"] = &domainbank.EnterpriseBank{ID: "eb-1"}
		svc := appbank.NewService(repo, fakeCatalog{}, audit)

		_, err := svc.Link(ctx, "user-1", "ent-1", "341", "Itaú")
		require.Error(t, err)
		require.Equal(t, apperr.CodeConflict, apperr.CodeOf(err))
		require.Nil(t, repo.createdLink)
		require.Empty(t, audit.entries)
	})

	t.Run("exige código e nome", func(t *testing.T) {
		svc := appbank.NewService(newFakeRepo(), fakeCatalog{}, &fakeAudit{})

		_, err := svc.Link(ctx, "user-1", "ent-1", "  ", "Itaú")
		require.Equal(t, apperr.CodeInvalid, apperr.CodeOf(err))

		_, err = svc.Link(ctx, "user-1", "ent-1", "341", " ")
		require.Equal(t, apperr.CodeInvalid, apperr.CodeOf(err))
	})
}

func TestService_Unlink(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("remove o vínculo da própria empresa", func(t *testing.T) {
		repo, audit := newFakeRepo(), &fakeAudit{}
		repo.byID["eb-1"] = &types.EnterpriseBankDetail{ID: "eb-1", EnterpriseID: "ent-1", EnterpriseName: "Acme"}
		svc := appbank.NewService(repo, fakeCatalog{}, audit)

		require.NoError(t, svc.Unlink(ctx, "user-1", "eb-1", "ent-1"))
		require.Equal(t, "eb-1", repo.deletedID)
		require.Contains(t, audit.entries[0], "Acme")
	})

	t.Run("404 quando o vínculo não existe", func(t *testing.T) {
		svc := appbank.NewService(newFakeRepo(), fakeCatalog{}, &fakeAudit{})

		err := svc.Unlink(ctx, "user-1", "sumiu", "ent-1")
		require.Equal(t, apperr.CodeNotFound, apperr.CodeOf(err))
	})

	t.Run("403 quando o vínculo é de outra empresa", func(t *testing.T) {
		repo := newFakeRepo()
		repo.byID["eb-1"] = &types.EnterpriseBankDetail{ID: "eb-1", EnterpriseID: "ent-2"}
		svc := appbank.NewService(repo, fakeCatalog{}, &fakeAudit{})

		err := svc.Unlink(ctx, "user-1", "eb-1", "ent-1")
		require.Equal(t, apperr.CodeForbidden, apperr.CodeOf(err))
		require.Empty(t, repo.deletedID)
	})
}

func TestService_ListCatalog(t *testing.T) {
	t.Parallel()

	t.Run("repassa a resposta externa sem transformar", func(t *testing.T) {
		body := json.RawMessage(`{"banks":[{"code":"341"}]}`)
		svc := appbank.NewService(newFakeRepo(), fakeCatalog{body: body}, &fakeAudit{})

		got, err := svc.ListCatalog(context.Background())
		require.NoError(t, err)
		require.JSONEq(t, string(body), string(got))
	})

	t.Run("propaga a falha da fonte externa", func(t *testing.T) {
		svc := appbank.NewService(newFakeRepo(), fakeCatalog{err: errors.New("timeout")}, &fakeAudit{})

		_, err := svc.ListCatalog(context.Background())
		require.Error(t, err)
	})
}

func TestService_ListByEnterprise(t *testing.T) {
	t.Parallel()

	repo := newFakeRepo()
	repo.rows = []types.EnterpriseBankRow{{ID: "eb-1", BankID: "bank-1", BankCode: "341", BankName: "Itaú"}}
	svc := appbank.NewService(repo, fakeCatalog{}, &fakeAudit{})

	rows, err := svc.ListByEnterprise(context.Background(), "ent-1")
	require.NoError(t, err)
	require.Len(t, rows, 1)
	require.Equal(t, "341", rows[0].BankCode)
}
