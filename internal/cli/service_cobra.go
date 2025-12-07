package cli

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/studiowebux/kubebuddy/internal/client"
	"github.com/studiowebux/kubebuddy/internal/domain"
)

func newServiceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "service",
		Short: "Manage services",
		Long:  `Manage services with resource specifications and placement rules`,
	}

	cmd.AddCommand(newServiceListCmd())
	cmd.AddCommand(newServiceGetCmd())
	cmd.AddCommand(newServiceCreateCmd())
	cmd.AddCommand(newServiceDeleteCmd())

	return cmd
}

func newServiceListCmd() *cobra.Command{
	return &cobra.Command{
		Use:   "list",
		Short: "List all services",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireAPIKey(cmd); err != nil {
				return err
			}

			c := client.New(endpoint, apiKey)
			services, err := c.ListServices(context.Background())
			if err != nil {
				return err
			}

			printJSON(services)
			return nil
		},
	}
}

func newServiceGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get service by ID",
		Args:  cobra.ExactArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			return completeServiceIDs(toComplete), cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireAPIKey(cmd); err != nil {
				return err
			}

			c := client.New(endpoint, apiKey)
			service, err := c.GetService(context.Background(), args[0])
			if err != nil {
				return err
			}

			printJSON(service)
			return nil
		},
	}

	return cmd
}

func newServiceCreateCmd() *cobra.Command {
	var (
		name      string
		minSpec   string
		maxSpec   string
		placement string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new service (upserts based on name)",
		Long:  `Create a new service. If a service with the same name exists, it updates it.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireAPIKey(cmd); err != nil {
				return err
			}

			service := &domain.Service{
				ID:      uuid.New().String(),
				Name:    name,
				MinSpec: make(domain.Resources),
				MaxSpec: make(domain.Resources),
			}

			// Parse min_spec JSON
			if minSpec != "" {
				if err := json.Unmarshal([]byte(minSpec), &service.MinSpec); err != nil {
					return fmt.Errorf("invalid min-spec JSON: %w", err)
				}
			}

			// Parse max_spec JSON
			if maxSpec != "" {
				if err := json.Unmarshal([]byte(maxSpec), &service.MaxSpec); err != nil {
					return fmt.Errorf("invalid max-spec JSON: %w", err)
				}
			}

			// Parse placement JSON
			if placement != "" {
				if err := json.Unmarshal([]byte(placement), &service.Placement); err != nil {
					return fmt.Errorf("invalid placement JSON: %w", err)
				}
			}

			c := client.New(endpoint, apiKey)
			result, err := c.CreateService(context.Background(), service)
			if err != nil {
				return err
			}

			printJSON(result)
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Service name (required)")
	cmd.Flags().StringVar(&minSpec, "min-spec", "", "Minimum resource spec as JSON (e.g. '{\"cpu\":2,\"ram_gb\":4}')")
	cmd.Flags().StringVar(&maxSpec, "max-spec", "", "Maximum resource spec as JSON (e.g. '{\"cpu\":8,\"ram_gb\":16}')")
	cmd.Flags().StringVar(&placement, "placement", "", "Placement rules as JSON")
	cmd.MarkFlagRequired("name")

	return cmd
}

func newServiceDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a service",
		Args:  cobra.ExactArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			return completeServiceIDs(toComplete), cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireAPIKey(cmd); err != nil {
				return err
			}

			c := client.New(endpoint, apiKey)
			if err := c.DeleteService(context.Background(), args[0]); err != nil {
				return err
			}

			fmt.Println("Service deleted successfully")
			return nil
		},
	}

	return cmd
}

// Helper function for service ID completion
func completeServiceIDs(toComplete string) []string {
	if apiKey == "" {
		return nil
	}

	c := client.New(endpoint, apiKey)
	services, err := c.ListServices(context.Background())
	if err != nil {
		return nil
	}

	var completions []string
	for _, service := range services {
		completions = append(completions, fmt.Sprintf("%s\t%s", service.ID, service.Name))
	}

	return completions
}
