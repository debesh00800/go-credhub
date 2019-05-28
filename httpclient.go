package credhub

import "net/http"

// HTTPClient is an interface that http.Client conforms to, and is useful for
// mocking purposes.
type HTTPClient interface {
	Get(url string) (resp *http.Response, err error)
	Do(req *http.Request) (*http.Response, error)
}
