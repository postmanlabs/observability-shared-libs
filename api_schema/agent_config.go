package api_schema

import "regexp"

// Reflects the version of the JSON encoding of ServiceAgentConfig. Increase the
// minor version number for backwards-compatible changes, and the major version
// number for non-backwards-compatible changes.
const ServiceAgentConfigVersion = "v0.1"

type ServiceAgentConfig struct {
	FormatVersion  string                `json:"version"`
	FieldsToRedact *FieldRedactionConfig `json:"fields_to_redact"`
}

type FieldRedactionConfig struct {
	FieldNames       []string         `json:"field_names"`
	FieldNameRegexps []*regexp.Regexp `json:"field_name_regexps"`
}

func NewServiceAgentConfig() *ServiceAgentConfig {
	return &ServiceAgentConfig{
		FormatVersion:  ServiceAgentConfigVersion,
		FieldsToRedact: NewFieldRedactionConfig(),
	}
}

func NewFieldRedactionConfig() *FieldRedactionConfig {
	return &FieldRedactionConfig{
		FieldNames:       []string{},
		FieldNameRegexps: []*regexp.Regexp{},
	}
}

// Returns a deep copy of this configuration.
func (config *ServiceAgentConfig) Clone() *ServiceAgentConfig {
	if config == nil {
		return nil
	}

	return &ServiceAgentConfig{
		FormatVersion:  config.FormatVersion,
		FieldsToRedact: config.FieldsToRedact.Clone(),
	}
}

// Returns a deep copy of this configuration.
func (config *FieldRedactionConfig) Clone() *FieldRedactionConfig {
	if config == nil {
		return nil
	}

	return &FieldRedactionConfig{
		FieldNames:       append([]string{}, config.FieldNames...),
		FieldNameRegexps: append([]*regexp.Regexp{}, config.FieldNameRegexps...),
	}
}
