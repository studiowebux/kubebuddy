package domain

import (
	"fmt"
	"time"
)

// IPType represents the type of IP address
type IPType string

const (
	IPTypePublic  IPType = "public"
	IPTypePrivate IPType = "private"
)

// IPState represents the state of an IP address
type IPState string

const (
	IPStateAvailable IPState = "available"
	IPStateAssigned  IPState = "assigned"
	IPStateReserved  IPState = "reserved"
)

// IPAddress represents an IP address resource
type IPAddress struct {
	ID         string    `json:"id"`
	Address    string    `json:"address"`
	Type       IPType    `json:"type"`
	CIDR       string    `json:"cidr"`
	Gateway    string    `json:"gateway,omitempty"`
	DNSServers []string  `json:"dns_servers,omitempty"`
	Provider   string    `json:"provider"`
	Region     string    `json:"region"`
	VLAN       string    `json:"vlan,omitempty"`
	Notes      string    `json:"notes,omitempty"`
	State      IPState   `json:"state"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// ComputeIP represents an IP assignment to a compute
type ComputeIP struct {
	ID            string    `json:"id"`
	ComputeID     string    `json:"compute_id"`
	IPID          string    `json:"ip_id"`
	InterfaceName string    `json:"interface_name,omitempty"`
	IsPrimary     bool      `json:"is_primary"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// DNSRecordType represents DNS record types
type DNSRecordType string

const (
	DNSRecordTypeA     DNSRecordType = "A"
	DNSRecordTypeAAAA  DNSRecordType = "AAAA"
	DNSRecordTypeCNAME DNSRecordType = "CNAME"
	DNSRecordTypePTR   DNSRecordType = "PTR"
)

// DNSRecord represents a DNS record
type DNSRecord struct {
	ID        string        `json:"id"`
	Name      string        `json:"name"`
	Type      DNSRecordType `json:"type"`
	Value     string        `json:"value"`
	IPID      string        `json:"ip_id,omitempty"` // Optional link to IP
	TTL       int           `json:"ttl"`
	Zone      string        `json:"zone"`
	Notes     string        `json:"notes,omitempty"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
}

// Protocol represents network protocols
type Protocol string

const (
	ProtocolTCP  Protocol = "tcp"
	ProtocolUDP  Protocol = "udp"
	ProtocolICMP Protocol = "icmp"
	ProtocolAll  Protocol = "all"
)

// PortAssignment represents a port mapping
type PortAssignment struct {
	ID           string    `json:"id"`
	AssignmentID string    `json:"assignment_id"`
	IPID         string    `json:"ip_id"`
	Port         int       `json:"port"`
	Protocol     Protocol  `json:"protocol"`
	ServicePort  int       `json:"service_port"` // Original service port
	Description  string    `json:"description,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

// FirewallAction represents firewall rule actions
type FirewallAction string

const (
	FirewallActionAllow FirewallAction = "ALLOW"
	FirewallActionDeny  FirewallAction = "DENY"
)

// FirewallRule represents a firewall rule definition
type FirewallRule struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Action      FirewallAction `json:"action"`
	Protocol    Protocol       `json:"protocol"`
	Source      string         `json:"source"`      // CIDR, IP, or "any"
	Destination string         `json:"destination"` // CIDR, IP, or "any"
	PortStart   *int           `json:"port_start,omitempty"`
	PortEnd     *int           `json:"port_end,omitempty"`
	Description string         `json:"description,omitempty"`
	Priority    int            `json:"priority"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// ComputeFirewallRule represents a firewall rule assignment to a compute
type ComputeFirewallRule struct {
	ID        string    `json:"id"`
	ComputeID string    `json:"compute_id"`
	RuleID    string    `json:"rule_id"`
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
}

// IsPortRange checks if the firewall rule is for a port range
func (f *FirewallRule) IsPortRange() bool {
	return f.PortStart != nil && f.PortEnd != nil && *f.PortEnd > *f.PortStart
}

// IsSinglePort checks if the firewall rule is for a single port
func (f *FirewallRule) IsSinglePort() bool {
	return f.PortStart != nil && (f.PortEnd == nil || *f.PortEnd == *f.PortStart)
}

// GetPortRange returns the port range as a string
func (f *FirewallRule) GetPortRange() string {
	if f.PortStart == nil {
		return "any"
	}
	if f.PortEnd == nil || *f.PortEnd == *f.PortStart {
		return fmt.Sprintf("%d", *f.PortStart)
	}
	return fmt.Sprintf("%d-%d", *f.PortStart, *f.PortEnd)
}

// PortRequirement defines port requirements for a service
type PortRequirement struct {
	Port        int      `json:"port"`
	Protocol    Protocol `json:"protocol"`
	Description string   `json:"description,omitempty"`
}
