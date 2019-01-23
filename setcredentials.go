package credhub

import (
	"bytes"
	"encoding/json"
	"net/http"
)

// Set adds a credential in Credhub.
func (c *Client) Set(credential Credential, mode OverwriteMode, additionalPermissions []Permission) (*Credential, error) {
	reqBody := struct {
		Credential
		Mode                  OverwriteMode `json:"mode,omitempty"`
		AdditionalPermissions []Permission  `json:"additional_permissions,omitempty"`
	}{
		Credential: credential,
	}

	if c.IsV1API() {
		reqBody.Mode = mode
		reqBody.AdditionalPermissions = additionalPermissions
	} else {
		c.Log.Println("[WARNING] mode is ignored on v2 credhub servers and will be ignored")
		c.Log.Println("[WARNING] additional_permissions can not be set directly on v2 servers, please use AddPermissions instead")
	}

	// an error can't occur since everything being marshalled is valid according to
	// the encoding/json spec
	buf, _ := json.Marshal(reqBody)

	var req *http.Request
	req, err := http.NewRequest("PUT", c.url+"/api/v1/data", bytes.NewBuffer(buf))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")

	resp, err := c.hc.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	cred := new(Credential)
	unmarshaller := json.NewDecoder(resp.Body)
	err = unmarshaller.Decode(&cred)

	return cred, err
}
