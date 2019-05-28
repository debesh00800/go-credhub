package credhub

import (
	"crypto/tls"
	"net/http"
	"os"
)

// NewCFAppAuthClient creates a CFAppAuthClient with client cert info injected
// into the given transport
func NewCFAppAuthClient(tr *http.Transport) (HTTPClient, error) {
	client := &CFAppAuthClient{}

	if err := client.loadTLS(tr); err != nil {
		return nil, err
	}

	return client, nil
}

// CFAppAuthClient wraps an HTTPClient and handles mTLS authentication
type CFAppAuthClient struct {
	hc HTTPClient
}

// Get will do an HTTP Request to the specified URL using the HTTP GET method
func (c *CFAppAuthClient) Get(url string) (resp *http.Response, err error) {
	return c.hc.Get(url)
}

// Do will perform the HTTP Request specified with the underlying HTTPClient
func (c *CFAppAuthClient) Do(req *http.Request) (*http.Response, error) {
	return c.hc.Do(req)
}

func (c *CFAppAuthClient) loadTLS(tr *http.Transport) error {
	var modifiedTransport *http.Transport

	if tr == nil {
		modifiedTransport = copyTransport(http.DefaultTransport.(*http.Transport))
	} else {
		modifiedTransport = copyTransport(tr)
	}

	cert, err := tls.LoadX509KeyPair(
		os.Getenv("CF_INSTANCE_CERT"),
		os.Getenv("CF_INSTANCE_KEY"),
	)

	if err != nil {
		return err
	}

	if modifiedTransport.TLSClientConfig == nil {
		modifiedTransport.TLSClientConfig = &tls.Config{}
	}

	if modifiedTransport.TLSClientConfig.Certificates == nil {
		modifiedTransport.TLSClientConfig.Certificates = []tls.Certificate{cert}
	} else {
		modifiedTransport.TLSClientConfig.Certificates = append(modifiedTransport.TLSClientConfig.Certificates, cert)
	}

	modifiedTransport.TLSClientConfig.BuildNameToCertificate()

	c.hc = &http.Client{Transport: modifiedTransport}

	return nil
}

func copyTransport(tr *http.Transport) *http.Transport {
	copy := &http.Transport{
		Dial:                   tr.Dial,
		DialContext:            tr.DialContext,
		DialTLS:                tr.DialTLS,
		DisableCompression:     tr.DisableCompression,
		DisableKeepAlives:      tr.DisableKeepAlives,
		ExpectContinueTimeout:  tr.ExpectContinueTimeout,
		IdleConnTimeout:        tr.IdleConnTimeout,
		MaxIdleConns:           tr.MaxIdleConns,
		MaxIdleConnsPerHost:    tr.MaxIdleConnsPerHost,
		MaxResponseHeaderBytes: tr.MaxResponseHeaderBytes,
		Proxy:                  tr.Proxy,
		ProxyConnectHeader:     tr.ProxyConnectHeader,
		ResponseHeaderTimeout:  tr.ResponseHeaderTimeout,
		TLSClientConfig:        tr.TLSClientConfig,
		TLSHandshakeTimeout:    tr.TLSHandshakeTimeout,
		TLSNextProto:           tr.TLSNextProto,
	}

	copyExtraTransportFields(tr, copy)
	return copy
}
