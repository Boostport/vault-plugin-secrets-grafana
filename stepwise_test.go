package vault_plugin_secrets_grafana

// Stepwise test temporary disabled due to https://github.com/hashicorp/vault-testing-stepwise/issues/10
/*
import (
	"fmt"
	"os"
	"sync"
	"testing"

	stepwise "github.com/hashicorp/vault-testing-stepwise"
	dockerEnvironment "github.com/hashicorp/vault-testing-stepwise/environments/docker"
	"github.com/hashicorp/vault/api"
	"github.com/stretchr/testify/require"
)

var (
	stepwiseTestScopes = []string{"accesspolicies:delete", "accesspolicies:read", "accesspolicies:write", "stacks:read", "stack-service-accounts:write"}
	stepwiseTestRealms = fmt.Sprintf(`[{"type": "org", "identifier": "%s", "labelPolicies": []}]`, os.Getenv(envVarGrafanaCloudOrgIdentifier))
)

// TestAccUserToken runs a series of acceptance tests to check the
// end-to-end workflow of the backend. It creates a Vault Docker container
// and loads a temporary plugin.
func TestAccUserToken(t *testing.T) {
	t.Parallel()
	if !runAcceptanceTests {
		t.SkipNow()
	}
	envOptions := &stepwise.MountOptions{
		RegistryName:    "grafana",
		PluginType:      api.PluginTypeSecrets,
		PluginName:      "vault-plugin-secrets-grafana",
		MountPathPrefix: "grafana",
	}

	roleName := "vault-stepwise-user-role"

	cred := new(string)
	stepwise.Run(t, stepwise.Case{
		Precheck:    func() { testAccPreCheck(t) },
		Environment: dockerEnvironment.NewEnvironment("grafana", envOptions),
		Steps: []stepwise.Step{
			testAccConfig(t),
			testAccUserRole(t, roleName),
			testAccUserRoleRead(t, roleName),
			testAccUserCredRead(t, roleName, cred),
		},
	})
}

var initSetup sync.Once

func testAccPreCheck(t *testing.T) {
	initSetup.Do(func() {
		// Ensure test variables are set
		if v := os.Getenv(envVarGrafanaCloudToken); v == "" {
			t.Skip(fmt.Printf("%s not set", envVarGrafanaCloudToken))
		}
		if v := os.Getenv(envVarGrafanaCloudStackSlug); v == "" {
			t.Skip(fmt.Printf("%s not set", envVarGrafanaCloudStackSlug))
		}
		if v := os.Getenv(envVarGrafanaCloudRegion); v == "" {
			t.Skip(fmt.Printf("%s not set", envVarGrafanaCloudRegion))
		}
		if v := os.Getenv(envVarGrafanaCloudOrgIdentifier); v == "" {
			t.Skip(fmt.Printf("%s not set", envVarGrafanaCloudOrgIdentifier))
		}
	})
}

func testAccConfig(_ *testing.T) stepwise.Step {
	return stepwise.Step{
		Operation: stepwise.UpdateOperation,
		Path:      "config",
		Data: map[string]interface{}{
			"type":  GrafanaCloudType,
			"token": os.Getenv(envVarGrafanaCloudToken),
		},
	}
}

func testAccUserRole(t *testing.T, roleName string) stepwise.Step {
	return stepwise.Step{
		Operation: stepwise.UpdateOperation,
		Path:      "roles/" + roleName,
		Data: map[string]interface{}{
			"type":    roleCloudAccessPolicy,
			"region":  os.Getenv(envVarGrafanaCloudRegion),
			"scopes":  stepwiseTestScopes,
			"realms":  stepwiseTestRealms,
			"ttl":     "1m",
			"max_ttl": "5m",
		},
		Assert: func(resp *api.Secret, err error) error {
			require.Nil(t, err)
			require.Nil(t, resp)
			return nil
		},
	}
}

func testAccUserRoleRead(t *testing.T, roleName string) stepwise.Step {
	return stepwise.Step{
		Operation: stepwise.ReadOperation,
		Path:      "roles/" + roleName,
		Assert: func(resp *api.Secret, err error) error {
			require.Nil(t, err)
			require.NotNil(t, resp)
			require.Equal(t, roleCloudAccessPolicy, resp.Data["type"])
			require.Equal(t, os.Getenv(envVarGrafanaCloudRegion), resp.Data["region"])
			require.Equal(t, stepwiseTestScopes, resp.Data["scopes"])
			require.Equal(t, stepwiseTestRealms, resp.Data["realms"])
			return nil
		},
	}
}

func testAccUserCredRead(t *testing.T, roleName string, token *string) stepwise.Step {
	return stepwise.Step{
		Operation: stepwise.ReadOperation,
		Path:      "creds/" + roleName,
		Assert: func(resp *api.Secret, err error) error {
			require.Nil(t, err)
			require.NotNil(t, resp)
			require.NotEmpty(t, resp.Data["token"])
			*token = resp.Data["token"].(string)
			return nil
		},
	}
}
*/
