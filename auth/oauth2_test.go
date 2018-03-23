package auth_test

import (
	"strings"
	"testing"

	"net/url"

	"github.com/jghiloni/credhub-api/auth"
	sdktest "github.com/jghiloni/credhub-api/testing"

	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestOAuthClient(t *testing.T) {
	spec.Run(t, "OAuth2 Client", func(t *testing.T, when spec.G, it spec.S) {
		cs := sdktest.MockCredhubServer()
		it.Before(func() {
			RegisterTestingT(t)
		})

		it("should work if credentials are correct", func() {
			client, err := auth.NewOAuthClient(cs.URL, "user", "pass", false)
			Expect(client).To(Not(BeNil()))
			Expect(err).To(BeNil())

			r, err2 := client.Get(cs.URL + "/some-url")
			Expect(r).To(Not(BeNil()))
			Expect(r.StatusCode).To(Equal(200))
			Expect(err2).To(BeNil())
		})

		it("should not work if credentials are incorrect", func() {
			client, err := auth.NewOAuthClient(cs.URL, "baduser", "badpass", false)
			Expect(client).To(Not(BeNil()))
			Expect(err).To(BeNil())

			r, err2 := client.Get(cs.URL + "/some-url")
			Expect(r).To(BeNil())
			urlerr, ok := err2.(*url.Error)
			Expect(ok).To(BeTrue())
			Expect(urlerr).To(Not(BeNil()))
			Expect(strings.HasSuffix(urlerr.Error(), "401 Unauthorized"))
		})
	}, spec.Report(report.Terminal{}))
}
