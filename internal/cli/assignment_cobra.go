package cli

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/studiowebux/kubebuddy/internal/client"
	"github.com/studiowebux/kubebuddy/internal/domain"
	"github.com/studiowebux/kubebuddy/internal/storage"
)

func newAssignmentCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "assignment",
		Short: "Manage service assignments",
		Long:  `Manage service-to-compute assignments`,
	}

	cmd.AddCommand(newAssignmentListCmd())
	cmd.AddCommand(newAssignmentCreateCmd())
	cmd.AddCommand(newAssignmentDeleteCmd())

	return cmd
}

func newAssignmentListCmd() *cobra.Command {
	var computeID string
	var serviceID string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all assignments",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireAPIKey(cmd); err != nil {
				return err
			}

			c := client.New(endpoint, apiKey)
			assignments, err := c.ListAssignments(context.Background(), storage.AssignmentFilters{
				ComputeID: computeID,
				ServiceID: serviceID,
			})
			if err != nil {
				return err
			}

			printJSON(assignments)
			return nil
		},
	}

	cmd.Flags().StringVar(&computeID, "compute", "", "Filter by compute ID")
	cmd.Flags().StringVar(&serviceID, "service", "", "Filter by service ID")

	cmd.RegisterFlagCompletionFunc("compute", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return completeComputeIDs(toComplete), cobra.ShellCompDirectiveNoFileComp
	})

	cmd.RegisterFlagCompletionFunc("service", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return completeServiceIDs(toComplete), cobra.ShellCompDirectiveNoFileComp
	})

	return cmd
}

func newAssignmentCreateCmd() *cobra.Command {
	var (
		serviceID string
		computeID string
		force     bool
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new assignment",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireAPIKey(cmd); err != nil {
				return err
			}

			assignment := &domain.Assignment{
				ID:        uuid.New().String(),
				ServiceID: serviceID,
				ComputeID: computeID,
			}

			c := client.New(endpoint, apiKey)
			result, err := c.CreateAssignment(context.Background(), assignment, force)
			if err != nil {
				return err
			}

			printJSON(result)
			return nil
		},
	}

	cmd.Flags().StringVar(&serviceID, "service", "", "Service ID (required)")
	cmd.Flags().StringVar(&computeID, "compute", "", "Compute ID (required)")
	cmd.Flags().BoolVar(&force, "force", false, "Force assignment even if resources insufficient")

	cmd.MarkFlagRequired("service")
	cmd.MarkFlagRequired("compute")

	cmd.RegisterFlagCompletionFunc("service", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return completeServiceIDs(toComplete), cobra.ShellCompDirectiveNoFileComp
	})

	cmd.RegisterFlagCompletionFunc("compute", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return completeComputeIDs(toComplete), cobra.ShellCompDirectiveNoFileComp
	})

	return cmd
}

func newAssignmentDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete an assignment",
		Args:  cobra.ExactArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			return completeAssignmentIDs(toComplete), cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireAPIKey(cmd); err != nil {
				return err
			}

			c := client.New(endpoint, apiKey)
			if err := c.DeleteAssignment(context.Background(), args[0]); err != nil {
				return err
			}

			fmt.Println("Assignment deleted successfully")
			return nil
		},
	}

	return cmd
}

func completeAssignmentIDs(toComplete string) []string {
	if apiKey == "" {
		return nil
	}

	c := client.New(endpoint, apiKey)
	assignments, err := c.ListAssignments(context.Background(), storage.AssignmentFilters{})
	if err != nil {
		return nil
	}

	var completions []string
	for _, assignment := range assignments {
		completions = append(completions, assignment.ID)
	}

	return completions
}
