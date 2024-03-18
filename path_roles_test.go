package vault_plugin_secrets_grafana

import (
	"context"
	"strconv"
	"testing"

	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	cloudAccessPolicyRoleName = "CloudAccessPolicyRole"
	cloudAccessPolicyRegion   = "us-east-2"
	cloudAccessPolicyRealms   = `[{"type": "org", "identifier": "123456", "labelPolicies": []}]`
	serviceAccountRoleName    = "ServiceAccountRole"

	serviceAccountStack = "test"
	serviceAccountRole  = "Admin"
	testTTL             = int64(120)
	testMaxTTL          = int64(3600)
)

var (
	cloudAccessPolicyScopes = []string{"logs:read"}
)

func TestCloudAccessPolicyRole(t *testing.T) {
	b, s := getTestBackend(t)

	err := testConfigCreate(b, s, map[string]interface{}{
		"type":  GrafanaCloudType,
		"token": "abcd",
	})
	assert.NoError(t, err)

	t.Run("List All Roles (Cloud Access Policy)", func(t *testing.T) {
		for i := 1; i <= 10; i++ {
			_, err := testTokenRoleCreate(t, b, s,
				cloudAccessPolicyRoleName+strconv.Itoa(i),
				map[string]interface{}{
					"type":    roleCloudAccessPolicy,
					"region":  cloudAccessPolicyRegion,
					"scopes":  cloudAccessPolicyScopes,
					"realms":  cloudAccessPolicyRealms,
					"ttl":     testTTL,
					"max_ttl": testMaxTTL,
				})
			require.NoError(t, err)
		}

		resp, err := testTokenRoleList(t, b, s)
		require.NoError(t, err)
		require.Len(t, resp.Data["keys"].([]string), 10)
	})

	t.Run("Create User Role - pass", func(t *testing.T) {
		resp, err := testTokenRoleCreate(t, b, s, cloudAccessPolicyRoleName, map[string]interface{}{
			"type":    roleCloudAccessPolicy,
			"region":  cloudAccessPolicyRegion,
			"scopes":  cloudAccessPolicyScopes,
			"realms":  cloudAccessPolicyRealms,
			"ttl":     testTTL,
			"max_ttl": testMaxTTL,
		})

		require.Nil(t, err)
		require.Nil(t, resp.Error())
		require.Nil(t, resp)
	})

	t.Run("Create User Role - fail on invalid type", func(t *testing.T) {
		typeValues := map[string]interface{}{
			"Invalid type": "invalid",
			"Blank type":   "",
			"Not a string": 100,
		}
		for d, v := range typeValues {
			t.Run(d, func(t *testing.T) {
				resp, err := testTokenRoleCreate(t, b, s, cloudAccessPolicyRoleName, map[string]interface{}{
					"type":    v,
					"region":  cloudAccessPolicyRegion,
					"scopes":  cloudAccessPolicyScopes,
					"realms":  cloudAccessPolicyRealms,
					"ttl":     testTTL,
					"max_ttl": testMaxTTL,
				})

				require.Nil(t, err)
				require.NotNil(t, resp)
				require.NotNil(t, resp.Error())
			})
		}
	})

	t.Run("Create User Role - fail on invalid region", func(t *testing.T) {
		ttlValues := map[string]interface{}{
			"Empty region": "",
		}
		for d, v := range ttlValues {
			t.Run(d, func(t *testing.T) {
				resp, err := testTokenRoleCreate(t, b, s, cloudAccessPolicyRoleName, map[string]interface{}{
					"type":    roleCloudAccessPolicy,
					"region":  v,
					"scopes":  cloudAccessPolicyScopes,
					"realms":  cloudAccessPolicyRealms,
					"ttl":     testTTL,
					"max_ttl": testMaxTTL,
				})
				require.Nil(t, err)
				require.NotNil(t, resp)
				require.NotNil(t, resp.Error())
			})
		}
	})

	t.Run("Create User Role - fail on invalid scopes", func(t *testing.T) {
		ttlValues := map[string]interface{}{
			"Empty string": "",
		}
		for d, v := range ttlValues {
			t.Run(d, func(t *testing.T) {
				resp, err := testTokenRoleCreate(t, b, s, cloudAccessPolicyRoleName, map[string]interface{}{
					"type":    roleCloudAccessPolicy,
					"region":  cloudAccessPolicyRegion,
					"scopes":  v,
					"realms":  cloudAccessPolicyRealms,
					"ttl":     testTTL,
					"max_ttl": testMaxTTL,
				})

				require.Nil(t, err)
				require.NotNil(t, resp)
				require.NotNil(t, resp.Error())
			})
		}
	})

	t.Run("Create User Role - fail on invalid realms", func(t *testing.T) {
		ttlValues := map[string]interface{}{
			"Number":       1,
			"Empty string": "",
			"String":       "test",
		}
		for d, v := range ttlValues {
			t.Run(d, func(t *testing.T) {
				resp, err := testTokenRoleCreate(t, b, s, cloudAccessPolicyRoleName, map[string]interface{}{
					"type":    roleCloudAccessPolicy,
					"region":  cloudAccessPolicyRegion,
					"scopes":  cloudAccessPolicyScopes,
					"realms":  v,
					"ttl":     testTTL,
					"max_ttl": testMaxTTL,
				})

				require.Nil(t, err)
				require.NotNil(t, resp)
				require.NotNil(t, resp.Error())
			})
		}
	})

	t.Run("Create User Role - fail on invalid TTL", func(t *testing.T) {
		ttlValues := map[string]interface{}{
			"Not a number":         "a",
			"Negative number":      -1,
			"Greater than max ttl": testMaxTTL + 10,
		}
		for d, v := range ttlValues {
			t.Run(d, func(t *testing.T) {
				resp, err := testTokenRoleCreate(t, b, s, cloudAccessPolicyRoleName, map[string]interface{}{
					"type":    roleCloudAccessPolicy,
					"region":  cloudAccessPolicyRegion,
					"scopes":  cloudAccessPolicyScopes,
					"realms":  cloudAccessPolicyRealms,
					"ttl":     v,
					"max_ttl": testMaxTTL,
				})

				require.Nil(t, err)
				require.NotNil(t, resp)
				require.NotNil(t, resp.Error())
			})
		}
	})

	t.Run("Create User Role - fail on invalid Max TTL", func(t *testing.T) {
		ttlValues := map[string]interface{}{
			"Not a number":    "a",
			"Negative number": -1,
			"Less than ttl":   testTTL - 10,
		}
		for d, v := range ttlValues {
			t.Run(d, func(t *testing.T) {
				resp, err := testTokenRoleCreate(t, b, s, cloudAccessPolicyRoleName, map[string]interface{}{
					"type":    roleCloudAccessPolicy,
					"region":  cloudAccessPolicyRegion,
					"scopes":  cloudAccessPolicyScopes,
					"realms":  cloudAccessPolicyRealms,
					"ttl":     testTTL,
					"max_ttl": v,
				})

				require.Nil(t, err)
				require.NotNil(t, resp)
				require.NotNil(t, resp.Error())
			})
		}
	})

	t.Run("Read User Role - existing", func(t *testing.T) {
		resp, err := testTokenRoleRead(t, b, s, cloudAccessPolicyRoleName)

		require.Nil(t, err)
		require.NotNil(t, resp)
		require.Nil(t, resp.Error())
		require.Equal(t, resp.Data["type"], roleCloudAccessPolicy)
		require.Equal(t, resp.Data["region"], cloudAccessPolicyRegion)
		require.Equal(t, resp.Data["scopes"], cloudAccessPolicyScopes)
		require.Equal(t, resp.Data["realms"], cloudAccessPolicyRealms)
	})

	t.Run("Read User Role - non existent", func(t *testing.T) {
		resp, err := testTokenRoleRead(t, b, s, "non-existent-role")

		require.Nil(t, err)
		require.Nil(t, resp)
	})

	t.Run("Update User Role", func(t *testing.T) {
		resp, err := testTokenRoleUpdate(t, b, s, cloudAccessPolicyRoleName, map[string]interface{}{
			"ttl":     "1m",
			"max_ttl": "5h",
		})

		require.Nil(t, err)
		require.Nil(t, resp.Error())
		require.Nil(t, resp)
	})

	t.Run("Re-read User Role - existing", func(t *testing.T) {
		resp, err := testTokenRoleRead(t, b, s, cloudAccessPolicyRoleName)

		require.Nil(t, err)
		require.Nil(t, resp.Error())
		require.NotNil(t, resp)
		require.Equal(t, resp.Data["type"], roleCloudAccessPolicy)
		require.Equal(t, resp.Data["region"], cloudAccessPolicyRegion)
		require.Equal(t, resp.Data["scopes"], cloudAccessPolicyScopes)
		require.Equal(t, resp.Data["realms"], cloudAccessPolicyRealms)
	})

	t.Run("Delete User Role", func(t *testing.T) {
		_, err := testTokenRoleDelete(t, b, s, cloudAccessPolicyRoleName)

		require.NoError(t, err)
	})
}

func TestServiceAccountRole(t *testing.T) {
	b, s := getTestBackend(t)

	err := testConfigCreate(b, s, map[string]interface{}{
		"type":  GrafanaCloudType,
		"token": "abcd",
	})
	assert.NoError(t, err)

	t.Run("List All Roles", func(t *testing.T) {
		for i := 1; i <= 10; i++ {
			_, err := testTokenRoleCreate(t, b, s,
				serviceAccountRoleName+strconv.Itoa(i),
				map[string]interface{}{
					"type":    roleGrafanaServiceAccount,
					"stack":   serviceAccountStack,
					"role":    serviceAccountRole,
					"ttl":     testTTL,
					"max_ttl": testMaxTTL,
				})
			require.NoError(t, err)
		}

		resp, err := testTokenRoleList(t, b, s)
		require.NoError(t, err)
		require.Len(t, resp.Data["keys"].([]string), 10)
	})

	t.Run("Create User Role - pass", func(t *testing.T) {
		resp, err := testTokenRoleCreate(t, b, s, serviceAccountRoleName, map[string]interface{}{
			"type":    roleGrafanaServiceAccount,
			"stack":   serviceAccountStack,
			"role":    serviceAccountRole,
			"ttl":     testTTL,
			"max_ttl": testMaxTTL,
		})

		require.Nil(t, err)
		require.Nil(t, resp.Error())
		require.Nil(t, resp)
	})

	t.Run("Create User Role - fail on invalid type", func(t *testing.T) {
		typeValues := map[string]interface{}{
			"Invalid type": "invalid",
			"Blank type":   "",
			"Not a string": 100,
		}
		for d, v := range typeValues {
			t.Run(d, func(t *testing.T) {
				resp, err := testTokenRoleCreate(t, b, s, serviceAccountRoleName, map[string]interface{}{
					"type":    v,
					"stack":   serviceAccountStack,
					"role":    serviceAccountRole,
					"ttl":     testTTL,
					"max_ttl": testMaxTTL,
				})

				require.Nil(t, err)
				require.NotNil(t, resp)
				require.NotNil(t, resp.Error())
			})
		}
	})

	t.Run("Create User Role - fail on invalid stack", func(t *testing.T) {
		ttlValues := map[string]interface{}{
			"Empty stack": "",
		}
		for d, v := range ttlValues {
			t.Run(d, func(t *testing.T) {
				resp, err := testTokenRoleCreate(t, b, s, serviceAccountRoleName, map[string]interface{}{
					"type":    roleGrafanaServiceAccount,
					"stack":   v,
					"role":    serviceAccountRole,
					"ttl":     testTTL,
					"max_ttl": testMaxTTL,
				})
				require.Nil(t, err)
				require.NotNil(t, resp)
				require.NotNil(t, resp.Error())
			})
		}
	})

	t.Run("Create User Role - fail on invalid TTL", func(t *testing.T) {
		ttlValues := map[string]interface{}{
			"Not a number":         "a",
			"Negative number":      -1,
			"Greater than max ttl": testMaxTTL + 10,
		}
		for d, v := range ttlValues {
			t.Run(d, func(t *testing.T) {
				resp, err := testTokenRoleCreate(t, b, s, serviceAccountRoleName, map[string]interface{}{
					"type":    roleGrafanaServiceAccount,
					"stack":   serviceAccountStack,
					"role":    serviceAccountRole,
					"ttl":     v,
					"max_ttl": testMaxTTL,
				})

				require.Nil(t, err)
				require.NotNil(t, resp)
				require.NotNil(t, resp.Error())
			})
		}
	})

	t.Run("Create User Role - fail on invalid Max TTL", func(t *testing.T) {
		ttlValues := map[string]interface{}{
			"Not a number":    "a",
			"Negative number": -1,
			"Less than ttl":   testTTL - 10,
		}
		for d, v := range ttlValues {
			t.Run(d, func(t *testing.T) {
				resp, err := testTokenRoleCreate(t, b, s, serviceAccountRoleName, map[string]interface{}{
					"type":    roleGrafanaServiceAccount,
					"stack":   serviceAccountStack,
					"role":    serviceAccountRole,
					"ttl":     testTTL,
					"max_ttl": v,
				})

				require.Nil(t, err)
				require.NotNil(t, resp)
				require.NotNil(t, resp.Error())
			})
		}
	})

	t.Run("Read User Role - existing", func(t *testing.T) {
		resp, err := testTokenRoleRead(t, b, s, serviceAccountRoleName)

		require.Nil(t, err)
		require.NotNil(t, resp)
		require.Nil(t, resp.Error())
		require.Equal(t, resp.Data["type"], roleGrafanaServiceAccount)
		require.Equal(t, resp.Data["stack"], serviceAccountStack)
		require.Equal(t, resp.Data["role"], serviceAccountRole)
	})

	t.Run("Read User Role - non existent", func(t *testing.T) {
		resp, err := testTokenRoleRead(t, b, s, "non-existent-role")

		require.Nil(t, err)
		require.Nil(t, resp)
	})

	t.Run("Update User Role", func(t *testing.T) {
		resp, err := testTokenRoleUpdate(t, b, s, serviceAccountRoleName, map[string]interface{}{
			"ttl":     "1m",
			"max_ttl": "5h",
		})

		require.Nil(t, err)
		require.Nil(t, resp.Error())
		require.Nil(t, resp)
	})

	t.Run("Re-read User Role - existing", func(t *testing.T) {
		resp, err := testTokenRoleRead(t, b, s, serviceAccountRoleName)

		require.Nil(t, err)
		require.Nil(t, resp.Error())
		require.NotNil(t, resp)
		require.Equal(t, resp.Data["type"], roleGrafanaServiceAccount)
		require.Equal(t, resp.Data["stack"], serviceAccountStack)
		require.Equal(t, resp.Data["role"], serviceAccountRole)
	})

	t.Run("Delete User Role", func(t *testing.T) {
		_, err := testTokenRoleDelete(t, b, s, serviceAccountRoleName)

		require.NoError(t, err)
	})
}

// Utility function to create a role while, returning any response (including errors).
func testTokenRoleCreate(t *testing.T, b *grafanaBackend, s logical.Storage, roleName string, d map[string]interface{}) (*logical.Response, error) {
	t.Helper()
	resp, err := b.HandleRequest(context.Background(), &logical.Request{
		Operation: logical.CreateOperation,
		Path:      "roles/" + roleName,
		Data:      d,
		Storage:   s,
	})
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// Utility function to update a role while, returning any response (including errors).
func testTokenRoleUpdate(t *testing.T, b *grafanaBackend, s logical.Storage, roleName string, d map[string]interface{}) (*logical.Response, error) {
	t.Helper()
	resp, err := b.HandleRequest(context.Background(), &logical.Request{
		Operation: logical.UpdateOperation,
		Path:      "roles/" + roleName,
		Data:      d,
		Storage:   s,
	})
	if err != nil {
		return nil, err
	}

	if resp != nil && resp.IsError() {
		t.Fatal(resp.Error())
	}
	return resp, nil
}

// Utility function to read a role and return any errors.
func testTokenRoleRead(t *testing.T, b *grafanaBackend, s logical.Storage, vRole string) (*logical.Response, error) {
	t.Helper()

	return b.HandleRequest(context.Background(), &logical.Request{
		Operation: logical.ReadOperation,
		Path:      "roles/" + vRole,
		Storage:   s,
	})
}

// Utility function to list roles and return any errors.
func testTokenRoleList(t *testing.T, b *grafanaBackend, s logical.Storage) (*logical.Response, error) {
	t.Helper()
	return b.HandleRequest(context.Background(), &logical.Request{
		Operation: logical.ListOperation,
		Path:      "roles/",
		Storage:   s,
	})
}

// Utility function to delete a role and return any errors.
func testTokenRoleDelete(t *testing.T, b *grafanaBackend, s logical.Storage, vRole string) (*logical.Response, error) {
	t.Helper()
	return b.HandleRequest(context.Background(), &logical.Request{
		Operation: logical.DeleteOperation,
		Path:      "roles/" + vRole,
		Storage:   s,
	})
}
