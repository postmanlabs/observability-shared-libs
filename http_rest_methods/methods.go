package http_rest_methods

import (
	"encoding/json"
	"strings"

	"github.com/akitasoftware/go-utils/sets"
	"github.com/pkg/errors"
)

type HTTPMethod string

const (
	CONNECT HTTPMethod = "CONNECT"
	DELETE  HTTPMethod = "DELETE"
	GET     HTTPMethod = "GET"
	HEAD    HTTPMethod = "HEAD"
	OPTIONS HTTPMethod = "OPTIONS"
	PATCH   HTTPMethod = "PATCH"
	POST    HTTPMethod = "POST"
	PUT     HTTPMethod = "PUT"
	TRACE   HTTPMethod = "TRACE"
)

var AllMethods = sets.NewSet(
	CONNECT,
	DELETE,
	GET,
	HEAD,
	OPTIONS,
	PATCH,
	POST,
	PUT,
	TRACE,
)

func (m *HTTPMethod) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	var err error
	if *m, err = ParseHTTPMethod(s); err != nil {
		return err
	}
	return nil
}

func ParseHTTPMethod(s string) (HTTPMethod, error) {
	result := HTTPMethod(strings.ToUpper(s))
	if AllMethods.Contains(result) {
		return result, nil
	}

	return "", errors.Errorf("unknown rest method: %s", s)
}
