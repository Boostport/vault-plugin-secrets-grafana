package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type CreateServiceAccountInput struct {
	Name       string `json:"name"`
	Role       string `json:"role"`
	IsDisabled *bool  `json:"isDisabled,omitempty"`
}

type CreateServiceAccountTokenInput struct {
	Name             string `json:"name"`
	ServiceAccountID int64  `json:"-"`
	SecondsToLive    int64  `json:"secondsToLive,omitempty"`
}

type ServiceAccount struct {
	ID         int64      `json:"id"`
	Name       string     `json:"name"`
	Login      string     `json:"login"`
	OrgID      int64      `json:"orgId"`
	IsDisabled bool       `json:"isDisabled"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  *time.Time `json:"updatedAt"`
	AvatarURL  string     `json:"avatarUrl"`
	Role       string     `json:"role"`
	Teams      []string   `json:"teams"`
}

type ServiceAccountToken struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Key  string `json:"key"`
}

func (g *Grafana) CreateServiceAccount(input CreateServiceAccountInput) (ServiceAccount, error) {
	result := ServiceAccount{}

	data, err := json.Marshal(input)
	if err != nil {
		return result, fmt.Errorf("error marshalling input: %w", err)
	}

	err = g.do(http.MethodPost, "/api/serviceaccounts", nil, data, &result)

	if err != nil {
		return result, fmt.Errorf("error creating service account: %w", err)
	}

	return result, nil
}

func (g *Grafana) DeleteServiceAccount(serviceAccountID int64) error {
	err := g.do(http.MethodDelete, fmt.Sprintf("/api/serviceaccounts/%d", serviceAccountID), nil, nil, nil)

	if err != nil {
		return fmt.Errorf("error deleting service account: %w", err)
	}

	return nil
}

func (g *Grafana) CreateServiceAccountToken(input CreateServiceAccountTokenInput) (ServiceAccountToken, error) {

	result := ServiceAccountToken{}

	data, err := json.Marshal(input)
	if err != nil {
		return result, fmt.Errorf("error marshalling input: %w", err)
	}

	err = g.do(http.MethodPost, fmt.Sprintf("/api/serviceaccounts/%d/tokens", input.ServiceAccountID), nil, data, &result)

	if err != nil {
		return result, fmt.Errorf("error creating service account token: %w", err)
	}

	return result, nil
}
