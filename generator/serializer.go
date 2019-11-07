package generator

import (
	"encoding/gob"
	"io"
)

func init() {
	// register Prompt interface types
	gob.Register(&input{})
	gob.Register(&integer{})
	gob.Register(&confirm{})
	gob.Register(&list{})
	gob.Register(&choice{})
	gob.Register(&multiChoice{})
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