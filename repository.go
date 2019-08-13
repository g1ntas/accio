package main

type Registry struct {
	Repos []*LocalRepository
}

type LocalRepository struct {
	origin string
}

// registry := loadRegistry(registryPath)
// repo := NewLocalRepo("~/code/symfony-crud")
// registry.addRepo(repo)
// generators = ParseGenerators(repo)
// registry.save()

func NewLocalRepo(origin string) *LocalRepository {
	return &LocalRepository{origin}
}

func (r *LocalRepository) Origin() string {
	return r.origin
}

func (r *LocalRepository) Dest() string {
	return r.origin
}



