package main

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/studiowebux/kubebuddy/internal/api"
	"github.com/studiowebux/kubebuddy/internal/cli"
	"github.com/studiowebux/kubebuddy/internal/client"
	"github.com/studiowebux/kubebuddy/internal/domain"
	"github.com/studiowebux/kubebuddy/internal/storage"
	"github.com/studiowebux/kubebuddy/internal/storage/sqlite"
)

//go:embed webui
var webuiFS embed.FS

func main() {
	rootCmd := cli.NewRootCmd()

	// Add server command
	rootCmd.AddCommand(newServerCmd())

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func newServerCmd() *cobra.Command {
	var (
		dbPath         string
		port           string
		createAdminKey bool
		seedData       bool
		enableWebUI    bool
		webuiPort      string
	)

	cmd := &cobra.Command{
		Use:   "server",
		Short: "Start API server",
		Long: `Start the KubeBuddy API server

Environment Variables:
  KUBEBUDDY_DB                  Database file path (overridden by --db)
  KUBEBUDDY_PORT                Server port (overridden by --port)
  KUBEBUDDY_CREATE_ADMIN_KEY    Set to "true" to create admin key (overridden by --create-admin-key)
  KUBEBUDDY_SEED                Set to "true" to seed database (overridden by --seed)
  KUBEBUDDY_ADMIN_API_KEY       Required when using --create-admin-key flag`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load configuration from environment variables if not set via flags
			if !cmd.Flags().Changed("db") {
				if envDB := os.Getenv("KUBEBUDDY_DB"); envDB != "" {
					dbPath = envDB
				}
			}
			if !cmd.Flags().Changed("port") {
				if envPort := os.Getenv("KUBEBUDDY_PORT"); envPort != "" {
					port = envPort
				}
			}
			if !cmd.Flags().Changed("create-admin-key") {
				if envCreateAdmin := os.Getenv("KUBEBUDDY_CREATE_ADMIN_KEY"); envCreateAdmin == "true" {
					createAdminKey = true
				}
			}
			if !cmd.Flags().Changed("seed") {
				if envSeed := os.Getenv("KUBEBUDDY_SEED"); envSeed == "true" {
					seedData = true
				}
			}

			// Expand ~ in database path
			if strings.HasPrefix(dbPath, "~/") {
				usr, err := user.Current()
				if err != nil {
					return fmt.Errorf("failed to get current user: %w", err)
				}
				dbPath = filepath.Join(usr.HomeDir, dbPath[2:])
			}

			// If path ends with /, append default database filename
			if strings.HasSuffix(dbPath, "/") || strings.HasSuffix(dbPath, string(filepath.Separator)) {
				dbPath = filepath.Join(dbPath, "kubebuddy.db")
			}

			// Create directory if it doesn't exist
			dbDir := filepath.Dir(dbPath)
			if dbDir != "." && dbDir != "/" {
				if err := os.MkdirAll(dbDir, 0755); err != nil {
					return fmt.Errorf("failed to create database directory: %w", err)
				}
			}

			// Initialize storage
			store, err := sqlite.New(dbPath)
			if err != nil {
				return fmt.Errorf("failed to initialize database: %w", err)
			}
			defer store.Close()

			ctx := context.Background()

			// Create admin key if requested
			if createAdminKey {
				adminKey := os.Getenv("KUBEBUDDY_ADMIN_API_KEY")
				if adminKey == "" {
					return fmt.Errorf("KUBEBUDDY_ADMIN_API_KEY environment variable is required when using --create-admin-key")
				}

				fmt.Println("Creating admin API key...")
				if err := store.(*sqlite.SQLiteStorage).CreateAdminKey(ctx, adminKey); err != nil {
					return fmt.Errorf("failed to create admin key: %w", err)
				}
				fmt.Println("Admin API key created successfully")
			}

			// Seed data if requested
			if seedData {
				fmt.Println("Seeding database with sample data...")
				if err := store.(*sqlite.SQLiteStorage).Seed(ctx); err != nil {
					return fmt.Errorf("failed to seed database: %w", err)
				}
				fmt.Println("Database seeded successfully")
			}

			// Create and start API server
			server := api.NewServer(store, ":"+port)

			// Start WebUI if enabled
			var webuiServer *http.Server
			if enableWebUI {
				adminKey := os.Getenv("KUBEBUDDY_ADMIN_API_KEY")
				if adminKey == "" {
					return fmt.Errorf("KUBEBUDDY_ADMIN_API_KEY environment variable is required when using --webui")
				}

				webuiServer = startWebUI(webuiPort, "http://localhost:"+port, adminKey)
				fmt.Printf("WebUI server started on http://localhost:%s\n", webuiPort)
			}

			// Handle graceful shutdown
			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

			go func() {
				<-sigChan
				fmt.Println("\nShutting down servers...")
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				if webuiServer != nil {
					webuiServer.Shutdown(ctx)
				}

				if err := server.Shutdown(ctx); err != nil {
					fmt.Printf("Error during shutdown: %v\n", err)
				}
				os.Exit(0)
			}()

			return server.Start()
		},
	}

	cmd.Flags().StringVar(&dbPath, "db", "kubebuddy.db", "Database file path")
	cmd.Flags().StringVar(&port, "port", "8080", "Server port")
	cmd.Flags().BoolVar(&createAdminKey, "create-admin-key", false, "Create admin API key from KUBEBUDDY_ADMIN_API_KEY env var")
	cmd.Flags().BoolVar(&seedData, "seed", false, "Seed database with sample data")
	cmd.Flags().BoolVar(&enableWebUI, "webui", false, "Enable WebUI on separate port (requires KUBEBUDDY_ADMIN_API_KEY)")
	cmd.Flags().StringVar(&webuiPort, "webui-port", "8081", "WebUI server port")

	return cmd
}

// startWebUI starts the WebUI server in a goroutine
func startWebUI(port, apiEndpoint, apiKey string) *http.Server {
	mux := http.NewServeMux()

	// Serve static files
	webFS, err := fs.Sub(webuiFS, "webui")
	if err != nil {
		fmt.Printf("Failed to load web files: %v\n", err)
		return nil
	}
	mux.Handle("/", http.FileServer(http.FS(webFS)))

	// API routes
	c := client.New(apiEndpoint, apiKey)
	setupAPIRoutes(mux, c)

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      corsMiddleware(mux),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("WebUI server error: %v\n", err)
		}
	}()

	return server
}

// corsMiddleware adds CORS headers
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// setupAPIRoutes configures all API endpoints
func setupAPIRoutes(mux *http.ServeMux, c *client.Client) {
	ctx := context.Background()

	// Computes
	mux.HandleFunc("/api/computes", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			computes, err := c.ListComputes(ctx, storage.ComputeFilters{})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			respondJSON(w, computes)
		case "POST":
			var compute domain.Compute
			if err := json.NewDecoder(r.Body).Decode(&compute); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			result, err := c.CreateCompute(ctx, &compute)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			respondJSON(w, result)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/computes/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/api/computes/")

		switch r.Method {
		case "GET":
			compute, err := c.ResolveCompute(ctx, id)
			if err != nil {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
			respondJSON(w, compute)
		case "PUT":
			var compute domain.Compute
			if err := json.NewDecoder(r.Body).Decode(&compute); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			result, err := c.UpdateCompute(ctx, id, &compute)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			respondJSON(w, result)
		case "DELETE":
			if err := c.DeleteCompute(ctx, id); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			respondJSON(w, map[string]string{"message": "deleted"})
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Components
	mux.HandleFunc("/api/components", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			components, err := c.ListComponents(ctx, storage.ComponentFilters{})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			respondJSON(w, components)
		case "POST":
			var component domain.Component
			if err := json.NewDecoder(r.Body).Decode(&component); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			result, err := c.CreateComponent(ctx, &component)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			respondJSON(w, result)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Services
	mux.HandleFunc("/api/services", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			services, err := c.ListServices(ctx)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			respondJSON(w, services)
		case "POST":
			var service domain.Service
			if err := json.NewDecoder(r.Body).Decode(&service); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			result, err := c.CreateService(ctx, &service)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			respondJSON(w, result)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Assignments
	mux.HandleFunc("/api/assignments", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			assignments, err := c.ListAssignments(ctx, storage.AssignmentFilters{})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			respondJSON(w, assignments)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// IPs
	mux.HandleFunc("/api/ips", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			ips, err := c.ListIPAddresses(ctx, storage.IPAddressFilters{})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			respondJSON(w, ips)
		case "POST":
			var ip domain.IPAddress
			if err := json.NewDecoder(r.Body).Decode(&ip); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			result, err := c.CreateIPAddress(ctx, &ip)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			respondJSON(w, result)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/ips/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/api/ips/")
		switch r.Method {
		case "GET":
			ip, err := c.GetIPAddress(ctx, id)
			if err != nil {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
			respondJSON(w, ip)
		case "DELETE":
			if err := c.DeleteIPAddress(ctx, id); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			respondJSON(w, map[string]string{"message": "deleted"})
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// DNS
	mux.HandleFunc("/api/dns", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			dns, err := c.ListDNSRecords(ctx, storage.DNSRecordFilters{})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			respondJSON(w, dns)
		case "POST":
			var record domain.DNSRecord
			if err := json.NewDecoder(r.Body).Decode(&record); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			result, err := c.CreateDNSRecord(ctx, &record)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			respondJSON(w, result)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/dns/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/api/dns/")
		switch r.Method {
		case "DELETE":
			if err := c.DeleteDNSRecord(ctx, id); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			respondJSON(w, map[string]string{"message": "deleted"})
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Firewall
	mux.HandleFunc("/api/firewall", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			rules, err := c.ListFirewallRules(ctx, storage.FirewallRuleFilters{})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			respondJSON(w, rules)
		case "POST":
			var rule domain.FirewallRule
			if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			result, err := c.CreateFirewallRule(ctx, &rule)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			respondJSON(w, result)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/firewall/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/api/firewall/")
		switch r.Method {
		case "DELETE":
			if err := c.DeleteFirewallRule(ctx, id); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			respondJSON(w, map[string]string{"message": "deleted"})
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Ports
	mux.HandleFunc("/api/ports", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			ports, err := c.ListPortAssignments(ctx, storage.PortAssignmentFilters{})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			respondJSON(w, ports)
		case "POST":
			var port domain.PortAssignment
			if err := json.NewDecoder(r.Body).Decode(&port); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			result, err := c.CreatePortAssignment(ctx, &port)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			respondJSON(w, result)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/ports/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/api/ports/")
		switch r.Method {
		case "DELETE":
			if err := c.DeletePortAssignment(ctx, id); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			respondJSON(w, map[string]string{"message": "deleted"})
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Journal
	mux.HandleFunc("/api/journal", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			journal, err := c.ListJournal(ctx, storage.JournalFilters{})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			respondJSON(w, journal)
		case "POST":
			var entry domain.JournalEntry
			if err := json.NewDecoder(r.Body).Decode(&entry); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			result, err := c.CreateJournalEntry(ctx, &entry)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			respondJSON(w, result)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Admin API Keys
	mux.HandleFunc("/api/admin/apikeys", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			keys, err := c.ListAPIKeys(ctx)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			respondJSON(w, keys)
		case "POST":
			var req struct {
				Name        string `json:"name"`
				Scope       string `json:"scope"`
				Description string `json:"description"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			result, err := c.CreateAPIKey(ctx, client.CreateAPIKeyRequest{
				Name:        req.Name,
				Scope:       domain.APIKeyScope(req.Scope),
				Description: req.Description,
			})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			respondJSON(w, result)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/admin/apikeys/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/api/admin/apikeys/")
		switch r.Method {
		case "DELETE":
			if err := c.DeleteAPIKey(ctx, id); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			respondJSON(w, map[string]string{"message": "deleted"})
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Components delete endpoint
	mux.HandleFunc("/api/components/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/api/components/")
		switch r.Method {
		case "DELETE":
			if err := c.DeleteComponent(ctx, id); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			respondJSON(w, map[string]string{"message": "deleted"})
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Services delete endpoint
	mux.HandleFunc("/api/services/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/api/services/")
		switch r.Method {
		case "DELETE":
			if err := c.DeleteService(ctx, id); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			respondJSON(w, map[string]string{"message": "deleted"})
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Reports
	mux.HandleFunc("/api/reports/compute/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/api/reports/compute/")
		switch r.Method {
		case "GET":
			// Get compute
			compute, err := c.ResolveCompute(ctx, id)
			if err != nil {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}

			// Get component assignments
			componentAssignments, err := c.ListComponentAssignments(ctx, storage.ComputeComponentFilters{
				ComputeID: compute.ID,
			})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// Get service assignments
			assignments, err := c.ListAssignments(ctx, storage.AssignmentFilters{
				ComputeID: compute.ID,
			})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// Get journal entries
			journals, err := c.ListJournal(ctx, storage.JournalFilters{
				ComputeID: compute.ID,
			})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// Get IP assignments
			ipAssignments, err := c.ListIPAssignments(ctx, compute.ID, "")
			if err != nil {
				ipAssignments = []*domain.ComputeIP{}
			}

			// Build report
			report := map[string]interface{}{
				"compute":              compute,
				"component_assignments": componentAssignments,
				"service_assignments":   assignments,
				"journal_entries":       journals,
				"ip_assignments":        ipAssignments,
			}

			respondJSON(w, report)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
}

func respondJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
