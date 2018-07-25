package credhub

import (
	"bytes"
	"encoding/json"
	"net/http"
)

// AddPermissions adds permissions to a credential. Note that this method is *not* idempotent.
func (c *Client) AddPermissions(credentialName string, newPerms []Permission) ([]Permission, error) {
	type permbody struct {
		Name        string       `json:"credential_name"`
		Permissions []Permission `json:"permissions"`
	}

	request := permbody{
		Name:        credentialName,
		Permissions: newPerms,
	}

	body, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	var req *http.Request
	req, err = http.NewRequest("POST", c.url+"/api/v1/permissions", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")

	resp, err := c.hc.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var response permbody

	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&response)
	if err != nil {
		return nil, err
	}

	return response.Permissions, nil
}
