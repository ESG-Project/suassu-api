package technician

import (
	"errors"
	"strings"
)

// Technician é o sub-registro de um User cujo papel exige tratamento como
// técnico (ex.: título de papel "Técnico"). Espelha a tabela legada
// "Technician" (1:1 com "User").
type Technician struct {
	ID          string
	ProRegister *string
	Graduation  *string
	CTF         *string
	UserID      string
}

func NewTechnician(id, userID string) *Technician {
	return &Technician{ID: id, UserID: userID}
}

func (t *Technician) SetProRegister(v *string) { t.ProRegister = v }
func (t *Technician) SetGraduation(v *string)  { t.Graduation = v }
func (t *Technician) SetCTF(v *string)         { t.CTF = v }

func (t *Technician) Validate() error {
	if strings.TrimSpace(t.ID) == "" {
		return errors.New("id is required")
	}
	if strings.TrimSpace(t.UserID) == "" {
		return errors.New("userId is required")
	}
	return nil
}
