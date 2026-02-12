package repository

type RepositoryKeys interface {
	ValidKey(key string, rate int) bool
}
