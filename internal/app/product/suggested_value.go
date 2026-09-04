package product

import (
	"strconv"
	"strings"

	"github.com/ESG-Project/suassu-api/internal/apperr"
)

// NormalizeSuggestedValue replica normalizeSuggestedValue do user-crud: o
// valor é validado como número (aceitando o formato pt-BR "1.234,56"), mas o
// que se grava é a string original — é assim que os preços já estão no banco,
// e a proposta comercial depende desse formato ao renderizar.
//
// raw nil significa campo ausente; devolve nil para gravar NULL.
func NormalizeSuggestedValue(raw *string) (*string, error) {
	if raw == nil {
		return nil, nil
	}

	trimmed := strings.TrimSpace(*raw)
	if trimmed == "" || trimmed == "undefined" {
		return nil, nil
	}

	parseable := trimmed
	if strings.Contains(parseable, ",") {
		parseable = strings.ReplaceAll(parseable, ".", "")
		parseable = strings.ReplaceAll(parseable, ",", ".")
	}

	n, err := strconv.ParseFloat(parseable, 64)
	if err != nil {
		return nil, apperr.New(apperr.CodeInvalid, "Preço sugerido inválido.")
	}
	if n < 0 {
		return nil, apperr.New(apperr.CodeInvalid, "O preço sugerido não pode ser negativo.")
	}

	return raw, nil
}
