package serialization

import (
	"bytes"
	"encoding/gob"
)

func Serialize(source any) []byte {
	var b bytes.Buffer

	encoder := gob.NewEncoder(&b)
	encoder.Encode(source)

	return b.Bytes()
}

func Deserialize(source *bytes.Buffer, destination any) error {
	decoder := gob.NewDecoder(source)
	return decoder.Decode(destination)
}
