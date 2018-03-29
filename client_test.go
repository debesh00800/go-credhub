package credhub_test

import (
	"context"
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"

	credhub "github.com/jghiloni/credhub-api"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

func TestCredhubClient(t *testing.T) {
	spec.Run(t, "Credhub Client", func(t *testing.T, when spec.G, it spec.S) {
		var chClient *credhub.Client
		var server *httptest.Server

		findByGoodPath := func() {
			creds, err := chClient.FindByPath("/concourse/common")
			Expect(err).To(BeNil())
			Expect(len(creds)).To(Equal(3))
		}

		findByBadPath := func() {
			creds, err := chClient.FindByPath("/concourse/uncommon")
			Expect(err).To(HaveOccurred())
			Expect(len(creds)).To(Equal(0))
		}

		valueByNameTests := func(latest bool, num int) func() {
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

		passwordByNameTests := func(latest bool, num int) func() {
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

		jsonByNameTests := func(latest bool, num int) func() {
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

		getUserByName := func() {
			creds, err := chClient.GetAllByName("/concourse/common/sample-user")
			Expect(err).To(Not(HaveOccurred()))
			Expect(len(creds)).To(Equal(1))

			cred := creds[0]

			var val credhub.UserValueType
			val, err = credhub.UserValue(cred)
			Expect(err).To(Not(HaveOccurred()))
			Expect(val.Username).To(Equal("me"))
		}

		getSSHByName := func() {
			creds, err := chClient.GetAllByName("/concourse/common/sample-ssh")
			Expect(err).To(Not(HaveOccurred()))
			Expect(len(creds)).To(Equal(1))

			cred := creds[0]
			var val credhub.SSHValueType
			val, err = credhub.SSHValue(cred)
			Expect(err).To(Not(HaveOccurred()))
			Expect(val.PublicKey).To(HavePrefix("ssh-rsa"))
		}

		getRSAByName := func() {
			creds, err := chClient.GetAllByName("/concourse/common/sample-rsa")
			Expect(err).To(Not(HaveOccurred()))
			Expect(len(creds)).To(Equal(1))

			cred := creds[0]
			var val credhub.RSAValueType
			val, err = credhub.RSAValue(cred)
			Expect(err).To(Not(HaveOccurred()))
			Expect(val.PrivateKey).To(HavePrefix("-----BEGIN PRIVATE KEY-----"))
		}

		getNonexistentName := func() {
			_, err := chClient.GetAllByName("/concourse/common/not-real")
			Expect(err).To(HaveOccurred())
		}

		getCertificateByName := func() {
			creds, err := chClient.GetAllByName("/concourse/common/sample-certificate")
			Expect(err).To(Not(HaveOccurred()))
			Expect(len(creds)).To(Equal(1))

			cred := creds[0]
			var val credhub.CertificateValueType
			val, err = credhub.CertificateValue(cred)
			Expect(err).To(Not(HaveOccurred()))
			Expect(val.Certificate).To(HavePrefix("-----BEGIN CERTIFICATE-----"))
		}

		listAllByPath := func() {
			paths, err := chClient.ListAllPaths()
			Expect(err).To(Not(HaveOccurred()))
			Expect(paths).To(HaveLen(5))
		}

		getById := func() {
			cred, err := chClient.GetByID("1234")
			Expect(err).To(Not(HaveOccurred()))
			Expect(cred.Name).To(BeEquivalentTo("/by-id"))
		}

		setOverwriteCredential := func() {
			cred := credhub.Credential{
				Name: "/sample-set",
				Type: "user",
				Value: credhub.UserValueType{
					Username:     "me",
					Password:     "super-secret",
					PasswordHash: "somestring",
				},
			}

			newCred, err := chClient.Set(cred, credhub.Overwrite, nil)
			Expect(err).To(Not(HaveOccurred()))
			Expect(newCred.Created).To(Not(BeEmpty()))
			Expect(newCred.ID).To(Not(BeEmpty()))
		}

		setNoOverwriteCredential := func() {
			cred := credhub.Credential{
				Name: "/sample-set",
				Type: "user",
				Value: credhub.UserValueType{
					Username:     "me",
					Password:     "super-secret",
					PasswordHash: "somestring",
				},
			}

			newCred, err := chClient.Set(cred, credhub.NoOverwrite, nil)
			Expect(err).To(Not(HaveOccurred()))
			Expect(newCred.Created).To(Not(BeEmpty()))
			v, err := credhub.UserValue(newCred)
			Expect(err).To(Not(HaveOccurred()))
			Expect(v.Password).To(BeEquivalentTo("old"))
			Expect(newCred.ID).To(BeEquivalentTo("6ba7b810-9dad-11d1-80b4-00c04fd430c8"))
		}

		setConvergeCredentialWithoutDifference := func() {
			cred := credhub.Credential{
				Name: "/sample-set",
				Type: "user",
				Value: credhub.UserValueType{
					Username:     "me",
					Password:     "super-secret",
					PasswordHash: "somestring",
				},
			}

			newCred, err := chClient.Set(cred, credhub.Converge, nil)
			Expect(err).To(Not(HaveOccurred()))
			Expect(newCred.Created).To(Not(BeEmpty()))
			Expect(newCred.ID).To(BeEquivalentTo("6ba7b810-9dad-11d1-80b4-00c04fd430c8"))
		}

		setConvergeCredentialWithDifference := func() {
			cred := credhub.Credential{
				Name: "/sample-set",
				Type: "user",
				Value: credhub.UserValueType{
					Username:     "me",
					Password:     "new-super-secret",
					PasswordHash: "somestring",
				},
			}

			newCred, err := chClient.Set(cred, credhub.Converge, nil)
			Expect(err).To(Not(HaveOccurred()))
			Expect(newCred.Created).To(Not(BeEmpty()))
			Expect(newCred.ID).To(Not(BeEquivalentTo("6ba7b810-9dad-11d1-80b4-00c04fd430c8")))
		}

		deleteFoundCredential := func() {
			err := chClient.Delete("/some-cred")
			Expect(err).To(Not(HaveOccurred()))
		}

		deleteNotFoundCredential := func() {
			err := chClient.Delete("/some-other-cred")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(BeEquivalentTo("expected return code 204, got 404"))
		}

		it.Before(func() {
			RegisterTestingT(t)
			server = mockCredhubServer()
		})

		it.After(func() {
			server.Close()
		})

		when("Running with UAA Authorization", func() {
			it.Before(func() {
				var err error

				endpoint, err := credhub.UAAEndpoint(server.URL, true)
				Expect(err).To(Not(HaveOccurred()))

				sslcli := &http.Client{
					Transport: &http.Transport{
						TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
					},
				}

				ctx := context.WithValue(context.TODO(), oauth2.HTTPClient, sslcli)
				cfg := &clientcredentials.Config{
					ClientID:     "user",
					ClientSecret: "pass",
					Scopes:       []string{"credhub.read", "credhub.write"},
					TokenURL:     endpoint.TokenURL,
				}

				chClient = credhub.New(server.URL, cfg.Client(ctx))
			})

			when("Testing Find By Path", func() {
				it("should be able to find creds by path", findByGoodPath)
				it("should not be able to find creds with an unknown path", findByBadPath)
				it("should not be able to find creds with bad auth", func() {
					endpoint, err := credhub.UAAEndpoint(server.URL, true)
					Expect(err).To(Not(HaveOccurred()))

					sslcli := &http.Client{
						Transport: &http.Transport{
							TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
						},
					}

					ctx := context.WithValue(context.TODO(), oauth2.HTTPClient, sslcli)
					cfg := &clientcredentials.Config{
						ClientID:     "user",
						ClientSecret: "pass",
						Scopes:       []string{"credhub.read", "credhub.write"},
						TokenURL:     endpoint.TokenURL,
					}

					badClient := credhub.New(server.URL, cfg.Client(ctx))

					_, err = badClient.FindByPath("/some/path")
					Expect(err).To(HaveOccurred())
				})
			})

			when("Testing Get By Name", func() {
				it("should get a 'value' type credential", valueByNameTests(false, -1))
				it("should get a 'password' type credential", passwordByNameTests(false, -1))
				it("should get a 'json' type credential", jsonByNameTests(false, -1))
				it("should get a 'user' type credential", getUserByName)
				it("should get a 'ssh' type credential", getSSHByName)
				it("should get a 'rsa' type credential", getRSAByName)
				it("should get a 'certificate' type credential", getCertificateByName)
				it("should not get a credential that doesn't exist", getNonexistentName)
			})

			when("Testing Get Latest By Name", func() {
				it("should get a 'value' type credential", valueByNameTests(true, -1))
				it("should get a 'password' type credential", passwordByNameTests(true, -1))
				it("should get a 'json' type credential", jsonByNameTests(true, -1))
			})

			when("Testing Get Latest By Name", func() {
				it("should get a 'value' type credential", valueByNameTests(false, 2))
				it("should get a 'password' type credential", passwordByNameTests(false, 2))
				it("should get a 'json' type credential", jsonByNameTests(false, 2))
			})

			when("Testing List All Paths", func() {
				it("should list all paths", listAllByPath)
			})

			when("Testing Get By ID", func() {
				it("should get an item with a valid ID", getById)
			})

			when("Testing Set Credential", func() {
				it("should receive the same item it sent, but with a timestamp", setOverwriteCredential)
				it("should receive an old credential", setNoOverwriteCredential)
				it("should receive an old credential if converging without changes", setConvergeCredentialWithoutDifference)
				it("should receive a new credential if converging with changes", setConvergeCredentialWithDifference)
			})

			when("Testing Generate Credential", func() {
				it("should generate a password credential", func() {
					params := make(map[string]interface{})
					params["length"] = 30
					cred, err := chClient.Generate("/example-generated", "password", params)
					Expect(err).To(Not(HaveOccurred()))
					Expect(cred.Type).To(Equal(credhub.Password))
					Expect(cred.Value).To(BeAssignableToTypeOf("expected"))
					Expect(cred.Value).To(HaveLen(30))
				})
			})

			when("Testing Regenerate Credential", func() {
				it("should regenerate a password credential", func() {
					cred, err := chClient.Regenerate("/example-password")
					Expect(err).To(Not(HaveOccurred()))
					Expect(cred.Type).To(Equal(credhub.Password))
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
				it("should delete a credential that it can find", deleteFoundCredential)
				it("should fail to delete a credential that it cannot find", deleteNotFoundCredential)
			})
		})
	}, spec.Report(report.Terminal{}))
}
