// Package scalar contém tipos de decodificação para campos que o front envia
// ora como string, ora como número — reflexo de o user-crud aplicar
// String(value) a eles antes de gravar.
package scalar

import (
	"encoding/json"
	"strconv"
)

// Text aceita string, número ou booleano no JSON e guarda a forma textual.
//
// Use sempre como ponteiro (*Text): campo ausente ou null deixa o ponteiro
// nil, o que permite distinguir "não informado" de "informado como vazio".
type Text struct{ Value string }

func (t *Text) UnmarshalJSON(b []byte) error {
	if len(b) > 0 && b[0] == '"' {
		var s string
		if err := json.Unmarshal(b, &s); err != nil {
			return err
		}
		t.Value = s
		return nil
	}
	// Números passam pelo formato mais curto que os representa, como faz o
	// String() do JavaScript (1500.0 -> "1500"); demais literais vão crus.
	if f, err := strconv.ParseFloat(string(b), 64); err == nil {
		t.Value = strconv.FormatFloat(f, 'f', -1, 64)
		return nil
	}
	t.Value = string(b)
	return nil
}

func (t *Text) MarshalJSON() ([]byte, error) {
	if t == nil {
		return []byte("null"), nil
	}
	return json.Marshal(t.Value)
}

// StringPtr devolve nil quando o campo não veio no corpo.
func (t *Text) StringPtr() *string {
	if t == nil {
		return nil
	}
	v := t.Value
	return &v
}
