package postgres

import (
	"context"
	"database/sql"

	sqlc "github.com/ESG-Project/suassu-api/internal/infra/db/sqlc/gen"
)

// Adaptador para permitir *sql.DB e *sql.Tx via sqlc.DBTX
type dbtx = sqlc.DBTX

// Interface para permitir mocking nos testes
type TxManagerInterface interface {
	RunInTx(ctx context.Context, fn func(r Repos) error) error
}

type TxManager struct{ DB *sql.DB }

// Garantir que TxManager implementa TxManagerInterface
var _ TxManagerInterface = (*TxManager)(nil)

type Repos struct {
	Users         func() *UserRepo
	Addresses     func() *AddressRepo
	Enterprises   func() *EnterpriseRepo
	Roles         func() *RoleRepo
	Permissions   func() *PermissionRepo
	Products      func() *ProductRepo
	Parameters    func() *ParameterRepo
	Features      func() *FeatureRepo
	PhytoAnalyses func() *PhytoAnalysisRepo
	Specimens     func() *SpecimenRepo
	Species       func() *SpeciesRepo
	Clients       func() *ClientRepo
	Technicians   func() *TechnicianRepo
	Logs          func() *LogRepo
}

func (m *TxManager) RunInTx(ctx context.Context, fn func(r Repos) error) error {
	tx, err := m.DB.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}

	r := Repos{
		Users:         func() *UserRepo { return NewUserRepoFrom(tx) },
		Addresses:     func() *AddressRepo { return NewAddressRepoFrom(tx) },
		Enterprises:   func() *EnterpriseRepo { return NewEnterpriseRepoFrom(tx) },
		Roles:         func() *RoleRepo { return NewRoleRepoFrom(tx) },
		Permissions:   func() *PermissionRepo { return NewPermissionRepoFrom(tx) },
		Products:      func() *ProductRepo { return NewProductRepoFrom(tx) },
		Parameters:    func() *ParameterRepo { return NewParameterRepoFrom(tx) },
		Features:      func() *FeatureRepo { return NewFeatureRepoFrom(tx) },
		PhytoAnalyses: func() *PhytoAnalysisRepo { return NewPhytoAnalysisRepoFrom(tx) },
		Specimens:     func() *SpecimenRepo { return NewSpecimenRepoFrom(tx) },
		Species:       func() *SpeciesRepo { return NewSpeciesRepoFrom(tx) },
		Clients:       func() *ClientRepo { return NewClientRepoFrom(tx) },
		Technicians:   func() *TechnicianRepo { return NewTechnicianRepoFrom(tx) },
		Logs:          func() *LogRepo { return NewLogRepoFrom(tx) },
	}

	if err := fn(r); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}
