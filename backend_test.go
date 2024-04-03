package vault_plugin_secrets_grafana

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/Boostport/vault-plugin-secrets-grafana/client"
	"github.com/hashicorp/go-hclog"
	log "github.com/hashicorp/go-hclog"
	"github.com/hashicorp/vault/sdk/helper/logging"
	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/require"
)

const (
	envVarRunAccTests               = "VAULT_ACC"
	envVarGrafanaCloudToken         = "TEST_GRAFANA_CLOUD_TOKEN"
	envVarGrafanaCloudStackSlug     = "TEST_GRAFANA_CLOUD_STACK_SLUG"
	envVarGrafanaCloudRegion        = "TEST_GRAFANA_CLOUD_REGION"
	envVarGrafanaCloudOrgIdentifier = "TEST_GRAFANA_CLOUD_ORG_IDENTIFIER"
	customGrafanaRoleName           = "test-custom-role"
)

func getTestBackend(tb testing.TB) (*grafanaBackend, logical.Storage) {
	tb.Helper()

	config := logical.TestBackendConfig()
	config.StorageView = new(logical.InmemStorage)
	config.Logger = hclog.NewNullLogger()
	config.System = logical.TestSystemView()

	b, err := Factory("test")(context.Background(), config)
	if err != nil {
		tb.Fatal(err)
	}

	return b.(*grafanaBackend), config.StorageView
}

// runAcceptanceTests will separate unit tests from
// acceptance tests, which will make active requests
// to your target API.
var runAcceptanceTests = os.Getenv(envVarRunAccTests) == "1"

// testCloudEnv creates an object to store and track testing environment
// resources.
type testCloudEnv struct {
	Token          string
	CloudStackSlug string
	CloudRegion    string
	OrgIdentifier  string

	Backend logical.Backend
	Context context.Context
	Storage logical.Storage

	// SecretToken tracks the API token, for checking rotations.
	SecretToken string

	// Tokens tracks the generated tokens, to make sure we clean up.
	AccessPolicyIDs   []string
	ServiceAccountIDs []int64
}

// AddConfig adds the configuration to the test backend.
// Make sure data includes all of the configuration
// attributes you need and the `config` path!
func (e *testCloudEnv) AddConfig(t *testing.T) {
	req := &logical.Request{
		Operation: logical.CreateOperation,
		Path:      "config",
		Storage:   e.Storage,
		Data: map[string]interface{}{
			"type":  GrafanaCloudType,
			"token": e.Token,
		},
	}
	resp, err := e.Backend.HandleRequest(e.Context, req)
	require.Nil(t, resp)
	require.Nil(t, err)
}

func (e *testCloudEnv) GetInstanceEnv(t *testing.T) *testInstanceEnv {

	b := e.Backend.(*grafanaBackend)
	c, err := b.getClient(e.Context, e.Storage)
	if err != nil {
		t.Fatal("error getting client")
	}

	stack, err := c.StackBySlug(e.CloudStackSlug)
	if err != nil {
		t.Fatalf("unexpected error getting stack: %s", err)
	}

	t.Run("add instance service account", e.AddServiceAccountRole)
	t.Run("add instance service token", e.ReadServiceAccountToken)

	ctx := context.Background()

	maxLease, _ := time.ParseDuration("60s")
	defaultLease, _ := time.ParseDuration("30s")
	conf := &logical.BackendConfig{
		System: &logical.StaticSystemView{
			DefaultLeaseTTLVal: defaultLease,
			MaxLeaseTTLVal:     maxLease,
		},
		Logger: logging.NewVaultLogger(log.Debug),
	}

	instanceBackend, err := Factory("test")(ctx, conf)
	if err != nil {
		t.Fatalf("unexpected error creating instance backend: %s", err)
	}

	instanceTestEnv := &testInstanceEnv{
		Token:   e.SecretToken,
		Backend: instanceBackend,
		Context: ctx,
		Storage: &logical.InmemStorage{},
	}

	req := &logical.Request{
		Operation: logical.CreateOperation,
		Path:      "config",
		Storage:   instanceTestEnv.Storage,
		Data: map[string]interface{}{
			"type":  GrafanaType,
			"token": e.SecretToken,
			"url":   stack.URL,
		},
	}

	resp, err := e.Backend.HandleRequest(e.Context, req)
	require.Nil(t, resp)
	require.Nil(t, err)

	return instanceTestEnv
}

func (e *testCloudEnv) AddAccessPolicyRole(t *testing.T) {
	req := &logical.Request{
		Operation: logical.UpdateOperation,
		Path:      "roles/test-access-policy",
		Storage:   e.Storage,
		Data: map[string]interface{}{
			"type":   roleCloudAccessPolicy,
			"region": e.CloudRegion,
			"scopes": []string{"accesspolicies:delete", "accesspolicies:read", "accesspolicies:write", "stacks:read", "stack-service-accounts:write"},
			"realms": fmt.Sprintf(`[{"type": "org", "identifier": "%s", "labelPolicies": []}]`, e.OrgIdentifier),
		},
	}
	resp, err := e.Backend.HandleRequest(e.Context, req)
	require.Nil(t, resp)
	require.Nil(t, err)
}

func (e *testCloudEnv) AddServiceAccountRole(t *testing.T) {
	req := &logical.Request{
		Operation: logical.UpdateOperation,
		Path:      "roles/test-service-account",
		Storage:   e.Storage,
		Data: map[string]interface{}{
			"type":  roleGrafanaServiceAccount,
			"stack": e.CloudStackSlug,
			"role":  "Admin",
		},
	}
	resp, err := e.Backend.HandleRequest(e.Context, req)
	require.Nil(t, resp)
	require.Nil(t, err)
}

func (e *testCloudEnv) ReadAccessPolicyToken(t *testing.T) {
	req := &logical.Request{
		Operation: logical.ReadOperation,
		Path:      "creds/test-access-policy",
		Storage:   e.Storage,
	}
	resp, err := e.Backend.HandleRequest(e.Context, req)
	require.Nil(t, err)
	require.NotNil(t, resp)
	require.True(t, resp.Secret.InternalData["is_cloud"].(bool))
	require.NotEmpty(t, resp.Secret.InternalData["access_policy_id"])
	require.NotNil(t, resp.Secret)
	require.NotEmpty(t, resp.Data["token"])

	if e.SecretToken != "" {
		require.NotEqual(t, e.SecretToken, resp.Data["token"])
	}

	e.SecretToken = resp.Data["token"].(string)

	e.AccessPolicyIDs = append(e.AccessPolicyIDs, resp.Secret.InternalData["access_policy_id"].(string))
}

func (e *testCloudEnv) ReadServiceAccountToken(t *testing.T) {
	req := &logical.Request{
		Operation: logical.ReadOperation,
		Path:      "creds/test-service-account",
		Storage:   e.Storage,
	}
	resp, err := e.Backend.HandleRequest(e.Context, req)
	require.Nil(t, err)
	require.NotNil(t, resp)
	require.True(t, resp.Secret.InternalData["is_cloud"].(bool))
	require.NotEmpty(t, resp.Secret.InternalData["service_account_id"])
	require.NotEmpty(t, resp.Secret.InternalData["stack"])
	require.NotNil(t, resp.Secret)
	require.NotEmpty(t, resp.Data["token"])

	if e.SecretToken != "" {
		require.NotEqual(t, e.SecretToken, resp.Data["token"])
	}

	e.SecretToken = resp.Data["token"].(string)

	e.ServiceAccountIDs = append(e.ServiceAccountIDs, resp.Secret.InternalData["service_account_id"].(int64))
}

func (e *testCloudEnv) VerifyNumberOfIssuedCredentials(t *testing.T) {
	if len(e.AccessPolicyIDs) != 2 {
		t.Fatalf("expected 2 access policies, got: %d", len(e.AccessPolicyIDs))
	}

	if len(e.ServiceAccountIDs) != 2 {
		t.Fatalf("expected 2 service accounts, got: %d", len(e.AccessPolicyIDs))
	}
}

func (e *testCloudEnv) CleanupCreds(t *testing.T) {

	if len(e.AccessPolicyIDs) <= 0 && len(e.ServiceAccountIDs) <= 0 {
		return
	}

	b := e.Backend.(*grafanaBackend)
	client, err := b.getClient(e.Context, e.Storage)
	if err != nil {
		t.Fatal("error getting client")
	}

	if len(e.AccessPolicyIDs) > 0 {
		for _, id := range e.AccessPolicyIDs {
			err = client.DeleteCloudAccessPolicy(e.CloudRegion, id)
			if err != nil {
				t.Fatalf("unexpected error deleting access policy: %s", err)
			}
		}
	}

	if len(e.ServiceAccountIDs) > 0 {
		for _, id := range e.ServiceAccountIDs {
			err = client.DeleteGrafanaServiceAccountFromCloud(e.CloudStackSlug, id)
			if err != nil {
				t.Fatalf("unexpected error deleting service account: %s", err)
			}
		}
	}
}

type testInstanceEnv struct {
	Token string

	Backend logical.Backend
	Context context.Context
	Storage logical.Storage

	// SecretToken tracks the API token, for checking rotations.
	SecretToken string

	// CustomRoleIDs tracks the created custom roles, to make sure we clean up.
	CustomRoleIDs []string

	// Tokens tracks the generated tokens, to make sure we clean up.
	ServiceAccountIDs []int64
}

func (e *testInstanceEnv) AddCustomGrafanaRole(t *testing.T) {
	b := e.Backend.(*grafanaBackend)
	c, err := b.getClient(e.Context, e.Storage)
	if err != nil {
		t.Fatal("error getting client")
	}

	input := client.RoleInput{
		Name: customGrafanaRoleName,
	}

	resp, err := c.CreateCustomRole(input)
	if err != nil {
		t.Fatalf("unexpected error creating custom role: %s", err)
	}

	require.Nil(t, err)
	require.NotNil(t, resp)

	e.CustomRoleIDs = append(e.CustomRoleIDs, resp.UID)
}

func (e *testInstanceEnv) AddServiceAccountRoleWithCustomGrafanaRoles(t *testing.T) {
	req := &logical.Request{
		Operation: logical.UpdateOperation,
		Path:      "roles/test-service-account-with-roles",
		Storage:   e.Storage,
		Data: map[string]interface{}{
			"role":       "Admin",
			"rbac_roles": []string{customGrafanaRoleName},
		},
	}
	resp, err := e.Backend.HandleRequest(e.Context, req)
	require.Nil(t, resp)
	require.Nil(t, err)
}

func (e *testInstanceEnv) ReadServiceAccountToken(t *testing.T) {
	req := &logical.Request{
		Operation: logical.ReadOperation,
		Path:      "creds/test-service-account-with-roles",
		Storage:   e.Storage,
	}
	resp, err := e.Backend.HandleRequest(e.Context, req)

	require.Nil(t, err)
	require.NotNil(t, resp)
	require.False(t, resp.Secret.InternalData["is_cloud"].(bool))
	require.NotEmpty(t, resp.Secret.InternalData["service_account_id"])
	require.NotNil(t, resp.Secret)
	require.NotEmpty(t, resp.Data["token"])

	if e.SecretToken != "" {
		require.NotEqual(t, e.SecretToken, resp.Data["token"])
	}

	e.SecretToken = resp.Data["token"].(string)

	e.ServiceAccountIDs = append(e.ServiceAccountIDs, resp.Secret.InternalData["service_account_id"].(int64))
}

func (e *testInstanceEnv) CleanupCustomRoles(t *testing.T) {
	if len(e.CustomRoleIDs) == 0 {
		t.Fatalf("expected 1 custom role, got: %d", len(e.CustomRoleIDs))
	}

	b := e.Backend.(*grafanaBackend)
	c, err := b.getClient(e.Context, e.Storage)
	if err != nil {
		t.Fatal("error getting client")
	}

	for _, id := range e.CustomRoleIDs {
		err = c.DeleteCustomRole(id)
		if err != nil {
			t.Fatalf("unexpected error deleting custom role: %s", err)
		}
	}
}

func (e *testInstanceEnv) CleanupCreds(t *testing.T) {

	if len(e.ServiceAccountIDs) <= 0 {
		return
	}

	b := e.Backend.(*grafanaBackend)
	client, err := b.getClient(e.Context, e.Storage)
	if err != nil {
		t.Fatal("error getting client")
	}

	for _, id := range e.ServiceAccountIDs {
		err = client.DeleteServiceAccount(id)
		if err != nil {
			t.Fatalf("unexpected error deleting service account: %s", err)
		}
	}
}
