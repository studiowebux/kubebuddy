package storage

import (
	"context"
	"time"

	"github.com/studiowebux/kubebuddy/internal/domain"
)

// Storage is the main storage interface
type Storage interface {
	Close() error
	Computes() ComputeRepository
	Services() ServiceRepository
	Assignments() AssignmentRepository
	Journal() JournalRepository
	APIKeys() APIKeyRepository
	Components() ComponentRepository
	ComputeComponents() ComputeComponentRepository
}

// ComputeRepository handles compute resource persistence
type ComputeRepository interface {
	Create(ctx context.Context, compute *domain.Compute) error
	Get(ctx context.Context, id string) (*domain.Compute, error)
	GetByNameProviderRegionType(ctx context.Context, name, provider, region, computeType string) (*domain.Compute, error)
	List(ctx context.Context, filters ComputeFilters) ([]*domain.Compute, error)
	Update(ctx context.Context, compute *domain.Compute) error
	Delete(ctx context.Context, id string) error
}

// ComputeFilters for querying computes
type ComputeFilters struct {
	Type     string
	Provider string
	Region   string
	State    string
	Tags     map[string]string
}

// ServiceRepository handles service persistence
type ServiceRepository interface {
	Create(ctx context.Context, service *domain.Service) error
	Get(ctx context.Context, id string) (*domain.Service, error)
	GetByName(ctx context.Context, name string) (*domain.Service, error)
	List(ctx context.Context) ([]*domain.Service, error)
	Update(ctx context.Context, service *domain.Service) error
	Delete(ctx context.Context, id string) error
}

// AssignmentRepository handles assignment persistence
type AssignmentRepository interface {
	Create(ctx context.Context, assignment *domain.Assignment) error
	Get(ctx context.Context, id string) (*domain.Assignment, error)
	GetByComputeAndService(ctx context.Context, computeID, serviceID string) (*domain.Assignment, error)
	List(ctx context.Context, filters AssignmentFilters) ([]*domain.Assignment, error)
	Update(ctx context.Context, assignment *domain.Assignment) error
	Delete(ctx context.Context, id string) error
	DeleteByService(ctx context.Context, serviceID string) error
	DeleteByCompute(ctx context.Context, computeID string) error
}

// AssignmentFilters for querying assignments
type AssignmentFilters struct {
	ServiceID string
	ComputeID string
}

// JournalRepository handles journal entry persistence
type JournalRepository interface {
	Create(ctx context.Context, entry *domain.JournalEntry) error
	Get(ctx context.Context, id string) (*domain.JournalEntry, error)
	List(ctx context.Context, filters JournalFilters) ([]*domain.JournalEntry, error)
	Delete(ctx context.Context, id string) error
}

// JournalFilters for querying journal entries
type JournalFilters struct {
	ComputeID string
	Category  string
	From      *time.Time
	To        *time.Time
	Limit     int
}

// APIKeyRepository handles API key persistence
type APIKeyRepository interface {
	Create(ctx context.Context, key *domain.APIKey) error
	Get(ctx context.Context, id string) (*domain.APIKey, error)
	GetByKeyHash(ctx context.Context, keyHash string) (*domain.APIKey, error)
	List(ctx context.Context) ([]*domain.APIKey, error)
	Delete(ctx context.Context, id string) error
}

// ComponentRepository handles component persistence
type ComponentRepository interface {
	Create(ctx context.Context, component *domain.Component) error
	Get(ctx context.Context, id string) (*domain.Component, error)
	GetByManufacturerAndModel(ctx context.Context, manufacturer, model string) (*domain.Component, error)
	List(ctx context.Context, filters ComponentFilters) ([]*domain.Component, error)
	Update(ctx context.Context, component *domain.Component) error
	Delete(ctx context.Context, id string) error
}

// ComponentFilters for querying components
type ComponentFilters struct {
	Type         string
	Manufacturer string
}

// ComputeComponentFilters for querying component assignments
type ComputeComponentFilters struct {
	ComputeID   string
	ComponentID string
}

// ComputeComponentRepository handles compute-component assignment persistence
type ComputeComponentRepository interface {
	Assign(ctx context.Context, assignment *domain.ComputeComponent) error
	Unassign(ctx context.Context, id string) error
	ListByCompute(ctx context.Context, computeID string) ([]*domain.ComputeComponent, error)
	ListByComponent(ctx context.Context, componentID string) ([]*domain.ComputeComponent, error)
}
