package vault_plugin_secrets_grafana

import (
	"context"
	"os"
	"testing"
	"time"

	log "github.com/hashicorp/go-hclog"
	"github.com/hashicorp/vault/sdk/helper/logging"
	"github.com/hashicorp/vault/sdk/logical"
)

func newCloudAcceptanceTestEnv() (*testCloudEnv, error) {
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
	b, err := Factory("test")(ctx, conf)
	if err != nil {
		return nil, err
	}
	return &testCloudEnv{
		Token:          os.Getenv(envVarGrafanaCloudToken),
		CloudStackSlug: os.Getenv(envVarGrafanaCloudStackSlug),
		CloudRegion:    os.Getenv(envVarGrafanaCloudRegion),
		OrgIdentifier:  os.Getenv(envVarGrafanaCloudOrgIdentifier),

		Backend: b,
		Context: ctx,
		Storage: &logical.InmemStorage{},
	}, nil
}

func TestCloudAcceptanceToken(t *testing.T) {
	if !runAcceptanceTests {
		t.SkipNow()
	}

	acceptanceTestEnv, err := newCloudAcceptanceTestEnv()
	if err != nil {
		t.Fatal(err)
	}

	t.Run("add config", acceptanceTestEnv.AddConfig)
	t.Run("add access policy role", acceptanceTestEnv.AddAccessPolicyRole)
	t.Run("add service account role", acceptanceTestEnv.AddServiceAccountRole)
	t.Run("read access policy cred", acceptanceTestEnv.ReadAccessPolicyToken)
	t.Run("read access policy cred", acceptanceTestEnv.ReadAccessPolicyToken)
	t.Run("read service account cred", acceptanceTestEnv.ReadServiceAccountToken)
	t.Run("read service account cred", acceptanceTestEnv.ReadServiceAccountToken)
	t.Run("verify number of issued tokens", acceptanceTestEnv.VerifyNumberOfIssuedCredentials)
	t.Run("cleanup creds", acceptanceTestEnv.CleanupCreds)
}

func TestGrafanaInstanceToken(t *testing.T) {
	if !runAcceptanceTests {
		t.SkipNow()
	}

	acceptanceTestEnv, err := newCloudAcceptanceTestEnv()
	if err != nil {
		t.Fatal(err)
	}

	t.Run("add config", acceptanceTestEnv.AddConfig)
	t.Run("add service account role", acceptanceTestEnv.AddServiceAccountRole)
	t.Run("read service account cred", acceptanceTestEnv.ReadServiceAccountToken)

	instanceTestEnv := acceptanceTestEnv.GetInstanceEnv(t)
	t.Run("add custom grafana role", instanceTestEnv.AddCustomGrafanaRole)
	t.Run("add service account role", instanceTestEnv.AddServiceAccountRoleWithCustomGrafanaRoles)
	t.Run("read service account cred", instanceTestEnv.ReadServiceAccountToken)
	t.Run("read service account cred", instanceTestEnv.ReadServiceAccountToken)

	t.Run("cleanup instance custom roles", instanceTestEnv.CleanupCustomRoles)
	t.Run("cleanup instance creds", instanceTestEnv.CleanupCreds)
	t.Run("cleanup cloud creds", acceptanceTestEnv.CleanupCreds)
}
