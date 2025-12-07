package cli

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/studiowebux/kubebuddy/internal/client"
	"github.com/studiowebux/kubebuddy/internal/domain"
	"github.com/studiowebux/kubebuddy/internal/storage"
)

func newComponentCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "component",
		Short: "Manage hardware components",
		Long:  `Manage hardware components (CPU, RAM, Storage, GPU, NIC, PSU)`,
	}

	cmd.AddCommand(newComponentListCmd())
	cmd.AddCommand(newComponentGetCmd())
	cmd.AddCommand(newComponentCreateCmd())
	cmd.AddCommand(newComponentDeleteCmd())
	cmd.AddCommand(newComponentAssignCmd())
	cmd.AddCommand(newComponentUnassignCmd())
	cmd.AddCommand(newComponentListAssignmentsCmd())

	return cmd
}

func newComponentListCmd() *cobra.Command {
	var (
		componentType string
		manufacturer  string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List components",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireAPIKey(cmd); err != nil {
				return err
			}

			filters := storage.ComponentFilters{
				Type:         componentType,
				Manufacturer: manufacturer,
			}

			c := client.New(endpoint, apiKey)
			components, err := c.ListComponents(context.Background(), filters)
			if err != nil {
				return err
			}

			printJSON(components)
			return nil
		},
	}

	cmd.Flags().StringVar(&componentType, "type", "", "Filter by component type")
	cmd.Flags().StringVar(&manufacturer, "manufacturer", "", "Filter by manufacturer")

	cmd.RegisterFlagCompletionFunc("type", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return domain.ComponentTypes(), cobra.ShellCompDirectiveNoFileComp
	})

	return cmd
}

func newComponentGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get [id]",
		Short: "Get component details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireAPIKey(cmd); err != nil {
				return err
			}

			c := client.New(endpoint, apiKey)
			component, err := c.GetComponent(context.Background(), args[0])
			if err != nil {
				return err
			}

			printJSON(component)
			return nil
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				return completeComponentIDs(toComplete), cobra.ShellCompDirectiveNoFileComp
			}
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
	}

	return cmd
}

func newComponentCreateCmd() *cobra.Command {
	var (
		name         string
		componentType string
		manufacturer string
		model        string
		specs        string
		notes        string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new component",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireAPIKey(cmd); err != nil {
				return err
			}

			specsMap := make(map[string]interface{})
			if specs != "" {
				if err := json.Unmarshal([]byte(specs), &specsMap); err != nil {
					return err
				}
			}

			component := &domain.Component{
				ID:           uuid.New().String(),
				Name:         name,
				Type:         domain.ComponentType(componentType),
				Manufacturer: manufacturer,
				Model:        model,
				Specs:        specsMap,
				Notes:        notes,
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			}

			c := client.New(endpoint, apiKey)
			result, err := c.CreateComponent(context.Background(), component)
			if err != nil {
				return err
			}

			printJSON(result)
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Component name (required)")
	cmd.Flags().StringVar(&componentType, "type", "", "Component type (required)")
	cmd.Flags().StringVar(&manufacturer, "manufacturer", "", "Manufacturer (required)")
	cmd.Flags().StringVar(&model, "model", "", "Model (required)")
	cmd.Flags().StringVar(&specs, "specs", "", "Specs as JSON (e.g. '{\"cores\":8,\"ghz\":3.5}')")
	cmd.Flags().StringVar(&notes, "notes", "", "Notes")

	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("type")
	cmd.MarkFlagRequired("manufacturer")
	cmd.MarkFlagRequired("model")

	cmd.RegisterFlagCompletionFunc("type", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return domain.ComponentTypes(), cobra.ShellCompDirectiveNoFileComp
	})

	return cmd
}

func newComponentDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete [id]",
		Short: "Delete a component",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireAPIKey(cmd); err != nil {
				return err
			}

			c := client.New(endpoint, apiKey)
			if err := c.DeleteComponent(context.Background(), args[0]); err != nil {
				return err
			}

			printJSON(map[string]string{"message": "component deleted successfully"})
			return nil
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				return completeComponentIDs(toComplete), cobra.ShellCompDirectiveNoFileComp
			}
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
	}

	return cmd
}

func newComponentAssignCmd() *cobra.Command {
	var (
		computeID   string
		componentID string
		quantity    int
		slot        string
		serialNo    string
		notes       string
		raidLevel   string
		raidGroup   string
	)

	cmd := &cobra.Command{
		Use:   "assign",
		Short: "Assign a component to a compute",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireAPIKey(cmd); err != nil {
				return err
			}

			assignment := &domain.ComputeComponent{
				ID:          uuid.New().String(),
				ComputeID:   computeID,
				ComponentID: componentID,
				Quantity:    quantity,
				Slot:        slot,
				SerialNo:    serialNo,
				Notes:       notes,
				RaidLevel:   domain.RaidLevel(raidLevel),
				RaidGroup:   raidGroup,
				CreatedAt:   time.Now(),
			}

			c := client.New(endpoint, apiKey)
			result, err := c.AssignComponent(context.Background(), assignment)
			if err != nil {
				return err
			}

			printJSON(result)
			return nil
		},
	}

	cmd.Flags().StringVar(&computeID, "compute", "", "Compute ID (required)")
	cmd.Flags().StringVar(&componentID, "component", "", "Component ID (required)")
	cmd.Flags().IntVar(&quantity, "quantity", 1, "Quantity")
	cmd.Flags().StringVar(&slot, "slot", "", "Physical slot (e.g., CPU1, DIMM0-3)")
	cmd.Flags().StringVar(&serialNo, "serial", "", "Serial number")
	cmd.Flags().StringVar(&notes, "notes", "", "Notes")
	cmd.Flags().StringVar(&raidLevel, "raid", "", "RAID level for storage (raid0, raid1, raid5, raid6, raid10)")
	cmd.Flags().StringVar(&raidGroup, "raid-group", "", "RAID group ID (storage components in same group form RAID array)")

	cmd.MarkFlagRequired("compute")
	cmd.MarkFlagRequired("component")

	cmd.RegisterFlagCompletionFunc("compute", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return completeComputeIDs(toComplete), cobra.ShellCompDirectiveNoFileComp
	})

	cmd.RegisterFlagCompletionFunc("component", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return completeComponentIDs(toComplete), cobra.ShellCompDirectiveNoFileComp
	})

	return cmd
}

func newComponentUnassignCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unassign [assignment-id]",
		Short: "Unassign a component from a compute",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireAPIKey(cmd); err != nil {
				return err
			}

			c := client.New(endpoint, apiKey)
			if err := c.UnassignComponent(context.Background(), args[0]); err != nil {
				return err
			}

			printJSON(map[string]string{"message": "component unassigned successfully"})
			return nil
		},
	}

	return cmd
}

func newComponentListAssignmentsCmd() *cobra.Command {
	var computeID string
	var componentID string

	cmd := &cobra.Command{
		Use:   "list-assignments",
		Short: "List component assignments",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireAPIKey(cmd); err != nil {
				return err
			}

			filters := storage.ComputeComponentFilters{
				ComputeID:   computeID,
				ComponentID: componentID,
			}

			c := client.New(endpoint, apiKey)
			assignments, err := c.ListComponentAssignments(context.Background(), filters)
			if err != nil {
				return err
			}

			printJSON(assignments)
			return nil
		},
	}

	cmd.Flags().StringVar(&computeID, "compute", "", "Filter by compute ID")
	cmd.Flags().StringVar(&componentID, "component", "", "Filter by component ID")

	cmd.RegisterFlagCompletionFunc("compute", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return completeComputeIDs(toComplete), cobra.ShellCompDirectiveNoFileComp
	})

	cmd.RegisterFlagCompletionFunc("component", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return completeComponentIDs(toComplete), cobra.ShellCompDirectiveNoFileComp
	})

	return cmd
}

func completeComponentIDs(toComplete string) []string {
	if apiKey == "" {
		return nil
	}

	c := client.New(endpoint, apiKey)
	components, err := c.ListComponents(context.Background(), storage.ComponentFilters{})
	if err != nil {
		return nil
	}

	var completions []string

	for _, component := range components {
		// Format: ID \t Name (Manufacturer Model)
		completions = append(completions, component.ID+"\t"+component.Name+" ("+component.Manufacturer+" "+component.Model+")")
	}

	return completions
}
