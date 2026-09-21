package example


import (
	"context"
	"encoding/json"
	"fmt"
	"go_project_structure/common_pkg/http_client"
	"net/http"
)

type Client struct {
	baseURL string
	client  *http.Client
}

func NewClient(baseURL string, cfg http_client.ClientConfig) *Client {
	return &Client{
		baseURL: baseURL,
		client:  http_client.NewClient(cfg),
	}
}

func (c *Client) GetExample(ctx context.Context, id int) (*GetExampleResponse, error) {
	url := fmt.Sprintf("%s/api/examples/%d", c.baseURL, id)

	req, err := http.NewRequestWithContext(
        ctx,
		http.MethodGet,
		url,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GetExample request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"GetExample returned unexpected status: %s",
			resp.Status,
		)
	}

	var example GetExampleResponse
	if err := json.NewDecoder(resp.Body).Decode(&example); err != nil {
		return nil, fmt.Errorf("failed to decode response body: %w", err)
	}

	return &example, nil
}
