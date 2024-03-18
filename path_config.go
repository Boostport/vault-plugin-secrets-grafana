package vault_plugin_secrets_grafana

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

const (
	configStoragePath      = "config"
	defaultGrafanaCloudURL = "https://grafana.com"
	GrafanaCloudType       = "cloud"
	GrafanaType            = "grafana"
)

type grafanaConfig struct {
	Type  string `json:"type"`
	Token string `json:"token"`
	URL   string `json:"url,omitempty"`
}

func (c *grafanaConfig) validate() error {
	if c.Type != GrafanaCloudType && c.Type != GrafanaType {
		return fmt.Errorf("type must be either '%s' or '%s'", GrafanaCloudType, GrafanaType)
	}

	if c.Token == "" {
		return errors.New("token must not be empty")
	}

	if c.Type == GrafanaType {
		if c.URL == "" {
			return errors.New("url must not be empty")
		}

		if u, err := url.ParseRequestURI(c.URL); err != nil || !u.IsAbs() {
			return fmt.Errorf("invalid url in configuration: %s", c.URL)
		}
	} else if c.Type == GrafanaCloudType {
		if c.URL == "" {
			c.URL = defaultGrafanaCloudURL
		}
	}

	return nil
}

func pathConfig(b *grafanaBackend) *framework.Path {
	return &framework.Path{
		Pattern: "config",
		Fields: map[string]*framework.FieldSchema{
			"type": {
				Type:        framework.TypeString,
				Description: "The type of Grafana instance to generate tokens for. Either 'cloud' or 'grafana'",
				Required:    true,
				DisplayAttrs: &framework.DisplayAttributes{
					Name:      "Type",
					Sensitive: false,
				},
			},
			"token": {
				Type:        framework.TypeString,
				Description: "The token to use for authentication",
				DisplayAttrs: &framework.DisplayAttributes{
					Name:      "Token",
					Sensitive: true,
				},
			},
			"url": {
				Type:        framework.TypeString,
				Description: "The URL of the Grafana Cloud or Grafana instance to connect to",
				DisplayAttrs: &framework.DisplayAttributes{
					Name:      "URL",
					Sensitive: false,
				},
			},
		},
		Operations: map[logical.Operation]framework.OperationHandler{
			logical.ReadOperation: &framework.PathOperation{
				Callback: b.pathConfigRead,
			},
			logical.CreateOperation: &framework.PathOperation{
				Callback: b.pathConfigWrite,
			},
			logical.UpdateOperation: &framework.PathOperation{
				Callback: b.pathConfigWrite,
			},
			logical.DeleteOperation: &framework.PathOperation{
				Callback: b.pathConfigDelete,
			},
		},
		ExistenceCheck:  b.pathConfigExistenceCheck,
		HelpSynopsis:    pathConfigHelpSynopsis,
		HelpDescription: pathConfigHelpDescription,
	}
}

func (b *grafanaBackend) pathConfigExistenceCheck(ctx context.Context, req *logical.Request, d *framework.FieldData) (bool, error) {
	out, err := req.Storage.Get(ctx, req.Path)
	if err != nil {
		return false, fmt.Errorf("existence check failed: %w", err)
	}

	return out != nil, nil
}

func (b *grafanaBackend) pathConfigRead(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	config, err := getConfig(ctx, req.Storage)
	if err != nil {
		return nil, err
	}

	if config == nil {
		return nil, nil
	}

	return &logical.Response{
		Data: map[string]interface{}{
			"type":  config.Type,
			"token": config.Token,
			"url":   config.URL,
		},
	}, nil
}

func (b *grafanaBackend) pathConfigWrite(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	config, err := getConfig(ctx, req.Storage)
	if err != nil {
		return nil, err
	}

	createOperation := req.Operation == logical.CreateOperation

	if config == nil {
		if !createOperation {
			return nil, errors.New("config not found during update operation")
		}
		config = new(grafanaConfig)
	}

	if configTypeRaw, ok := data.GetOk("type"); ok {
		config.Type = configTypeRaw.(string)

	}
	if token, ok := data.GetOk("token"); ok {
		config.Token = token.(string)
	}

	if configURL, ok := data.GetOk("url"); ok {
		config.URL = configURL.(string)
	}

	if err := config.validate(); err != nil {
		return nil, err
	}

	entry, err := logical.StorageEntryJSON(configStoragePath, config)
	if err != nil {
		return nil, err
	}

	if err := req.Storage.Put(ctx, entry); err != nil {
		return nil, err
	}

	// reset the client so the next invocation will pick up the new configuration
	b.reset()

	return nil, nil
}

func (b *grafanaBackend) pathConfigDelete(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	err := req.Storage.Delete(ctx, configStoragePath)

	if err == nil {
		b.reset()
	}

	return nil, err
}

func getConfig(ctx context.Context, s logical.Storage) (*grafanaConfig, error) {
	entry, err := s.Get(ctx, configStoragePath)
	if err != nil {
		return nil, fmt.Errorf("error reading mount configuration: %w", err)
	}

	if entry == nil {
		return nil, nil
	}

	config := new(grafanaConfig)
	if err := entry.DecodeJSON(&config); err != nil {
		return nil, fmt.Errorf("error reading root configuration: %w", err)
	}

	return config, nil
}

const (
	pathConfigHelpSynopsis    = `Configure the Grafana backend.`
	pathConfigHelpDescription = `
The Grafana secret backend requires a token to manage tokens that it issues.
`
)
