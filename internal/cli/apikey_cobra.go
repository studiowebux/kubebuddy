package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/studiowebux/kubebuddy/internal/client"
	"github.com/studiowebux/kubebuddy/internal/domain"
)

func newAPIKeyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apikey",
		Short: "Manage API keys (admin only)",
		Long:  `Manage API keys for authentication`,
	}

	cmd.AddCommand(newAPIKeyListCmd())
	cmd.AddCommand(newAPIKeyCreateCmd())
	cmd.AddCommand(newAPIKeyDeleteCmd())

	return cmd
}

func newAPIKeyListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all API keys",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireAPIKey(cmd); err != nil {
				return err
			}

			c := client.New(endpoint, apiKey)
			keys, err := c.ListAPIKeys(context.Background())
			if err != nil {
				return err
			}

			printJSON(keys)
			return nil
		},
	}
}

func newAPIKeyCreateCmd() *cobra.Command {
	var (
		name        string
		scope       string
		description string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new API key",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireAPIKey(cmd); err != nil {
				return err
			}

			req := client.CreateAPIKeyRequest{
				Name:        name,
				Scope:       domain.APIKeyScope(scope),
				Description: description,
			}

			c := client.New(endpoint, apiKey)
			result, err := c.CreateAPIKey(context.Background(), req)
			if err != nil {
				return err
			}

			fmt.Printf("API Key created successfully!\n")
			fmt.Printf("ID: %s\n", result.APIKey.ID)
			fmt.Printf("Name: %s\n", result.APIKey.Name)
			fmt.Printf("Scope: %s\n", result.APIKey.Scope)
			fmt.Printf("Key: %s\n", result.Key)
			fmt.Printf("\nIMPORTANT: Save this key now, it will not be shown again!\n")

			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "API key name (required)")
	cmd.Flags().StringVar(&scope, "scope", "readonly", "API key scope (admin, readwrite, readonly)")
	cmd.Flags().StringVar(&description, "description", "", "API key description")

	cmd.MarkFlagRequired("name")

	cmd.RegisterFlagCompletionFunc("scope", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"admin", "readwrite", "readonly"}, cobra.ShellCompDirectiveNoFileComp
	})

	return cmd
}

func newAPIKeyDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete an API key",
		Args:  cobra.ExactArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			return completeAPIKeyIDs(toComplete), cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireAPIKey(cmd); err != nil {
				return err
			}

			c := client.New(endpoint, apiKey)
			if err := c.DeleteAPIKey(context.Background(), args[0]); err != nil {
				return err
			}

			fmt.Println("API key deleted successfully")
			return nil
		},
	}

	return cmd
}

func completeAPIKeyIDs(toComplete string) []string {
	if apiKey == "" {
		return nil
	}

	c := client.New(endpoint, apiKey)
	keys, err := c.ListAPIKeys(context.Background())
	if err != nil {
		return nil
	}

	var completions []string
	for _, key := range keys {
		completions = append(completions, fmt.Sprintf("%s\t%s", key.ID, key.Name))
	}

	return completions
}
