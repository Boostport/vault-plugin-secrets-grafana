package vault_plugin_secrets_grafana

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/assert"
)

const (
	token     = "123456"
	configURL = "http://localhost:19090"
)

func TestConfig(t *testing.T) {
	b, reqStorage := getTestBackend(t)

	t.Run("Test Configuration", func(t *testing.T) {

		t.Run("Create Configuration (Cloud) - empty token", func(t *testing.T) {
			err := testConfigCreate(b, reqStorage, map[string]interface{}{
				"type":  GrafanaCloudType,
				"token": "",
			})
			assert.Error(t, err)
		})

		t.Run("Create Configuration (Grafana) - empty token", func(t *testing.T) {
			err := testConfigCreate(b, reqStorage, map[string]interface{}{
				"type":  GrafanaType,
				"token": "",
				"url":   configURL,
			})
			assert.Error(t, err)
		})

		t.Run("Create Configuration (Grafana) - empty url", func(t *testing.T) {
			err := testConfigCreate(b, reqStorage, map[string]interface{}{
				"type":  GrafanaType,
				"token": token,
				"url":   "",
			})
			assert.Error(t, err)
		})

		t.Run("Create Configuration (Grafana) - invalid url", func(t *testing.T) {
			err := testConfigCreate(b, reqStorage, map[string]interface{}{
				"type":  GrafanaType,
				"token": token,
				"url":   "/addd",
			})
			assert.Error(t, err)
		})

		t.Run("Create Configuration (Cloud) - pass", func(t *testing.T) {
			err := testConfigCreate(b, reqStorage, map[string]interface{}{
				"type":  GrafanaCloudType,
				"token": token,
			})
			assert.NoError(t, err)
		})

		t.Run("Read Configuration (Cloud) - pass", func(t *testing.T) {
			err := testConfigRead(b, reqStorage, map[string]interface{}{
				"type":  GrafanaCloudType,
				"token": token,
				"url":   defaultGrafanaCloudURL,
			})
			assert.NoError(t, err)
		})

		t.Run("Update Configuration (Cloud - set token) - pass", func(t *testing.T) {
			err := testConfigUpdate(b, reqStorage, map[string]interface{}{
				"type":  GrafanaCloudType,
				"token": "abcd",
			})
			assert.NoError(t, err)
		})

		t.Run("Read Updated Configuration (Cloud - set token) - pass", func(t *testing.T) {
			err := testConfigRead(b, reqStorage, map[string]interface{}{
				"type":  GrafanaCloudType,
				"token": "abcd",
				"url":   defaultGrafanaCloudURL,
			})
			assert.NoError(t, err)
		})

		t.Run("Update Configuration (Cloud - set type) - pass", func(t *testing.T) {
			err := testConfigUpdate(b, reqStorage, map[string]interface{}{
				"type":  GrafanaType,
				"url":   configURL,
				"token": "abcd",
			})
			assert.NoError(t, err)
		})

		t.Run("Read Updated Configuration (Cloud - set type) - pass", func(t *testing.T) {
			err := testConfigRead(b, reqStorage, map[string]interface{}{
				"type":  GrafanaType,
				"url":   configURL,
				"token": "abcd",
			})
			assert.NoError(t, err)
		})

		t.Run("Delete Configuration (Cloud) - pass", func(t *testing.T) {
			err := testConfigDelete(b, reqStorage)
			assert.NoError(t, err)
		})

		t.Run("Create Configuration (Grafana) - pass", func(t *testing.T) {
			err := testConfigCreate(b, reqStorage, map[string]interface{}{
				"type":  GrafanaType,
				"token": token,
				"url":   configURL,
			})
			assert.NoError(t, err)
		})

		t.Run("Read Configuration (Grafana) - pass", func(t *testing.T) {
			err := testConfigRead(b, reqStorage, map[string]interface{}{
				"type":  GrafanaType,
				"token": token,
				"url":   configURL,
			})
			assert.NoError(t, err)
		})

		t.Run("Update Configuration (Grafana - set token and url) - pass", func(t *testing.T) {
			err := testConfigUpdate(b, reqStorage, map[string]interface{}{
				"type":  GrafanaCloudType,
				"url":   "https://test.com:19090",
				"token": "abcd",
			})
			assert.NoError(t, err)
		})

		t.Run("Read Updated Configuration (Grafana - set token and url) - pass", func(t *testing.T) {
			err := testConfigRead(b, reqStorage, map[string]interface{}{
				"type":  GrafanaCloudType,
				"url":   "https://test.com:19090",
				"token": "abcd",
			})
			assert.NoError(t, err)
		})

		t.Run("Update Configuration (Grafana - set type) - pass", func(t *testing.T) {
			err := testConfigUpdate(b, reqStorage, map[string]interface{}{
				"type":  GrafanaCloudType,
				"token": token,
				"url":   "",
			})
			assert.NoError(t, err)
		})

		t.Run("Read Updated Configuration (Grafana - set type) - pass", func(t *testing.T) {
			err := testConfigRead(b, reqStorage, map[string]interface{}{
				"type":  GrafanaCloudType,
				"token": token,
				"url":   defaultGrafanaCloudURL,
			})
			assert.NoError(t, err)
		})

		t.Run("Delete Configuration (Grafana) - pass", func(t *testing.T) {
			err := testConfigDelete(b, reqStorage)
			assert.NoError(t, err)
		})
	})
}

func testConfigCreate(b logical.Backend, s logical.Storage, d map[string]interface{}) error {
	resp, err := b.HandleRequest(context.Background(), &logical.Request{
		Operation: logical.CreateOperation,
		Path:      configStoragePath,
		Data:      d,
		Storage:   s,
	})
	if err != nil {
		return err
	}

	if resp != nil && resp.IsError() {
		return resp.Error()
	}
	return nil
}

func testConfigDelete(b logical.Backend, s logical.Storage) error {
	resp, err := b.HandleRequest(context.Background(), &logical.Request{
		Operation: logical.DeleteOperation,
		Path:      configStoragePath,
		Storage:   s,
	})
	if err != nil {
		return err
	}

	if resp != nil && resp.IsError() {
		return resp.Error()
	}
	return nil
}

func testConfigUpdate(b logical.Backend, s logical.Storage, d map[string]interface{}) error {
	resp, err := b.HandleRequest(context.Background(), &logical.Request{
		Operation: logical.UpdateOperation,
		Path:      configStoragePath,
		Data:      d,
		Storage:   s,
	})
	if err != nil {
		return err
	}

	if resp != nil && resp.IsError() {
		return resp.Error()
	}
	return nil
}

func testConfigRead(b logical.Backend, s logical.Storage, expected map[string]interface{}) error {
	resp, err := b.HandleRequest(context.Background(), &logical.Request{
		Operation: logical.ReadOperation,
		Path:      configStoragePath,
		Storage:   s,
	})
	if err != nil {
		return err
	}

	if resp == nil && expected == nil {
		return nil
	}

	if resp.IsError() {
		return resp.Error()
	}

	if len(expected) != len(resp.Data) {
		return fmt.Errorf("read data mismatch (expected %d values, got %d)", len(expected), len(resp.Data))
	}

	for k, expectedV := range expected {
		actualV, ok := resp.Data[k]

		if !ok {
			return fmt.Errorf(`expected data["%s"] = %v but was not included in read output"`, k, expectedV)
		} else if expectedV != actualV {
			return fmt.Errorf(`expected data["%s"] = %v, instead got %v"`, k, expectedV, actualV)
		}
	}

	return nil
}
