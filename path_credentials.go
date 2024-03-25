package vault_plugin_secrets_grafana

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Boostport/vault-plugin-secrets-grafana/client"
	"github.com/google/uuid"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

func pathCredentials(b *grafanaBackend) *framework.Path {
	return &framework.Path{
		Pattern: "creds/" + framework.GenericNameRegex("name"),
		Fields: map[string]*framework.FieldSchema{
			"name": {
				Type:        framework.TypeLowerCaseString,
				Description: "Name of the role",
				Required:    true,
			},
		},
		Operations: map[logical.Operation]framework.OperationHandler{
			logical.ReadOperation: &framework.PathOperation{
				Callback: b.pathCredentialsRead,
			},
			logical.UpdateOperation: &framework.PathOperation{
				Callback: b.pathCredentialsRead,
			},
		},
		HelpSynopsis:    pathCredentialsHelpSyn,
		HelpDescription: pathCredentialsHelpDesc,
	}
}

func (b *grafanaBackend) pathCredentialsRead(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	roleName := d.Get("name").(string)

	return b.createUserCreds(ctx, req, roleName)
}

func (b *grafanaBackend) createUserCreds(ctx context.Context, req *logical.Request, roleName string) (*logical.Response, error) {
	role, err := b.getRole(ctx, req.Storage, roleName)
	if err != nil {
		return nil, fmt.Errorf("error retrieving role: %w", err)
	}

	if role == nil {
		return nil, errors.New("error retrieving role: role is nil")
	}

	config, err := getConfig(ctx, req.Storage)
	if err != nil {
		return nil, fmt.Errorf("error reading config: %w", err)
	}

	if err := role.validate(config.Type); err != nil {
		return logical.ErrorResponse("role configuration not compatible with mount configuration: %w", err.Error()), nil
	}

	token, err := b.createToken(ctx, req.Storage, config.Type, role)
	if err != nil {
		return nil, err
	}

	// The response is divided into two objects (1) internal data and (2) data.
	// If you want to reference any information in your code, you need to
	// store it in internal data!
	resp := b.Secret(grafanaTokenType).Response(map[string]interface{}{
		"token": token.Token,
	}, map[string]interface{}{
		"is_cloud":           token.IsCloud,
		"stack":              token.Stack,
		"region":             token.Region,
		"access_policy_id":   token.AccessPolicyID,
		"service_account_id": token.ServiceAccountID,
		"vault_role":         roleName,
	})

	if role.TTL > 0 {
		resp.Secret.TTL = role.TTL
	}

	if role.MaxTTL > 0 {
		resp.Secret.MaxTTL = role.MaxTTL
	}

	return resp, nil
}

func (b *grafanaBackend) createToken(ctx context.Context, s logical.Storage, configType string, roleEntry *grafanaRoleEntry) (*grafanaToken, error) {
	c, err := b.getClient(ctx, s)
	if err != nil {
		return nil, err
	}

	credentialName := fmt.Sprintf("vault-%s", uuid.New())

	if configType == GrafanaCloudType {
		if roleEntry.Type == roleCloudAccessPolicy {
			return createCloudAccessPolicyToken(c, credentialName, roleEntry)
		} else if roleEntry.Type == roleGrafanaServiceAccount {
			return createCloudServiceAccountToken(c, credentialName, roleEntry)
		}
	} else if configType == GrafanaType {
		return createServiceAccountToken(c, credentialName, roleEntry)
	}

	return nil, errors.New("cannot create token due to inconsistent mount configuration and role configuration")
}

func createCloudAccessPolicyToken(c *client.Grafana, credentialName string, roleEntry *grafanaRoleEntry) (*grafanaToken, error) {

	cloudAccessPolicyInput := client.CreateCloudAccessPolicyInput{
		Name:        credentialName,
		DisplayName: credentialName,
		Scopes:      roleEntry.Scopes,
	}

	if roleEntry.Realms != "" {
		realms, err := realmsStringToStruct(roleEntry.Realms)

		if err != nil {
			return nil, fmt.Errorf("error converting realms string to struct: %w", err)
		}

		cloudAccessPolicyInput.Realms = realms
	}

	if len(roleEntry.AllowedSubnets) > 0 {
		cloudAccessPolicyInput.Conditions.AllowedSubnets = roleEntry.AllowedSubnets
	}

	cloudAccessPolicy, err := c.CreateCloudAccessPolicy(roleEntry.Region, cloudAccessPolicyInput)

	if err != nil {
		return nil, fmt.Errorf("error creating cloud access policy: %w", err)
	}

	token, err := c.CreateCloudAccessPolicyToken(roleEntry.Region, client.CreateCloudAccessPolicyTokenInput{
		AccessPolicyID: cloudAccessPolicy.ID,
		Name:           credentialName,
		DisplayName:    credentialName,
	})

	if err != nil {
		err := c.DeleteCloudAccessPolicy(roleEntry.Region, cloudAccessPolicy.ID)

		if err != nil {
			return nil, fmt.Errorf("error deleting cloud access policy after error creating token: %w", err)
		}

		return nil, fmt.Errorf("error creating cloud access policy token: %w", err)
	}

	return &grafanaToken{
		IsCloud:        true,
		Token:          token.Token,
		Region:         roleEntry.Region,
		AccessPolicyID: cloudAccessPolicy.ID,
	}, nil
}

func createCloudServiceAccountToken(c *client.Grafana, credentialName string, roleEntry *grafanaRoleEntry) (*grafanaToken, error) {
	serviceAccount, err := c.CreateGrafanaServiceAccountFromCloud(roleEntry.Stack, client.CreateServiceAccountInput{
		Name: credentialName,
		Role: roleEntry.Role,
	})

	if err != nil {
		return nil, fmt.Errorf("error creating service account: %w", err)
	}

	token, err := c.CreateGrafanaServiceAccountTokenFromCloud(roleEntry.Stack, client.CreateServiceAccountTokenInput{
		Name:             credentialName,
		ServiceAccountID: serviceAccount.ID,
	})

	if err != nil {
		err := c.DeleteGrafanaServiceAccountFromCloud(roleEntry.Stack, serviceAccount.ID)

		if err != nil {
			return nil, fmt.Errorf("error deleting service account after error creating token: %w", err)
		}

		return nil, fmt.Errorf("error creating service account token: %w", err)
	}

	return &grafanaToken{
		IsCloud:          true,
		Token:            token.Key,
		Stack:            roleEntry.Stack,
		ServiceAccountID: serviceAccount.ID,
	}, nil
}

func createServiceAccountToken(c *client.Grafana, credentialName string, roleEntry *grafanaRoleEntry) (*grafanaToken, error) {
	serviceAccount, err := c.CreateServiceAccount(client.CreateServiceAccountInput{
		Name: credentialName,
		Role: roleEntry.Role,
	})

	if err != nil {
		return nil, fmt.Errorf("error creating service account: %w", err)
	}

	if len(roleEntry.RBACRoles) > 0 {
		roleUIDs, err := customRBACRoleNamesToIDs(c, roleEntry.RBACRoles)

		if err != nil {
			err := deleteServiceAccount(c, serviceAccount.ID)

			if err != nil {
				return nil, fmt.Errorf("error deleting service account after error converting role names to IDs: %w", err)
			}
			return nil, fmt.Errorf("error converting role names to IDs: %w", err)
		}

		err = c.SetServiceAccountRoleAssignments(client.ServiceAccountRoleAssignmentsInput{
			ServiceAccountID: serviceAccount.ID,
			RoleUIDs:         roleUIDs,
		})

		if err != nil {
			err := deleteServiceAccount(c, serviceAccount.ID)

			if err != nil {
				return nil, fmt.Errorf("error deleting service account after error setting role assignments: %w", err)

			}

			return nil, fmt.Errorf("error setting service account role assignments: %w", err)
		}
	}

	token, err := c.CreateServiceAccountToken(client.CreateServiceAccountTokenInput{
		Name:             credentialName,
		ServiceAccountID: serviceAccount.ID,
	})

	if err != nil {
		err := deleteServiceAccount(c, serviceAccount.ID)

		if err != nil {
			return nil, fmt.Errorf("error deleting service account after error creating token: %w", err)
		}
		return nil, fmt.Errorf("error creating service account token: %w", err)
	}

	return &grafanaToken{
		IsCloud:          false,
		Token:            token.Key,
		ServiceAccountID: serviceAccount.ID,
	}, nil
}

func deleteServiceAccount(c *client.Grafana, serviceAccountID int64) error {
	err := c.DeleteServiceAccount(serviceAccountID)

	if err != nil {
		return fmt.Errorf("error deleting service account: %w", err)
	}

	return nil
}

func customRBACRoleNamesToIDs(c *client.Grafana, roleNames []string) ([]string, error) {
	var roleIDs []string

	allRoles, err := c.GetAllRoles()

	if err != nil {
		return nil, fmt.Errorf("error getting all roles: %w", err)
	}

	for _, roleName := range roleNames {
		found := false
		for _, role := range allRoles {
			if role.Name == roleName && role.UID != "" {
				roleIDs = append(roleIDs, role.UID)
				found = true
				break
			}
		}

		if !found {
			return nil, fmt.Errorf("rbac role does not exist: %s", roleName)
		}
	}

	return roleIDs, nil
}

func realmsStringToStruct(realms string) ([]client.CloudAccessPolicyRealm, error) {
	var result []client.CloudAccessPolicyRealm

	if err := json.Unmarshal([]byte(realms), &result); err != nil {
		return result, fmt.Errorf("unable to unmarshall realms string: %w", err)
	}

	return result, nil
}

const pathCredentialsHelpSyn = `
Generate a Grafana Cloud or Grafana token from a specific Vault role.
`

const pathCredentialsHelpDesc = `
This path generates a Grafana Cloud or Grafana tokens
based on a particular role.
`
