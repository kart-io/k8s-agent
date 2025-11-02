// Package serializers provides various serialization implementations.
package serializers

import (
	"encoding/json"
)

// JSONSerializer implements JSON serialization using encoding/json.
type JSONSerializer struct{}

// NewJSONSerializer creates a new JSON serializer.
func NewJSONSerializer() *JSONSerializer {
	return &JSONSerializer{}
}

// Marshal serializes a value to JSON bytes.
func (s *JSONSerializer) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// Unmarshal deserializes JSON bytes to a value.
func (s *JSONSerializer) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// Name returns the serializer name.
func (s *JSONSerializer) Name() string {
	return "json"
}
