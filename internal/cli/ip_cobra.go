package cli

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/studiowebux/kubebuddy/internal/client"
	"github.com/studiowebux/kubebuddy/internal/domain"
	"github.com/studiowebux/kubebuddy/internal/storage"
)

func newIPCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ip",
		Short: "Manage IP addresses",
		Long:  `Manage IP addresses and their assignments to computes`,
	}

	cmd.AddCommand(newIPListCmd())
	cmd.AddCommand(newIPGetCmd())
	cmd.AddCommand(newIPCreateCmd())
	cmd.AddCommand(newIPDeleteCmd())
	cmd.AddCommand(newIPAssignCmd())
	cmd.AddCommand(newIPUnassignCmd())
	cmd.AddCommand(newIPListAssignmentsCmd())

	return cmd
}

func newIPListCmd() *cobra.Command {
	var (
		ipType   string
		provider string
		region   string
		state    string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List IP addresses",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireAPIKey(cmd); err != nil {
				return err
			}

			filters := storage.IPAddressFilters{
				Type:     ipType,
				Provider: provider,
				Region:   region,
				State:    state,
			}

			c := client.New(endpoint, apiKey)
			ips, err := c.ListIPAddresses(context.Background(), filters)
			if err != nil {
				return err
			}

			printJSON(ips)
			return nil
		},
	}

	cmd.Flags().StringVar(&ipType, "type", "", "Filter by IP type (public, private)")
	cmd.Flags().StringVar(&provider, "provider", "", "Filter by provider")
	cmd.Flags().StringVar(&region, "region", "", "Filter by region")
	cmd.Flags().StringVar(&state, "state", "", "Filter by state (available, assigned, reserved)")

	cmd.RegisterFlagCompletionFunc("type", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"public", "private"}, cobra.ShellCompDirectiveNoFileComp
	})

	cmd.RegisterFlagCompletionFunc("state", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"available", "assigned", "reserved"}, cobra.ShellCompDirectiveNoFileComp
	})

	return cmd
}

func newIPGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get [id]",
		Short: "Get IP address details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireAPIKey(cmd); err != nil {
				return err
			}

			c := client.New(endpoint, apiKey)
			ip, err := c.GetIPAddress(context.Background(), args[0])
			if err != nil {
				return err
			}

			printJSON(ip)
			return nil
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				return completeIPIDs(toComplete), cobra.ShellCompDirectiveNoFileComp
			}
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
	}

	return cmd
}

func newIPCreateCmd() *cobra.Command {
	var (
		address    string
		ipType     string
		cidr       string
		gateway    string
		dnsServers string
		provider   string
		region     string
		notes      string
		state      string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new IP address",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireAPIKey(cmd); err != nil {
				return err
			}

			var dnsServerList []string
			if dnsServers != "" {
				dnsServerList = strings.Split(dnsServers, ",")
			}

			ipState := domain.IPStateAvailable
			if state != "" {
				ipState = domain.IPState(state)
			}

			ip := &domain.IPAddress{
				ID:         uuid.New().String(),
				Address:    address,
				Type:       domain.IPType(ipType),
				CIDR:       cidr,
				Gateway:    gateway,
				DNSServers: dnsServerList,
				Provider:   provider,
				Region:     region,
				Notes:      notes,
				State:      ipState,
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			}

			c := client.New(endpoint, apiKey)
			result, err := c.CreateIPAddress(context.Background(), ip)
			if err != nil {
				return err
			}

			printJSON(result)
			return nil
		},
	}

	cmd.Flags().StringVar(&address, "address", "", "IP address (required)")
	cmd.Flags().StringVar(&ipType, "type", "", "IP type: public or private (required)")
	cmd.Flags().StringVar(&cidr, "cidr", "", "CIDR notation (e.g., 192.168.1.0/24) (required)")
	cmd.Flags().StringVar(&gateway, "gateway", "", "Gateway address")
	cmd.Flags().StringVar(&dnsServers, "dns", "", "DNS servers (comma-separated)")
	cmd.Flags().StringVar(&provider, "provider", "", "Provider (required)")
	cmd.Flags().StringVar(&region, "region", "", "Region (required)")
	cmd.Flags().StringVar(&notes, "notes", "", "Notes")
	cmd.Flags().StringVar(&state, "state", "available", "State (available, assigned, reserved)")

	cmd.MarkFlagRequired("address")
	cmd.MarkFlagRequired("type")
	cmd.MarkFlagRequired("cidr")
	cmd.MarkFlagRequired("provider")
	cmd.MarkFlagRequired("region")

	cmd.RegisterFlagCompletionFunc("type", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"public", "private"}, cobra.ShellCompDirectiveNoFileComp
	})

	cmd.RegisterFlagCompletionFunc("state", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"available", "assigned", "reserved"}, cobra.ShellCompDirectiveNoFileComp
	})

	return cmd
}

func newIPDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete [id]",
		Short: "Delete an IP address",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireAPIKey(cmd); err != nil {
				return err
			}

			c := client.New(endpoint, apiKey)
			if err := c.DeleteIPAddress(context.Background(), args[0]); err != nil {
				return err
			}

			printJSON(map[string]string{"message": "IP address deleted successfully"})
			return nil
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				return completeIPIDs(toComplete), cobra.ShellCompDirectiveNoFileComp
			}
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
	}

	return cmd
}

func newIPAssignCmd() *cobra.Command {
	var (
		computeID string
		ipID      string
		isPrimary bool
	)

	cmd := &cobra.Command{
		Use:   "assign",
		Short: "Assign an IP address to a compute",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireAPIKey(cmd); err != nil {
				return err
			}

			assignment := &domain.ComputeIP{
				ID:        uuid.New().String(),
				ComputeID: computeID,
				IPID:      ipID,
				IsPrimary: isPrimary,
				CreatedAt: time.Now(),
			}

			c := client.New(endpoint, apiKey)
			result, err := c.AssignIP(context.Background(), assignment)
			if err != nil {
				return err
			}

			printJSON(result)
			return nil
		},
	}

	cmd.Flags().StringVar(&computeID, "compute", "", "Compute ID (required)")
	cmd.Flags().StringVar(&ipID, "ip", "", "IP address ID (required)")
	cmd.Flags().BoolVar(&isPrimary, "primary", false, "Set as primary IP")

	cmd.MarkFlagRequired("compute")
	cmd.MarkFlagRequired("ip")

	cmd.RegisterFlagCompletionFunc("compute", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return completeComputeIDs(toComplete), cobra.ShellCompDirectiveNoFileComp
	})

	cmd.RegisterFlagCompletionFunc("ip", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return completeIPIDs(toComplete), cobra.ShellCompDirectiveNoFileComp
	})

	return cmd
}

func newIPUnassignCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unassign [assignment-id]",
		Short: "Unassign an IP address from a compute",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireAPIKey(cmd); err != nil {
				return err
			}

			c := client.New(endpoint, apiKey)
			if err := c.UnassignIP(context.Background(), args[0]); err != nil {
				return err
			}

			printJSON(map[string]string{"message": "IP unassigned successfully"})
			return nil
		},
	}

	return cmd
}

func newIPListAssignmentsCmd() *cobra.Command {
	var (
		computeID string
		ipID      string
	)

	cmd := &cobra.Command{
		Use:   "list-assignments",
		Short: "List IP address assignments",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireAPIKey(cmd); err != nil {
				return err
			}

			c := client.New(endpoint, apiKey)
			assignments, err := c.ListIPAssignments(context.Background(), computeID, ipID)
			if err != nil {
				return err
			}

			printJSON(assignments)
			return nil
		},
	}

	cmd.Flags().StringVar(&computeID, "compute", "", "Filter by compute ID")
	cmd.Flags().StringVar(&ipID, "ip", "", "Filter by IP address ID")

	cmd.RegisterFlagCompletionFunc("compute", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return completeComputeIDs(toComplete), cobra.ShellCompDirectiveNoFileComp
	})

	cmd.RegisterFlagCompletionFunc("ip", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return completeIPIDs(toComplete), cobra.ShellCompDirectiveNoFileComp
	})

	return cmd
}

func completeIPIDs(toComplete string) []string {
	if apiKey == "" {
		return nil
	}

	c := client.New(endpoint, apiKey)
	ips, err := c.ListIPAddresses(context.Background(), storage.IPAddressFilters{})
	if err != nil {
		return nil
	}

	var completions []string
	for _, ip := range ips {
		// Format: ID \t Address (Type - Provider/Region)
		completions = append(completions, ip.ID+"\t"+ip.Address+" ("+string(ip.Type)+" - "+ip.Provider+"/"+ip.Region+")")
	}

	return completions
}
