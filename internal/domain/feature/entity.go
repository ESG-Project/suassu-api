package feature

import (
	"errors"
	"strings"
)

type Feature struct {
	ID   string
	Name string
}

func NewFeature(id, name string) *Feature {
	return &Feature{
		ID:   id,
		Name: name,
	}
}

func (f *Feature) Validate() error {
	if strings.TrimSpace(f.ID) == "" {
		return errors.New("id is required")
	}
	if strings.TrimSpace(f.Name) == "" {
		return errors.New("name is required")
	}
	return nil
}
