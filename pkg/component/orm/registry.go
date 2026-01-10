package orm

import (
	"context"
	"fmt"
	"strings"
)

// ExistsChecker provides existence checking for entities.
// All BaseRepositoryInterface[T] implement this interface.
type ExistsChecker interface {
	Exists(ctx context.Context, conditions map[string]any) (bool, error)
	EntityName() string
}

// RepositoryRegistry stores repositories for lookup by entity name.
type RepositoryRegistry struct {
	repositories map[string]ExistsChecker
}

// NewRepositoryRegistry creates a new repository registry.
func NewRepositoryRegistry() *RepositoryRegistry {
	return &RepositoryRegistry{
		repositories: make(map[string]ExistsChecker),
	}
}

// NewRepositoryRegistryFromRepos creates a repository registry from a slice of repositories.
func NewRepositoryRegistryFromRepos(repos []ExistsChecker) *RepositoryRegistry {
	registry := &RepositoryRegistry{
		repositories: make(map[string]ExistsChecker),
	}
	for _, repo := range repos {
		normalizedEntityName := strings.ToLower(repo.EntityName())
		registry.repositories[normalizedEntityName] = repo
	}

	return registry
}

// Register adds a repository to the registry.
func Register[T any](r *RepositoryRegistry, repo BaseRepositoryInterface[T]) {
	normalizedEntityName := strings.ToLower(repo.EntityName())
	r.repositories[normalizedEntityName] = repo
}

// GetRepository retrieves a repository by entity name for type-safe access.
func GetRepository[T any](r *RepositoryRegistry, entityName string) (BaseRepositoryInterface[T], error) {
	normalizedEntityName := strings.ToLower(entityName)
	repo, exists := r.repositories[normalizedEntityName]
	if !exists {
		return nil, fmt.Errorf("repository for entity %s not found", entityName)
	}

	typedRepo, ok := repo.(BaseRepositoryInterface[T])
	if !ok {
		return nil, fmt.Errorf("repository for entity %s does not implement BaseRepositoryInterface[%T]", entityName, *new(T))
	}

	return typedRepo, nil
}

// GetExistsChecker retrieves a repository by entity name for existence checking.
func (r *RepositoryRegistry) GetExistsChecker(entityName string) (ExistsChecker, error) {
	normalizedEntityName := strings.ToLower(entityName)
	repo, exists := r.repositories[normalizedEntityName]
	if !exists {
		return nil, fmt.Errorf("repository for entity %s not found", entityName)
	}

	return repo, nil
}
