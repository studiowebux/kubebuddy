package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/studiowebux/kubebuddy/internal/client"
	"github.com/studiowebux/kubebuddy/internal/storage"
)

func newReportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "report",
		Short: "Generate system reports",
		Long:  `Generate detailed reports about the system state`,
	}

	cmd.AddCommand(newReportComputeCmd())

	return cmd
}

func newReportComputeCmd() *cobra.Command {
	var computeID string
	var detailedJournal bool

	cmd := &cobra.Command{
		Use:   "compute [id]",
		Short: "Generate markdown report for a compute",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireAPIKey(cmd); err != nil {
				return err
			}

			c := client.New(endpoint, apiKey)

			// If ID provided as argument, use it
			if len(args) > 0 {
				computeID = args[0]
			}

			// If no ID specified, generate reports for all computes
			if computeID == "" {
				computes, err := c.ListComputes(context.Background(), storage.ComputeFilters{})
				if err != nil {
					return err
				}

				for i, compute := range computes {
					if i > 0 {
						fmt.Println("\n---\n")
					}
					if err := printComputeReport(c, compute.ID, detailedJournal); err != nil {
						return err
					}
				}
				return nil
			}

			// Generate report for specific compute
			return printComputeReport(c, computeID, detailedJournal)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				return completeComputeIDs(toComplete), cobra.ShellCompDirectiveNoFileComp
			}
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
	}

	cmd.Flags().BoolVar(&detailedJournal, "journal", false, "Show detailed journal entries with full content")

	return cmd
}

// storageInfo holds information about storage components
type storageInfo struct {
	size      float64
	quantity  int
	raidLevel string
	name      string
}

// Helper to extract float values from component specs with multiple possible keys
func getSpecFloat(specs map[string]interface{}, keys ...string) float64 {
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

func printComputeReport(c *client.Client, computeID string, detailedJournal bool) error {
	ctx := context.Background()

	// Get compute
	compute, err := c.GetCompute(ctx, computeID)
	if err != nil {
		return err
	}

	// Get assigned components
	components, err := c.ListComponentAssignments(ctx, storage.ComputeComponentFilters{
		ComputeID: computeID,
	})
	if err != nil {
		return err
	}

	// Get assignments (services on this compute)
	assignments, err := c.ListAssignments(ctx, storage.AssignmentFilters{
		ComputeID: computeID,
	})
	if err != nil {
		return err
	}

	// Get journal entries
	journals, err := c.ListJournal(ctx, storage.JournalFilters{
		ComputeID: computeID,
	})
	if err != nil {
		return err
	}

	// Print markdown report
	fmt.Printf("# %s\n\n", compute.Name)
	fmt.Printf("**Type:** %s  \n", compute.Type)
	fmt.Printf("**Provider:** %s  \n", compute.Provider)
	fmt.Printf("**Region:** %s  \n", compute.Region)
	fmt.Printf("**State:** %s  \n", compute.State)

	// Tags
	if len(compute.Tags) > 0 {
		fmt.Printf("\n**Tags:**\n")
		for k, v := range compute.Tags {
			fmt.Printf("- `%s`: %s\n", k, v)
		}
	}

	// Hardware Components
	if len(components) > 0 {
		fmt.Printf("\n## Hardware Components\n\n")
		for _, cc := range components {
			// Get component details
			comp, err := c.GetComponent(ctx, cc.ComponentID)
			if err != nil {
				continue
			}

			fmt.Printf("### %s\n\n", comp.Name)
			fmt.Printf("- **Type:** %s\n", comp.Type)
			fmt.Printf("- **Manufacturer:** %s\n", comp.Manufacturer)
			fmt.Printf("- **Model:** %s\n", comp.Model)
			fmt.Printf("- **Quantity:** %d\n", cc.Quantity)

			if cc.Slot != "" {
				fmt.Printf("- **Slot:** %s\n", cc.Slot)
			}
			if cc.SerialNo != "" {
				fmt.Printf("- **Serial:** %s\n", cc.SerialNo)
			}
			if cc.RaidLevel != "" && cc.RaidLevel != "none" {
				fmt.Printf("- **RAID:** %s\n", cc.RaidLevel)
				if cc.RaidGroup != "" {
					fmt.Printf("- **RAID Group:** %s\n", cc.RaidGroup)
				}
			}

			// Specs
			if len(comp.Specs) > 0 {
				fmt.Printf("- **Specs:**\n")
				for k, v := range comp.Specs {
					fmt.Printf("  - %s: %v\n", k, v)
				}
			}

			if cc.Notes != "" {
				fmt.Printf("- **Notes:** %s\n", cc.Notes)
			}
			fmt.Println()
		}
	}

	// Service Assignments
	if len(assignments) > 0 {
		fmt.Printf("## Assigned Services\n\n")
		for _, assignment := range assignments {
			// Get service details
			service, err := c.GetService(ctx, assignment.ServiceID)
			if err != nil {
				continue
			}

			fmt.Printf("### %s\n\n", service.Name)

			// Allocated resources
			if len(assignment.Allocated) > 0 {
				fmt.Printf("**Allocated Resources:**\n")
				for k, v := range assignment.Allocated {
					fmt.Printf("- %s: %v\n", k, v)
				}
			}
			fmt.Println()
		}
	}

	// Resource Summary
	if len(components) > 0 {
		fmt.Printf("## Resource Summary\n\n")

		var totalCores int
		var totalMemoryGB float64
		var totalVRAMGB float64
		var totalStorageGB float64

		// Group storage by RAID configuration
		raidGroups := make(map[string][]*storageInfo)
		var nonRaidStorage []*storageInfo

		for _, cc := range components {
			comp, err := c.GetComponent(ctx, cc.ComponentID)
			if err != nil {
				continue
			}

			compType := string(comp.Type)
			switch compType {
			case "cpu":
				// Try multiple field names: threads, cores, thread_count, core_count
				threads := getSpecFloat(comp.Specs, "threads", "thread_count", "cores", "core_count")
				if threads > 0 {
					totalCores += int(threads) * cc.Quantity
				}
			case "ram", "memory":
				// Try multiple field names: capacity_gb, size, size_gb, memory_gb, memory
				// Memory can be in GB or MB (if > 1024, assume MB and convert)
				memValue := getSpecFloat(comp.Specs, "capacity_gb", "size", "size_gb", "memory_gb", "memory")
				if memValue > 0 {
					// If value > 1024, assume it's in MB and convert to GB
					if memValue > 1024 {
						memValue = memValue / 1024
					}
					totalMemoryGB += memValue * float64(cc.Quantity)
				}
			case "gpu":
				// Try multiple field names: vram_gb, vram, memory_gb, video_memory_gb, memory
				// Memory can be in GB or MB (if > 1024, assume MB and convert)
				vramValue := getSpecFloat(comp.Specs, "vram_gb", "vram", "memory_gb", "video_memory_gb", "memory")
				if vramValue > 0 {
					// If value > 1024, assume it's in MB and convert to GB
					if vramValue > 1024 {
						vramValue = vramValue / 1024
					}
					totalVRAMGB += vramValue * float64(cc.Quantity)
				}
			case "storage", "nvme", "ssd", "hdd":
				// Handle storage with RAID grouping
				storageValue := getSpecFloat(comp.Specs, "size", "capacity_gb", "storage_gb", "capacity")
				if storageValue > 0 {
					si := &storageInfo{
						size:      storageValue,
						quantity:  cc.Quantity,
						raidLevel: string(cc.RaidLevel),
						name:      comp.Name,
					}

					// Group by RAID configuration
					if cc.RaidLevel != "" && cc.RaidLevel != "none" && cc.RaidGroup != "" {
						if raidGroups[cc.RaidGroup] == nil {
							raidGroups[cc.RaidGroup] = make([]*storageInfo, 0)
						}
						raidGroups[cc.RaidGroup] = append(raidGroups[cc.RaidGroup], si)
					} else {
						nonRaidStorage = append(nonRaidStorage, si)
					}
				}
			}
		}

		// Calculate total storage with RAID
		for _, group := range raidGroups {
			capacity := calculateRaidCapacity(group)
			totalStorageGB += capacity
		}
		for _, si := range nonRaidStorage {
			totalStorageGB += si.size * float64(si.quantity)
		}

		// Calculate allocated resources from assignments
		var allocatedCores int
		var allocatedMemoryMB float64
		var allocatedVRAMMB float64
		var allocatedStorageGB float64

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

		// Convert totals to same units for comparison (MB for memory/vram)
		totalMemoryMB := totalMemoryGB * 1024
		totalVRAMMB := totalVRAMGB * 1024

		if totalCores > 0 {
			utilPct := 0.0
			if totalCores > 0 {
				utilPct = (float64(allocatedCores) / float64(totalCores)) * 100
			}
			fmt.Printf("- **Cores:** %d (%.1f%% allocated)\n", totalCores, utilPct)
		}
		if totalMemoryGB > 0 {
			utilPct := 0.0
			if totalMemoryMB > 0 {
				utilPct = (allocatedMemoryMB / totalMemoryMB) * 100
			}
			fmt.Printf("- **Memory:** %.0f GB (%.1f%% allocated)\n", totalMemoryGB, utilPct)
		}
		if totalVRAMGB > 0 {
			utilPct := 0.0
			if totalVRAMMB > 0 {
				utilPct = (allocatedVRAMMB / totalVRAMMB) * 100
			}
			fmt.Printf("- **VRAM:** %.0f GB (%.1f%% allocated)\n", totalVRAMGB, utilPct)
		}
		if totalStorageGB > 0 {
			utilPct := 0.0
			if totalStorageGB > 0 {
				utilPct = (allocatedStorageGB / totalStorageGB) * 100
			}
			fmt.Printf("- **Storage:** %.0f GB (%.1f%% allocated)\n", totalStorageGB, utilPct)

			// Show storage breakdown with RAID info
			if len(raidGroups) > 0 || len(nonRaidStorage) > 0 {
				fmt.Printf("\n### Storage Configuration\n\n")

				// Show RAID arrays
				for groupID, group := range raidGroups {
					capacity := calculateRaidCapacity(group)
					raidLevel := group[0].raidLevel
					diskCount := 0
					for _, si := range group {
						diskCount += si.quantity
					}
					fmt.Printf("**RAID Group: %s (%s)**\n", groupID, raidLevel)
					fmt.Printf("- Disks: %d\n", diskCount)
					fmt.Printf("- Effective Capacity: %.0f GB\n", capacity)
					fmt.Printf("- Components:\n")
					for _, si := range group {
						fmt.Printf("  - %dx %s (%.0f GB each)\n", si.quantity, si.name, si.size)
					}
					fmt.Println()
				}

				// Show non-RAID storage
				if len(nonRaidStorage) > 0 {
					fmt.Printf("**Non-RAID Storage**\n")
					total := 0.0
					for _, si := range nonRaidStorage {
						capacity := si.size * float64(si.quantity)
						fmt.Printf("- %dx %s = %.0f GB\n", si.quantity, si.name, capacity)
						total += capacity
					}
					fmt.Printf("- Total: %.0f GB\n", total)
					fmt.Println()
				}
			}
		}
		fmt.Println()
	}

	// Journal Entries
	if len(journals) > 0 {
		fmt.Printf("## Journal Entries\n\n")

		// Prepare data and calculate column widths
		type tableRow struct {
			date      string
			category  string
			createdBy string
			content   string
		}

		rows := make([]tableRow, 0, len(journals))
		maxDateWidth := len("Date")
		maxCategoryWidth := len("Category")
		maxCreatedByWidth := len("Created By")
		maxContentWidth := len("Content")

		for _, journal := range journals {
			createdBy := journal.CreatedBy
			if createdBy == "" {
				createdBy = "-"
			}

			// Escape pipe characters and replace newlines
			content := strings.ReplaceAll(journal.Content, "|", "\\|")
			content = strings.ReplaceAll(content, "\n", " ")

			// Truncate very long content for table
			truncatedContent := content
			if len(truncatedContent) > 80 {
				truncatedContent = truncatedContent[:77] + "..."
			}

			row := tableRow{
				date:      journal.CreatedAt.Format("2006-01-02 15:04"),
				category:  journal.Category,
				createdBy: createdBy,
				content:   truncatedContent,
			}
			rows = append(rows, row)

			// Update max widths
			if len(row.date) > maxDateWidth {
				maxDateWidth = len(row.date)
			}
			if len(row.category) > maxCategoryWidth {
				maxCategoryWidth = len(row.category)
			}
			if len(row.createdBy) > maxCreatedByWidth {
				maxCreatedByWidth = len(row.createdBy)
			}
			if len(row.content) > maxContentWidth {
				maxContentWidth = len(row.content)
			}
		}

		// Print header
		fmt.Printf("| %-*s | %-*s | %-*s | %-*s |\n",
			maxDateWidth, "Date",
			maxCategoryWidth, "Category",
			maxCreatedByWidth, "Created By",
			maxContentWidth, "Content")

		// Print separator
		fmt.Printf("|-%s-|-%s-|-%s-|-%s-|\n",
			strings.Repeat("-", maxDateWidth),
			strings.Repeat("-", maxCategoryWidth),
			strings.Repeat("-", maxCreatedByWidth),
			strings.Repeat("-", maxContentWidth))

		// Print rows
		for _, row := range rows {
			fmt.Printf("| %-*s | %-*s | %-*s | %-*s |\n",
				maxDateWidth, row.date,
				maxCategoryWidth, row.category,
				maxCreatedByWidth, row.createdBy,
				maxContentWidth, row.content)
		}

		fmt.Println()

		// If detailed journal flag is set, show full entries
		if detailedJournal {
			fmt.Printf("### Detailed Entries\n\n")
			for i, journal := range journals {
				if i > 0 {
					fmt.Println()
				}

				createdBy := journal.CreatedBy
				if createdBy == "" {
					createdBy = "-"
				}

				fmt.Printf("**Entry %d**\n\n", i+1)
				fmt.Printf("- **Date:** %s\n", journal.CreatedAt.Format("2006-01-02 15:04:05"))
				fmt.Printf("- **Category:** %s\n", journal.Category)
				fmt.Printf("- **Created By:** %s\n", createdBy)
				fmt.Printf("\n**Content:**\n\n%s\n", journal.Content)
			}
			fmt.Println()
		}
	}

	return nil
}

// Calculate RAID capacity based on RAID level
func calculateRaidCapacity(storage []*storageInfo) float64 {
	if len(storage) == 0 {
		return 0
	}

	raidLevel := storage[0].raidLevel

	// Collect all disk sizes
	var disks []float64
	for _, si := range storage {
		for i := 0; i < si.quantity; i++ {
			disks = append(disks, si.size)
		}
	}

	if len(disks) == 0 {
		return 0
	}

	switch raidLevel {
	case "raid0":
		// RAID 0: Sum of all disks
		total := 0.0
		for _, size := range disks {
			total += size
		}
		return total

	case "raid1":
		// RAID 1: Size of smallest disk (mirroring)
		smallest := disks[0]
		for _, size := range disks {
			if size < smallest {
				smallest = size
			}
		}
		return smallest

	case "raid5":
		// RAID 5: (n-1) * smallest disk
		if len(disks) < 3 {
			total := 0.0
			for _, size := range disks {
				total += size
			}
			return total
		}
		smallest := disks[0]
		for _, size := range disks {
			if size < smallest {
				smallest = size
			}
		}
		return float64(len(disks)-1) * smallest

	case "raid6":
		// RAID 6: (n-2) * smallest disk
		if len(disks) < 4 {
			total := 0.0
			for _, size := range disks {
				total += size
			}
			return total
		}
		smallest := disks[0]
		for _, size := range disks {
			if size < smallest {
				smallest = size
			}
		}
		return float64(len(disks)-2) * smallest

	case "raid10":
		// RAID 10: Sum / 2 (mirrored stripes)
		if len(disks) < 4 || len(disks)%2 != 0 {
			total := 0.0
			for _, size := range disks {
				total += size
			}
			return total
		}
		total := 0.0
		for _, size := range disks {
			total += size
		}
		return total / 2.0

	default:
		// Unknown RAID level, sum all disks
		total := 0.0
		for _, size := range disks {
			total += size
		}
		return total
	}
}
