package credhub

import (
	"fmt"
	"net/http"
	"net/url"
)

// DeletePermissions deletes permissions from a credential. Note that this method
// is *not* idempotent
func (c *Client) DeletePermissions(credentialName, actorID string) error {
	chURL := c.url + "/api/v1/permissions"

	req, err := http.NewRequest("DELETE", chURL, nil)
	if err != nil {
		return err
	}

	params := make(url.Values)
	params.Add("credential_name", credentialName)
	params.Add("actor", actorID)
	req.URL.RawQuery = params.Encode()

	resp, err := c.hc.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != 204 {
		return fmt.Errorf("expected return code 204, got %d", resp.StatusCode)
	}

	return nil
}
