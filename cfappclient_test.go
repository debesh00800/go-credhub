package credhub_test

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	credhub "github.com/cloudfoundry-community/go-credhub"
)

func mtlsTestServer() *httptest.Server {
	server := httptest.NewUnstartedServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, "Hello world")
		}),
	)
	server.TLS = &tls.Config{
		ClientAuth: tls.RequireAnyClientCert,
	}
	server.StartTLS()

	return server
}

func beforeCFAppClientTest() *httptest.Server {
	os.Setenv("CF_INSTANCE_KEY", "testdata/tls/key")
	os.Setenv("CF_INSTANCE_CERT", "testdata/tls/cert")

	return mtlsTestServer()
}

func afterCFAppClientTest(server *httptest.Server) {
	server.Close()

	os.Unsetenv("CF_INSTANCE_KEY")
	os.Unsetenv("CF_INSTANCE_CERT")
}

func TestCFAppClient_Get(t *testing.T) {
	server := beforeCFAppClientTest()
	defer afterCFAppClientTest(server)

	client, err := credhub.NewCFAppAuthClient(server.Client().Transport.(*http.Transport))
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.Get(server.URL)
	if err != nil {
		t.Fatal(err)
	}
}

func TestCFAppClient_Do(t *testing.T) {
	server := beforeCFAppClientTest()
	defer afterCFAppClientTest(server)

	client, err := credhub.NewCFAppAuthClient(server.Client().Transport.(*http.Transport))
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodGet, server.URL, nil)
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
}

func TestMTLS_Server(t *testing.T) {
	server := beforeCFAppClientTest()
	defer afterCFAppClientTest(server)

	_, err := server.Client().Get(server.URL)
	if err == nil {
		t.Fatal("error should have occurred")
	}
}
