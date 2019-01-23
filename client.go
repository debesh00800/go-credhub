package credhub

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// Client interacts with the Credhub API. It provides methods for all available
// endpoints
type Client struct {
	url  string
	hc   HTTPClient
	isV1 bool
}

// New creates a new Credhub client. You must bring an *http.Client that will
// negotiate authentication and authorization for you. See the examples for more
// information.
func New(credhubURL string, hc HTTPClient) (*Client, error) {
	c := &Client{
		url: credhubURL,
		hc:  hc,
	}

	err := c.setVersion()
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Client) setVersion() error {
	resp, err := c.hc.Get(fmt.Sprintf("%s/version", c.url))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Expected 200 OK, got %d %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	var body map[string]string
	err = json.NewDecoder(resp.Body).Decode(&body)
	if err != nil {
		return err
	}

	c.isV1 = strings.HasPrefix(body["version"], "1.")

	return nil
}

// IsV1API returns true if the credhub API is version 1.x
func (c *Client) IsV1API() bool {
	return c.isV1
}
