package testing

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jghiloni/credhub-api/client"
)

// MockCredhubServer will create a mock server that is useful for unit testing
func MockCredhubServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/info" {
			infoHandler(w, r)
			return
		}

		authHeader := r.Header.Get("Authorization")
		if strings.ToLower(authHeader) != "bearer abcd" {
			w.WriteHeader(401)
			return
		}

		switch strings.ToLower(r.Method) {
		case "get":
			getHandler(w, r)
		case "post":
			postHandler(w, r)
		}
	}))
}

func getHandler(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/some-url":
		w.Write([]byte("Hello world"))
	case "/api/v1/data":
		path := r.FormValue("path")
		name := r.FormValue("name")
		paths := r.FormValue("paths")

		switch {
		case path != "" && name == "":
			returnFromFile("bypath", path, w, r)
		case path == "" && name != "":
			returnFromFile("byname", name, w, r)
		case paths == "true":
			returnFromFile("", "allpaths", w, r)
		default:
			w.WriteHeader(400)
		}
	case "/api/v1/data/1234":
		returnFromFile("byid", "1234", w, r)
	default:
		w.WriteHeader(404)
	}
}

func postHandler(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/api/v1/data":
		var generateBody struct {
			Name   string                 `json:"name"`
			Type   client.CredentialType  `json:"type"`
			Params map[string]interface{} `json:"parameters"`
		}

		var cred client.Credential
		buf, _ := ioutil.ReadAll(r.Body)
		if err := json.Unmarshal(buf, &cred); err != nil {
			w.WriteHeader(400)
		}

		if cred.Value == nil {
			if err := json.Unmarshal(buf, &generateBody); err != nil {
				w.WriteHeader(400)
			} else if generateBody.Params != nil {
				cred.Name = generateBody.Name
				cred.Type = generateBody.Type
				cred.Value = "1234567890asdfghjkl;ZXCVBNM<$P"
			} else {
				w.WriteHeader(400)
			}
		}
		t := time.Now()
		cred.Created = t.Format(time.RFC3339)
		buf, e := json.Marshal(cred)
		if e != nil {
			w.WriteHeader(500)
		}

		w.Write(buf)
	}
}

func infoHandler(w http.ResponseWriter, r *http.Request) {
	body := make(map[string]interface{})

	url := make(map[string]string)
	url["url"] = mockUaaServer().URL

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
	filePath := path.Join("fixtures", query, value+".json")
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

	params := r.URL.Query()
	name := params.Get("name")
	if name != "" {
		currentStr := params.Get("current")
		versionsStr := params.Get("versions")

		var ret struct {
			Data []client.Credential `json:"data"`
		}

		if err = json.Unmarshal(buf, &ret); err != nil {
			w.WriteHeader(500)
		}

		sort.Slice(ret.Data, func(i, j int) bool {
			less := strings.Compare(ret.Data[i].Created, ret.Data[j].Created)
			return less > 0
		})

		current, _ := strconv.ParseBool(currentStr)
		if current {
			data := ret.Data[0:1]
			ret.Data = data
		} else {
			nv, _ := strconv.Atoi(versionsStr)
			if nv > 0 {
				data := ret.Data[0:nv]
				ret.Data = data
			}
		}

		buf, _ = json.Marshal(ret)
	}

	w.WriteHeader(200)
	w.Header().Add("Content-Type", "applicaton/json")
	w.Write(buf)
}
