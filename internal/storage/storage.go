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
	IPAddresses() IPAddressRepository
	ComputeIPs() ComputeIPRepository
	DNSRecords() DNSRecordRepository
	PortAssignments() PortAssignmentRepository
	FirewallRules() FirewallRuleRepository
	ComputeFirewallRules() ComputeFirewallRuleRepository
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

// IPAddressRepository handles IP address persistence
type IPAddressRepository interface {
	Create(ctx context.Context, ip *domain.IPAddress) error
	Get(ctx context.Context, id string) (*domain.IPAddress, error)
	GetByAddress(ctx context.Context, address string) (*domain.IPAddress, error)
	List(ctx context.Context, filters IPAddressFilters) ([]*domain.IPAddress, error)
	Update(ctx context.Context, ip *domain.IPAddress) error
	Delete(ctx context.Context, id string) error
}

// IPAddressFilters for querying IP addresses
type IPAddressFilters struct {
	Type     string
	Provider string
	Region   string
	State    string
}

// ComputeIPRepository handles IP address assignments to computes
type ComputeIPRepository interface {
	Assign(ctx context.Context, assignment *domain.ComputeIP) error
	Unassign(ctx context.Context, id string) error
	UnassignByIP(ctx context.Context, ipID string) error
	GetByComputeAndIP(ctx context.Context, computeID, ipID string) (*domain.ComputeIP, error)
	List(ctx context.Context) ([]*domain.ComputeIP, error)
	ListByCompute(ctx context.Context, computeID string) ([]*domain.ComputeIP, error)
	ListByIP(ctx context.Context, ipID string) ([]*domain.ComputeIP, error)
	GetPrimaryIP(ctx context.Context, computeID string) (*domain.ComputeIP, error)
	UpdatePrimary(ctx context.Context, id string, isPrimary bool) error
}

// DNSRecordRepository handles DNS record persistence
type DNSRecordRepository interface {
	Create(ctx context.Context, record *domain.DNSRecord) error
	Get(ctx context.Context, id string) (*domain.DNSRecord, error)
	GetByNameTypeZone(ctx context.Context, name, recordType, zone string) (*domain.DNSRecord, error)
	List(ctx context.Context, filters DNSRecordFilters) ([]*domain.DNSRecord, error)
	Update(ctx context.Context, record *domain.DNSRecord) error
	Delete(ctx context.Context, id string) error
}

// DNSRecordFilters for querying DNS records
type DNSRecordFilters struct {
	Type   string
	Zone   string
	IPID   string
	Name   string
}

// PortAssignmentRepository handles port assignment persistence
type PortAssignmentRepository interface {
	Create(ctx context.Context, assignment *domain.PortAssignment) error
	Get(ctx context.Context, id string) (*domain.PortAssignment, error)
	GetByIPPortProtocol(ctx context.Context, ipID string, port int, protocol string) (*domain.PortAssignment, error)
	List(ctx context.Context, filters PortAssignmentFilters) ([]*domain.PortAssignment, error)
	Update(ctx context.Context, assignment *domain.PortAssignment) error
	Delete(ctx context.Context, id string) error
	DeleteByAssignment(ctx context.Context, assignmentID string) error
}

// PortAssignmentFilters for querying port assignments
type PortAssignmentFilters struct {
	AssignmentID string
	IPID         string
	Protocol     string
}

// FirewallRuleRepository handles firewall rule persistence
type FirewallRuleRepository interface {
	Create(ctx context.Context, rule *domain.FirewallRule) error
	Get(ctx context.Context, id string) (*domain.FirewallRule, error)
	GetByName(ctx context.Context, name string) (*domain.FirewallRule, error)
	List(ctx context.Context, filters FirewallRuleFilters) ([]*domain.FirewallRule, error)
	Update(ctx context.Context, rule *domain.FirewallRule) error
	Delete(ctx context.Context, id string) error
}

// FirewallRuleFilters for querying firewall rules
type FirewallRuleFilters struct {
	Action   string
	Protocol string
}

// ComputeFirewallRuleRepository handles firewall rule assignments to computes
type ComputeFirewallRuleRepository interface {
	Assign(ctx context.Context, assignment *domain.ComputeFirewallRule) error
	Unassign(ctx context.Context, id string) error
	ListByCompute(ctx context.Context, computeID string) ([]*domain.ComputeFirewallRule, error)
	ListByRule(ctx context.Context, ruleID string) ([]*domain.ComputeFirewallRule, error)
	UpdateEnabled(ctx context.Context, id string, enabled bool) error
}
