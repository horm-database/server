package codec

import (
	"github.com/horm-database/common/json"
)

func init() {
	RegisterSerializer(SerializationTypeJSON, &JSONSerialization{})
}

// JSONSerialization provides json serialization mode.
type JSONSerialization struct{}

// Deserialize json unmarshal the input bytes into body.
func (s *JSONSerialization) Deserialize(in []byte, body interface{}) error {
	return json.Api.Unmarshal(in, body)
}

// Serialize json marshal the body into output bytes
func (s *JSONSerialization) Serialize(body interface{}) ([]byte, error) {
	return json.MarshalBase(body)
}
