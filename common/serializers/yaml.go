// Package serializers provides various serialization implementations.
package serializers

import (
	"gopkg.in/yaml.v3"
)

// YAMLSerializer implements YAML serialization using gopkg.in/yaml.v3.
// YAML is a human-readable data serialization format, ideal for configuration files.
type YAMLSerializer struct{}

// NewYAMLSerializer creates a new YAML serializer.
func NewYAMLSerializer() *YAMLSerializer {
	return &YAMLSerializer{}
}

// Marshal serializes a value to YAML bytes.
func (s *YAMLSerializer) Marshal(v interface{}) ([]byte, error) {
	return yaml.Marshal(v)
}

// Unmarshal deserializes YAML bytes to a value.
func (s *YAMLSerializer) Unmarshal(data []byte, v interface{}) error {
	return yaml.Unmarshal(data, v)
}

// Name returns the serializer name.
func (s *YAMLSerializer) Name() string {
	return "yaml"
}
