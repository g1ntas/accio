package main

type registry struct {
	Repos []*LocalRepository
}

type LocalRepository struct {
	origin string
}

// registry := loadRegistry(registryPath)
// repo := NewLocalRepo("~/code/symfony-crud")
// repo.Clone(registry.Dir)
// registry.add(repo)
// generators = ParseGenerators(repo)
// saveRegistry(registry)

var Registry *registry

func init() {
	// todo: load or create registry
	var err error
	Registry, err = loadRegistry("")
	if err != nil {
		// todo: do sth
	}
}

func loadRegistry(path string) (*registry, error) {
	// todo: check if file exists
	// todo: 	create if doesnt
	// todo: 	load data if it does
}

func (r *registry) Save() error {

}

func (r *registry) add(repo *LocalRepository) {
	r.Repos = append(r.Repos, repo)
}

func NewLocalRepo(origin string) *LocalRepository {
	return &LocalRepository{origin: origin}
}

func (r *LocalRepository) Origin() string {
	return r.origin
}

func (r *LocalRepository) Dest() string {
	return r.origin
}