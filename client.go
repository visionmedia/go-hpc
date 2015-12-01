package hpc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// Config for the RPC client.
type Config struct {
	HTTPClient *http.Client // Custom HTTP client
	URL        string       // URL end-point for RPC services
}

// Client lets you make calls to an HTTP-RPC end-point.
type Client struct {
	*Config
}

// NewClient returns a new client.
func NewClient(config *Config) *Client {
	return &Client{
		Config: config,
	}
}

// NewConfig returns configuration for the given `url`.
func NewConfig(url string) *Config {
	return &Config{
		URL:        url,
		HTTPClient: http.DefaultClient,
	}
}

// Call a method.
func (c *Client) Call(service, method string, in interface{}, out interface{}) error {
	url := fmt.Sprintf("%s/%s/%s", c.URL, service, method)

	b, err := json.Marshal(in)
	if err != nil {
		return err
	}

	res, err := c.HTTPClient.Post(url, "application/json", bytes.NewBuffer(b))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	return json.NewDecoder(res.Body).Decode(out)
}
