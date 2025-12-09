package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"sort"

	_ "github.com/mattn/go-sqlite3"
	"github.com/studiowebux/kubebuddy/internal/storage"
)

// SQLiteStorage implements the Storage interface
type SQLiteStorage struct {
	db *sql.DB

	computes          *computeRepo
	services          *serviceRepo
	assignments       *assignmentRepo
	journal           *journalRepo
	apikeys           *apikeyRepo
	components        *componentRepo
	computeComponents *computeComponentRepo
	ipAddresses          *ipAddressRepo
	computeIPs           *computeIPRepo
	dnsRecords           *dnsRecordRepo
	portAssignments      *portAssignmentRepo
	firewallRules        *firewallRuleRepo
	computeFirewallRules *computeFirewallRuleRepo
}

// New creates a new SQLite storage instance
func New(dataSourceName string) (storage.Storage, error) {
	db, err := sql.Open("sqlite3", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	s := &SQLiteStorage{
		db: db,
	}

	// Initialize repositories
	s.computes = &computeRepo{db: db}
	s.services = &serviceRepo{db: db}
	s.assignments = &assignmentRepo{db: db}
	s.journal = &journalRepo{db: db}
	s.apikeys = &apikeyRepo{db: db}
	s.components = &componentRepo{db: db}
	s.computeComponents = &computeComponentRepo{db: db}
	s.ipAddresses = &ipAddressRepo{db: db}
	s.computeIPs = &computeIPRepo{db: db}
	s.dnsRecords = &dnsRecordRepo{db: db}
	s.portAssignments = &portAssignmentRepo{db: db}
	s.firewallRules = &firewallRuleRepo{db: db}
	s.computeFirewallRules = &computeFirewallRuleRepo{db: db}

	// Run migrations
	if err := s.migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return s, nil
}

// Close closes the database connection
func (s *SQLiteStorage) Close() error {
	return s.db.Close()
}

// Computes returns the compute repository
func (s *SQLiteStorage) Computes() storage.ComputeRepository {
	return s.computes
}

// Services returns the service repository
func (s *SQLiteStorage) Services() storage.ServiceRepository {
	return s.services
}

// Assignments returns the assignment repository
func (s *SQLiteStorage) Assignments() storage.AssignmentRepository {
	return s.assignments
}

// Journal returns the journal repository
func (s *SQLiteStorage) Journal() storage.JournalRepository {
	return s.journal
}

// APIKeys returns the API key repository
func (s *SQLiteStorage) APIKeys() storage.APIKeyRepository {
	return s.apikeys
}

// Components returns the component repository
func (s *SQLiteStorage) Components() storage.ComponentRepository {
	return s.components
}

// ComputeComponents returns the compute-component assignment repository
func (s *SQLiteStorage) ComputeComponents() storage.ComputeComponentRepository {
	return s.computeComponents
}

// IPAddresses returns the IP address repository
func (s *SQLiteStorage) IPAddresses() storage.IPAddressRepository {
	return s.ipAddresses
}

// ComputeIPs returns the compute-IP assignment repository
func (s *SQLiteStorage) ComputeIPs() storage.ComputeIPRepository {
	return s.computeIPs
}

// DNSRecords returns the DNS record repository
func (s *SQLiteStorage) DNSRecords() storage.DNSRecordRepository {
	return s.dnsRecords
}

// PortAssignments returns the port assignment repository
func (s *SQLiteStorage) PortAssignments() storage.PortAssignmentRepository {
	return s.portAssignments
}

// FirewallRules returns the firewall rule repository
func (s *SQLiteStorage) FirewallRules() storage.FirewallRuleRepository {
	return s.firewallRules
}

// ComputeFirewallRules returns the compute-firewall rule assignment repository
func (s *SQLiteStorage) ComputeFirewallRules() storage.ComputeFirewallRuleRepository {
	return s.computeFirewallRules
}

// migrate runs database migrations
func (s *SQLiteStorage) migrate() error {
	ctx := context.Background()

	// Create migrations table
	_, err := s.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS migrations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			version INTEGER UNIQUE NOT NULL,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get sorted migration versions
	versions := make([]int, 0, len(migrations))
	for version := range migrations {
		versions = append(versions, version)
	}
	sort.Ints(versions)

	// Run migrations in order
	for _, version := range versions {
		migration := migrations[version]

		// Check if migration already applied
		var count int
		err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM migrations WHERE version = ?", version).Scan(&count)
		if err != nil {
			return fmt.Errorf("failed to check migration version %d: %w", version, err)
		}

		if count > 0 {
			continue
		}

		// Run migration
		if _, err := s.db.ExecContext(ctx, migration); err != nil {
			return fmt.Errorf("failed to run migration version %d: %w", version, err)
		}

		// Mark as applied
		if _, err := s.db.ExecContext(ctx, "INSERT INTO migrations (version) VALUES (?)", version); err != nil {
			return fmt.Errorf("failed to mark migration version %d as applied: %w", version, err)
		}
	}

	return nil
}

// migrations contains all database schema migrations
var migrations = map[int]string{
	1: `
		-- Computes table
		CREATE TABLE computes (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			type TEXT NOT NULL,
			provider TEXT NOT NULL,
			region TEXT NOT NULL,
			tags TEXT NOT NULL, -- JSON
			state TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL
		);

		CREATE INDEX idx_computes_type ON computes(type);
		CREATE INDEX idx_computes_provider ON computes(provider);
		CREATE INDEX idx_computes_region ON computes(region);
		CREATE INDEX idx_computes_state ON computes(state);
	`,
	2: `
		-- Services table
		CREATE TABLE services (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			min_spec TEXT NOT NULL, -- JSON
			max_spec TEXT NOT NULL, -- JSON
			placement TEXT NOT NULL, -- JSON
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL
		);
	`,
	3: `
		-- Assignments table
		CREATE TABLE assignments (
			id TEXT PRIMARY KEY,
			service_id TEXT NOT NULL,
			compute_id TEXT NOT NULL,
			allocated TEXT NOT NULL, -- JSON
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL,
			FOREIGN KEY (service_id) REFERENCES services(id) ON DELETE CASCADE,
			FOREIGN KEY (compute_id) REFERENCES computes(id) ON DELETE CASCADE
		);

		CREATE INDEX idx_assignments_service ON assignments(service_id);
		CREATE INDEX idx_assignments_compute ON assignments(compute_id);
	`,
	4: `
		-- Journal entries table
		CREATE TABLE journal_entries (
			id TEXT PRIMARY KEY,
			compute_id TEXT NOT NULL,
			category TEXT NOT NULL,
			content TEXT NOT NULL,
			created_by TEXT DEFAULT '',
			created_at TIMESTAMP NOT NULL,
			FOREIGN KEY (compute_id) REFERENCES computes(id) ON DELETE CASCADE
		);

		CREATE INDEX idx_journal_compute ON journal_entries(compute_id);
		CREATE INDEX idx_journal_category ON journal_entries(category);
		CREATE INDEX idx_journal_created ON journal_entries(created_at);
	`,
	5: `
		-- API keys table
		CREATE TABLE api_keys (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			key_hash TEXT NOT NULL UNIQUE,
			scope TEXT NOT NULL,
			description TEXT,
			created_by TEXT,
			created_at TIMESTAMP NOT NULL,
			expires_at TIMESTAMP
		);

		CREATE INDEX idx_apikeys_scope ON api_keys(scope);
	`,
	6: `
		-- Components table
		CREATE TABLE components (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			type TEXT NOT NULL,
			manufacturer TEXT NOT NULL,
			model TEXT NOT NULL,
			specs TEXT NOT NULL, -- JSON
			notes TEXT,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL
		);

		CREATE INDEX idx_components_type ON components(type);
		CREATE INDEX idx_components_manufacturer ON components(manufacturer);
	`,
	7: `
		-- Compute-Component assignments table
		CREATE TABLE compute_components (
			id TEXT PRIMARY KEY,
			compute_id TEXT NOT NULL,
			component_id TEXT NOT NULL,
			quantity INTEGER NOT NULL DEFAULT 1,
			slot TEXT,
			serial_no TEXT,
			notes TEXT,
			raid_level TEXT DEFAULT '',
			raid_group TEXT DEFAULT '',
			created_at TIMESTAMP NOT NULL,
			FOREIGN KEY (compute_id) REFERENCES computes(id) ON DELETE CASCADE,
			FOREIGN KEY (component_id) REFERENCES components(id) ON DELETE CASCADE
		);

		CREATE INDEX idx_compute_components_compute ON compute_components(compute_id);
		CREATE INDEX idx_compute_components_component ON compute_components(component_id);
	`,
	8: `
		-- IP addresses table
		CREATE TABLE ip_addresses (
			id TEXT PRIMARY KEY,
			address TEXT NOT NULL UNIQUE,
			type TEXT NOT NULL,
			cidr TEXT NOT NULL,
			gateway TEXT,
			dns_servers TEXT,
			provider TEXT NOT NULL,
			region TEXT NOT NULL,
			notes TEXT,
			state TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL
		);

		CREATE INDEX idx_ip_addresses_type ON ip_addresses(type);
		CREATE INDEX idx_ip_addresses_provider ON ip_addresses(provider);
		CREATE INDEX idx_ip_addresses_region ON ip_addresses(region);
		CREATE INDEX idx_ip_addresses_state ON ip_addresses(state);

		-- Compute-IP assignments table
		CREATE TABLE compute_ips (
			id TEXT PRIMARY KEY,
			compute_id TEXT NOT NULL,
			ip_id TEXT NOT NULL,
			is_primary INTEGER DEFAULT 0,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL,
			FOREIGN KEY (compute_id) REFERENCES computes(id) ON DELETE CASCADE,
			FOREIGN KEY (ip_id) REFERENCES ip_addresses(id) ON DELETE CASCADE
		);

		CREATE INDEX idx_compute_ips_compute ON compute_ips(compute_id);
		CREATE INDEX idx_compute_ips_ip ON compute_ips(ip_id);
		CREATE UNIQUE INDEX idx_compute_ips_unique ON compute_ips(compute_id, ip_id);
	`,
	9: `
		-- DNS records table
		CREATE TABLE dns_records (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			type TEXT NOT NULL,
			value TEXT NOT NULL,
			ip_id TEXT,
			ttl INTEGER DEFAULT 3600,
			zone TEXT NOT NULL,
			notes TEXT,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL,
			FOREIGN KEY (ip_id) REFERENCES ip_addresses(id) ON DELETE SET NULL
		);

		CREATE INDEX idx_dns_records_name ON dns_records(name);
		CREATE INDEX idx_dns_records_type ON dns_records(type);
		CREATE INDEX idx_dns_records_zone ON dns_records(zone);
		CREATE INDEX idx_dns_records_ip ON dns_records(ip_id);
		CREATE UNIQUE INDEX idx_dns_records_unique ON dns_records(name, type, zone);
	`,
	10: `
		-- Port assignments table
		CREATE TABLE port_assignments (
			id TEXT PRIMARY KEY,
			assignment_id TEXT NOT NULL,
			ip_id TEXT NOT NULL,
			port INTEGER NOT NULL,
			protocol TEXT NOT NULL,
			service_port INTEGER NOT NULL,
			description TEXT,
			created_at TIMESTAMP NOT NULL,
			FOREIGN KEY (assignment_id) REFERENCES assignments(id) ON DELETE CASCADE,
			FOREIGN KEY (ip_id) REFERENCES ip_addresses(id) ON DELETE CASCADE
		);

		CREATE INDEX idx_port_assignments_assignment ON port_assignments(assignment_id);
		CREATE INDEX idx_port_assignments_ip ON port_assignments(ip_id);
		CREATE INDEX idx_port_assignments_port ON port_assignments(port);
		CREATE UNIQUE INDEX idx_port_assignments_unique ON port_assignments(ip_id, port, protocol);
	`,
	11: `
		-- Firewall rules table
		CREATE TABLE firewall_rules (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			action TEXT NOT NULL,
			protocol TEXT NOT NULL,
			source TEXT NOT NULL,
			destination TEXT NOT NULL,
			port_start INTEGER,
			port_end INTEGER,
			description TEXT,
			priority INTEGER DEFAULT 100,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL
		);

		CREATE INDEX idx_firewall_rules_name ON firewall_rules(name);
		CREATE INDEX idx_firewall_rules_priority ON firewall_rules(priority);

		-- Compute-firewall rule assignments table
		CREATE TABLE compute_firewall_rules (
			id TEXT PRIMARY KEY,
			compute_id TEXT NOT NULL,
			rule_id TEXT NOT NULL,
			enabled INTEGER DEFAULT 1,
			created_at TIMESTAMP NOT NULL,
			FOREIGN KEY (compute_id) REFERENCES computes(id) ON DELETE CASCADE,
			FOREIGN KEY (rule_id) REFERENCES firewall_rules(id) ON DELETE CASCADE
		);

		CREATE INDEX idx_compute_firewall_rules_compute ON compute_firewall_rules(compute_id);
		CREATE INDEX idx_compute_firewall_rules_rule ON compute_firewall_rules(rule_id);
		CREATE UNIQUE INDEX idx_compute_firewall_rules_unique ON compute_firewall_rules(compute_id, rule_id);
	`,
	12: `
		-- This migration is no longer needed as migration 8 already creates compute_ips with updated_at
		-- Keeping for backwards compatibility with existing databases
		-- No-op migration
		SELECT 1;
	`,
	13: `
		-- Add billing fields to computes table
		ALTER TABLE computes ADD COLUMN monthly_cost REAL;
		ALTER TABLE computes ADD COLUMN annual_cost REAL;
		ALTER TABLE computes ADD COLUMN contract_end_date TIMESTAMP;
		ALTER TABLE computes ADD COLUMN next_renewal_date TIMESTAMP;
	`,
}
