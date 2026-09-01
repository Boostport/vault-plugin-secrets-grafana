package vault_plugin_secrets_grafana

import (
	"context"
	"encoding/json"
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
			serviceAccountID, err := serviceAccountIDFromInternalData(req.Secret.InternalData["service_account_id"])
			if err != nil {
				return nil, err
			}

			err = client.DeleteGrafanaServiceAccountFromCloud(stack, serviceAccountID)

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
		// InternalData round-trips through JSON when the lease is persisted, so
		// the id arrives as float64 or json.Number, not the int64 it was stored
		// as. A bare int64 assertion panics and kills the plugin process.
		serviceAccountID, err := serviceAccountIDFromInternalData(req.Secret.InternalData["service_account_id"])
		if err != nil {
			return nil, err
		}
		err = client.DeleteServiceAccount(serviceAccountID)

		if err != nil {
			return nil, fmt.Errorf("error deleting grafana service account: %w", err)
		}
	}

	return nil, nil
}

func serviceAccountIDFromInternalData(val any) (int64, error) {
	switch v := val.(type) {
	case int64:
		return v, nil
	case float64:
		return int64(v), nil
	case json.Number:
		return v.Int64()
	default:
		return 0, fmt.Errorf("secret internal data has unexpected service_account_id type %T", val)
	}
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
