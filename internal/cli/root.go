package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	endpoint string
	apiKey   string
	// Version can be set at build time using ldflags
	Version = "0.0.4"
)

// NewRootCmd creates the root command
func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:     "kubebuddy",
		Short:   "KubeBuddy - Capacity Planning System",
		Long: `KubeBuddy is a capacity planning system for managing compute resources and services.

Environment Variables:
  KUBEBUDDY_ENDPOINT    API endpoint (default: http://localhost:8080, overridden by --endpoint)
  KUBEBUDDY_API_KEY     API key for authentication (overridden by --api-key)`,
		Version: Version,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Get from env if not set via flags
			if endpoint == "" {
				endpoint = os.Getenv("KUBEBUDDY_ENDPOINT")
				if endpoint == "" {
					endpoint = "http://localhost:8080"
				}
			}
			if apiKey == "" {
				apiKey = os.Getenv("KUBEBUDDY_API_KEY")
			}
		},
	}

	// Global flags
	rootCmd.PersistentFlags().StringVar(&endpoint, "endpoint", "", "API endpoint (default: http://localhost:8080 or KUBEBUDDY_ENDPOINT)")
	rootCmd.PersistentFlags().StringVar(&apiKey, "api-key", "", "API key for authentication (default: KUBEBUDDY_API_KEY)")

	// Add subcommands
	rootCmd.AddCommand(newComputeCmd())
	rootCmd.AddCommand(newServiceCmd())
	rootCmd.AddCommand(newAssignmentCmd())
	rootCmd.AddCommand(newPlanCmd())
	rootCmd.AddCommand(newJournalCmd())
	rootCmd.AddCommand(newAPIKeyCmd())
	rootCmd.AddCommand(newComponentCmd())
	rootCmd.AddCommand(newIPCmd())
	rootCmd.AddCommand(newDNSCmd())
	rootCmd.AddCommand(newPortCmd())
	rootCmd.AddCommand(newFirewallCmd())
	rootCmd.AddCommand(newReportCmd())

	return rootCmd
}

// Execute runs the root command
func Execute() error {
	return NewRootCmd().Execute()
}

// Helper function to check if API key is set
func requireAPIKey(cmd *cobra.Command) error {
	if apiKey == "" {
		return fmt.Errorf("API key required. Set KUBEBUDDY_API_KEY environment variable or use --api-key flag")
	}
	return nil
}

// Helper to print JSON
func printJSON(v interface{}) {
	data, _ := json.MarshalIndent(v, "", "  ")
	fmt.Println(string(data))
}

// Helper to parse tags from comma-separated key=value pairs
func parseTags(tagsStr string) map[string]string {
	tags := make(map[string]string)
	if tagsStr == "" {
		return tags
	}

	pairs := splitTags(tagsStr)
	for _, pair := range pairs {
		kv := splitKeyValue(pair)
		if len(kv) == 2 {
			tags[kv[0]] = kv[1]
		}
	}
	return tags
}

func splitTags(s string) []string {
	var parts []string
	var current string
	for _, r := range s {
		if r == ',' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(r)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}

func splitKeyValue(s string) []string {
	for i, r := range s {
		if r == '=' {
			return []string{s[:i], s[i+1:]}
		}
	}
	return []string{s}
}
