package http_client

import (
	"net/http"
	"time"
)

type ClientConfig struct {
	Timeout time.Duration
}

func DefaultConfig() ClientConfig {
	return ClientConfig{
		Timeout: 5 * time.Second,
	}
}

func NewClient(cfg ClientConfig) *http.Client {
	return &http.Client{
		Timeout: cfg.Timeout,
	}
}
