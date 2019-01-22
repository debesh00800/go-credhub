// +build !go1.11

package credhub

import "net/http"

func copyExtraTransportFields(src, dst *http.Transport) {}
