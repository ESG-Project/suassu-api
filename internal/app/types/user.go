package types

// UserWithDetails representa um usuário com todas as informações para a camada de aplicação
type UserWithDetails struct {
	ID           string
	Name         string
	Email        string
	Document     string
	Phone        *string
	EnterpriseID string
	Address      *UserAddress
	Role         *UserRole
	Enterprise   *UserEnterprise
	Permissions  []*UserPermission
}

// UserAddress representa o endereço do usuário na camada de aplicação
type UserAddress struct {
	ID           string
	ZipCode      string
	State        string
	City         string
	Neighborhood string
	Street       string
	Num          string
	Latitude     *string
	Longitude    *string
	AddInfo      *string
}

// UserRole representa o papel do usuário na camada de aplicação
type UserRole struct {
	ID    string
	Title string
}

// UserEnterprise representa a empresa do usuário na camada de aplicação
type UserEnterprise struct {
	ID          string
	Name        string
	CNPJ        string
	Email       string
	FantasyName *string
	Phone       *string
	AddressID   *string
	Address     *UserAddress
}

// UserPermission representa uma permissão do usuário na camada de aplicação
type UserPermission struct {
	ID          string
	FeatureID   string
	FeatureName string
	Create      bool
	Read        bool
	Update      bool
	Delete      bool
}

// UserPermissions representa as permissões do usuário na camada de aplicação
type UserPermissions struct {
	ID          string
	Name        string
	RoleTitle   string
	Permissions []*UserPermission
}
