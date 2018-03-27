package auth_test

import (
	"context"
	"strings"
	"testing"

	"net/http"
	"net/url"

	"github.com/jghiloni/credhub-api/auth"
	apitest "github.com/jghiloni/credhub-api/internal/testing"
	"golang.org/x/oauth2/clientcredentials"

	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestOAuthClient(t *testing.T) {
	spec.Run(t, "OAuth2 Client", func(t *testing.T, when spec.G, it spec.S) {
		cs := apitest.MockCredhubServer()
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
	ctx := context.Background()
	endpoint, err := auth.UAAEndpoint(cu, true)
	if err != nil {
		return nil, err
	}

	cfg := &clientcredentials.Config{
		ClientID:     ci,
		ClientSecret: cs,
		TokenURL:     endpoint.TokenURL,
		Scopes:       []string{"credhub.read", "credhub.write"},
	}

	return cfg.Client(ctx), nil
}
