package main

import (
	"os"

	"github.com/Boostport/vault-plugin-secrets-grafana/client"
)

func main() {
	serviceAccountToken := os.Getenv("GRAFANA_SERVICE_ACCOUNT_TOKEN")

	if serviceAccountToken == "" {
		panic("GRAFANA_SERVICE_ACCOUNT_TOKEN not set")
	}

	grafanaInstanceURL := os.Getenv("GRAFANA_INSTANCE_URL")

	if grafanaInstanceURL == "" {
		panic("GRAFANA_INSTANCE_URL not set")
	}

	c, err := client.New(grafanaInstanceURL, serviceAccountToken)

	if err != nil {
		panic(err)
	}

	_, err = c.GetHomeDashboard()

	if err != nil {
		panic(err)
	}
}
