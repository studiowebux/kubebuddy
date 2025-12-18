package cli

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/studiowebux/kubebuddy/internal/client"
	"github.com/studiowebux/kubebuddy/internal/domain"
	"github.com/studiowebux/kubebuddy/internal/storage"
)

func newComputeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "compute",
		Short: "Manage compute resources",
		Long:  `Manage compute resources (baremetal, VPS, VM)`,
	}

	cmd.AddCommand(newComputeListCmd())
	cmd.AddCommand(newComputeGetCmd())
	cmd.AddCommand(newComputeCreateCmd())
	cmd.AddCommand(newComputeUpdateCmd())
	cmd.AddCommand(newComputeDeleteCmd())

	return cmd
}

func newComputeListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all compute resources",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireAPIKey(cmd); err != nil {
				return err
			}

			c := client.New(endpoint, apiKey)
			computes, err := c.ListComputes(context.Background(), storage.ComputeFilters{})
			if err != nil {
				return err
			}

			printJSON(computes)
			return nil
		},
	}
}

func newComputeGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <id|name>",
		Short: "Get compute resource by ID or name",
		Args:  cobra.ExactArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			return completeComputeIDs(toComplete), cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveKeepOrder
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireAPIKey(cmd); err != nil {
				return err
			}

			c := client.New(endpoint, apiKey)
			compute, err := c.ResolveCompute(context.Background(), args[0])
			if err != nil {
				return err
			}

			printJSON(compute)
			return nil
		},
	}

	return cmd
}

func newComputeCreateCmd() *cobra.Command {
	var (
		name            string
		computeType     string
		provider        string
		region          string
		tags            string
		monthlyCost     float64
		annualCost      float64
		contractEnd     string
		renewalDate     string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new compute resource",
		Long:  `Create a new compute resource. Use 'component assign' to add hardware components.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireAPIKey(cmd); err != nil {
				return err
			}

			compute := &domain.Compute{
				ID:        uuid.New().String(),
				Name:      name,
				Type:      domain.ComputeType(computeType),
				Provider:  provider,
				Region:    region,
				Tags:      parseTags(tags),
				Resources: make(domain.Resources),
				State:     domain.ComputeStateActive,
			}

			if cmd.Flags().Changed("monthly-cost") {
				compute.MonthlyCost = &monthlyCost
			}
			if cmd.Flags().Changed("annual-cost") {
				compute.AnnualCost = &annualCost
			}
			if contractEnd != "" {
				t, err := time.Parse("2006-01-02", contractEnd)
				if err != nil {
					return fmt.Errorf("invalid contract-end date format (use YYYY-MM-DD): %w", err)
				}
				compute.ContractEndDate = &t
			}
			if renewalDate != "" {
				t, err := time.Parse("2006-01-02", renewalDate)
				if err != nil {
					return fmt.Errorf("invalid renewal-date format (use YYYY-MM-DD): %w", err)
				}
				compute.NextRenewalDate = &t
			}

			c := client.New(endpoint, apiKey)
			result, err := c.CreateCompute(context.Background(), compute)
			if err != nil {
				return err
			}

			printJSON(result)
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Compute name (required)")
	cmd.Flags().StringVar(&computeType, "type", "baremetal", "Compute type (baremetal, vps, vm)")
	cmd.Flags().StringVar(&provider, "provider", "", "Provider name (required)")
	cmd.Flags().StringVar(&region, "region", "", "Region (required)")
	cmd.Flags().StringVar(&tags, "tags", "", "Tags as key=value pairs, comma-separated (e.g., env=prod,zone=us-east)")
	cmd.Flags().Float64Var(&monthlyCost, "monthly-cost", 0, "Monthly cost")
	cmd.Flags().Float64Var(&annualCost, "annual-cost", 0, "Annual cost")
	cmd.Flags().StringVar(&contractEnd, "contract-end", "", "Contract end date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&renewalDate, "renewal-date", "", "Next renewal date (YYYY-MM-DD)")

	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("provider")
	cmd.MarkFlagRequired("region")

	// Add completion for type flag
	cmd.RegisterFlagCompletionFunc("type", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"baremetal", "vps", "vm"}, cobra.ShellCompDirectiveNoFileComp
	})

	// Add completion for provider flag
	cmd.RegisterFlagCompletionFunc("provider", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return completeProviders(), cobra.ShellCompDirectiveNoFileComp
	})

	// Add completion for region flag
	cmd.RegisterFlagCompletionFunc("region", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return completeRegions(), cobra.ShellCompDirectiveNoFileComp
	})

	return cmd
}

func newComputeUpdateCmd() *cobra.Command {
	var (
		name            string
		computeType     string
		provider        string
		region          string
		tags            string
		state           string
		monthlyCost     float64
		annualCost      float64
		contractEnd     string
		renewalDate     string
	)

	cmd := &cobra.Command{
		Use:   "update <id|name>",
		Short: "Update a compute resource by ID or name",
		Args:  cobra.ExactArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			return completeComputeIDs(toComplete), cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveKeepOrder
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireAPIKey(cmd); err != nil {
				return err
			}

			c := client.New(endpoint, apiKey)

			// Get existing compute
			existing, err := c.ResolveCompute(context.Background(), args[0])
			if err != nil {
				return err
			}

			// Update only specified fields
			if name != "" {
				existing.Name = name
			}
			if computeType != "" {
				existing.Type = domain.ComputeType(computeType)
			}
			if provider != "" {
				existing.Provider = provider
			}
			if region != "" {
				existing.Region = region
			}
			if tags != "" {
				existing.Tags = parseTags(tags)
			}
			if state != "" {
				existing.State = domain.ComputeState(state)
			}
			if cmd.Flags().Changed("monthly-cost") {
				existing.MonthlyCost = &monthlyCost
			}
			if cmd.Flags().Changed("annual-cost") {
				existing.AnnualCost = &annualCost
			}
			if contractEnd != "" {
				t, err := time.Parse("2006-01-02", contractEnd)
				if err != nil {
					return fmt.Errorf("invalid contract-end date format (use YYYY-MM-DD): %w", err)
				}
				existing.ContractEndDate = &t
			}
			if renewalDate != "" {
				t, err := time.Parse("2006-01-02", renewalDate)
				if err != nil {
					return fmt.Errorf("invalid renewal-date format (use YYYY-MM-DD): %w", err)
				}
				existing.NextRenewalDate = &t
			}

			result, err := c.UpdateCompute(context.Background(), existing.ID, existing)
			if err != nil {
				return err
			}

			printJSON(result)
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Compute name")
	cmd.Flags().StringVar(&computeType, "type", "", "Compute type (baremetal, vps, vm)")
	cmd.Flags().StringVar(&provider, "provider", "", "Provider name")
	cmd.Flags().StringVar(&region, "region", "", "Region")
	cmd.Flags().StringVar(&tags, "tags", "", "Tags as key=value pairs, comma-separated (e.g., env=prod,zone=us-east)")
	cmd.Flags().StringVar(&state, "state", "", "State (active, inactive, maintenance)")
	cmd.Flags().Float64Var(&monthlyCost, "monthly-cost", 0, "Monthly cost")
	cmd.Flags().Float64Var(&annualCost, "annual-cost", 0, "Annual cost")
	cmd.Flags().StringVar(&contractEnd, "contract-end", "", "Contract end date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&renewalDate, "renewal-date", "", "Next renewal date (YYYY-MM-DD)")

	// Add completion for type flag
	cmd.RegisterFlagCompletionFunc("type", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"baremetal", "vps", "vm"}, cobra.ShellCompDirectiveNoFileComp
	})

	// Add completion for state flag
	cmd.RegisterFlagCompletionFunc("state", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"active", "inactive", "maintenance"}, cobra.ShellCompDirectiveNoFileComp
	})

	// Add completion for provider flag
	cmd.RegisterFlagCompletionFunc("provider", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return completeProviders(), cobra.ShellCompDirectiveNoFileComp
	})

	// Add completion for region flag
	cmd.RegisterFlagCompletionFunc("region", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return completeRegions(), cobra.ShellCompDirectiveNoFileComp
	})

	return cmd
}

func newComputeDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id|name>",
		Short: "Delete a compute resource by ID or name",
		Args:  cobra.ExactArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			return completeComputeIDs(toComplete), cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveKeepOrder
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireAPIKey(cmd); err != nil {
				return err
			}

			c := client.New(endpoint, apiKey)
			compute, err := c.ResolveCompute(context.Background(), args[0])
			if err != nil {
				return err
			}

			if err := c.DeleteCompute(context.Background(), compute.ID); err != nil {
				return err
			}

			fmt.Println("Compute deleted successfully")
			return nil
		},
	}

	return cmd
}

// Helper function for compute ID completion
func completeComputeIDs(toComplete string) []string {
	if apiKey == "" {
		return nil
	}

	c := client.New(endpoint, apiKey)
	computes, err := c.ListComputes(context.Background(), storage.ComputeFilters{})
	if err != nil {
		return nil
	}

	// Handle comma-separated values - only suggest computes not already selected
	prefix := ""
	lastCommaIdx := strings.LastIndex(toComplete, ",")
	alreadySelected := make(map[string]bool)

	if lastCommaIdx >= 0 {
		// Extract already selected computes
		prefix = toComplete[:lastCommaIdx+1]
		for _, name := range strings.Split(toComplete[:lastCommaIdx], ",") {
			alreadySelected[strings.TrimSpace(name)] = true
		}
	}

	var completions []string
	for _, compute := range computes {
		// Skip already selected computes
		if !alreadySelected[compute.Name] {
			completions = append(completions, prefix+compute.Name)
		}
	}

	// Sort alphabetically by name
	sort.Strings(completions)

	return completions
}

// Helper function for region completion
func completeRegions() []string {
	if apiKey == "" {
		return nil
	}

	c := client.New(endpoint, apiKey)

	regions := make(map[string]bool)

	computes, err := c.ListComputes(context.Background(), storage.ComputeFilters{})
	if err == nil {
		for _, compute := range computes {
			if compute.Region != "" {
				regions[compute.Region] = true
			}
		}
	}

	ips, err := c.ListIPAddresses(context.Background(), storage.IPAddressFilters{})
	if err == nil {
		for _, ip := range ips {
			if ip.Region != "" {
				regions[ip.Region] = true
			}
		}
	}

	var completions []string
	for region := range regions {
		completions = append(completions, region)
	}

	sort.Strings(completions)
	return completions
}

// Helper function for provider completion
func completeProviders() []string {
	if apiKey == "" {
		return nil
	}

	c := client.New(endpoint, apiKey)

	providers := make(map[string]bool)

	computes, err := c.ListComputes(context.Background(), storage.ComputeFilters{})
	if err == nil {
		for _, compute := range computes {
			if compute.Provider != "" {
				providers[compute.Provider] = true
			}
		}
	}

	ips, err := c.ListIPAddresses(context.Background(), storage.IPAddressFilters{})
	if err == nil {
		for _, ip := range ips {
			if ip.Provider != "" {
				providers[ip.Provider] = true
			}
		}
	}

	var completions []string
	for provider := range providers {
		completions = append(completions, provider)
	}

	sort.Strings(completions)
	return completions
}
