// Package serializers provides various serialization implementations.
package serializers

import (
	"github.com/vmihailenco/msgpack/v5"
)

// MsgpackSerializer implements Msgpack serialization using vmihailenco/msgpack.
// Msgpack is a binary serialization format that is faster and more compact than JSON.
// Typical performance: 3-5x faster than JSON, 30-50% smaller size.
type MsgpackSerializer struct{}

// NewMsgpackSerializer creates a new Msgpack serializer.
func NewMsgpackSerializer() *MsgpackSerializer {
	return &MsgpackSerializer{}
}

// Marshal serializes a value to Msgpack bytes.
func (s *MsgpackSerializer) Marshal(v interface{}) ([]byte, error) {
	return msgpack.Marshal(v)
}

// Unmarshal deserializes Msgpack bytes to a value.
func (s *MsgpackSerializer) Unmarshal(data []byte, v interface{}) error {
	return msgpack.Unmarshal(data, v)
}

// Name returns the serializer name.
func (s *MsgpackSerializer) Name() string {
	return "msgpack"
}
