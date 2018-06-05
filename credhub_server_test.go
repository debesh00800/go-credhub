package credhub_test

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jghiloni/credhub-api"
	uuid "github.com/nu7hatch/gouuid"
)

type credentialFile map[string][]credhub.Credential

// MockCredhubServer will create a mock server that is useful for unit testing
func mockCredhubServer() *httptest.Server {
	return httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
		case "put":
			putHandler(w, r)
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
			creds, err = returnCredentialsFromFile("bypath", path, key, w, r)
		case name != "":
			creds, err = returnCredentialsFromFile("byname", name, key, w, r)
			if os.IsNotExist(err) {
				w.WriteHeader(404)
				return
			}
		case paths == "true":
			// creds, err = returnFromFile("", "allpaths", w, r)
			directWriteFile("testdata/credentials/allpaths.json", w, r)
			return
		case nameLike != "":
			key = "credentials"
			creds, err = returnCredentialsFromFile("bypath", "/concourse/common", key, w, r)
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
	case "/api/v1/permissions":
		name := r.FormValue("credential_name")

		if name == "/add-permission-credential" {
			fileName := "testdata/permissions/add-permissions/cred.json"
			if _, err = os.Stat(fileName); os.IsNotExist(err) {
				err = copyFile("testdata/permissions/add-permissions/base.json", fileName)
			}

			if err != nil {
				w.WriteHeader(500)
				return
			}

			name = "/add-permissions/cred"
		}

		directWriteFile(path.Join("testdata/permissions", name+".json"), w, r)
		return
	case "/api/v1/data/1234":
		directWriteFile("testdata/credentials/byid/1234.json", w, r)
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
				return
			} else if generateBody.Params != nil {
				cred.Name = generateBody.Name
				cred.Type = generateBody.Type
				cred.Value = "1234567890asdfghjkl;ZXCVBNM<$P"
			} else {
				w.WriteHeader(400)
				return
			}
		}
		t := time.Now()
		cred.Created = t.Format(time.RFC3339)
		buf, e := json.Marshal(cred)
		if e != nil {
			w.WriteHeader(500)
			return
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
			return
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
	case "/api/v1/permissions":
		type permbody struct {
			Name        string               `json:"credential_name"`
			Permissions []credhub.Permission `json:"permissions"`
		}

		var body permbody

		buf, _ := ioutil.ReadAll(r.Body)
		if err := json.Unmarshal(buf, &body); err != nil {
			w.WriteHeader(400)
			return
		}

		if body.Name == "/add-permission-credential" {
			fp, err := os.OpenFile("testdata/permissions/add-permissions/cred.json", os.O_RDWR, 0644)
			if os.IsNotExist(err) {
				w.WriteHeader(404)
				return
			} else if err != nil {
				w.WriteHeader(500)
				return
			}
			defer fp.Close()

			var buf []byte
			buf, err = ioutil.ReadAll(fp)
			if err != nil {
				w.WriteHeader(500)
				return
			}

			var existing permbody
			if err = json.Unmarshal(buf, &existing); err != nil {
				w.WriteHeader(500)
				return
			}

			existing.Permissions = append(existing.Permissions, body.Permissions...)
			outbuf, err := json.Marshal(existing)
			if err != nil {
				w.WriteHeader(500)
				return
			}

			fp.WriteAt(outbuf, 0)
			w.Write(outbuf)
			return
		} else {
			w.WriteHeader(404)
			return
		}
	}
}

func putHandler(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/api/v1/data":
		var cred credhub.Credential
		var req struct {
			credhub.Credential
			Mode                  credhub.OverwriteMode `json:"mode"`
			AdditionalPermissions []credhub.Permission  `json:"additonal_permissions,omitempty"`
		}
		buf, _ := ioutil.ReadAll(r.Body)
		if err := json.Unmarshal(buf, &req); err != nil {
			w.WriteHeader(400)
		}

		cred.Name = req.Name
		cred.Type = req.Type
		cred.Value = req.Value

		switch req.Mode {
		case credhub.Overwrite:
			guid, err := uuid.NewV4()
			if err != nil {
				w.WriteHeader(500)
			}
			cred.ID = guid.String()
		case credhub.NoOverwrite:
			cred.ID = "6ba7b810-9dad-11d1-80b4-00c04fd430c8"
			cred.Value = credhub.UserValueType{
				Username:     "me",
				Password:     "old",
				PasswordHash: "old-hash",
			}
		case credhub.Converge:
			v, err := credhub.UserValue(cred)
			if err != nil {
				w.WriteHeader(400)
				return
			}

			if v.Password == "super-secret" {
				cred.ID = "6ba7b810-9dad-11d1-80b4-00c04fd430c8"
			} else {
				guid, err := uuid.NewV4()
				if err != nil {
					w.WriteHeader(500)
				}
				cred.ID = guid.String()
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

func returnPermissionsFromFile(credentialName string) ([]credhub.Permission, error) {
	filePath := path.Join("testdata/permissions", credentialName+".json")
	buf, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	retBody := struct {
		CN          string               `json:"credential_name"`
		Permissions []credhub.Permission `json:"permissions"`
	}{}

	if err = json.Unmarshal(buf, &retBody); err != nil {
		return nil, err
	}

	return retBody.Permissions, nil
}

func returnCredentialsFromFile(query, value, key string, w http.ResponseWriter, r *http.Request) ([]credhub.Credential, error) {
	filePath := path.Join("testdata/credentials", query, value+".json")
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
	log.SetFlags(log.Lshortfile | log.Ldate | log.Ltime | log.LUTC)
	buf, err := ioutil.ReadFile(path)
	if os.IsNotExist(err) {
		w.WriteHeader(404)
		return
	} else if err != nil {
		w.WriteHeader(500)
		return
	}

	w.Write(buf)
}

func copyFile(src, dst string) error {
	var in, out *os.File
	var err error
	if in, err = os.Open(src); err != nil {
		return err
	}
	// defer in.Close()
	defer func() {
		in.Close()
	}()

	if out, err = os.Create(dst); err != nil {
		return err
	}
	//defer out.Close()
	defer func() {
		out.Close()
	}()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}

	return nil
}
