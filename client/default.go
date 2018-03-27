package client

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

	"github.com/jghiloni/credhub-api/auth"
)

type client struct {
	url string
	hc  *http.Client
}

var errNotImpl = errors.New("unimplemented")

// New creates a new Credhub client with OAuth2 authentication
func New(credhubURL, clientID, clientSecret string, skipTLSVerify bool) (Credhub, error) {
	var cli *http.Client
	var err error
	if cli, err = auth.NewOAuthClient(credhubURL, clientID, clientSecret, skipTLSVerify); err != nil {
		return nil, err
	}

	return &client{
		url: credhubURL,
		hc:  cli,
	}, nil
}

func (c *client) ListAllPaths() ([]string, error) {
	var retBody struct {
		Paths []struct {
			Path string `json:"path"`
		} `json:"paths"`
	}

	resp, err := c.hc.Get(c.url + "/api/v1/data?paths=true")
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == 404 {
		return nil, errors.New("path not found")
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

func (c *client) GetByID(id string) (Credential, error) {
	var cred Credential

	resp, err := c.hc.Get(c.url + "/api/v1/data/" + id)
	if err != nil {
		return cred, err
	}

	if resp.StatusCode == 404 {
		return cred, errors.New("path not found")
	}

	marshaller := json.NewDecoder(resp.Body)

	if err = marshaller.Decode(&cred); err != nil {
		return cred, err
	}

	return cred, nil
}

func (c *client) GetAllByName(name string) ([]Credential, error) {
	return c.getByName(name, false, -1)
}

func (c *client) GetVersionsByName(name string, numVersions int) ([]Credential, error) {
	return c.getByName(name, false, numVersions)
}

func (c *client) GetLatestByName(name string) (Credential, error) {
	creds, err := c.getByName(name, true, -1)
	if err != nil {
		return Credential{}, err
	}

	return creds[0], nil
}

func (c *client) Set(credential Credential) (Credential, error) {
	buf, err := json.Marshal(credential)
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

func (c *client) Generate(name string, credentialType CredentialType, parameters map[string]interface{}) (Credential, error) {
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

func (c *client) Regenerate(name string) (Credential, error) {
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

func (c *client) Delete(name string) error { return errNotImpl }

func (c *client) FindByPath(path string) ([]Credential, error) {
	var retBody struct {
		Credentials []Credential `json:"credentials"`
	}

	resp, err := c.hc.Get(c.url + "/api/v1/data?path=" + path)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == 404 {
		return nil, errors.New("path not found")
	}

	buf, _ := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(buf, &retBody)
	return retBody.Credentials, err
}

func (c *client) FindByPartialName(partialName string) ([]Credential, error) {
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

func (c *client) GetPermissions(credentialName string) ([]Permission, error) {
	return nil, errNotImpl
}

func (c *client) AddPermissions(credentialName string, newPerms []Permission) ([]Permission, error) {
	return nil, errNotImpl
}

func (c *client) DeletePermissions(credentialName, actorID string) error {
	return errNotImpl
}

func (c *client) getByName(name string, latest bool, numVersions int) ([]Credential, error) {
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
