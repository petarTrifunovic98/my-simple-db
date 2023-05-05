package serialization

import (
	"encoding/json"
	"io"
)

func Serialize(source any, destination io.Writer) {
	encoder := json.NewEncoder(destination)
	encoder.Encode(source)
	// encoder := gob.NewEncoder(destination)
	// encoder.Encode(source)
}

func Deserialize(source io.Reader, destination any) error {
	decoder := json.NewDecoder(source)
	return decoder.Decode(destination)
	// decoder := gob.NewDecoder(source)
	// return decoder.Decode(destination)
}
