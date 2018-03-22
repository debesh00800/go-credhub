package auth

import (
	"crypto/tls"
	"encoding/json"
	"net/http"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

// NewOAuthClient will return a pointer to an http.Client that has an appropriate OAuth2 client credentials
//   token from the UAA server bound to the Credhub Server specified by credhubURL
func NewOAuthClient(credhubURL, clientID, clientSecret string, skipTLSVerify bool) (*http.Client, error) {
	baseClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: skipTLSVerify},
		},
	}

	ctx := context.WithValue(context.TODO(), oauth2.HTTPClient, baseClient)
	config := &clientcredentials.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       []string{"credhub.read", "credhub.write"},
	}

	r, err := baseClient.Get(credhubURL + "/info")
	if err != nil {
		return nil, err
	}

	var body struct {
		AuthServer struct {
			URL string `json:"url"`
		} `json:"auth-server"`
	}
	decoder := json.NewDecoder(r.Body)

	if err = decoder.Decode(&body); err != nil {
		return nil, err
	}

	config.TokenURL = body.AuthServer.URL + "/oauth/token"
	return config.Client(ctx), nil
}
