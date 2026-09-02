package types

// PermissionWithEnterprise representa uma permissão junto do enterpriseId do
// seu papel, usado para checar se a permission pertence à empresa do ator.
type PermissionWithEnterprise struct {
	ID           string
	FeatureID    string
	RoleID       string
	Create       bool
	Read         bool
	Update       bool
	Delete       bool
	EnterpriseID string
}
