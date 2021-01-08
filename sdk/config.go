package sdk

import (
	"time"

	"github.com/marmotedu/medu-sdk-go/sdk/log"
)

type Config struct {
	Scheme   string
	Endpoint string
	Timeout  time.Duration
	LogLevel log.Level
}

var defaultEndpoint = "www.medu-api.com"

// NewConfig returns a pointer of Config
// scheme only accepts http or https
// endpoint is the host to access, the connection could not be created if it's error
func NewConfig() *Config {
	return &Config{
		Scheme:   SchemeHTTP,
		Timeout:  30 * time.Second,
		LogLevel: log.WarnLevel,
	}
}

func (c *Config) WithScheme(scheme string) *Config {
	c.Scheme = scheme
	return c
}

func (c *Config) WithEndpoint(endpoint string) *Config {
	c.Endpoint = endpoint
	return c
}

func (c *Config) WithTimeout(timeout time.Duration) *Config {
	c.Timeout = timeout
	return c
}

func (c *Config) WithLogLevel(level log.Level) *Config {
	c.LogLevel = level
	return c
}
