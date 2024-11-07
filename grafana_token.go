package vault_plugin_secrets_grafana

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

const (
	grafanaTokenType = "grafana_token"
)

type grafanaToken struct {
	IsCloud          bool   `json:"is_cloud"`
	Token            string `json:"token"`
	Stack            string `json:"stack"`              // For Grafana Cloud service accounts
	Region           string `json:"region"`             // For Grafana Cloud access policies
	AccessPolicyID   string `json:"access_policy_id"`   // For Grafana Cloud access policies
	ServiceAccountID int64  `json:"service_account_id"` // For Grafana Cloud and Grafana service accounts
}

func (b *grafanaBackend) grafanaToken() *framework.Secret {
	return &framework.Secret{
		Type: grafanaTokenType,
		Fields: map[string]*framework.FieldSchema{
			"token": {
				Type:        framework.TypeString,
				Description: "Grafana Token",
			},
		},
		Revoke: b.tokenRevoke,
		Renew:  b.tokenRenew,
	}
}

func (b *grafanaBackend) tokenRevoke(ctx context.Context, req *logical.Request, _ *framework.FieldData) (*logical.Response, error) {
	client, err := b.getClient(ctx, req.Storage)
	if err != nil {
		return nil, fmt.Errorf("error getting client: %w", err)
	}

	isCloud := false

	if val, ok := req.Secret.InternalData["is_cloud"]; ok {
		isCloud = val.(bool)
	}

	if isCloud {
		stack := ""

		if val, ok := req.Secret.InternalData["stack"]; ok {
			stack = val.(string)
		}

		if stack != "" {
			serviceAccountID := int64(req.Secret.InternalData["service_account_id"].(float64))

			err := client.DeleteGrafanaServiceAccountFromCloud(stack, serviceAccountID)

			if err != nil {
				return nil, fmt.Errorf("error deleting grafana cloud service account: %w", err)
			}
		} else {
			accessPolicyID := req.Secret.InternalData["access_policy_id"].(string)
			region := req.Secret.InternalData["region"].(string)
			err := client.DeleteCloudAccessPolicy(region, accessPolicyID)

			if err != nil {
				return nil, fmt.Errorf("error deleting grafana cloud access policy: %w", err)
			}
		}

	} else {
		serviceAccountID := req.Secret.InternalData["service_account_id"].(int64)
		err := client.DeleteServiceAccount(serviceAccountID)

		if err != nil {
			return nil, fmt.Errorf("error deleting grafana service account: %w", err)
		}
	}

	return nil, nil
}

func (b *grafanaBackend) tokenRenew(ctx context.Context, req *logical.Request, _ *framework.FieldData) (*logical.Response, error) {
	roleRaw, ok := req.Secret.InternalData["vault_role"]
	if !ok {
		return nil, fmt.Errorf("secret is missing role internal data")
	}

	// get the role entry
	role := roleRaw.(string)
	roleEntry, err := b.getRole(ctx, req.Storage, role)
	if err != nil {
		return nil, fmt.Errorf("error retrieving role: %w", err)
	}

	if roleEntry == nil {
		return nil, errors.New("error retrieving role: role is nil")
	}

	resp := &logical.Response{Secret: req.Secret}

	if roleEntry.TTL > 0 {
		resp.Secret.TTL = roleEntry.TTL
	}
	if roleEntry.MaxTTL > 0 {
		resp.Secret.MaxTTL = roleEntry.MaxTTL
	}

	return resp, nil
}
