package client

//go:generate counterfeiter . Credhub

// Credhub - the main entry point
type Credhub interface {
	FindByPath(path string) ([]Credential, error)
	GetAllByName(name string) ([]Credential, error)
	GetVersionsByName(name string, numVersions int) ([]Credential, error)
	GetLatestByName(name string) (Credential, error)
	Set(credential Credential) (Credential, error)
}
