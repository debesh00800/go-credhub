package credhub_test

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

	"github.com/jghiloni/credhub-api"
)

type credentialFile map[string][]credhub.Credential

// MockCredhubServer will create a mock server that is useful for unit testing
func mockCredhubServer() *httptest.Server {
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
		case "delete":
			deleteHandler(w, r)
		}
	}))
}

func getHandler(w http.ResponseWriter, r *http.Request) {
	ret := make(map[string]interface{})
	key := "data"

	var creds []credhub.Credential
	var err error
	switch r.URL.Path {
	case "/some-url":
		w.Write([]byte("Hello world"))
		return
	case "/api/v1/data":
		path := r.FormValue("path")
		name := r.FormValue("name")
		paths := r.FormValue("paths")
		nameLike := r.FormValue("name-like")

		switch {
		case path != "":
			key = "credentials"
			creds, err = returnFromFile("bypath", path, key, w, r)
		case name != "":
			creds, err = returnFromFile("byname", name, key, w, r)
		case paths == "true":
			// creds, err = returnFromFile("", "allpaths", w, r)
			directWriteFile("fixtures/allpaths.json", w, r)
			return
		case nameLike != "":
			key = "credentials"
			creds, err = returnFromFile("bypath", "/concourse/common", key, w, r)
			idxs := make([]int, 0, len(creds))
			for idx, cred := range creds {
				if !strings.Contains(strings.ToLower(cred.Name), strings.ToLower(nameLike)) {
					// get the list of bad indices in high to low order so as to most easily delete them
					idxs = append([]int{idx}, idxs...)
				}
			}

			for _, i := range idxs {
				creds = append(creds[:i], creds[i+1:]...)
			}
		default:
			w.WriteHeader(400)
			return
		}
	case "/api/v1/data/1234":
		directWriteFile("fixtures/byid/1234.json", w, r)
		return
	default:
		w.WriteHeader(404)
		return
	}

	if err != nil {
		w.WriteHeader(400)
		return
	}

	ret[key] = creds

	buf, err := json.Marshal(ret)
	if err != nil {
		w.WriteHeader(500)
		return
	}

	w.WriteHeader(200)
	w.Write(buf)
}

func postHandler(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/api/v1/data":
		var generateBody struct {
			Name   string                 `json:"name"`
			Type   credhub.CredentialType `json:"type"`
			Params map[string]interface{} `json:"parameters"`
		}

		var cred credhub.Credential
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
	case "/api/v1/data/regenerate":
		var body struct {
			Name string `json:"name"`
		}

		var cred credhub.Credential
		buf, _ := ioutil.ReadAll(r.Body)
		if err := json.Unmarshal(buf, &body); err != nil {
			w.WriteHeader(400)
		}

		cred.Name = body.Name
		cred.Type = credhub.Password
		cred.Value = "P$<MNBVCXZ;lkjhgfdsa0987654321"
		cred.Created = time.Now().Format(time.RFC3339)
		buf, e := json.Marshal(cred)
		if e != nil {
			w.WriteHeader(500)
		}

		w.Write(buf)
	}
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/api/v1/data":
		name := r.URL.Query().Get("name")
		if name == "/some-cred" {
			w.WriteHeader(204)
			return
		}

		fallthrough
	default:
		w.WriteHeader(404)
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

func returnFromFile(query, value, key string, w http.ResponseWriter, r *http.Request) ([]credhub.Credential, error) {
	filePath := path.Join("fixtures", query, value+".json")
	buf, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var creds []credhub.Credential

	params := r.URL.Query()
	name := params.Get("name")

	ret := make(credentialFile)

	if err = json.Unmarshal(buf, &ret); err != nil {
		return nil, err
	}

	creds = ret[key]

	if name != "" {
		currentStr := params.Get("current")
		versionsStr := params.Get("versions")

		sort.Slice(ret[key], func(i, j int) bool {
			less := strings.Compare(ret[key][i].Created, ret[key][j].Created)
			return less > 0
		})

		current, _ := strconv.ParseBool(currentStr)
		if current {
			data := ret[key][0:1]
			ret[key] = data
		} else {
			nv, _ := strconv.Atoi(versionsStr)
			if nv > 0 {
				data := ret[key][0:nv]
				ret[key] = data
			}
		}

		creds = ret[key]
	}

	return creds, nil
}

func directWriteFile(path string, w http.ResponseWriter, r *http.Request) {
	buf, err := ioutil.ReadFile(path)
	if os.IsNotExist(err) {
		w.WriteHeader(404)
		return
	} else if err != nil {
		w.WriteHeader(500)
	}

	w.Write(buf)
}
