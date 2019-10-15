package gob

import (
	"encoding/gob"
	"io"

	"github.com/g1ntas/accio/generators"
)

func init() {
	// register prompt interface types
	gob.Register(&generators.Input{})
	gob.Register(&generators.Integer{})
	gob.Register(&generators.Confirm{})
	gob.Register(&generators.List{})
	gob.Register(&generators.Choice{})
	gob.Register(&generators.MultiChoice{})
	gob.Register(&generators.File{})
}

func Unserialize(r io.Reader) (*generators.Registry, error) {
	reg := generators.NewRegistry()
	dec := gob.NewDecoder(r)
	err := dec.Decode(reg)
	if err != nil {
		return nil, err
	}
	return reg, err
}

func Serialize(w io.Writer, reg *generators.Registry) error {
	enc := gob.NewEncoder(w)
	err := enc.Encode(reg)
	if err != nil {
		return err
	}
	return nil
}