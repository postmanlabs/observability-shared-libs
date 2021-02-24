package spec_summary

// Summarizes a spec along different dimensions that can be used to filter for
// parts of the spec.
type Summary struct {
	Authentications map[string]struct{} `json:"authentications"`
	HTTPMethods     map[string]struct{} `json:"http_methods"`
	Paths           map[string]struct{} `json:"paths"`
	Params          map[string]struct{} `json:"params"`
	Properties      map[string]struct{} `json:"properties"`
	ResponseCodes   map[int32]struct{}  `json:"response_codes"`
	DataFormats     map[string]struct{} `json:"data_formats"`
	DataKinds       map[string]struct{} `json:"data_kinds"`
	DataTypes       map[string]struct{} `json:"data_types"`
}
