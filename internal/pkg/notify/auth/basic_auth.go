package auth

import (
	"encoding/base64"
)

// encodeBasicAuth encodes basic auth credentials to base64
func (a *BasicAuth) encodeBasicAuth() string {
	credentials := a.Username + ":" + a.Password
	return base64.StdEncoding.EncodeToString([]byte(credentials))
}
