package spec_summary

// Summarizes a spec along different dimensions that can be used to filter for
// parts of the spec.
type Summary struct {
	Authentications map[string]int `json:"authentications"`
	HTTPMethods     map[string]int `json:"http_methods"`
	Paths           map[string]int `json:"paths"`
	Params          map[string]int `json:"params"`
	Properties      map[string]int `json:"properties"`
	ResponseCodes   map[string]int `json:"response_codes"`
	DataFormats     map[string]int `json:"data_formats"`
	DataKinds       map[string]int `json:"data_kinds"`
	DataTypes       map[string]int `json:"data_types"`
}
