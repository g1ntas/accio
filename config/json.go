package config

type GeneratorJSON struct {
	Dest        string `json:"dest"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type RepositoryJSON struct {
	Origin     string                    `json:"origin"`
	Generators map[string]*GeneratorJSON `json:"generators"`
}

type RegistryJSON struct {
	Repos []*RepositoryJSON `json:"repositories"`
}