package cli

import (
	"context"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/studiowebux/kubebuddy/internal/client"
	"github.com/studiowebux/kubebuddy/internal/domain"
	"github.com/studiowebux/kubebuddy/internal/storage"
)

func newJournalCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "journal",
		Short: "Manage journal entries",
		Long:  `Manage per-compute journal entries for maintenance, incidents, and deployments`,
	}

	cmd.AddCommand(newJournalListCmd())
	cmd.AddCommand(newJournalAddCmd())

	return cmd
}

func newJournalListCmd() *cobra.Command {
	var computeID string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List journal entries",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireAPIKey(cmd); err != nil {
				return err
			}

			filters := storage.JournalFilters{
				ComputeID: computeID,
			}

			c := client.New(endpoint, apiKey)
			entries, err := c.ListJournalEntries(context.Background(), filters)
			if err != nil {
				return err
			}

			printJSON(entries)
			return nil
		},
	}

	cmd.Flags().StringVar(&computeID, "compute", "", "Filter by compute ID")

	cmd.RegisterFlagCompletionFunc("compute", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return completeComputeIDs(toComplete), cobra.ShellCompDirectiveNoFileComp
	})

	return cmd
}

func newJournalAddCmd() *cobra.Command {
	var (
		computeID string
		category  string
		content   string
	)

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a journal entry",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireAPIKey(cmd); err != nil {
				return err
			}

			entry := &domain.JournalEntry{
				ID:        uuid.New().String(),
				ComputeID: computeID,
				Category:  category,
				Content:   content,
			}

			c := client.New(endpoint, apiKey)
			result, err := c.CreateJournalEntry(context.Background(), entry)
			if err != nil {
				return err
			}

			printJSON(result)
			return nil
		},
	}

	cmd.Flags().StringVar(&computeID, "compute", "", "Compute ID (required)")
	cmd.Flags().StringVar(&category, "category", "other", "Entry category")
	cmd.Flags().StringVar(&content, "content", "", "Entry content (plain text or markdown, required)")

	cmd.MarkFlagRequired("compute")
	cmd.MarkFlagRequired("content")

	// Add completion for compute flag
	cmd.RegisterFlagCompletionFunc("compute", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return completeComputeIDs(toComplete), cobra.ShellCompDirectiveNoFileComp
	})

	// Add completion for category flag
	cmd.RegisterFlagCompletionFunc("category", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{
			"maintenance",
			"incident",
			"deployment",
			"hardware",
			"network",
			"other",
		}, cobra.ShellCompDirectiveNoFileComp
	})

	return cmd
}
