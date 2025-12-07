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

func newDNSCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dns",
		Short: "Manage DNS records",
		Long:  `Manage DNS records (A, AAAA, CNAME, PTR)`,
	}

	cmd.AddCommand(newDNSListCmd())
	cmd.AddCommand(newDNSGetCmd())
	cmd.AddCommand(newDNSCreateCmd())
	cmd.AddCommand(newDNSDeleteCmd())

	return cmd
}

func newDNSListCmd() *cobra.Command {
	var (
		recordType string
		zone       string
		ipID       string
		name       string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List DNS records",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireAPIKey(cmd); err != nil {
				return err
			}

			filters := storage.DNSRecordFilters{
				Type: recordType,
				Zone: zone,
				IPID: ipID,
				Name: name,
			}

			c := client.New(endpoint, apiKey)
			records, err := c.ListDNSRecords(context.Background(), filters)
			if err != nil {
				return err
			}

			printJSON(records)
			return nil
		},
	}

	cmd.Flags().StringVar(&recordType, "type", "", "Filter by record type (A, AAAA, CNAME, PTR)")
	cmd.Flags().StringVar(&zone, "zone", "", "Filter by zone")
	cmd.Flags().StringVar(&ipID, "ip", "", "Filter by IP address ID")
	cmd.Flags().StringVar(&name, "name", "", "Filter by name (partial match)")

	cmd.RegisterFlagCompletionFunc("type", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"A", "AAAA", "CNAME", "PTR"}, cobra.ShellCompDirectiveNoFileComp
	})

	cmd.RegisterFlagCompletionFunc("ip", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return completeIPIDs(toComplete), cobra.ShellCompDirectiveNoFileComp
	})

	return cmd
}

func newDNSGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get [id]",
		Short: "Get DNS record details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireAPIKey(cmd); err != nil {
				return err
			}

			c := client.New(endpoint, apiKey)
			record, err := c.GetDNSRecord(context.Background(), args[0])
			if err != nil {
				return err
			}

			printJSON(record)
			return nil
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				return completeDNSIDs(toComplete), cobra.ShellCompDirectiveNoFileComp
			}
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
	}

	return cmd
}

func newDNSCreateCmd() *cobra.Command {
	var (
		name       string
		recordType string
		value      string
		zone       string
		ttl        int
		ipID       string
		notes      string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new DNS record",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireAPIKey(cmd); err != nil {
				return err
			}

			if ttl == 0 {
				ttl = 3600 // Default TTL
			}

			record := &domain.DNSRecord{
				ID:        uuid.New().String(),
				Name:      name,
				Type:      domain.DNSRecordType(recordType),
				Value:     value,
				IPID:      ipID,
				TTL:       ttl,
				Zone:      zone,
				Notes:     notes,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			c := client.New(endpoint, apiKey)
			result, err := c.CreateDNSRecord(context.Background(), record)
			if err != nil {
				return err
			}

			printJSON(result)
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "DNS record name (e.g., www.example.com) (required)")
	cmd.Flags().StringVar(&recordType, "type", "", "Record type: A, AAAA, CNAME, PTR (required)")
	cmd.Flags().StringVar(&value, "value", "", "Record value (IP or hostname) (required)")
	cmd.Flags().StringVar(&zone, "zone", "", "DNS zone (e.g., example.com) (required)")
	cmd.Flags().IntVar(&ttl, "ttl", 3600, "TTL in seconds")
	cmd.Flags().StringVar(&ipID, "ip", "", "Link to IP address ID (optional)")
	cmd.Flags().StringVar(&notes, "notes", "", "Notes")

	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("type")
	cmd.MarkFlagRequired("value")
	cmd.MarkFlagRequired("zone")

	cmd.RegisterFlagCompletionFunc("type", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"A", "AAAA", "CNAME", "PTR"}, cobra.ShellCompDirectiveNoFileComp
	})

	cmd.RegisterFlagCompletionFunc("ip", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return completeIPIDs(toComplete), cobra.ShellCompDirectiveNoFileComp
	})

	return cmd
}

func newDNSDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete [id]",
		Short: "Delete a DNS record",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireAPIKey(cmd); err != nil {
				return err
			}

			c := client.New(endpoint, apiKey)
			if err := c.DeleteDNSRecord(context.Background(), args[0]); err != nil {
				return err
			}

			printJSON(map[string]string{"message": "DNS record deleted successfully"})
			return nil
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				return completeDNSIDs(toComplete), cobra.ShellCompDirectiveNoFileComp
			}
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
	}

	return cmd
}

func completeDNSIDs(toComplete string) []string {
	if apiKey == "" {
		return nil
	}

	c := client.New(endpoint, apiKey)
	records, err := c.ListDNSRecords(context.Background(), storage.DNSRecordFilters{})
	if err != nil {
		return nil
	}

	var completions []string
	for _, record := range records {
		// Format: ID \t Name (Type - Zone)
		completions = append(completions, record.ID+"\t"+record.Name+" ("+string(record.Type)+" - "+record.Zone+")")
	}

	return completions
}
