package legacyuserhttp

import _ "embed"

// clientRegisterTemplateXLSX é o mesmo arquivo servido pelo user-crud em
// GET /user/import-file/enterprise (download estático, sem geração dinâmica).
//
//go:embed assets/client_register_template.xlsx
var clientRegisterTemplateXLSX []byte
