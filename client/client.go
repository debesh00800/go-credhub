package client

//go:generate counterfeiter . Credhub

// Credhub - the main entry point
type Credhub interface {
	ListAllCredentials() ([]Credential, error)

	GetByID(id string) (Credential, error)
	GetAllByName(name string) ([]Credential, error)
	GetVersionsByName(name string, numVersions int) ([]Credential, error)
	GetLatestByName(name string) (Credential, error)

	Set(credential Credential) (Credential, error)

	Generate(name string, credentialType CredentialType, parameters map[string]interface{}) (Credential, error)
	Regenerate(name string) (Credential, error)

	Delete(name string) error

	FindByPath(path string) ([]Credential, error)
	FindByPartialName(partialName string) ([]Credential, error)

	GetPermissions(credentialName string) ([]Permission, error)
	AddPermissions(credentialName string, newPerms []Permission) ([]Permission, error)
	DeletePermissions(credentialName, actorID string) error
}
