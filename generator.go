package main

type Generator struct {
	Name string
	Description string
	Dest string
	Prompts []interface{}
}

func ParseGenerators(repo *LocalRepository) {
	// todo: check if exists and parse .accio.yml config in repo.Dest()
	// todo: parse config and create Generator struct from it
	// todo: return all parsed generators
}