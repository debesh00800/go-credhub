package client

// Credhub - the main entry point
//go:generate counterfeiter . Credhub
type Credhub interface {
	FindByPath(path string) ([]Credential, error)
	GetByName(name string) ([]Credential, error)
	GetLatestByName(name string) (Credential, error)
	Set(credential Credential) (Credential, error)
}
