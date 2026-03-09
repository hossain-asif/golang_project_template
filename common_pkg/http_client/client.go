package http_client

import (
	"time"

	"github.com/gojek/heimdall/v7"
	"github.com/gojek/heimdall/v7/hystrix"
)

type ClientConfig struct {
	BaseURL                string
	Timeout                time.Duration
	RetryCount             int
	ErrorPercentThreshold  int
	MaxConcurrentRequests  int
	SleepWindow            int
	RequestVolumeThreshold int
}

func DefaultConfig(baseURL string) ClientConfig {
	return ClientConfig{
		BaseURL:                baseURL,
		Timeout:                5 * time.Second,
		RetryCount:             2,
		ErrorPercentThreshold:  25,
		MaxConcurrentRequests:  50,
		SleepWindow:            1000,
		RequestVolumeThreshold: 10,
	}
}

func NewHystrixClient(commandName string, cfg ClientConfig) *hystrix.Client {
	return hystrix.NewClient(
		hystrix.WithHTTPTimeout(cfg.Timeout),
		hystrix.WithCommandName(commandName),
		hystrix.WithHystrixTimeout(cfg.Timeout),
		hystrix.WithMaxConcurrentRequests(cfg.MaxConcurrentRequests),
		hystrix.WithErrorPercentThreshold(cfg.ErrorPercentThreshold),
		hystrix.WithSleepWindow(cfg.SleepWindow),
		hystrix.WithRequestVolumeThreshold(cfg.RequestVolumeThreshold),
		hystrix.WithRetryCount(cfg.RetryCount),
		hystrix.WithRetrier(heimdall.NewRetrier(
			heimdall.NewExponentialBackoff(
				50*time.Millisecond,
				500*time.Millisecond,
				2,
				50*time.Millisecond,
			),
		)),
	)
}
