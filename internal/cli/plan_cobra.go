package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/studiowebux/kubebuddy/internal/client"
	"github.com/studiowebux/kubebuddy/internal/domain"
	"github.com/studiowebux/kubebuddy/internal/storage"
)

// Helper to extract float values from component specs with multiple possible keys
func getComponentSpecFloat(specs map[string]interface{}, keys ...string) float64 {
	for _, key := range keys {
		if val, ok := specs[key]; ok {
			switch v := val.(type) {
			case float64:
				return v
			case int:
				return float64(v)
			}
		}
	}
	return 0
}

// Helper to convert interface{} to int (handles both int and float64)
func getIntValue(val interface{}) int {
	switch v := val.(type) {
	case int:
		return v
	case float64:
		return int(v)
	default:
		return 0
	}
}

func newPlanCmd() *cobra.Command {
	var jsonOutput bool
	var computeID string
	var assignFlag bool
	var forceFlag bool

	cmd := &cobra.Command{
		Use:   "plan <service-id>",
		Short: "Plan capacity for a service",
		Long:  `Evaluate capacity and get placement recommendations for a service`,
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
			ctx := context.Background()

			// Resolve service ID or name
			service, err := c.ResolveService(ctx, args[0])
			if err != nil {
				return fmt.Errorf("failed to resolve service: %w", err)
			}

			// Resolve compute ID or name if specified
			resolvedComputeID := ""
			if computeID != "" {
				compute, err := c.ResolveCompute(ctx, computeID)
				if err != nil {
					return fmt.Errorf("failed to resolve compute: %w", err)
				}
				resolvedComputeID = compute.ID
			}

			request := domain.PlanRequest{
				ServiceID: service.ID,
				Constraints: domain.Constraints{
					ComputeID: resolvedComputeID,
				},
			}

			result, err := c.PlanCapacity(context.Background(), request)
			if err != nil {
				return err
			}

			// JSON output if requested
			if jsonOutput {
				printJSON(result)
				return nil
			}

			// User-friendly output

			fmt.Printf("# Capacity Planning: %s\n\n", service.Name)

			if result.Feasible {
				fmt.Printf("✓ Feasible - Found %d candidate(s)\n\n", len(result.Candidates))

				for i, candidate := range result.Candidates {
					fmt.Printf("## Candidate %d: %s\n\n", i+1, candidate.Compute.Name)
					fmt.Printf("- **Provider:** %s\n", candidate.Compute.Provider)
					fmt.Printf("- **Region:** %s\n", candidate.Compute.Region)
					fmt.Printf("- **Score:** %.1f\n", candidate.Score)
					fmt.Printf("- **Utilization After:** %.0f%%\n", candidate.UtilizationAfter*100)

					// Use RAID-aware resources from compute object (already calculated by API)
					totalResources := candidate.Compute.Resources

					// Get assignments for this compute to show allocated resources
					assignments, err := c.ListAssignments(ctx, storage.AssignmentFilters{
						ComputeID: candidate.Compute.ID,
					})

					var allocatedCores int
					var allocatedMemoryMB float64
					var allocatedVRAMMB float64
					var allocatedStorageGB float64

					if err == nil {
						for _, assignment := range assignments {
							if cores, ok := assignment.Allocated["cores"]; ok {
								switch v := cores.(type) {
								case int:
									allocatedCores += v
								case float64:
									allocatedCores += int(v)
								}
							}
							if mem, ok := assignment.Allocated["memory"]; ok {
								switch v := mem.(type) {
								case int:
									allocatedMemoryMB += float64(v)
								case float64:
									allocatedMemoryMB += v
								}
							}
							if vram, ok := assignment.Allocated["vram"]; ok {
								switch v := vram.(type) {
								case int:
									allocatedVRAMMB += float64(v)
								case float64:
									allocatedVRAMMB += v
								}
							}
							if nvme, ok := assignment.Allocated["nvme"]; ok {
								switch v := nvme.(type) {
								case int:
									allocatedStorageGB += float64(v)
								case float64:
									allocatedStorageGB += v
								}
							}
						}
					}

					// Convert allocated to GB for display
					allocatedMemoryGB := allocatedMemoryMB / 1024
					allocatedVRAMGB := allocatedVRAMMB / 1024

					fmt.Println()
					fmt.Println("**Hardware:**")
					if cores, ok := totalResources["cores"]; ok {
						totalCores := getIntValue(cores)
						if totalCores > 0 {
							fmt.Printf("- Cores: %d / %d\n", allocatedCores, totalCores)
						}
					}
					if memory, ok := totalResources["memory"]; ok {
						totalMemoryGB := getIntValue(memory)
						if totalMemoryGB > 0 {
							fmt.Printf("- Memory: %.0f GB / %d GB\n", allocatedMemoryGB, totalMemoryGB)
						}
					}
					if vram, ok := totalResources["vram"]; ok {
						totalVRAMGB := getIntValue(vram)
						if totalVRAMGB > 0 {
							fmt.Printf("- VRAM: %.0f GB / %d GB\n", allocatedVRAMGB, totalVRAMGB)
						}
					}
					if nvme, ok := totalResources["nvme"]; ok {
						totalStorageGB := getIntValue(nvme)
						if totalStorageGB > 0 {
							fmt.Printf("- Storage: %.0f GB / %d GB\n", allocatedStorageGB, totalStorageGB)
						}
					}

					fmt.Println()
				}
			} else {
				fmt.Println("✗ Not feasible - No suitable compute found")
				if result.Message != "" {
					fmt.Printf("\n%s\n", result.Message)
				}

				// Show available computes and their resources
				fmt.Println("\n**Available Computes:**")
				computes, err := c.ListComputes(ctx, storage.ComputeFilters{})
				if err == nil {
					for _, compute := range computes {
						// Filter to specific compute if requested
						if resolvedComputeID != "" && compute.ID != resolvedComputeID {
							continue
						}
						fmt.Printf("\n### %s\n", compute.Name)
						fmt.Printf("- **State:** %s\n", compute.State)
						fmt.Printf("- **Provider:** %s\n", compute.Provider)
						fmt.Printf("- **Region:** %s\n", compute.Region)

						// Calculate RAID-aware resources from components
						componentAssignments, err := c.ListComponentAssignments(ctx, storage.ComputeComponentFilters{
							ComputeID: compute.ID,
						})
						if err != nil {
							continue
						}

						// Load component details
						components := make([]*domain.Component, 0)
						for _, ca := range componentAssignments {
							comp, err := c.GetComponent(ctx, ca.ComponentID)
							if err == nil {
								components = append(components, comp)
							}
						}

						// Calculate total resources with RAID support
						totalResources := compute.GetTotalResourcesFromComponents(components, componentAssignments)

						// Get assignments for this compute to show allocated resources
						assignments, err := c.ListAssignments(ctx, storage.AssignmentFilters{
							ComputeID: compute.ID,
						})

						var allocatedCores int
						var allocatedMemoryMB float64
						var allocatedVRAMMB float64
						var allocatedStorageGB float64

						if err == nil {
							for _, assignment := range assignments {
								if cores, ok := assignment.Allocated["cores"]; ok {
									switch v := cores.(type) {
									case int:
										allocatedCores += v
									case float64:
										allocatedCores += int(v)
									}
								}
								if mem, ok := assignment.Allocated["memory"]; ok {
									switch v := mem.(type) {
									case int:
										allocatedMemoryMB += float64(v)
									case float64:
										allocatedMemoryMB += v
									}
								}
								if vram, ok := assignment.Allocated["vram"]; ok {
									switch v := vram.(type) {
									case int:
										allocatedVRAMMB += float64(v)
									case float64:
										allocatedVRAMMB += v
									}
								}
								if nvme, ok := assignment.Allocated["nvme"]; ok {
									switch v := nvme.(type) {
									case int:
										allocatedStorageGB += float64(v)
									case float64:
										allocatedStorageGB += v
									}
								}
							}
						}

						// Convert allocated to GB for display
						allocatedMemoryGB := allocatedMemoryMB / 1024
						allocatedVRAMGB := allocatedVRAMMB / 1024

						fmt.Println("\n**Hardware:**")
						if cores, ok := totalResources["cores"]; ok {
							totalCores := getIntValue(cores)
							if totalCores > 0 {
								fmt.Printf("- Cores: %d / %d\n", allocatedCores, totalCores)
							}
						}
						if memory, ok := totalResources["memory"]; ok {
							totalMemoryGB := getIntValue(memory)
							if totalMemoryGB > 0 {
								fmt.Printf("- Memory: %.0f GB / %d GB\n", allocatedMemoryGB, totalMemoryGB)
							}
						}
						if vram, ok := totalResources["vram"]; ok {
							totalVRAMGB := getIntValue(vram)
							if totalVRAMGB > 0 {
								fmt.Printf("- VRAM: %.0f GB / %d GB\n", allocatedVRAMGB, totalVRAMGB)
							}
						}
						if nvme, ok := totalResources["nvme"]; ok {
							totalStorageGB := getIntValue(nvme)
							if totalStorageGB > 0 {
								fmt.Printf("- Storage: %.0f GB / %d GB\n", allocatedStorageGB, totalStorageGB)
							}
						}
					}
				}

				if len(result.Recommendations) > 0 {
					fmt.Println("\n**Recommendations:**")
					for _, rec := range result.Recommendations {
						fmt.Printf("- %s: %d x %s\n", rec.Rationale, rec.Quantity, rec.Type)
						if len(rec.Spec) > 0 {
							for k, v := range rec.Spec {
								fmt.Printf("  - %s: %v\n", k, v)
							}
						}
					}
				}
			}

			// Handle assignment creation if --assign flag is used
			if assignFlag {
				if result.Feasible && len(result.Candidates) > 0 {
					// Assign to best candidate (first in list)
					targetCompute := result.Candidates[0].Compute
					fmt.Printf("\nCreating assignment on %s...\n", targetCompute.Name)

					assignment := &domain.Assignment{
						ServiceID: service.ID,
						ComputeID: targetCompute.ID,
						Allocated: service.MaxSpec,
					}

					created, err := c.CreateAssignment(ctx, assignment, false)
					if err != nil {
						return fmt.Errorf("failed to create assignment: %w", err)
					}
					fmt.Printf("✓ Assignment created: %s\n", created.ID)
				} else if !result.Feasible && forceFlag && resolvedComputeID != "" {
					// Force assignment to specific compute
					fmt.Printf("\n⚠ Forcing assignment despite insufficient resources...\n")

					assignment := &domain.Assignment{
						ServiceID: service.ID,
						ComputeID: resolvedComputeID,
						Allocated: service.MaxSpec,
					}

					created, err := c.CreateAssignment(ctx, assignment, true)
					if err != nil {
						return fmt.Errorf("failed to force assignment: %w", err)
					}
					fmt.Printf("✓ Assignment created (forced): %s\n", created.ID)
				} else if !result.Feasible && forceFlag {
					return fmt.Errorf("--force requires --compute to be specified when no candidates are available")
				} else {
					return fmt.Errorf("no suitable compute found, use --force --compute <id> to override")
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	cmd.Flags().StringVar(&computeID, "compute", "", "Plan for specific compute ID or name (optional)")
	cmd.Flags().BoolVar(&assignFlag, "assign", false, "Create assignment on best candidate")
	cmd.Flags().BoolVar(&forceFlag, "force", false, "Force assignment even if resources insufficient (requires --assign)")

	// Add auto-completion for --compute flag
	cmd.RegisterFlagCompletionFunc("compute", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return completeComputeIDs(toComplete), cobra.ShellCompDirectiveNoFileComp
	})

	return cmd
}
