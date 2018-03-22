package testing

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"strings"
)

// MockCredhubServer will create a mock server that is useful for unit testing
func MockCredhubServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uaa := mockUaaServer()
		if r.URL.Path == "/info" {
			infoHandler(uaa.URL, w, r)
			return
		}

		authHeader := r.Header.Get("Authorization")
		if strings.ToLower(authHeader) != "bearer abcd" {
			w.WriteHeader(401)
			return
		}
		switch r.URL.Path {
		case "/some-url":
			w.Write([]byte("Hello world"))
		case "/api/v1/data":
			path := r.FormValue("path")
			name := r.FormValue("name")

			switch {
			case path != "" && name == "":
				returnFromFile("bypath", path, w, r)
			case path == "" && name != "":
				returnFromFile("byname", name, w, r)
			default:
				w.WriteHeader(400)
			}
		}
	}))
}

func infoHandler(uaaURL string, w http.ResponseWriter, r *http.Request) {
	body := make(map[string]interface{})

	url := make(map[string]string)
	url["url"] = uaaURL

	body["auth-server"] = url

	var out []byte
	var err error
	if out, err = json.Marshal(body); err != nil {
		w.WriteHeader(500)
		return
	}

	w.Write(out)
}

func returnFromFile(query, value string, w http.ResponseWriter, r *http.Request) {
	filePath := path.Join("fixtures", query, "fake-"+value+".json")
	buf, err := ioutil.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			w.WriteHeader(404)
		} else {
			w.WriteHeader(500)
		}

		w.Header().Add("content-type", "application/json")
		w.Write([]byte("{}"))

		return
	}

	w.WriteHeader(200)
	w.Header().Add("Content-Type", "applicaton/json")
	w.Write(buf)
}
