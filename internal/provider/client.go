package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const defaultBaseURL = "https://api.openai.com"

// OpenAIClient is a minimal HTTP client for the OpenAI platform admin API.
type OpenAIClient struct {
	apiKey  string
	orgID   string
	baseURL string
	http    *http.Client
}

func newClient(apiKey, orgID, baseURL string) *OpenAIClient {
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	return &OpenAIClient{
		apiKey:  apiKey,
		orgID:   orgID,
		baseURL: baseURL,
		http:    &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *OpenAIClient) do(ctx context.Context, method, path string, body any) ([]byte, int, error) {
	var buf io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, 0, fmt.Errorf("marshal request: %w", err)
		}
		buf = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, buf)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	if c.orgID != "" {
		req.Header.Set("OpenAI-Organization", c.orgID)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}
	if resp.StatusCode >= 400 {
		return nil, resp.StatusCode, fmt.Errorf("API error %d: %s", resp.StatusCode, string(data))
	}
	return data, resp.StatusCode, nil
}

// ── Project (workspace equivalent in OpenAI) ─────────────────────────────

type Project struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Status      string `json:"status"`
	CreatedAt   int64  `json:"created_at"`
	ArchivedAt  *int64 `json:"archived_at"`
}

func (c *OpenAIClient) CreateProject(ctx context.Context, name string) (*Project, error) {
	data, _, err := c.do(ctx, http.MethodPost, "/v1/organization/projects", map[string]string{"name": name})
	if err != nil {
		return nil, err
	}
	var p Project
	return &p, json.Unmarshal(data, &p)
}

func (c *OpenAIClient) GetProject(ctx context.Context, id string) (*Project, error) {
	data, _, err := c.do(ctx, http.MethodGet, "/v1/organization/projects/"+id, nil)
	if err != nil {
		return nil, err
	}
	var p Project
	return &p, json.Unmarshal(data, &p)
}

func (c *OpenAIClient) UpdateProject(ctx context.Context, id, name string) (*Project, error) {
	data, _, err := c.do(ctx, http.MethodPost, "/v1/organization/projects/"+id, map[string]string{"name": name})
	if err != nil {
		return nil, err
	}
	var p Project
	return &p, json.Unmarshal(data, &p)
}

func (c *OpenAIClient) ArchiveProject(ctx context.Context, id string) error {
	_, _, err := c.do(ctx, http.MethodPost, "/v1/organization/projects/"+id+"/archive", nil)
	return err
}

// ── API Key ────────────────────────────────────────────────────────────────

type APIKey struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	RedactedValue string `json:"redacted_value"`
	Value      *string `json:"value,omitempty"` // only on creation
	CreatedAt  int64   `json:"created_at"`
	Owner      struct {
		Type string `json:"type"`
	} `json:"owner"`
}

func (c *OpenAIClient) CreateAPIKey(ctx context.Context, projectID, name string) (*APIKey, error) {
	path := fmt.Sprintf("/v1/organization/projects/%s/api_keys", projectID)
	data, _, err := c.do(ctx, http.MethodPost, path, map[string]string{"name": name})
	if err != nil {
		return nil, err
	}
	var k APIKey
	return &k, json.Unmarshal(data, &k)
}

func (c *OpenAIClient) GetAPIKey(ctx context.Context, projectID, keyID string) (*APIKey, error) {
	path := fmt.Sprintf("/v1/organization/projects/%s/api_keys/%s", projectID, keyID)
	data, _, err := c.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	var k APIKey
	return &k, json.Unmarshal(data, &k)
}

func (c *OpenAIClient) DeleteAPIKey(ctx context.Context, projectID, keyID string) error {
	path := fmt.Sprintf("/v1/organization/projects/%s/api_keys/%s", projectID, keyID)
	_, _, err := c.do(ctx, http.MethodDelete, path, nil)
	return err
}

// ── Service Account ────────────────────────────────────────────────────────

type ServiceAccount struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Role      string `json:"role"`
	CreatedAt int64  `json:"created_at"`
}

func (c *OpenAIClient) CreateServiceAccount(ctx context.Context, projectID, name string) (*ServiceAccount, error) {
	path := fmt.Sprintf("/v1/organization/projects/%s/service_accounts", projectID)
	data, _, err := c.do(ctx, http.MethodPost, path, map[string]string{"name": name})
	if err != nil {
		return nil, err
	}
	var sa ServiceAccount
	return &sa, json.Unmarshal(data, &sa)
}

func (c *OpenAIClient) GetServiceAccount(ctx context.Context, projectID, saID string) (*ServiceAccount, error) {
	path := fmt.Sprintf("/v1/organization/projects/%s/service_accounts/%s", projectID, saID)
	data, _, err := c.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	var sa ServiceAccount
	return &sa, json.Unmarshal(data, &sa)
}

func (c *OpenAIClient) DeleteServiceAccount(ctx context.Context, projectID, saID string) error {
	path := fmt.Sprintf("/v1/organization/projects/%s/service_accounts/%s", projectID, saID)
	_, _, err := c.do(ctx, http.MethodDelete, path, nil)
	return err
}
