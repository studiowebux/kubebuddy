package sqlite

import (
	"context"
	"database/sql"
	"fmt"

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

	// Run migrations
	for version, migration := range migrations {
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
}
