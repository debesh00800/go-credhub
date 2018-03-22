package client

import (
	"encoding/json"
	"errors"
)

// CredentialType is the list of valid types of credentials Credhub supports
type CredentialType string

const (
	// Value - A generic value
	Value CredentialType = "value"
	// Password - A password that can be (re-)generated
	Password CredentialType = "password"
	// User - A username, password, and password hash
	User CredentialType = "user"
	// JSON - An arbitrary block of JSON
	JSON CredentialType = "json"
	// RSA - A public/private key pair
	RSA CredentialType = "rsa"
	// SSH - An SSH private key, public key (in OpenSSH format), and public key fingerprint
	SSH CredentialType = "ssh"
	// Certificate - A private key, associated certificate, and CA
	Certificate CredentialType = "certificate"
)

type Credential struct {
	Name         string         `json:"name"`
	Created      string         `json:"version_created_at"`
	Type         CredentialType `json:"type,omitempty"`
	Value        interface{}    `json:"value,omitempty"`
	remarshalled bool
}

type Permission struct {
	Actor      string   `json:"actor"`
	Operations []string `json:"operations"`
}

type UserValueType struct {
	Username     string `json:"username"`
	Password     string `json:"password"`
	PasswordHash string `json:"password_hash"`
}

type RSAValueType struct {
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key"`
}

type SSHValueType struct {
	PublicKey            string `json:"public_key"`
	PrivateKey           string `json:"private_key"`
	PublicKeyFingerprint string `json:"public_key_fingerprint"`
}

type CertificateValueType struct {
	CA          string `json:"ca"`
	PrivateKey  string `json:"private_key"`
	Certificate string `json:"certificate"`
}

func UserValue(cred Credential) (UserValueType, error) {
	def := UserValueType{}
	switch cred.Type {
	case User:
		if !cred.remarshalled {
			val := UserValueType{}
			buf, err := json.Marshal(cred.Value)
			if err != nil {
				return def, err
			}

			err = json.Unmarshal(buf, &val)
			if err != nil {
				return def, err
			}

			cred.Value = val
			cred.remarshalled = true
		}

		return cred.Value.(UserValueType), nil
	default:
		return def, errors.New(`only "user" type credentials have UserValueType values`)
	}
}

func RSAValue(cred Credential) (RSAValueType, error) {
	def := RSAValueType{}
	switch cred.Type {
	case RSA:
		if !cred.remarshalled {
			val := RSAValueType{}
			buf, err := json.Marshal(cred.Value)
			if err != nil {
				return def, err
			}

			err = json.Unmarshal(buf, &val)
			if err != nil {
				return def, err
			}

			cred.Value = val
			cred.remarshalled = true

		}

		return cred.Value.(RSAValueType), nil
	default:
		return def, errors.New(`only "rsa" type credentials have RSAValueType values`)
	}
}

func SSHValue(cred Credential) (SSHValueType, error) {
	def := SSHValueType{}
	switch cred.Type {
	case SSH:
		if !cred.remarshalled {
			val := SSHValueType{}
			buf, err := json.Marshal(cred.Value)
			if err != nil {
				return def, err
			}

			err = json.Unmarshal(buf, &val)
			if err != nil {
				return def, err
			}

			cred.Value = val
			cred.remarshalled = true
		}

		return cred.Value.(SSHValueType), nil
	default:
		return def, errors.New(`only "ssh" type credentials have SSHValueType values`)
	}
}

func CertificateValue(cred Credential) (CertificateValueType, error) {
	def := CertificateValueType{}
	switch cred.Type {
	case Certificate:
		if !cred.remarshalled {
			val := CertificateValueType{}
			buf, err := json.Marshal(cred.Value)
			if err != nil {
				return def, err
			}

			err = json.Unmarshal(buf, &val)
			if err != nil {
				return def, err
			}

			cred.Value = val
			cred.remarshalled = true
		}

		return cred.Value.(CertificateValueType), nil
	default:
		return def, errors.New(`only "certificate" type credentials have CertificateValueType values`)
	}
}
