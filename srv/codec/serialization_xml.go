package codec

import (
	"fmt"

	"github.com/horm-database/common/types"
)

func init() {
	RegisterSerializer(SerializationTypeXML, &XMLSerialization{})
}

// XMLSerialization export xml body
type XMLSerialization struct{}

// Deserialize the input xml into body
func (j *XMLSerialization) Deserialize(in []byte, body interface{}) error {
	body = in
	return nil
}

// Serialize the body to output xml.
func (j *XMLSerialization) Serialize(body interface{}) ([]byte, error) {
	return types.StringToBytes(fmt.Sprintf("%v", body)), nil
}
