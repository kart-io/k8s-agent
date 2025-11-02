// Package serializers provides serialization abstractions and implementations.
package serializers

// Serializer defines the interface for serializing and deserializing values.
// This allows components to support different serialization formats (JSON, Msgpack, YAML, Protobuf, etc.)
type Serializer interface {
	// Marshal serializes a value to bytes.
	Marshal(v interface{}) ([]byte, error)

	// Unmarshal deserializes bytes to a value.
	Unmarshal(data []byte, v interface{}) error

	// Name returns the serializer name for logging and metrics.
	Name() string
}
