package cli

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/studiowebux/kubebuddy/internal/client"
	"github.com/studiowebux/kubebuddy/internal/domain"
	"github.com/studiowebux/kubebuddy/internal/storage"
)

func newFirewallCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "firewall",
		Short: "Manage firewall rules",
		Long:  `Manage firewall rules and assignments to computes`,
	}

	cmd.AddCommand(newFirewallListCmd())
	cmd.AddCommand(newFirewallGetCmd())
	cmd.AddCommand(newFirewallCreateCmd())
	cmd.AddCommand(newFirewallDeleteCmd())
	cmd.AddCommand(newFirewallAssignCmd())
	cmd.AddCommand(newFirewallUnassignCmd())
	cmd.AddCommand(newFirewallListAssignmentsCmd())

	return cmd
}

func newFirewallListCmd() *cobra.Command {
	var (
		action   string
		protocol string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List firewall rules",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireAPIKey(cmd); err != nil {
				return err
			}

			filters := storage.FirewallRuleFilters{
				Action:   action,
				Protocol: protocol,
			}

			c := client.New(endpoint, apiKey)
			rules, err := c.ListFirewallRules(context.Background(), filters)
			if err != nil {
				return err
			}

			printJSON(rules)
			return nil
		},
	}

	cmd.Flags().StringVar(&action, "action", "", "Filter by action (ALLOW, DENY)")
	cmd.Flags().StringVar(&protocol, "protocol", "", "Filter by protocol (tcp, udp, icmp, all)")

	cmd.RegisterFlagCompletionFunc("action", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"ALLOW", "DENY"}, cobra.ShellCompDirectiveNoFileComp
	})

	cmd.RegisterFlagCompletionFunc("protocol", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"tcp", "udp", "icmp", "all"}, cobra.ShellCompDirectiveNoFileComp
	})

	return cmd
}

func newFirewallGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get [id]",
		Short: "Get firewall rule details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireAPIKey(cmd); err != nil {
				return err
			}

			c := client.New(endpoint, apiKey)
			rule, err := c.GetFirewallRule(context.Background(), args[0])
			if err != nil {
				return err
			}

			printJSON(rule)
			return nil
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				return completeFirewallRuleIDs(toComplete), cobra.ShellCompDirectiveNoFileComp
			}
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
	}

	return cmd
}

func newFirewallCreateCmd() *cobra.Command {
	var (
		name        string
		action      string
		protocol    string
		source      string
		destination string
		portStart   int
		portEnd     int
		description string
		priority    int
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new firewall rule",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireAPIKey(cmd); err != nil {
				return err
			}

			if priority == 0 {
				priority = 100 // Default priority
			}

			rule := &domain.FirewallRule{
				ID:          uuid.New().String(),
				Name:        name,
				Action:      domain.FirewallAction(action),
				Protocol:    domain.Protocol(protocol),
				Source:      source,
				Destination: destination,
				Description: description,
				Priority:    priority,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}

			if portStart > 0 {
				rule.PortStart = &portStart
			}
			if portEnd > 0 {
				rule.PortEnd = &portEnd
			}

			c := client.New(endpoint, apiKey)
			result, err := c.CreateFirewallRule(context.Background(), rule)
			if err != nil {
				return err
			}

			printJSON(result)
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Rule name (required, unique)")
	cmd.Flags().StringVar(&action, "action", "", "Action: ALLOW or DENY (required)")
	cmd.Flags().StringVar(&protocol, "protocol", "", "Protocol: tcp, udp, icmp, all (required)")
	cmd.Flags().StringVar(&source, "source", "", "Source CIDR, IP, or 'any' (required)")
	cmd.Flags().StringVar(&destination, "destination", "", "Destination CIDR, IP, or 'any' (required)")
	cmd.Flags().IntVar(&portStart, "port-start", 0, "Port start (0 for any)")
	cmd.Flags().IntVar(&portEnd, "port-end", 0, "Port end (0 for single port)")
	cmd.Flags().StringVar(&description, "description", "", "Description")
	cmd.Flags().IntVar(&priority, "priority", 100, "Priority (lower = higher priority)")

	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("action")
	cmd.MarkFlagRequired("protocol")
	cmd.MarkFlagRequired("source")
	cmd.MarkFlagRequired("destination")

	cmd.RegisterFlagCompletionFunc("action", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"ALLOW", "DENY"}, cobra.ShellCompDirectiveNoFileComp
	})

	cmd.RegisterFlagCompletionFunc("protocol", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"tcp", "udp", "icmp", "all"}, cobra.ShellCompDirectiveNoFileComp
	})

	return cmd
}

func newFirewallDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete [id]",
		Short: "Delete a firewall rule",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireAPIKey(cmd); err != nil {
				return err
			}

			c := client.New(endpoint, apiKey)
			if err := c.DeleteFirewallRule(context.Background(), args[0]); err != nil {
				return err
			}

			printJSON(map[string]string{"message": "firewall rule deleted successfully"})
			return nil
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				return completeFirewallRuleIDs(toComplete), cobra.ShellCompDirectiveNoFileComp
			}
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
	}

	return cmd
}

func newFirewallAssignCmd() *cobra.Command {
	var (
		computeID string
		ruleID    string
		enabled   bool
	)

	cmd := &cobra.Command{
		Use:   "assign",
		Short: "Assign firewall rule to compute",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireAPIKey(cmd); err != nil {
				return err
			}

			assignment := &domain.ComputeFirewallRule{
				ID:        uuid.New().String(),
				ComputeID: computeID,
				RuleID:    ruleID,
				Enabled:   enabled,
				CreatedAt: time.Now(),
			}

			c := client.New(endpoint, apiKey)
			result, err := c.AssignFirewallRule(context.Background(), assignment)
			if err != nil {
				return err
			}

			printJSON(result)
			return nil
		},
	}

	cmd.Flags().StringVar(&computeID, "compute", "", "Compute ID (required)")
	cmd.Flags().StringVar(&ruleID, "rule", "", "Firewall rule ID (required)")
	cmd.Flags().BoolVar(&enabled, "enabled", true, "Enable rule (default: true)")

	cmd.MarkFlagRequired("compute")
	cmd.MarkFlagRequired("rule")

	cmd.RegisterFlagCompletionFunc("compute", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return completeComputeIDs(toComplete), cobra.ShellCompDirectiveNoFileComp
	})

	cmd.RegisterFlagCompletionFunc("rule", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return completeFirewallRuleIDs(toComplete), cobra.ShellCompDirectiveNoFileComp
	})

	return cmd
}

func newFirewallUnassignCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unassign [assignment-id]",
		Short: "Unassign firewall rule from compute",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireAPIKey(cmd); err != nil {
				return err
			}

			c := client.New(endpoint, apiKey)
			if err := c.UnassignFirewallRule(context.Background(), args[0]); err != nil {
				return err
			}

			printJSON(map[string]string{"message": "firewall rule unassigned successfully"})
			return nil
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				return completeFirewallAssignmentIDs(toComplete), cobra.ShellCompDirectiveNoFileComp
			}
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
	}

	return cmd
}

func newFirewallListAssignmentsCmd() *cobra.Command {
	var (
		computeID string
		ruleID    string
	)

	cmd := &cobra.Command{
		Use:   "list-assignments",
		Short: "List firewall rule assignments",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireAPIKey(cmd); err != nil {
				return err
			}

			c := client.New(endpoint, apiKey)
			assignments, err := c.ListComputeFirewallRules(context.Background(), computeID, ruleID)
			if err != nil {
				return err
			}

			printJSON(assignments)
			return nil
		},
	}

	cmd.Flags().StringVar(&computeID, "compute", "", "Filter by compute ID")
	cmd.Flags().StringVar(&ruleID, "rule", "", "Filter by firewall rule ID")

	cmd.RegisterFlagCompletionFunc("compute", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return completeComputeIDs(toComplete), cobra.ShellCompDirectiveNoFileComp
	})

	cmd.RegisterFlagCompletionFunc("rule", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return completeFirewallRuleIDs(toComplete), cobra.ShellCompDirectiveNoFileComp
	})

	return cmd
}

func completeFirewallRuleIDs(toComplete string) []string {
	if apiKey == "" {
		return nil
	}

	c := client.New(endpoint, apiKey)
	rules, err := c.ListFirewallRules(context.Background(), storage.FirewallRuleFilters{})
	if err != nil {
		return nil
	}

	var completions []string
	for _, rule := range rules {
		// Format: ID \t Name (Action Protocol)
		completions = append(completions, rule.ID+"\t"+rule.Name+" ("+string(rule.Action)+" "+string(rule.Protocol)+")")
	}

	return completions
}

func completeFirewallAssignmentIDs(toComplete string) []string {
	if apiKey == "" {
		return nil
	}

	// Get all assignments by querying with empty filters would require API changes
	// For now, return empty list
	return nil
}
