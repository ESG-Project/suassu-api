package scalar_test

import (
	"encoding/json"
	"testing"

	"github.com/ESG-Project/suassu-api/internal/http/dto/scalar"
	"github.com/stretchr/testify/require"
)

type payload struct {
	Value *scalar.Text `json:"value"`
}

func decode(t *testing.T, body string) payload {
	t.Helper()
	var p payload
	require.NoError(t, json.Unmarshal([]byte(body), &p))
	return p
}

func TestText_UnmarshalJSON(t *testing.T) {
	t.Parallel()

	t.Run("string vem como está", func(t *testing.T) {
		require.Equal(t, "1.234,56", *decode(t, `{"value":"1.234,56"}`).Value.StringPtr())
	})

	t.Run("número vira texto na forma mais curta", func(t *testing.T) {
		require.Equal(t, "1500", *decode(t, `{"value":1500}`).Value.StringPtr())
		require.Equal(t, "1500", *decode(t, `{"value":1500.0}`).Value.StringPtr())
		require.Equal(t, "12.5", *decode(t, `{"value":12.5}`).Value.StringPtr())
	})

	t.Run("campo ausente e null ficam nil", func(t *testing.T) {
		require.Nil(t, decode(t, `{}`).Value.StringPtr())
		require.Nil(t, decode(t, `{"value":null}`).Value.StringPtr())
	})

	t.Run("string vazia é informada, não ausente", func(t *testing.T) {
		got := decode(t, `{"value":""}`).Value.StringPtr()
		require.NotNil(t, got)
		require.Equal(t, "", *got)
	})
}
