package role

import (
	"errors"
	"strings"
)

type Role struct {
	ID           string
	Title        string
	EnterpriseID string
}

func NewRole(id, title, enterpriseID string) *Role {
	return &Role{
		ID:           id,
		Title:        title,
		EnterpriseID: enterpriseID,
	}
}

func (r *Role) Validate() error {
	if strings.TrimSpace(r.ID) == "" {
		return errors.New("id is required")
	}
	if strings.TrimSpace(r.Title) == "" {
		return errors.New("title is required")
	}
	if strings.TrimSpace(r.EnterpriseID) == "" {
		return errors.New("enterpriseId is required")
	}
	return nil
}
