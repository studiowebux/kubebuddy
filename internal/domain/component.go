package domain

import (
	"time"
)

// ComponentType represents the type of component
type ComponentType string

const (
	ComponentTypeCPU     ComponentType = "cpu"
	ComponentTypeRAM     ComponentType = "ram"
	ComponentTypeStorage ComponentType = "storage"
	ComponentTypeGPU     ComponentType = "gpu"
	ComponentTypeNIC     ComponentType = "nic"
	ComponentTypePSU     ComponentType = "psu"
	ComponentTypeOther   ComponentType = "other"
)

// Component represents a hardware component
type Component struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Type         ComponentType          `json:"type"`
	Manufacturer string                 `json:"manufacturer"`
	Model        string                 `json:"model"`
	Specs        map[string]interface{} `json:"specs"` // Flexible specs like cores, ghz, capacity, etc.
	Notes        string                 `json:"notes,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// RaidLevel represents RAID configuration
type RaidLevel string

const (
	RaidLevelNone  RaidLevel = "none"
	RaidLevel0     RaidLevel = "raid0"  // Striping - Total = sum of disks
	RaidLevel1     RaidLevel = "raid1"  // Mirroring - Total = smallest disk
	RaidLevel5     RaidLevel = "raid5"  // Striping with parity - Total = (n-1) * smallest
	RaidLevel6     RaidLevel = "raid6"  // Striping with double parity - Total = (n-2) * smallest
	RaidLevel10    RaidLevel = "raid10" // Mirrored stripes - Total = sum / 2
)

// ComputeComponent represents a component assigned to a compute resource
type ComputeComponent struct {
	ID          string    `json:"id"`
	ComputeID   string    `json:"compute_id"`
	ComponentID string    `json:"component_id"`
	Quantity    int       `json:"quantity"`
	Slot        string    `json:"slot,omitempty"`        // Physical slot/position (e.g., "CPU1", "DIMM0-3", "Bay 0")
	SerialNo    string    `json:"serial_no,omitempty"`   // Serial number for tracking
	Notes       string    `json:"notes,omitempty"`       // Installation notes
	RaidLevel   RaidLevel `json:"raid_level,omitempty"`  // RAID configuration for storage
	RaidGroup   string    `json:"raid_group,omitempty"`  // Group ID for RAID arrays
	CreatedAt   time.Time `json:"created_at"`
}

// GetTotalResources calculates total resources from assigned components
func (c *Compute) GetTotalResourcesFromComponents(components []*Component, assignments []*ComputeComponent) Resources {
	resources := make(Resources)

	// Group storage assignments by RAID group for special handling
	raidGroups := make(map[string][]*storageAssignment)
	var nonRaidStorage []*storageAssignment

	for _, assignment := range assignments {
		if assignment.ComputeID != c.ID {
			continue
		}

		// Find the component
		var component *Component
		for _, comp := range components {
			if comp.ID == assignment.ComponentID {
				component = comp
				break
			}
		}

		if component == nil {
			continue
		}

		// Aggregate resources based on component type
		quantity := assignment.Quantity
		compType := string(component.Type)

		switch compType {
		case "cpu":
			// Aggregate CPU cores/threads - try multiple field names
			coresValue := getSpecFloat(component.Specs, "threads", "thread_count", "cores", "core_count")
			if coresValue > 0 {
				existing := getFloatValue(resources, "cores")
				resources["cores"] = int(existing + (coresValue * float64(quantity)))
			}
		case "ram", "memory":
			// Aggregate RAM capacity - try multiple field names
			// Fields ending in _gb are in GB, others are in MB
			memValue := 0.0
			if val := getSpecFloat(component.Specs, "capacity_gb", "size_gb", "memory_gb"); val > 0 {
				memValue = val * 1024 // Convert GB to MB
			} else if val := getSpecFloat(component.Specs, "memory", "size"); val > 0 {
				memValue = val // Already in MB
			}
			if memValue > 0 {
				existing := getFloatValue(resources, "memory")
				resources["memory"] = int(existing + (memValue * float64(quantity)))
			}
		case "storage", "nvme", "ssd", "hdd":
			// Handle storage with RAID support
			storageValue := getSpecFloat(component.Specs, "size", "capacity_gb", "storage_gb", "capacity")
			if storageValue > 0 {
				sa := &storageAssignment{
					size:     storageValue,
					quantity: quantity,
				}

				// Group by RAID configuration
				if assignment.RaidLevel != "" && assignment.RaidLevel != RaidLevelNone && assignment.RaidGroup != "" {
					// Part of a RAID array
					if raidGroups[assignment.RaidGroup] == nil {
						raidGroups[assignment.RaidGroup] = make([]*storageAssignment, 0)
					}
					sa.raidLevel = assignment.RaidLevel
					raidGroups[assignment.RaidGroup] = append(raidGroups[assignment.RaidGroup], sa)
				} else {
					// Non-RAID storage
					nonRaidStorage = append(nonRaidStorage, sa)
				}
			}
		case "gpu":
			// Count GPUs and aggregate VRAM
			existing := getFloatValue(resources, "gpu")
			resources["gpu"] = int(existing) + quantity

			// Aggregate VRAM - try multiple field names
			// Fields ending in _gb are in GB, others are in MB
			vramValue := 0.0
			if val := getSpecFloat(component.Specs, "vram_gb", "memory_gb", "video_memory_gb"); val > 0 {
				vramValue = val * 1024 // Convert GB to MB
			} else if val := getSpecFloat(component.Specs, "vram", "memory"); val > 0 {
				vramValue = val // Already in MB
			}
			if vramValue > 0 {
				existing := getFloatValue(resources, "vram")
				resources["vram"] = int(existing + (vramValue * float64(quantity)))
			}
		case "nic":
			// Aggregate network bandwidth
			if speedGbps, ok := component.Specs["speed_gbps"].(float64); ok {
				existing := getFloatValue(resources, "bandwidth_gbps")
				resources["bandwidth_gbps"] = existing + (speedGbps * float64(quantity))
			}
		}
	}

	// Calculate total storage with RAID support
	totalStorage := 0.0

	// Calculate RAID arrays
	for _, group := range raidGroups {
		capacity := calculateRaidCapacity(group)
		totalStorage += capacity
	}

	// Add non-RAID storage
	for _, sa := range nonRaidStorage {
		totalStorage += sa.size * float64(sa.quantity)
	}

	if totalStorage > 0 {
		resources["nvme"] = int(totalStorage)
	}

	return resources
}

type storageAssignment struct {
	size      float64
	quantity  int
	raidLevel RaidLevel
}

func calculateRaidCapacity(assignments []*storageAssignment) float64 {
	if len(assignments) == 0 {
		return 0
	}

	// All assignments in a group should have the same RAID level
	raidLevel := assignments[0].raidLevel

	// Collect all disk sizes
	var disks []float64
	for _, sa := range assignments {
		for i := 0; i < sa.quantity; i++ {
			disks = append(disks, sa.size)
		}
	}

	if len(disks) == 0 {
		return 0
	}

	switch raidLevel {
	case RaidLevel0:
		// RAID 0: Sum of all disks
		total := 0.0
		for _, size := range disks {
			total += size
		}
		return total

	case RaidLevel1:
		// RAID 1: Size of smallest disk (mirroring)
		smallest := disks[0]
		for _, size := range disks {
			if size < smallest {
				smallest = size
			}
		}
		return smallest

	case RaidLevel5:
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

	case RaidLevel6:
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

	case RaidLevel10:
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

// Helper to safely get float value from resources
func getFloatValue(resources Resources, key string) float64 {
	if val, ok := resources[key]; ok {
		switch v := val.(type) {
		case int:
			return float64(v)
		case float64:
			return v
		}
	}
	return 0
}

// ComponentTypes returns all valid component types
func ComponentTypes() []string {
	return []string{
		string(ComponentTypeCPU),
		string(ComponentTypeRAM),
		string(ComponentTypeStorage),
		string(ComponentTypeGPU),
		string(ComponentTypeNIC),
		string(ComponentTypePSU),
		string(ComponentTypeOther),
	}
}
