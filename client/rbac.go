package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type ServiceAccountRoleAssignmentsInput struct {
	ServiceAccountID int64    `json:"-"`
	Global           bool     `json:"global"`
	RoleUIDs         []string `json:"roleUids"`
	IncludeHidden    bool     `json:"includeHidden"`
}

type Role struct {
	Version     int64        `json:"version"`
	UID         string       `json:"uid,omitempty"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Global      bool         `json:"global"`
	Group       string       `json:"group"`
	DisplayName string       `json:"displayName"`
	Hidden      bool         `json:"hidden"`
	Permissions []Permission `json:"permissions,omitempty"`
}

type RoleInput Role

type Permission struct {
	Action string `json:"action"`
	Scope  string `json:"scope"`
}

func (g *Grafana) GetAllRoles() ([]Role, error) {
	var result []Role

	err := g.do(http.MethodGet, "/api/access-control/roles", nil, nil, &result)
	if err != nil {
		return nil, fmt.Errorf("error getting roles: %w", err)
	}

	return result, nil
}

func (g *Grafana) SetServiceAccountRoleAssignments(input ServiceAccountRoleAssignmentsInput) error {

	data, err := json.Marshal(input)
	if err != nil {
		return fmt.Errorf("error marshalling input: %w", err)
	}

	err = g.do(http.MethodPut, fmt.Sprintf("/api/access-control/users/%d/roles", input.ServiceAccountID), nil, data, nil)

	if err != nil {
		return fmt.Errorf("error setting service account role assignments: %w", err)
	}

	return nil
}

func (g *Grafana) CreateCustomRole(input RoleInput) (Role, error) {

	result := Role{}

	data, err := json.Marshal(input)
	if err != nil {
		return result, fmt.Errorf("error marshalling input: %w", err)
	}

	err = g.do(http.MethodPost, "/api/access-control/roles", nil, data, &result)

	if err != nil {
		return result, fmt.Errorf("error creating custom role: %w", err)
	}

	return result, nil
}

func (g *Grafana) DeleteCustomRole(roleUID string) error {
	err := g.do(http.MethodDelete, fmt.Sprintf("/api/access-control/roles/%s", roleUID), url.Values{
		"force": []string{"true"},
	}, nil, nil)

	if err != nil {
		return fmt.Errorf("error deleting custom role: %w", err)
	}

	return nil
}
