package client_test

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/jghiloni/credhub-api/auth"

	"github.com/jghiloni/credhub-api/client"
	apitest "github.com/jghiloni/credhub-api/internal/testing"

	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"

	"golang.org/x/oauth2/clientcredentials"
)

func TestCredhubClient(t *testing.T) {
	spec.Run(t, "Credhub Client", func(t *testing.T, when spec.G, it spec.S) {
		var chClient client.Credhub
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

			var val client.UserValueType
			val, err = client.UserValue(cred)
			Expect(err).To(Not(HaveOccurred()))
			Expect(val.Username).To(Equal("me"))
		}

		getSSHByName := func() {
			creds, err := chClient.GetAllByName("/concourse/common/sample-ssh")
			Expect(err).To(Not(HaveOccurred()))
			Expect(len(creds)).To(Equal(1))

			cred := creds[0]
			var val client.SSHValueType
			val, err = client.SSHValue(cred)
			Expect(err).To(Not(HaveOccurred()))
			Expect(val.PublicKey).To(HavePrefix("ssh-rsa"))
		}

		getRSAByName := func() {
			creds, err := chClient.GetAllByName("/concourse/common/sample-rsa")
			Expect(err).To(Not(HaveOccurred()))
			Expect(len(creds)).To(Equal(1))

			cred := creds[0]
			var val client.RSAValueType
			val, err = client.RSAValue(cred)
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
			var val client.CertificateValueType
			val, err = client.CertificateValue(cred)
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

		setCredential := func() {
			cred := client.Credential{
				Name: "/sample-set",
				Type: "user",
				Value: client.UserValueType{
					Username:     "me",
					Password:     "super-secret",
					PasswordHash: "somestring",
				},
			}

			newCred, err := chClient.Set(cred)
			Expect(err).To(Not(HaveOccurred()))
			Expect(newCred.Created).To(Not(BeEmpty()))
		}

		it.Before(func() {
			RegisterTestingT(t)
			server = apitest.MockCredhubServer()
		})

		when("Running with UAA Authorization", func() {
			it.Before(func() {
				var err error

				endpoint, err := auth.UAAEndpoint(server.URL, false)
				Expect(err).To(Not(HaveOccurred()))

				ctx := context.Background()
				cfg := &clientcredentials.Config{
					ClientID:     "user",
					ClientSecret: "pass",
					Scopes:       []string{"credhub.read", "credhub.write"},
					TokenURL:     endpoint.TokenURL,
				}

				chClient = client.New(server.URL, cfg.Client(ctx))
			})

			when("Testing Find By Path", func() {
				it("should be able to find creds by path", findByGoodPath)
				it("should not be able to find creds with an unknown path", findByBadPath)
				it("should not be able to find creds with bad auth", func() {
					endpoint, err := auth.UAAEndpoint(server.URL, false)
					Expect(err).To(Not(HaveOccurred()))

					ctx := context.Background()
					cfg := &clientcredentials.Config{
						ClientID:     "asdf",
						ClientSecret: "asdf",
						Scopes:       []string{"credhub.read", "credhub.write"},
						TokenURL:     endpoint.TokenURL,
					}

					badClient := client.New(server.URL, cfg.Client(ctx))

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
				it("should receive the same item it sent, but with a timestamp", setCredential)
			})

			when("Testing Generate Credential", func() {
				it("should generate a password credential", func() {
					params := make(map[string]interface{})
					params["length"] = 30
					cred, err := chClient.Generate("/example-generated", "password", params)
					Expect(err).To(Not(HaveOccurred()))
					Expect(cred.Type).To(Equal(client.Password))
					Expect(cred.Value).To(BeAssignableToTypeOf("expected"))
					Expect(cred.Value).To(HaveLen(30))
				})
			})

			when("Testing Regenerate Credential", func() {
				it("should regenerate a password credential", func() {
					cred, err := chClient.Regenerate("/example-password")
					Expect(err).To(Not(HaveOccurred()))
					Expect(cred.Type).To(Equal(client.Password))
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
		})
	}, spec.Report(report.Terminal{}))
}
