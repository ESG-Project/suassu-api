package client

import (
	"errors"
	"strings"
)

// Client é o sub-registro de um User cujo papel exige tratamento como
// cliente (ex.: título de papel "Cliente"). Espelha a tabela legada
// "Client" (1:1 com "User").
type Client struct {
	ID          string
	FantasyName *string
	UserID      string
}

func NewClient(id, userID string) *Client {
	return &Client{ID: id, UserID: userID}
}

func (c *Client) SetFantasyName(fantasyName *string) {
	c.FantasyName = fantasyName
}

func (c *Client) Validate() error {
	if strings.TrimSpace(c.ID) == "" {
		return errors.New("id is required")
	}
	if strings.TrimSpace(c.UserID) == "" {
		return errors.New("userId is required")
	}
	return nil
}
