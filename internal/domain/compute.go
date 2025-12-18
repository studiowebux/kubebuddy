package domain

import "time"

// ComputeType represents the type of compute resource
type ComputeType string

const (
	ComputeTypeBaremetal ComputeType = "baremetal"
	ComputeTypeVPS       ComputeType = "vps"
	ComputeTypeVM        ComputeType = "vm"
)

// ComputeState represents the operational state of a compute resource
type ComputeState string

const (
	ComputeStateActive        ComputeState = "active"
	ComputeStateMaintenance   ComputeState = "maintenance"
	ComputeStateDecommissioned ComputeState = "decommissioned"
)

// Resources represents dynamic resource attributes as key-value pairs
// Examples: {"cpu": 8, "ram_gb": 32, "nvme_gb": 500, "bandwidth_mbps": 1000}
type Resources map[string]interface{}

// Compute represents a compute resource (baremetal, VPS, or VM)
type Compute struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Type      ComputeType            `json:"type"`
	Provider  string                 `json:"provider"`
	Region    string                 `json:"region"`
	Tags      map[string]string      `json:"tags"`
	State     ComputeState           `json:"state"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`

	// Billing fields
	MonthlyCost      *float64   `json:"monthly_cost,omitempty"`
	AnnualCost       *float64   `json:"annual_cost,omitempty"`
	ContractEndDate  *time.Time `json:"contract_end_date,omitempty"`
	NextRenewalDate  *time.Time `json:"next_renewal_date,omitempty"`

	// Resources is computed from components and NOT persisted to database
	// Use GetTotalResourcesFromComponents to populate this field
	Resources Resources              `json:"-"`
}

// GetAllocatedResources calculates total allocated resources from assignments
// Uses service MaxSpec for each assignment, multiplied by assignment quantity
func (c *Compute) GetAllocatedResources(assignments []*Assignment, services map[string]*Service) Resources {
	allocated := make(Resources)

	for _, assignment := range assignments {
		if assignment.ComputeID == c.ID {
			// Look up service to get MaxSpec
			service, ok := services[assignment.ServiceID]
			if !ok {
				continue // Skip if service not found
			}

			quantity := assignment.Quantity
			if quantity == 0 {
				quantity = 1
			}

			// Add MaxSpec resources to allocated, multiplied by quantity
			for key, value := range service.MaxSpec {
				if existing, ok := allocated[key]; ok {
					// Sum numeric values
					switch v := value.(type) {
					case int:
						if e, ok := existing.(int); ok {
							allocated[key] = e + (v * quantity)
						}
					case float64:
						if e, ok := existing.(float64); ok {
							allocated[key] = e + (v * float64(quantity))
						}
					}
				} else {
					switch v := value.(type) {
					case int:
						allocated[key] = v * quantity
					case float64:
						allocated[key] = v * float64(quantity)
					default:
						allocated[key] = value
					}
				}
			}
		}
	}

	return allocated
}

// GetAvailableResources calculates available resources
func (c *Compute) GetAvailableResources(allocated Resources) Resources {
	available := make(Resources)

	for key, totalValue := range c.Resources {
		if alloc, ok := allocated[key]; ok {
			// Subtract allocated from total, handling type mismatches
			switch t := totalValue.(type) {
			case int:
				switch a := alloc.(type) {
				case int:
					available[key] = t - a
				case float64:
					available[key] = t - int(a)
				}
			case float64:
				switch a := alloc.(type) {
				case int:
					available[key] = t - float64(a)
				case float64:
					available[key] = t - a
				}
			}
		} else {
			available[key] = totalValue
		}
	}

	return available
}

// MatchesTags checks if compute has all required tags
func (c *Compute) MatchesTags(required map[string]string) bool {
	for key, value := range required {
		if c.Tags[key] != value {
			return false
		}
	}
	return true
}
