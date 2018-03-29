package credhub_test

import (
	"context"
	"crypto/tls"
	"strings"
	"testing"

	"net/http"
	"net/url"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"

	credhub "github.com/jghiloni/credhub-api"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestOAuthClient(t *testing.T) {
	spec.Run(t, "OAuth2 Client", func(t *testing.T, when spec.G, it spec.S) {
		cs := mockCredhubServer()
		it.Before(func() {
			RegisterTestingT(t)
		})

		it("should work if credentials are correct", func() {
			client, err := getClient(cs.URL, "user", "pass")
			Expect(client).To(Not(BeNil()))
			Expect(err).To(Not(HaveOccurred()))

			r, err2 := client.Get(cs.URL + "/some-url")
			Expect(r).To(Not(BeNil()))
			Expect(r.StatusCode).To(Equal(200))
			Expect(err2).To(BeNil())
		})

		it("should not work if credentials are incorrect", func() {
			client, err := getClient(cs.URL, "baduser", "badpass")
			Expect(client).To(Not(BeNil()))
			Expect(err).To(Not(HaveOccurred()))

			r, err2 := client.Get(cs.URL + "/some-url")
			Expect(r).To(BeNil())
			urlerr, ok := err2.(*url.Error)
			Expect(ok).To(BeTrue())
			Expect(urlerr).To(Not(BeNil()))
			Expect(strings.HasSuffix(urlerr.Error(), "401 Unauthorized"))
		})
	}, spec.Report(report.Terminal{}))
}

func getClient(cu, ci, cs string) (*http.Client, error) {
	endpoint, err := credhub.UAAEndpoint(cu, true)
	if err != nil {
		return nil, err
	}

	sslcli := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	ctx := context.WithValue(context.TODO(), oauth2.HTTPClient, sslcli)
	cfg := &clientcredentials.Config{
		ClientID:     ci,
		ClientSecret: cs,
		TokenURL:     endpoint.TokenURL,
		Scopes:       []string{"credhub.read", "credhub.write"},
	}

	return cfg.Client(ctx), nil
}
