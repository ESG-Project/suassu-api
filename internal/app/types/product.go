package types

// ProductParameterRef é o parâmetro aninhado em cada produto de
// GET /product/enterprise.
type ProductParameterRef struct {
	ID    string
	Title string
	Value *string
}

// ProductTypeRef é o tipo de produto aninhado em cada produto de
// GET /product/enterprise.
type ProductTypeRef struct {
	ID   string
	Type string
}

// ProductDetailRow é um produto com Parameter e TypeProduct já resolvidos.
// Parameter e Type são nil quando a FK correspondente é nula.
type ProductDetailRow struct {
	ID             string
	Name           string
	SuggestedValue *string
	EnterpriseID   string
	Deliverable    bool
	Parameter      *ProductParameterRef
	Type           *ProductTypeRef
}
