package generator

import (
	"encoding/gob"
	"io"
)

func init() {
	// register prompt interface types
	gob.Register(&Input{})
	gob.Register(&Integer{})
	gob.Register(&Confirm{})
	gob.Register(&List{})
	gob.Register(&Choice{})
	gob.Register(&MultiChoice{})
}

func Deserialize(r io.Reader) (*Registry, error) {
	reg := NewRegistry()
	dec := gob.NewDecoder(r)
	err := dec.Decode(reg)
	if err != nil {
		return nil, err
	}
	return reg, err
}

func Serialize(w io.Writer, reg *Registry) error {
	enc := gob.NewEncoder(w)
	err := enc.Encode(reg)
	if err != nil {
		return err
	}
	return nil
}