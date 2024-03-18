package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type CloudAccessPolicyLabelPolicy struct {
	Selector string `json:"selector"`
}

type CloudAccessPolicyRealm struct {
	Type          string                         `json:"type"`
	Identifier    string                         `json:"identifier"`
	LabelPolicies []CloudAccessPolicyLabelPolicy `json:"labelPolicies"`
}

type CloudAccessPolicyConditions struct {
	AllowedSubnets []string `json:"allowedSubnets,omitempty"`
}

type CreateCloudAccessPolicyInput struct {
	Name        string                       `json:"name"`
	DisplayName string                       `json:"displayName"`
	Scopes      []string                     `json:"scopes"`
	Realms      []CloudAccessPolicyRealm     `json:"realms"`
	Conditions  *CloudAccessPolicyConditions `json:"conditions,omitempty"`
}

type CreateCloudAccessPolicyTokenInput struct {
	AccessPolicyID string     `json:"accessPolicyId"`
	Name           string     `json:"name"`
	DisplayName    string     `json:"displayName,omitempty"`
	ExpiresAt      *time.Time `json:"expiresAt,omitempty"`
}

type CloudAccessPolicy struct {
	Name        string                   `json:"name"`
	DisplayName string                   `json:"displayName"`
	Scopes      []string                 `json:"scopes"`
	Realms      []CloudAccessPolicyRealm `json:"realms"`

	// The following fields are not part of the input, but are returned by the API.
	ID        string    `json:"id"`
	OrgID     string    `json:"orgId"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type CloudAccessPolicyToken struct {
	ID             string     `json:"id"`
	AccessPolicyID string     `json:"accessPolicyId"`
	Name           string     `json:"name"`
	DisplayName    string     `json:"displayName"`
	ExpiresAt      *time.Time `json:"expiresAt"`
	FirstUsedAt    time.Time  `json:"firstUsedAt"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      *time.Time `json:"updatedAt"`

	Token string `json:"token,omitempty"` // Only returned when creating a token.
}

func (g *Grafana) CreateCloudAccessPolicy(region string, input CreateCloudAccessPolicyInput) (CloudAccessPolicy, error) {

	result := CloudAccessPolicy{}

	data, err := json.Marshal(input)
	if err != nil {
		return result, fmt.Errorf("error marshalling input: %w", err)
	}

	err = g.do(http.MethodPost, "/api/v1/accesspolicies", url.Values{
		"region": []string{region},
	}, data, &result)

	if err != nil {
		return result, fmt.Errorf("error creating cloud access policy: %w", err)
	}

	return result, nil
}

func (g *Grafana) DeleteCloudAccessPolicy(region, cloudAccessPolicyID string) error {
	err := g.do(http.MethodDelete, fmt.Sprintf("/api/v1/accesspolicies/%s", cloudAccessPolicyID), url.Values{
		"region": []string{region},
	}, nil, nil)

	if err != nil {
		return fmt.Errorf("error deleting cloud access policy: %w", err)
	}

	return nil
}

func (g *Grafana) CreateCloudAccessPolicyToken(region string, input CreateCloudAccessPolicyTokenInput) (CloudAccessPolicyToken, error) {

	result := CloudAccessPolicyToken{}

	data, err := json.Marshal(input)
	if err != nil {
		return result, fmt.Errorf("error marshalling input: %w", err)
	}

	err = g.do(http.MethodPost, "/api/v1/tokens", url.Values{
		"region": []string{region},
	}, data, &result)

	if err != nil {
		return result, fmt.Errorf("error creating cloud access policy token: %w", err)
	}

	return result, nil
}
