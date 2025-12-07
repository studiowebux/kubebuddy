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

func newPortCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "port",
		Short: "Manage port assignments",
		Long:  `Manage port assignments for services`,
	}

	cmd.AddCommand(newPortListCmd())
	cmd.AddCommand(newPortGetCmd())
	cmd.AddCommand(newPortCreateCmd())
	cmd.AddCommand(newPortDeleteCmd())

	return cmd
}

func newPortListCmd() *cobra.Command {
	var (
		assignmentID string
		ipID         string
		protocol     string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List port assignments",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireAPIKey(cmd); err != nil {
				return err
			}

			filters := storage.PortAssignmentFilters{
				AssignmentID: assignmentID,
				IPID:         ipID,
				Protocol:     protocol,
			}

			c := client.New(endpoint, apiKey)
			assignments, err := c.ListPortAssignments(context.Background(), filters)
			if err != nil {
				return err
			}

			printJSON(assignments)
			return nil
		},
	}

	cmd.Flags().StringVar(&assignmentID, "assignment", "", "Filter by service assignment ID")
	cmd.Flags().StringVar(&ipID, "ip", "", "Filter by IP address ID")
	cmd.Flags().StringVar(&protocol, "protocol", "", "Filter by protocol (tcp, udp, icmp, all)")

	cmd.RegisterFlagCompletionFunc("assignment", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return completeAssignmentIDs(toComplete), cobra.ShellCompDirectiveNoFileComp
	})

	cmd.RegisterFlagCompletionFunc("ip", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return completeIPIDs(toComplete), cobra.ShellCompDirectiveNoFileComp
	})

	cmd.RegisterFlagCompletionFunc("protocol", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"tcp", "udp", "icmp", "all"}, cobra.ShellCompDirectiveNoFileComp
	})

	return cmd
}

func newPortGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get [id]",
		Short: "Get port assignment details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireAPIKey(cmd); err != nil {
				return err
			}

			c := client.New(endpoint, apiKey)
			assignment, err := c.GetPortAssignment(context.Background(), args[0])
			if err != nil {
				return err
			}

			printJSON(assignment)
			return nil
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				return completePortIDs(toComplete), cobra.ShellCompDirectiveNoFileComp
			}
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
	}

	return cmd
}

func newPortCreateCmd() *cobra.Command {
	var (
		assignmentID string
		ipID         string
		port         int
		protocol     string
		servicePort  int
		description  string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new port assignment",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireAPIKey(cmd); err != nil {
				return err
			}

			assignment := &domain.PortAssignment{
				ID:           uuid.New().String(),
				AssignmentID: assignmentID,
				IPID:         ipID,
				Port:         port,
				Protocol:     domain.Protocol(protocol),
				ServicePort:  servicePort,
				Description:  description,
				CreatedAt:    time.Now(),
			}

			c := client.New(endpoint, apiKey)
			result, err := c.CreatePortAssignment(context.Background(), assignment)
			if err != nil {
				return err
			}

			printJSON(result)
			return nil
		},
	}

	cmd.Flags().StringVar(&assignmentID, "assignment", "", "Service assignment ID (required)")
	cmd.Flags().StringVar(&ipID, "ip", "", "IP address ID (required)")
	cmd.Flags().IntVar(&port, "port", 0, "External port number (required)")
	cmd.Flags().StringVar(&protocol, "protocol", "tcp", "Protocol: tcp, udp, icmp, all")
	cmd.Flags().IntVar(&servicePort, "service-port", 0, "Internal service port (required)")
	cmd.Flags().StringVar(&description, "description", "", "Description")

	cmd.MarkFlagRequired("assignment")
	cmd.MarkFlagRequired("ip")
	cmd.MarkFlagRequired("port")
	cmd.MarkFlagRequired("service-port")

	cmd.RegisterFlagCompletionFunc("assignment", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return completeAssignmentIDs(toComplete), cobra.ShellCompDirectiveNoFileComp
	})

	cmd.RegisterFlagCompletionFunc("ip", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return completeIPIDs(toComplete), cobra.ShellCompDirectiveNoFileComp
	})

	cmd.RegisterFlagCompletionFunc("protocol", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"tcp", "udp", "icmp", "all"}, cobra.ShellCompDirectiveNoFileComp
	})

	return cmd
}

func newPortDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete [id]",
		Short: "Delete a port assignment",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireAPIKey(cmd); err != nil {
				return err
			}

			c := client.New(endpoint, apiKey)
			if err := c.DeletePortAssignment(context.Background(), args[0]); err != nil {
				return err
			}

			printJSON(map[string]string{"message": "port assignment deleted successfully"})
			return nil
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				return completePortIDs(toComplete), cobra.ShellCompDirectiveNoFileComp
			}
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
	}

	return cmd
}

func completePortIDs(toComplete string) []string {
	if apiKey == "" {
		return nil
	}

	c := client.New(endpoint, apiKey)
	assignments, err := c.ListPortAssignments(context.Background(), storage.PortAssignmentFilters{})
	if err != nil {
		return nil
	}

	var completions []string
	for _, assignment := range assignments {
		// Format: ID \t Port:ServicePort (Protocol)
		completions = append(completions, assignment.ID+"\t"+
			string(assignment.Protocol)+":"+
			string(rune(assignment.Port))+
			"->"+
			string(rune(assignment.ServicePort)))
	}

	return completions
}
