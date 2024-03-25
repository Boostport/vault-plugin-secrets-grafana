package vault_plugin_secrets_grafana

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/Boostport/vault-plugin-secrets-grafana/client"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

func Factory(version string) logical.Factory {
	return func(ctx context.Context, conf *logical.BackendConfig) (logical.Backend, error) {
		b := backend(version)
		if err := b.Setup(ctx, conf); err != nil {
			return nil, err
		}
		return b, nil
	}
}

type grafanaBackend struct {
	*framework.Backend
	lock   sync.RWMutex
	client *client.Grafana
}

func backend(version string) *grafanaBackend {
	var b grafanaBackend

	b.Backend = &framework.Backend{
		Help: strings.TrimSpace(backendHelp),
		PathsSpecial: &logical.Paths{
			SealWrapStorage: []string{
				"config",
				"role/*",
			},
		},
		Paths: framework.PathAppend(
			pathRole(&b),
			[]*framework.Path{
				pathConfig(&b),
				pathCredentials(&b),
			},
		),
		Secrets: []*framework.Secret{
			b.grafanaToken(),
		},
		BackendType: logical.TypeLogical,
		Invalidate:  b.invalidate,
	}

	if version != "" {
		b.Backend.RunningVersion = fmt.Sprintf("v%s", version)
	}

	return &b
}

func (b *grafanaBackend) reset() {
	b.lock.Lock()
	defer b.lock.Unlock()
	b.client = nil
}

func (b *grafanaBackend) invalidate(_ context.Context, key string) {
	if key == configStoragePath {
		b.reset()
	}
}

func (b *grafanaBackend) getClient(ctx context.Context, s logical.Storage) (*client.Grafana, error) {
	b.lock.RLock()
	unlockFunc := b.lock.RUnlock
	defer func() { unlockFunc() }()

	if b.client != nil {
		return b.client, nil
	}

	b.lock.RUnlock()
	b.lock.Lock()
	unlockFunc = b.lock.Unlock

	config, err := getConfig(ctx, s)
	if err != nil {
		return nil, err
	}

	if config == nil {
		config = new(grafanaConfig)
	}

	baseURL := strings.TrimSuffix(strings.ToLower(config.URL), "/")

	b.client, err = client.New(baseURL, config.Token)

	if err != nil {
		return nil, fmt.Errorf("error creating grafana client: %w", err)
	}

	return b.client, nil
}

const backendHelp = `
The Grafana secrets backend dynamically generates Grafana Cloud Access Policy tokens and Grafana Service Account tokens.
After mounting this backend, credentials to manage Grafana Cloud or Grafana tokens must be configured with the
"config/" endpoint.
`
