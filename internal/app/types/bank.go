package types

// EnterpriseBankRow é um vínculo empresa-banco já resolvido com os dados do
// banco do catálogo global.
type EnterpriseBankRow struct {
	ID       string
	BankID   string
	BankCode string
	BankName string
}

// EnterpriseBankDetail é um vínculo acrescido do nome da empresa dona —
// necessário porque a mensagem de auditoria da exclusão cita o nome da
// empresa, não o id.
type EnterpriseBankDetail struct {
	ID             string
	EnterpriseID   string
	BankID         string
	EnterpriseName string
}
