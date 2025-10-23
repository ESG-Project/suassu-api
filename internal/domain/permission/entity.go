package permission

import (
	"errors"
	"strings"
)

type Permission struct {
	ID        string
	FeatureID string
	RoleID    string
	Create    bool
	Read      bool
	Update    bool
	Delete    bool
}

func NewPermission(id, featureID, roleID string) *Permission {
	return &Permission{
		ID:        id,
		FeatureID: featureID,
		RoleID:    roleID,
		Create:    false,
		Read:      false,
		Update:    false,
		Delete:    false,
	}
}

func (p *Permission) Validate() error {
	if strings.TrimSpace(p.ID) == "" {
		return errors.New("id is required")
	}
	if strings.TrimSpace(p.FeatureID) == "" {
		return errors.New("featureId is required")
	}
	if strings.TrimSpace(p.RoleID) == "" {
		return errors.New("roleId is required")
	}
	return nil
}

func (p *Permission) SetPermissions(create, read, update, delete bool) {
	p.Create = create
	p.Read = read
	p.Update = update
	p.Delete = delete
}
