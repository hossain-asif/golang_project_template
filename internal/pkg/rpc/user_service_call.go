package rpc

import (
	"encoding/json"
	"fmt"
	"go_project_structure/common_pkg/http_client"
	"net/http"

	"github.com/gojek/heimdall/v7/hystrix"
)
type UserServiceClient struct {
    baseURL string
    client  *hystrix.Client
}

func NewUserServiceClient(cfg http_client.ClientConfig) *UserServiceClient {
    return &UserServiceClient{
        baseURL: cfg.BaseURL,
        client:  http_client.NewHystrixClient("user-service", cfg),
    }
}

func (c *UserServiceClient) GetUser(id int) (*User, error) {
    url := fmt.Sprintf("%s/api/users/%d", c.baseURL, id)

    res, err := c.client.Get(url, http.Header{})
    if err != nil {
        return nil, fmt.Errorf("GetUser failed: %w", err)
    }
    defer res.Body.Close()

    var user User
    if err := json.NewDecoder(res.Body).Decode(&user); err != nil {
        return nil, fmt.Errorf("GetUser decode failed: %w", err)
    }

    return &user, nil
}