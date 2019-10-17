package gob

import (
	"encoding/gob"
	"io"

	"github.com/g1ntas/accio/generator"
)

func init() {
	// register prompt interface types
	gob.Register(&generator.Input{})
	gob.Register(&generator.Integer{})
	gob.Register(&generator.Confirm{})
	gob.Register(&generator.List{})
	gob.Register(&generator.Choice{})
	gob.Register(&generator.MultiChoice{})
	gob.Register(&generator.File{})
}

func Unserialize(r io.Reader) (*generator.Registry, error) {
	reg := generator.NewRegistry()
	dec := gob.NewDecoder(r)
	err := dec.Decode(reg)
	if err != nil {
		return nil, err
	}
	return reg, err
}

func Serialize(w io.Writer, reg *generator.Registry) error {
	enc := gob.NewEncoder(w)
	err := enc.Encode(reg)
	if err != nil {
		return err
	}
	return nil
}