package credhub

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strings"
)

// Client interacts with the Credhub API. It provides methods for all available
// endpoints
type Client struct {
	url string
	hc  *http.Client
}

// New creates a new Credhub client. You must bring an *http.Client that will
// negotiate authentication and authorization for you. See the examples for more
// information.
func New(credhubURL string, hc *http.Client) *Client {
	return &Client{
		url: credhubURL,
		hc:  hc,
	}
}

// ListAllPaths lists all paths that have credentials that have that prefix.
// Use in conjunction with FindByPath() to list all credentials
func (c *Client) ListAllPaths() ([]string, error) {
	var retBody struct {
		Paths []struct {
			Path string `json:"path"`
		} `json:"paths"`
	}

	resp, err := c.hc.Get(c.url + "/api/v1/data?paths=true")
	if err != nil {
		return nil, err
	}

	marshaller := json.NewDecoder(resp.Body)

	if err = marshaller.Decode(&retBody); err != nil {
		return nil, err
	}

	paths := make([]string, 0, len(retBody.Paths))
	for _, path := range retBody.Paths {
		paths = append(paths, path.Path)
	}

	return paths, nil
}

// GetByID will look up a credental by its ID. Since each version of a named
// credential has a different ID, this will always return at most one value.
func (c *Client) GetByID(id string) (Credential, error) {
	var cred Credential

	resp, err := c.hc.Get(c.url + "/api/v1/data/" + id)
	if err != nil {
		return cred, err
	}

	if resp.StatusCode == 404 {
		return cred, errors.New("credential not found")
	}

	marshaller := json.NewDecoder(resp.Body)

	if err = marshaller.Decode(&cred); err != nil {
		return cred, err
	}

	return cred, nil
}

// GetAllByName will return all versions of a credential, sorted in descending
// order by their created date.
func (c *Client) GetAllByName(name string) ([]Credential, error) {
	return c.getByName(name, false, -1)
}

// GetVersionsByName will return the latest numVersions versions of a given
// credential, still sorted in descending order by their created date.
func (c *Client) GetVersionsByName(name string, numVersions int) ([]Credential, error) {
	return c.getByName(name, false, numVersions)
}

// GetLatestByName will return the current version of a credential. It will return
// at most one item.
func (c *Client) GetLatestByName(name string) (Credential, error) {
	creds, err := c.getByName(name, true, -1)
	if err != nil {
		return Credential{}, err
	}

	return creds[0], nil
}

// Set adds a credential in Credhub.
func (c *Client) Set(credential Credential, mode OverwriteMode, additionalPermissions []Permission) (Credential, error) {
	reqBody := struct {
		Credential
		Mode                  OverwriteMode `json:"mode"`
		AdditionalPermissions []Permission  `json:"additional_permissions,omitempty"`
	}{
		Credential: credential,
		Mode:       mode,
		AdditionalPermissions: additionalPermissions,
	}
	buf, err := json.Marshal(reqBody)
	if err != nil {
		return Credential{}, err
	}

	var req *http.Request
	req, err = http.NewRequest("PUT", c.url+"/api/v1/data", bytes.NewBuffer(buf))
	if err != nil {
		return Credential{}, err
	}

	req.Header.Add("Content-Type", "application/json")

	resp, err := c.hc.Do(req)
	if err != nil {
		return Credential{}, err
	}

	var cred Credential
	unmarshaller := json.NewDecoder(resp.Body)
	err = unmarshaller.Decode(&cred)

	return cred, err
}

// Generate will create a credential in Credhub. Currently does not work for the
// Value or JSON credential types. See https://credhub-api.cfapps.io/#generate-credentials
// for more information about available parameters.
func (c *Client) Generate(name string, credentialType CredentialType, parameters map[string]interface{}) (Credential, error) {
	reqBody := make(map[string]interface{})
	reqBody["name"] = name
	reqBody["type"] = credentialType
	reqBody["parameters"] = parameters

	buf, err := json.Marshal(reqBody)
	if err != nil {
		return Credential{}, err
	}

	var req *http.Request
	req, err = http.NewRequest("POST", c.url+"/api/v1/data", bytes.NewBuffer(buf))
	if err != nil {
		return Credential{}, err
	}

	req.Header.Add("Content-Type", "application/json")

	resp, err := c.hc.Do(req)
	if err != nil {
		return Credential{}, err
	}

	var cred Credential
	unmarshaller := json.NewDecoder(resp.Body)
	err = unmarshaller.Decode(&cred)

	return cred, err
}

// Regenerate will generate new values for credentials using the same parameters
// as the stored value. All RSA and SSH credentials may be regenerated. Password
// and user credentials must have been generated to enable regeneration.
// Statically set certificates may be regenerated if they are self-signed or if
// the CA name has been set to a stored CA certificate.
func (c *Client) Regenerate(name string) (Credential, error) {
	reqBody := struct {
		Name string `json:"name"`
	}{
		Name: name,
	}

	buf, err := json.Marshal(reqBody)
	if err != nil {
		return Credential{}, err
	}

	var req *http.Request
	req, err = http.NewRequest("POST", c.url+"/api/v1/data/regenerate", bytes.NewBuffer(buf))
	if err != nil {
		return Credential{}, err
	}

	req.Header.Add("Content-Type", "application/json")

	resp, err := c.hc.Do(req)
	if err != nil {
		return Credential{}, err
	}

	var cred Credential
	unmarshaller := json.NewDecoder(resp.Body)
	err = unmarshaller.Decode(&cred)

	return cred, err
}

// Delete deletes a credential by name
func (c *Client) Delete(name string) error {
	chURL := c.url + "/api/v1/data?name=" + name
	req, err := http.NewRequest("DELETE", chURL, nil)
	if err != nil {
		return err
	}

	resp, err := c.hc.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != 204 {
		return fmt.Errorf("expected return code 204, got %d", resp.StatusCode)
	}

	return nil
}

// FindByPath retrieves a list of stored credential names which are within the
// specified path. This method does not traverse sub-paths.
func (c *Client) FindByPath(path string) ([]Credential, error) {
	var retBody struct {
		Credentials []Credential `json:"credentials"`
	}

	resp, err := c.hc.Get(c.url + "/api/v1/data?path=" + path)
	if err != nil {
		return nil, err
	}

	buf, _ := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(buf, &retBody)
	return retBody.Credentials, err
}

// FindByPartialName retrieves a list of stored credential names which contain the search.
func (c *Client) FindByPartialName(partialName string) ([]Credential, error) {
	var retBody struct {
		Credentials []Credential `json:"credentials"`
	}

	resp, err := c.hc.Get(c.url + "/api/v1/data?name-like=" + partialName)
	if err != nil {
		return nil, err
	}

	marshaller := json.NewDecoder(resp.Body)

	err = marshaller.Decode(&retBody)
	return retBody.Credentials, err
}

// GetPermissions returns the permissions of a credential. Permissions consist of
// an actor (See https://github.com/cloudfoundry-incubator/credhub/blob/master/docs/authentication-identities.md
// for more information on actor identities) and Operations
func (c *Client) GetPermissions(credentialName string) ([]Permission, error) {
	params := make(url.Values)
	params.Add("credential_name", credentialName)

	resp, err := c.hc.Get(c.url + "/api/v1/permissions?" + params.Encode())
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == 404 {
		return nil, errors.New("credential not found")
	}

	retBody := struct {
		CN          string       `json:"credential_name"`
		Permissions []Permission `json:"permissions"`
	}{}

	marshaller := json.NewDecoder(resp.Body)

	err = marshaller.Decode(&retBody)
	return retBody.Permissions, err
}

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

// InterpolateCredentials will take a string representation of a VCAP_SERVICES
// json variable, and interpolate any services whose credentials block consists
// only of credhub-ref. It will return the interpolated JSON as a string
func (c *Client) InterpolateCredentials(vcapServices string) (string, error) {
	var err error

	type vcapService map[string]interface{}

	services := make(map[string][]vcapService)
	if err = json.Unmarshal([]byte(vcapServices), &services); err != nil {
		return "", err
	}

	for serviceType := range services {
		for i := range services[serviceType] {
			credRefIntf := services[serviceType][i]["credentials"]
			credRef, ok := credRefIntf.(map[string]interface{})
			if ok && len(credRef) == 1 {
				ref, ok := credRef["credhub-ref"]
				if ok {
					var resolvedCreds []Credential
					credName := ref.(string)
					resolvedCreds, err = c.getByName(credName, true, 1)
					if err != nil {
						return "", err
					}

					services[serviceType][i]["credentials"] = resolvedCreds[0].Value
				}
			}
		}
	}

	output, err := json.Marshal(services)
	if err != nil {
		return "", err
	}

	return string(output), nil

}

func (c *Client) getByName(name string, latest bool, numVersions int) ([]Credential, error) {
	var retBody struct {
		Data []Credential `json:"data"`
	}

	chURL := c.url + "/api/v1/data?"

	params := url.Values{}
	params.Add("name", name)

	if latest {
		params.Add("current", "true")
	}

	if numVersions > 0 {
		params.Add("versions", fmt.Sprint(numVersions))
	}

	chURL += params.Encode()
	resp, err := c.hc.Get(chURL)
	if err != nil {
		return retBody.Data, err
	}

	if resp.StatusCode == 404 {
		return nil, errors.New("Name Not Found")
	}

	marshaller := json.NewDecoder(resp.Body)

	err = marshaller.Decode(&retBody)
	if err != nil {
		return nil, err
	}

	data := retBody.Data
	sort.Slice(data, func(i, j int) bool {
		less := strings.Compare(data[i].Created, data[j].Created)
		// we want to sort in reverse order, so return the opposite of what you'd normally do
		return less > 0
	})

	return retBody.Data, err
}
