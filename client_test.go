package credhub_test

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"net/http"
	"os"
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"

	. "github.com/jghiloni/credhub-api"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/matchers"
)

func TestCredhubClient(t *testing.T) {
	spec.Run(t, "Credhub Client", func(t *testing.T, when spec.G, it spec.S) {
		server := mockCredhubServer()

		it.Before(func() {
			RegisterTestingT(t)
		})

		it.After(func() {
			server.Close()
		})

		getClient := func(ci, cs string, skip bool) *Client {
			endpoint, _ := UAAEndpoint(server.URL, true)
			var t *http.Transport
			if skip {
				t = &http.Transport{
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
				}
			} else {
				certs, _ := x509.SystemCertPool()
				certs.AddCert(server.Certificate())
				t = &http.Transport{
					TLSClientConfig: &tls.Config{RootCAs: certs},
				}
			}

			t.Proxy = http.ProxyFromEnvironment
			sslcli := &http.Client{Transport: t}

			ctx := context.WithValue(context.TODO(), oauth2.HTTPClient, sslcli)
			cfg := &clientcredentials.Config{
				ClientID:     ci,
				ClientSecret: cs,
				TokenURL:     endpoint.TokenURL,
				Scopes:       []string{"read", "write"},
			}
			return New(server.URL, cfg.Client(ctx))
		}

		findByGoodPath := func(chClient *Client) func() {
			return func() {
				creds, err := chClient.FindByPath("/concourse/common")
				Expect(err).To(BeNil())
				Expect(len(creds)).To(Equal(3))
			}
		}

		findByBadPath := func(chClient *Client) func() {
			return func() {
				creds, err := chClient.FindByPath("/concourse/uncommon")
				Expect(err).To(HaveOccurred())
				Expect(len(creds)).To(Equal(0))
			}
		}

		valueByNameTests := func(chClient *Client, latest bool, num int) func() {
			if latest {
				return func() {
					cred, err := chClient.GetLatestByName("/concourse/common/sample-value")
					Expect(err).To(Not(HaveOccurred()))
					Expect(cred.Value).To(BeEquivalentTo("sample2"))
				}
			} else if num <= 0 {
				return func() {
					creds, err := chClient.GetAllByName("/concourse/common/sample-value")
					Expect(err).To(Not(HaveOccurred()))
					Expect(len(creds)).To(Equal(3))
				}
			} else {
				return func() {
					creds, err := chClient.GetVersionsByName("/concourse/common/sample-value", num)
					Expect(err).To(Not(HaveOccurred()))
					Expect(len(creds)).To(Equal(num))
				}
			}
		}

		passwordByNameTests := func(chClient *Client, latest bool, num int) func() {
			if latest {
				return func() {
					cred, err := chClient.GetLatestByName("/concourse/common/sample-password")
					Expect(err).To(Not(HaveOccurred()))
					Expect(cred.Value).To(BeEquivalentTo("sample1"))
				}
			} else if num <= 0 {
				return func() {
					creds, err := chClient.GetAllByName("/concourse/common/sample-password")
					Expect(err).To(BeNil())
					Expect(len(creds)).To(Equal(3))
					Expect(creds[2].Value).To(BeEquivalentTo("sample2"))
				}
			} else {
				return func() {
					creds, err := chClient.GetVersionsByName("/concourse/common/sample-value", num)
					Expect(err).To(Not(HaveOccurred()))
					Expect(len(creds)).To(Equal(num))
				}
			}
		}

		jsonByNameTests := func(chClient *Client, latest bool, num int) func() {
			if latest {
				return func() {
					cred, err := chClient.GetLatestByName("/concourse/common/sample-json")
					Expect(err).To(Not(HaveOccurred()))

					rawVal := cred.Value
					val, ok := rawVal.(map[string]interface{})
					Expect(ok).To(BeTrue())

					Expect(val["foo"]).To(BeEquivalentTo("bar"))
				}
			} else if num <= 0 {
				return func() {
					creds, err := chClient.GetAllByName("/concourse/common/sample-json")
					Expect(err).To(Not(HaveOccurred()))
					Expect(len(creds)).To(Equal(3))

					intf := creds[2].Value
					val, ok := intf.([]interface{})
					Expect(ok).To(BeTrue())

					Expect(int(val[0].(float64))).To(Equal(1))
					Expect(int(val[1].(float64))).To(Equal(2))
				}
			} else {
				return func() {
					creds, err := chClient.GetVersionsByName("/concourse/common/sample-value", num)
					Expect(err).To(Not(HaveOccurred()))
					Expect(len(creds)).To(Equal(num))
				}
			}
		}

		getUserByName := func(chClient *Client) func() {
			return func() {
				creds, err := chClient.GetAllByName("/concourse/common/sample-user")
				Expect(err).To(Not(HaveOccurred()))
				Expect(len(creds)).To(Equal(1))

				cred := creds[0]

				var val UserValueType
				val, err = UserValue(cred)
				Expect(err).To(Not(HaveOccurred()))
				Expect(val.Username).To(Equal("me"))
			}
		}

		getSSHByName := func(chClient *Client) func() {
			return func() {
				creds, err := chClient.GetAllByName("/concourse/common/sample-ssh")
				Expect(err).To(Not(HaveOccurred()))
				Expect(len(creds)).To(Equal(1))

				cred := creds[0]
				var val SSHValueType
				val, err = SSHValue(cred)
				Expect(err).To(Not(HaveOccurred()))
				Expect(val.PublicKey).To(HavePrefix("ssh-rsa"))
			}
		}

		getRSAByName := func(chClient *Client) func() {
			return func() {
				creds, err := chClient.GetAllByName("/concourse/common/sample-rsa")
				Expect(err).To(Not(HaveOccurred()))
				Expect(len(creds)).To(Equal(1))

				cred := creds[0]
				var val RSAValueType
				val, err = RSAValue(cred)
				Expect(err).To(Not(HaveOccurred()))
				Expect(val.PrivateKey).To(HavePrefix("-----BEGIN PRIVATE KEY-----"))
			}
		}

		getNonexistentName := func(chClient *Client) func() {
			return func() {
				_, err := chClient.GetAllByName("/concourse/common/not-real")
				Expect(err).To(HaveOccurred())
			}
		}

		getCertificateByName := func(chClient *Client) func() {
			return func() {
				creds, err := chClient.GetAllByName("/concourse/common/sample-certificate")
				Expect(err).To(Not(HaveOccurred()))
				Expect(len(creds)).To(Equal(1))

				cred := creds[0]
				var val CertificateValueType
				val, err = CertificateValue(cred)
				Expect(err).To(Not(HaveOccurred()))
				Expect(val.Certificate).To(HavePrefix("-----BEGIN CERTIFICATE-----"))
			}
		}

		badConversionValueTests := func(chClient *Client) func() {
			return func() {
				var (
					cred Credential
					err  error
				)

				cred, err = chClient.GetLatestByName("/concourse/common/sample-rsa")
				Expect(err).NotTo(HaveOccurred())
				_, err = SSHValue(cred)
				Expect(err).To(HaveOccurred())

				cred, err = chClient.GetLatestByName("/concourse/common/sample-ssh")
				Expect(err).NotTo(HaveOccurred())
				_, err = UserValue(cred)
				Expect(err).To(HaveOccurred())

				cred, err = chClient.GetLatestByName("/concourse/common/sample-user")
				Expect(err).NotTo(HaveOccurred())
				_, err = CertificateValue(cred)
				Expect(err).To(HaveOccurred())

				cred, err = chClient.GetLatestByName("/concourse/common/sample-certificate")
				Expect(err).NotTo(HaveOccurred())
				_, err = RSAValue(cred)
				Expect(err).To(HaveOccurred())
			}
		}

		listAllByPath := func(chClient *Client) func() {
			return func() {
				paths, err := chClient.ListAllPaths()
				Expect(err).To(Not(HaveOccurred()))
				Expect(paths).To(HaveLen(5))
			}
		}

		getById := func(chClient *Client) func() {
			return func() {
				cred, err := chClient.GetByID("1234")
				Expect(err).To(Not(HaveOccurred()))
				Expect(cred.Name).To(BeEquivalentTo("/by-id"))

				badcred, err := chClient.GetByID("4567")
				Expect(err).To(HaveOccurred())
				Expect(badcred.Value).To(BeNil())
			}
		}

		setOverwriteCredential := func(chClient *Client) func() {
			return func() {
				cred := Credential{
					Name: "/sample-set",
					Type: "user",
					Value: UserValueType{
						Username:     "me",
						Password:     "super-secret",
						PasswordHash: "somestring",
					},
				}

				newCred, err := chClient.Set(cred, Overwrite, nil)
				Expect(err).To(Not(HaveOccurred()))
				Expect(newCred.Created).To(Not(BeEmpty()))
				Expect(newCred.ID).To(Not(BeEmpty()))
			}
		}

		setNoOverwriteCredential := func(chClient *Client) func() {
			return func() {
				cred := Credential{
					Name: "/sample-set",
					Type: "user",
					Value: UserValueType{
						Username:     "me",
						Password:     "super-secret",
						PasswordHash: "somestring",
					},
				}

				newCred, err := chClient.Set(cred, NoOverwrite, nil)
				Expect(err).To(Not(HaveOccurred()))
				Expect(newCred.Created).To(Not(BeEmpty()))
				v, err := UserValue(newCred)
				Expect(err).To(Not(HaveOccurred()))
				Expect(v.Password).To(BeEquivalentTo("old"))
				Expect(newCred.ID).To(BeEquivalentTo("6ba7b810-9dad-11d1-80b4-00c04fd430c8"))
			}
		}

		setConvergeCredentialWithoutDifference := func(chClient *Client) func() {
			return func() {
				cred := Credential{
					Name: "/sample-set",
					Type: "user",
					Value: UserValueType{
						Username:     "me",
						Password:     "super-secret",
						PasswordHash: "somestring",
					},
				}

				newCred, err := chClient.Set(cred, Converge, nil)
				Expect(err).To(Not(HaveOccurred()))
				Expect(newCred.Created).To(Not(BeEmpty()))
				Expect(newCred.ID).To(BeEquivalentTo("6ba7b810-9dad-11d1-80b4-00c04fd430c8"))
			}
		}

		setConvergeCredentialWithDifference := func(chClient *Client) func() {
			return func() {
				cred := Credential{
					Name: "/sample-set",
					Type: "user",
					Value: UserValueType{
						Username:     "me",
						Password:     "new-super-secret",
						PasswordHash: "somestring",
					},
				}

				newCred, err := chClient.Set(cred, Converge, nil)
				Expect(err).To(Not(HaveOccurred()))
				Expect(newCred.Created).To(Not(BeEmpty()))
				Expect(newCred.ID).To(Not(BeEquivalentTo("6ba7b810-9dad-11d1-80b4-00c04fd430c8")))
			}
		}

		deleteFoundCredential := func(chClient *Client) func() {
			return func() {
				err := chClient.Delete("/some-cred")
				Expect(err).To(Not(HaveOccurred()))
			}
		}

		deleteNotFoundCredential := func(chClient *Client) func() {
			return func() {
				err := chClient.Delete("/some-other-cred")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(BeEquivalentTo("expected return code 204, got 404"))
			}
		}

		getPermissions := func(chClient *Client) func() {
			return func() {
				perms, err := chClient.GetPermissions("/credential-with-permissions")
				Expect(err).NotTo(HaveOccurred())
				Expect(perms).To(HaveLen(3))

				perms, err = chClient.GetPermissions("/non-existent")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(BeEquivalentTo("credential not found"))
				Expect(perms).To(BeNil())
			}
		}

		modifyPermissions := func(chClient *Client) func() {
			return func() {
				var perms []Permission
				var err error

				perms, err = chClient.GetPermissions("/add-permission-credential")
				Expect(err).NotTo(HaveOccurred())
				Expect(perms).To(HaveLen(0))

				perms = append(perms, Permission{
					Actor:      "uaa-user:1234",
					Operations: []Operation{"read", "write", "delete"},
				})

				respPerms, err := chClient.AddPermissions("/add-permission-credential", perms)
				Expect(err).NotTo(HaveOccurred())
				Expect(respPerms).To(HaveLen(1))
				Expect(respPerms[0].Actor).To(Equal("uaa-user:1234"))

				err = chClient.DeletePermissions("/add-permission-credential", "some-non-existent-actor")
				Expect(err).NotTo(HaveOccurred())
				perms, err = chClient.GetPermissions("/add-permission-credential")
				Expect(err).NotTo(HaveOccurred())
				Expect(perms).To(HaveLen(1))
				Expect(perms[0].Actor).To(Equal("uaa-user:1234"))

				err = chClient.DeletePermissions("/add-permission-credential", "uaa-user:1234")
				Expect(err).NotTo(HaveOccurred())
				perms, err = chClient.GetPermissions("/add-permission-credential")
				Expect(err).NotTo(HaveOccurred())
				Expect(perms).To(HaveLen(0))
			}
		}

		interpolate := func(chClient *Client) func() {
			return func() {
				vcapServices := `
				{
					"p-config-server": [
					{
						"credentials": {
							"credhub-ref": "/service-cred-ref"
						},
						"label": "p-config-server",
						"name": "config-server",
						"plan": "standard",
						"provider": null,
						"syslog_drain_url": null,
						"tags": [
						"configuration",
						"spring-cloud"
						],
						"volume_mounts": []
					}
					]
				}
				`

				cred, err := chClient.GetLatestByName("/service-cred-ref")
				Expect(err).NotTo(HaveOccurred())

				interpolated, err := chClient.InterpolateCredentials(vcapServices)
				Expect(err).NotTo(HaveOccurred())
				Expect(vcapServicesDeepEnoughEquals(vcapServices, interpolated)).To(BeTrue())

				interpolatedObj := make(map[string][]map[string]interface{})
				err = json.Unmarshal([]byte(interpolated), &interpolatedObj)
				Expect(err).NotTo(HaveOccurred())

				resolvedCred := interpolatedObj["p-config-server"][0]["credentials"]
				Expect(resolvedCred).To(BeEquivalentTo(cred.Value))
			}
		}

		runTests := func(chClient *Client) func() {
			return func() {
				when("Testing Find By Path", func() {
					it("should be able to find creds by path", findByGoodPath(chClient))
					it("should not be able to find creds with an unknown path", findByBadPath(chClient))
				})

				when("Testing Get By Name", func() {
					it("should get a 'value' type credential", valueByNameTests(chClient, false, -1))
					it("should get a 'password' type credential", passwordByNameTests(chClient, false, -1))
					it("should get a 'json' type credential", jsonByNameTests(chClient, false, -1))
					it("should get a 'user' type credential", getUserByName(chClient))
					it("should get a 'ssh' type credential", getSSHByName(chClient))
					it("should get a 'rsa' type credential", getRSAByName(chClient))
					it("should get a 'certificate' type credential", getCertificateByName(chClient))
					it("should not get a credential that doesn't exist", getNonexistentName(chClient))
				})

				when("Testing Get Latest By Name", func() {
					it("should get a 'value' type credential", valueByNameTests(chClient, true, -1))
					it("should get a 'password' type credential", passwordByNameTests(chClient, true, -1))
					it("should get a 'json' type credential", jsonByNameTests(chClient, true, -1))
				})

				when("Testing Get Latest By Name", func() {
					it("should get a 'value' type credential", valueByNameTests(chClient, false, 2))
					it("should get a 'password' type credential", passwordByNameTests(chClient, false, 2))
					it("should get a 'json' type credential", jsonByNameTests(chClient, false, 2))
				})

				when("Testing List All Paths", func() {
					it("should list all paths", listAllByPath(chClient))
				})

				when("Testing Get By ID", func() {
					it("should get an item with a valid ID", getById(chClient))
				})

				when("Testing Set Credential", func() {
					it("should receive the same item it sent, but with a timestamp", setOverwriteCredential(chClient))
					it("should receive an old credential", setNoOverwriteCredential(chClient))
					it("should receive an old credential if converging without changes", setConvergeCredentialWithoutDifference(chClient))
					it("should receive a new credential if converging with changes", setConvergeCredentialWithDifference(chClient))
				})

				when("Testing Generate Credential", func() {
					it("should generate a password credential", func() {
						params := make(map[string]interface{})
						params["length"] = 30
						cred, err := chClient.Generate("/example-generated", "password", params)
						Expect(err).To(Not(HaveOccurred()))
						Expect(cred.Type).To(Equal(Password))
						Expect(cred.Value).To(BeAssignableToTypeOf("expected"))
						Expect(cred.Value).To(HaveLen(30))
					})
				})

				when("Testing Regenerate Credential", func() {
					it("should regenerate a password credential", func() {
						cred, err := chClient.Regenerate("/example-password")
						Expect(err).To(Not(HaveOccurred()))
						Expect(cred.Type).To(Equal(Password))
						Expect(cred.Value).To(BeAssignableToTypeOf("expected"))
						Expect(cred.Value).To(BeEquivalentTo("P$<MNBVCXZ;lkjhgfdsa0987654321"))
					})
				})

				when("Testing Find By Name", func() {
					it("should return names with 'password' in them", func() {
						creds, err := chClient.FindByPartialName("password")
						Expect(err).To(Not(HaveOccurred()))
						Expect(creds).To(HaveLen(2))
						for _, cred := range creds {
							Expect(cred.Name).To(ContainSubstring("password"))
						}
					})
				})

				when("Testing Delete", func() {
					it("should delete a credential that it can find", deleteFoundCredential(chClient))
					it("should fail to delete a credential that it cannot find", deleteNotFoundCredential(chClient))
				})

				when("Testing Bad Value Conversions", func() {
					it("should fail every time", badConversionValueTests(chClient))
				})

				when("Testing Get Permissions", func() {
					it("should find permissions for an existing credential", getPermissions(chClient))
				})

				when("Testing Modify Permissions", func() {
					it.After(func() {
						err := os.Remove("testdata/permissions/add-permissions/cred.json")
						Expect(err).NotTo(HaveOccurred())
					})

					it("should allow permissions to be added and deleted", modifyPermissions(chClient))
				})

				when("Testing interpolation", func() {
					it("should have values from credhub in VCAP_SERVICES", interpolate(chClient))
				})
			}
		}

		when("Running with UAA Authorization", func() {
			when("skipping TLS validation", runTests(getClient("user", "pass", true)))
			when("not skipping TLS validation", runTests(getClient("user", "pass", false)))
		})
	}, spec.Report(report.Terminal{}))
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
