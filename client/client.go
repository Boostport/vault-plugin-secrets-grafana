package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"time"
)

const (
	defaultTimeout = 30 * time.Second
)

type Grafana struct {
	client      *http.Client
	bearerToken string
	baseURL     url.URL
}

func New(baseURL, bearerToken string) (*Grafana, error) {

	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("error parsing base URL: %w", err)
	}

	return &Grafana{
		client:      &http.Client{},
		bearerToken: bearerToken,
		baseURL:     *u,
	}, nil
}

func (g *Grafana) do(method, requestPath string, query url.Values, body []byte, responseStruct interface{}) error {
	requestURL := g.baseURL
	requestURL.Path = path.Join(requestURL.Path, requestPath)
	requestURL.RawQuery = query.Encode()

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, method, requestURL.String(), bytes.NewBuffer(body))

	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", g.bearerToken))
	req.Header.Add("Content-Type", "application/json")

	resp, err := g.client.Do(req)

	if err != nil {
		return fmt.Errorf("error making request: %w", err)
	}

	bodyContents, err := io.ReadAll(resp.Body)

	if err != nil {
		return fmt.Errorf("error reading response body: %w", err)
	}

	resp.Body.Close()

	if resp.StatusCode > 299 {
		return fmt.Errorf("error response from server (%d): %s", resp.StatusCode, bodyContents)
	}

	if responseStruct == nil {
		return nil
	}

	err = json.Unmarshal(bodyContents, responseStruct)
	if err != nil {
		return fmt.Errorf("error decoding json response: %w", err)
	}

	return nil
}
