package endpoints

import "github.com/akitasoftware/akita-libs/http_rest_methods"

// An endpoint whose host and path template have not been normalized.
type Endpoint struct {
	HTTPMethod   http_rest_methods.HTTPMethod `json:"method"`
	Host         Host                         `json:"host"`
	PathTemplate PathTemplate                 `json:"path_template"`
}
