package client_test

import (
	"net/http/httptest"
	"testing"

	"github.com/jghiloni/credhub-sdk/client"
	sdktest "github.com/jghiloni/credhub-sdk/testing"

	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
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

		getValueByName := func() {
			creds, err := chClient.GetByName("/concourse/common/sample-value")
			Expect(err).To(BeNil())
			Expect(len(creds)).To(Equal(3))
		}

		getPasswordByName := func() {
			creds, err := chClient.GetByName("/concourse/common/sample-password")
			Expect(err).To(BeNil())
			Expect(len(creds)).To(Equal(3))
			Expect(creds[0].Value).To(BeEquivalentTo("sample1"))
		}

		getJSONByName := func() {
			creds, err := chClient.GetByName("/concourse/common/sample-json")
			Expect(err).To(BeNil())
			Expect(len(creds)).To(Equal(3))

			intf := creds[2].Value
			val, ok := intf.([]interface{})
			Expect(ok).To(BeTrue())

			Expect(int(val[0].(float64))).To(Equal(1))
			Expect(int(val[1].(float64))).To(Equal(2))
		}

		getUserByName := func() {
			creds, err := chClient.GetByName("/concourse/common/sample-user")
			Expect(err).To(Not(HaveOccurred()))
			Expect(len(creds)).To(Equal(1))

			cred := creds[0]

			var val client.UserValueType
			val, err = client.UserValue(cred)
			Expect(err).To(Not(HaveOccurred()))
			Expect(val.Username).To(Equal("me"))
		}

		getSSHByName := func() {
			creds, err := chClient.GetByName("/concourse/common/sample-ssh")
			Expect(err).To(Not(HaveOccurred()))
			Expect(len(creds)).To(Equal(1))

			cred := creds[0]
			var val client.SSHValueType
			val, err = client.SSHValue(cred)
			Expect(err).To(Not(HaveOccurred()))
			Expect(val.PublicKey).To(HavePrefix("ssh-rsa"))
		}

		getRSAByName := func() {
			creds, err := chClient.GetByName("/concourse/common/sample-rsa")
			Expect(err).To(Not(HaveOccurred()))
			Expect(len(creds)).To(Equal(1))

			cred := creds[0]
			var val client.RSAValueType
			val, err = client.RSAValue(cred)
			Expect(err).To(Not(HaveOccurred()))
			Expect(val.PrivateKey).To(HavePrefix("-----BEGIN PRIVATE KEY-----"))
		}

		getNonexistentName := func() {
			_, err := chClient.GetByName("/concourse/common/not-real")
			Expect(err).To(HaveOccurred())
		}

		getCertificateByName := func() {
			creds, err := chClient.GetByName("/concourse/common/sample-certificate")
			Expect(err).To(Not(HaveOccurred()))
			Expect(len(creds)).To(Equal(1))

			cred := creds[0]
			var val client.CertificateValueType
			val, err = client.CertificateValue(cred)
			Expect(err).To(Not(HaveOccurred()))
			Expect(val.Certificate).To(HavePrefix("-----BEGIN CERTIFICATE-----"))
		}

		it.Before(func() {
			RegisterTestingT(t)
			server = sdktest.MockCredhubServer()
		})

		when("Running with UAA Authorization", func() {
			it.Before(func() {
				var err error
				chClient, err = client.New(server.URL, "user", "pass", false)
				Expect(err).To(Not(HaveOccurred()))
			})

			when("Testing Find By Path", func() {
				it("should be able to find creds by path", findByGoodPath)
				it("should not be able to find creds with an unknown path", findByBadPath)
				it("should not be able to find creds with bad auth", func() {
					badClient, err := client.New(server.URL, "asdf", "asdf", true)
					Expect(err).To(BeNil())

					_, err = badClient.FindByPath("/some/path")
					Expect(err).To(HaveOccurred())
				})
			})

			when("Testing Get By Name", func() {
				it("should get a 'value' type credential", getValueByName)
				it("should get a 'password' type credential", getPasswordByName)
				it("should get a 'json' type credential", getJSONByName)
				it("should get a 'user' type credential", getUserByName)
				it("should get a 'ssh' type credential", getSSHByName)
				it("should get a 'rsa' type credential", getRSAByName)
				it("should get a 'certificate' type credential", getCertificateByName)
				it("should not get a credential that doesn't exist", getNonexistentName)
			})
		})
	}, spec.Report(report.Terminal{}))
}
