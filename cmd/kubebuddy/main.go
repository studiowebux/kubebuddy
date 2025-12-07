package main

import (
	"context"
	"fmt"
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
	"github.com/studiowebux/kubebuddy/internal/storage/sqlite"
)

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
	)

	cmd := &cobra.Command{
		Use:   "server",
		Short: "Start API server",
		Long:  `Start the KubeBuddy API server`,
		RunE: func(cmd *cobra.Command, args []string) error {
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
				adminKey := os.Getenv("ADMIN_API_KEY")
				if adminKey == "" {
					return fmt.Errorf("ADMIN_API_KEY environment variable is required when using --create-admin-key")
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

			// Handle graceful shutdown
			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

			go func() {
				<-sigChan
				fmt.Println("\nShutting down server...")
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
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
	cmd.Flags().BoolVar(&createAdminKey, "create-admin-key", false, "Create admin API key from ADMIN_API_KEY env var")
	cmd.Flags().BoolVar(&seedData, "seed", false, "Seed database with sample data")

	return cmd
}
