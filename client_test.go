package credhub_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"

	credhub "github.com/cloudfoundry-community/go-credhub"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/matchers"
)

type authRoundTripper struct {
	orig http.RoundTripper
}

func (a *authRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	r.Header.Add("authorization", "bearer abcd")
	return a.orig.RoundTrip(r)
}

type unauthorizedRoundTripper struct{}

func (u *unauthorizedRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	resp := httptest.NewRecorder()
	if r.URL.Path == "/version" {
		resp.WriteString(`{"version" : "1.9.1"}`)
	} else {
		resp.WriteHeader(http.StatusUnauthorized)
	}
	return resp.Result(), nil
}

type errorRoundTripper struct {
}

func (e *errorRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	if r.URL.Path == "/version" {
		w := httptest.NewRecorder()
		w.WriteString(`{"version" : "1.9.1"}`)
		return w.Result(), nil
	}
	return nil, errors.New("test-error")
}

func getAuthenticatedClient(hc *http.Client) *http.Client {
	tr := &authRoundTripper{
		orig: hc.Transport,
	}

	hc.Transport = tr
	return hc
}

func TestInvalidValueTypeConversion(t *testing.T) {
	spec.Run(t, "InvalidValueTypeConversion", testInvalidValueTypeConversion, spec.Report(report.Terminal{}))
}

func testInvalidValueTypeConversion(t *testing.T, when spec.G, it spec.S) {
	var (
		server   *httptest.Server
		chClient *credhub.Client
	)

	it.Before(func() {
		var err error
		RegisterTestingT(t)
		server = mockCredhubServer()
		chClient, err = credhub.New(server.URL, getAuthenticatedClient(server.Client()))
		Expect(err).NotTo(HaveOccurred())
	})

	it.After(func() {
		server.Close()
	})

	when("converting fetched credentials to the wrong value types", func() {
		it("fails", func() {
			var (
				cred *credhub.Credential
				err  error
			)

			cred, err = chClient.GetLatestByName("/concourse/common/sample-rsa")
			Expect(err).NotTo(HaveOccurred())
			_, err = credhub.SSHValue(*cred)
			Expect(err).To(HaveOccurred())

			cred, err = chClient.GetLatestByName("/concourse/common/sample-ssh")
			Expect(err).NotTo(HaveOccurred())
			_, err = credhub.UserValue(*cred)
			Expect(err).To(HaveOccurred())

			cred, err = chClient.GetLatestByName("/concourse/common/sample-user")
			Expect(err).NotTo(HaveOccurred())
			_, err = credhub.CertificateValue(*cred)
			Expect(err).To(HaveOccurred())

			cred, err = chClient.GetLatestByName("/concourse/common/sample-certificate")
			Expect(err).NotTo(HaveOccurred())
			_, err = credhub.RSAValue(*cred)
			Expect(err).To(HaveOccurred())
		})
	})

	when("getting the value from an invalid credential", func() {
		var cred credhub.Credential

		it.Before(func() {
			cred = credhub.Credential{
				Name:    "/test",
				ID:      "1234",
				Created: "today",
				Value: map[float32]float32{
					8.67: 53.09,
				},
			}
		})

		when("converting to user type", func() {
			it("fails", func() {
				cred.Type = credhub.User
				v, err := credhub.UserValue(cred)
				Expect(err).To(HaveOccurred())
				Expect(v).To(BeZero())
			})
		})

		when("converting to rsa type", func() {
			it("fails", func() {
				cred.Type = credhub.RSA
				v, err := credhub.RSAValue(cred)
				Expect(err).To(HaveOccurred())
				Expect(v).To(BeZero())
			})
		})

		when("converting to ssh type", func() {
			it("fails", func() {
				cred.Type = credhub.SSH
				v, err := credhub.SSHValue(cred)
				Expect(err).To(HaveOccurred())
				Expect(v).To(BeZero())
			})
		})

		when("converting to certificate type", func() {
			it("fails", func() {
				cred.Type = credhub.Certificate
				v, err := credhub.CertificateValue(cred)
				Expect(err).To(HaveOccurred())
				Expect(v).To(BeZero())
			})
		})
	})

	when("getting the value from a cred whose type and value don't match", func() {
		var cred credhub.Credential

		it.Before(func() {
			cred = credhub.Credential{
				Name:    "/test",
				ID:      "1234",
				Created: "today",
			}
		})

		when("converting to user type", func() {
			it("fails", func() {
				cred.Type = credhub.User
				cred.Value = map[string]interface{}{
					"username": "foo",
					"extra":    "bad",
				}
				v, err := credhub.UserValue(cred)
				Expect(err).To(HaveOccurred())
				Expect(v).To(BeZero())
			})
		})

		when("converting to rsa type", func() {
			it("fails", func() {
				cred.Type = credhub.RSA
				cred.Value = map[string]interface{}{
					"public_key": "foo",
					"extra":      "bad",
				}
				v, err := credhub.RSAValue(cred)
				Expect(err).To(HaveOccurred())
				Expect(v).To(BeZero())
			})
		})

		when("converting to ssh type", func() {
			it("fails", func() {
				cred.Type = credhub.SSH
				cred.Value = map[string]interface{}{
					"public_key": "foo",
					"extra":      "bad",
				}
				v, err := credhub.SSHValue(cred)
				Expect(err).To(HaveOccurred())
				Expect(v).To(BeZero())
			})
		})

		when("converting to certificate type", func() {
			it("fails", func() {
				cred.Type = credhub.Certificate
				cred.Value = map[string]interface{}{
					"certificate": "foo",
					"extra":       "bad",
				}
				v, err := credhub.CertificateValue(cred)
				Expect(err).To(HaveOccurred())
				Expect(v).To(BeZero())
			})
		})
	})
}

func TestCheckVersion(t *testing.T) {
	v1Server := versionServer(true)
	defer v1Server.Close()

	cli, err := credhub.New(v1Server.URL, v1Server.Client())
	if err != nil {
		t.Fatal(err)
	}
	v := cli.IsV1API()
	if !v {
		t.Fatal("Expected IsV1API to be true, got false")
	}

	v2Server := versionServer(false)
	defer v2Server.Close()

	cli2, err := credhub.New(v2Server.URL, v2Server.Client())
	if err != nil {
		t.Fatal(err)
	}
	v = cli2.IsV1API()
	if v {
		t.Fatal("Expected IsV1API to be false, got true")
	}
}

func vcapServicesDeepEnoughEquals(a, b string) bool {
	var err error

	actual := new(map[string][]map[string]interface{})
	expected := new(map[string][]map[string]interface{})

	if err = json.Unmarshal([]byte(a), actual); err != nil {
		return false
	}

	if err = json.Unmarshal([]byte(b), expected); err != nil {
		return false
	}

	if err = normalizeCredentials(actual); err != nil {
		return false
	}

	if err = normalizeCredentials(expected); err != nil {
		return false
	}

	matcher := &BeEquivalentToMatcher{
		Expected: *expected,
	}

	equal, err := matcher.Match(*actual)
	return equal && err == nil
}

func normalizeCredentials(vcap *map[string][]map[string]interface{}) error {
	for serviceType := range *vcap {
		for i := range (*vcap)[serviceType] {
			if _, ok := (*vcap)[serviceType][i]["credentials"]; ok {
				(*vcap)[serviceType][i]["credentials"] = "TEST-NORMALIZATION"
			}
		}
	}

	return nil
}

func versionServer(isV1 bool) *httptest.Server {
	return httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var v string
		if isV1 {
			v = "1.9.1"
		} else {
			v = "2.0.2"
		}

		fmt.Fprintf(w, `{"version": %q}`, v)
	}))
}
